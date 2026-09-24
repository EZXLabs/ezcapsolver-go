package ezcapsolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// EzCapSolverClient talks to the EzCaptchaSolver API.
//
// One client covers both endpoints. Asynchronous task types are created and
// polled, synchronous ones answer inline, and the convenience methods hide that
// difference: each task type has a method taking its request model and
// returning its solution model.
//
// A client is safe for concurrent use and holds a connection pool, so build one
// and share it rather than making one per request.
type EzCapSolverClient struct {
	clientKey  string
	config     ClientConfig
	httpClient *http.Client
	// ownsHTTP records whether the client built its own http.Client. A caller's
	// own client may be shared with the rest of their program, so Close must not
	// reach into it.
	ownsHTTP bool
	logger   *slog.Logger
}

// NewClient builds a client, reading the client key from the EZCAPTCHA_API_KEY
// environment variable unless [WithClientKey] supplies one.
//
// Configuration is validated here rather than on the first request, so a bad
// setting surfaces before anything is billed. An invalid setting comes back as
// an error wrapping [ErrConfig], with the offending setting named in the message.
func NewClient(options ...Option) (*EzCapSolverClient, error) {
	settings := &builder{config: defaultConfig()}
	for _, option := range options {
		option(settings)
	}

	if settings.config.clientKey == "" {
		settings.config.clientKey = os.Getenv(DefaultClientKeyEnv)
	}
	if err := settings.config.validate(); err != nil {
		return nil, err
	}

	httpClient := settings.httpClient
	ownsHTTP := httpClient == nil
	if ownsHTTP {
		built, err := buildHTTPClient(settings.config)
		if err != nil {
			return nil, err
		}
		httpClient = built
	}

	logger := settings.logger
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}

	return &EzCapSolverClient{
		clientKey:  settings.config.clientKey,
		config:     settings.config,
		httpClient: httpClient,
		ownsHTTP:   ownsHTTP,
		logger:     logger,
	}, nil
}

// buildHTTPClient produces the client used when the caller supplied none.
//
// Timeout is deliberately left at zero: the SDK applies a per-request deadline
// through the context, and the asynchronous and synchronous endpoints get
// different budgets. A client-wide timeout would cut the longer one short.
func buildHTTPClient(config ClientConfig) (*http.Client, error) {
	transport, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return nil, fmt.Errorf(
			"%w: the default HTTP transport has been replaced; pass your own client with WithHTTPClient",
			ErrConfig,
		)
	}
	cloned := transport.Clone()
	if config.proxy != "" {
		proxyURL, err := url.Parse(config.proxy)
		if err != nil {
			return nil, fmt.Errorf("%w: SDK proxy URL is not parseable: %w", ErrConfig, err)
		}
		cloned.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Transport: cloned}, nil
}

// Close releases the connections this client is holding idle.
//
// [NewClient] clones the default transport rather than sharing it, so each
// client keeps a connection pool of its own. Long-lived programs that build one
// client and keep it need not call this; programs that build clients per key or
// per tenant should, or the idle connections sit on file descriptors until the
// transport's own IdleConnTimeout expires.
//
// It does nothing when the HTTP client came from [WithHTTPClient]: that one may
// be shared with the rest of the program, and disposing of it is not the SDK's
// call. In-flight requests are unaffected either way, and the client stays
// usable afterwards — a later request simply dials again.
func (c *EzCapSolverClient) Close() {
	if c.ownsHTTP {
		c.httpClient.CloseIdleConnections()
	}
}

// Config returns the effective configuration. Printing it is safe: the client
// key and the SDK proxy are replaced with a placeholder.
func (c *EzCapSolverClient) Config() ClientConfig { return c.config }

// CreateTask creates an asynchronous task and returns immediately, without
// waiting for a result.
//
// The task type is injected into the serialized parameters, so no task model
// carries it as a field. params is any value that marshals to a JSON object: one
// of this package's task models, or a plain map for a type the SDK does not
// model yet.
//
// Most callers want [EzCapSolverClient.Solve], or the convenience method for
// their task type, which create and wait in one call.
//
// Creating a task is billed and is not idempotent. If this returns a transport
// error, the service may still have accepted the task.
func (c *EzCapSolverClient) CreateTask(
	ctx context.Context, taskType TaskType, params any,
) (*CreateTaskResponse, error) {
	c.logger.DebugContext(ctx, "creating task", "task_type", string(taskType))

	request := createTaskRequest{
		ClientKey: c.clientKey,
		AppID:     c.config.AppID,
		Task:      taskPayload{taskType: taskType, params: params},
	}
	var response CreateTaskResponse
	url := endpoint(c.config.AsyncBaseURL, pathCreateTask)
	if err := c.post(ctx, url, request, c.config.Timeout, &response); err != nil {
		return nil, err
	}
	if strings.TrimSpace(response.TaskID) == "" {
		return nil, &UnexpectedResponseError{
			Reason: "successful createTask response does not contain a task ID",
		}
	}

	c.logger.InfoContext(ctx, "task created",
		"task_id", response.TaskID,
		"task_type", string(taskType),
	)
	return &response, nil
}

// GetTaskResult queries an asynchronous task once, without waiting.
//
// A task that has not finished comes back with [TaskStatusProcessing] and no
// solution. Use [EzCapSolverClient.WaitForResult] to poll until it does.
func (c *EzCapSolverClient) GetTaskResult(ctx context.Context, taskID string) (*TaskResult, error) {
	request := queryTaskResultRequest{ClientKey: c.clientKey, TaskID: taskID}
	var result TaskResult
	url := endpoint(c.config.AsyncBaseURL, pathGetTaskResult)
	if err := c.post(ctx, url, request, c.config.Timeout, &result); err != nil {
		return nil, err
	}
	if err := validateStatus(result.Status); err != nil {
		return nil, err
	}
	return &result, nil
}

// WaitForResult polls an existing task until it finishes, using the client's
// polling settings.
func (c *EzCapSolverClient) WaitForResult(ctx context.Context, taskID string) (*TaskResult, error) {
	return c.WaitForResultWith(ctx, taskID, c.config.Polling)
}

// WaitForResultWith polls an existing task with settings for this call only.
//
// Task types differ widely in how long they take, so a single client-wide
// budget does not fit every one of them.
//
// When the budget runs out it returns a [*PollingExhaustedError], matchable with
// errors.Is against [ErrPollingExhausted]. The task may still finish afterwards,
// and the service holds its result for five minutes after creation — so waiting
// again on the same id is worth more than creating a second, billed task.
func (c *EzCapSolverClient) WaitForResultWith(
	ctx context.Context, taskID string, polling PollingConfig,
) (*TaskResult, error) {
	if err := polling.validate(); err != nil {
		return nil, err
	}

	for attempt := 1; attempt <= polling.MaxAttempts; attempt++ {
		// Wait before every query, the first one included: a task that was just
		// created is still queued for a worker and would only report processing.
		select {
		case <-ctx.Done():
			return nil, &TransportError{Op: "waiting for task " + taskID, Err: ctx.Err()}
		case <-time.After(polling.Interval):
		}

		c.logger.DebugContext(ctx, "polling task result", "task_id", taskID, "attempt", attempt)
		result, err := c.GetTaskResult(ctx, taskID)
		if err != nil {
			// Throttling says nothing about the task, which is still queued.
			// Spend the attempt and poll again rather than failing a task that
			// has already been billed. The attempt is spent on purpose:
			// interval x attempts is what keeps the whole wait inside the five
			// minute window the result is held for.
			var apiErr *APIError
			if errors.As(err, &apiErr) && apiErr.IsRateLimited() {
				c.logger.WarnContext(ctx, "polling throttled, retrying",
					"task_id", taskID,
					"attempt", attempt,
					"error_code", apiErr.ErrorCode,
				)
				continue
			}
			return nil, err
		}

		switch result.Status {
		case TaskStatusReady:
			c.logger.InfoContext(ctx, "task completed", "task_id", taskID, "attempts", attempt)
			return result, nil
		case TaskStatusProcessing:
			continue
		default:
			// A failed task arrives with errorId 1 and is already an APIError by
			// now, so reaching this means the envelope contradicted itself.
			return nil, &UnexpectedResponseError{Reason: fmt.Sprintf(
				"task %q reported status %q without an API error", taskID, result.Status,
			)}
		}
	}

	return nil, &PollingExhaustedError{
		TaskID:   taskID,
		Attempts: polling.MaxAttempts,
		Interval: polling.Interval,
	}
}

// Solve creates an asynchronous task and waits for its result, returning the
// solution as raw JSON.
//
// This is the escape hatch for a task type the SDK does not model: pass the type
// name and a map of parameters. For a modelled type, use its convenience method
// and get a typed solution, or [SolveAs] to decode into your own struct.
//
// When waiting fails, the task and request identifiers are attached to the
// error, so a failure can still be traced back to the task it belongs to.
func (c *EzCapSolverClient) Solve(
	ctx context.Context, taskType TaskType, params any,
) (*Solved[json.RawMessage], error) {
	return c.SolveWith(ctx, taskType, params, c.config.Polling)
}

// SolveWith creates an asynchronous task and waits with settings for this call
// only.
func (c *EzCapSolverClient) SolveWith(
	ctx context.Context, taskType TaskType, params any, polling PollingConfig,
) (*Solved[json.RawMessage], error) {
	created, err := c.CreateTask(ctx, taskType, params)
	if err != nil {
		return nil, err
	}

	result, err := c.WaitForResultWith(ctx, created.TaskID, polling)
	if err != nil {
		return nil, tagWithTask(err, created.TaskID, created.RequestID)
	}

	raw := cloneRaw(result.Solution)
	return &Solved[json.RawMessage]{
		Solution:  raw,
		Raw:       raw,
		TaskID:    created.TaskID,
		RequestID: firstNonEmpty(result.RequestID, created.RequestID),
	}, nil
}

// SyncSolve runs a task on the synchronous endpoint and returns its solution as
// raw JSON.
//
// It is the counterpart of [EzCapSolverClient.Solve]: same arguments, same result
// type, only the endpoint differs. [Solved.TaskID] carries the identifier this
// endpoint assigns and returns alongside the result.
//
// This is the escape hatch for a task type the SDK does not model. For a
// modelled type, use its SyncSolveX convenience method and get a typed
// solution, or [SyncSolveAs] to decode into your own struct.
func (c *EzCapSolverClient) SyncSolve(
	ctx context.Context, taskType TaskType, params any,
) (*Solved[json.RawMessage], error) {
	result, err := c.CreateSyncTask(ctx, taskType, params)
	if err != nil {
		return nil, err
	}

	// Mirror Solve exactly: hand back the raw value without decoding it, so an
	// empty solution reaches the caller rather than becoming an error here.
	raw := cloneRaw(result.Solution)
	return &Solved[json.RawMessage]{
		Solution:  raw,
		Raw:       raw,
		TaskID:    result.TaskID,
		RequestID: result.RequestID,
	}, nil
}

// CreateSyncTask executes a synchronous task and returns its terminal result.
//
// This is the low-level primitive that [EzCapSolverClient.SyncSolve] builds on,
// and the synchronous counterpart of [EzCapSolverClient.CreateTask]. Prefer
// SyncSolve unless you need the raw [TaskResult].
//
// The synchronous endpoint blocks until the worker answers and returns no task
// ID, so it gets its own, much longer timeout — see [WithSyncTimeout].
func (c *EzCapSolverClient) CreateSyncTask(
	ctx context.Context, taskType TaskType, params any,
) (*TaskResult, error) {
	c.logger.DebugContext(ctx, "creating sync task", "task_type", string(taskType))

	request := createTaskRequest{
		ClientKey: c.clientKey,
		AppID:     c.config.AppID,
		Task:      taskPayload{taskType: taskType, params: params},
	}
	var result TaskResult
	url := endpoint(c.config.SyncBaseURL, pathCreateSyncTask)
	if err := c.post(ctx, url, request, c.config.SyncTimeout, &result); err != nil {
		return nil, err
	}
	if err := validateStatus(result.Status); err != nil {
		return nil, err
	}
	if result.IsProcessing() {
		return nil, &UnexpectedResponseError{
			Reason: "createSyncTask returned status `processing`, which has no task ID to follow up on",
		}
	}

	c.logger.InfoContext(ctx, "sync task completed", "task_type", string(taskType))
	return &result, nil
}

// Balance returns the account balance.
func (c *EzCapSolverClient) Balance(ctx context.Context) (float64, error) {
	c.logger.DebugContext(ctx, "querying account balance")

	var response balanceResponse
	url := endpoint(c.config.AsyncBaseURL, pathGetBalance)
	request := queryBalanceRequest{ClientKey: c.clientKey}
	if err := c.post(ctx, url, request, c.config.Timeout, &response); err != nil {
		return 0, err
	}
	if response.Balance == nil {
		return 0, &UnexpectedResponseError{
			Reason: "successful balance response does not contain a balance",
		}
	}
	return *response.Balance, nil
}

// SolveAs creates an asynchronous task, waits for it, and decodes the solution
// into T.
//
// It is the typed counterpart of [EzCapSolverClient.Solve]: use it for a task type
// this release does not model, with a struct of your own.
//
//	type myShape struct {
//		Token string `json:"token"`
//	}
//	solved, err := ezcapsolver.SolveAs[myShape](ctx, client, "BrandNewTaskType", params)
//
// It is a function rather than a method because Go methods cannot take type
// parameters.
func SolveAs[T any](
	ctx context.Context, client *EzCapSolverClient, taskType TaskType, params any,
) (*Solved[T], error) {
	return solveTyped[T](ctx, client, taskType, params, client.config.Polling)
}

// SyncSolveAs runs a task on the synchronous endpoint and decodes the solution
// into T.
//
// It is the typed counterpart of [EzCapSolverClient.SyncSolve], and the
// synchronous-endpoint counterpart of [SolveAs].
//
// It is a function rather than a method because Go methods cannot take type
// parameters.
func SyncSolveAs[T any](
	ctx context.Context, client *EzCapSolverClient, taskType TaskType, params any,
) (*Solved[T], error) {
	return solveSyncTyped[T](ctx, client, taskType, params)
}

// solveTyped backs [SolveAs] and every polling convenience method.
func solveTyped[T any](
	ctx context.Context,
	client *EzCapSolverClient,
	taskType TaskType,
	params any,
	polling PollingConfig,
) (*Solved[T], error) {
	solved, err := client.SolveWith(ctx, taskType, params, polling)
	if err != nil {
		return nil, err
	}

	var solution T
	if err := decodeSolution(solved.Raw, &solution); err != nil {
		return nil, err
	}
	return &Solved[T]{
		Solution:  solution,
		Raw:       solved.Raw,
		TaskID:    solved.TaskID,
		RequestID: solved.RequestID,
	}, nil
}

// solveSyncTyped backs every synchronous convenience method.
//
// The returned [Solved.TaskID] comes from the response: the synchronous
// endpoint assigns an identifier and returns it with the result.
func solveSyncTyped[T any](
	ctx context.Context, client *EzCapSolverClient, taskType TaskType, params any,
) (*Solved[T], error) {
	result, err := client.CreateSyncTask(ctx, taskType, params)
	if err != nil {
		return nil, err
	}
	raw := cloneRaw(result.Solution)

	var solution T
	if err := decodeSolution(raw, &solution); err != nil {
		return nil, err
	}
	return &Solved[T]{
		Solution:  solution,
		Raw:       raw,
		TaskID:    result.TaskID,
		RequestID: result.RequestID,
	}, nil
}

// decodeSolution turns a raw solution into a model, keeping the raw value on the
// error when it does not fit. That value is the whole point: a decode failure
// means the worker returned a shape this release does not model, and seeing it
// is what makes the failure diagnosable.
func decodeSolution(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return &UnexpectedResponseError{Reason: missingSolutionReason}
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return &SolutionDecodeError{Raw: raw, Err: err}
	}
	return nil
}

// validateStatus rejects a status outside the three the service defines.
//
// An empty value means the field was absent. Request-level failures do omit it,
// but those carry an error envelope and never reach here, so its absence in a
// success response is a contract break — and reporting that is more useful than
// quietly substituting a status.
func validateStatus(status TaskStatus) error {
	if status == "" {
		return &UnexpectedResponseError{
			Reason: "successful task response does not contain a status field",
		}
	}
	if !status.IsValid() {
		return &UnexpectedResponseError{
			Reason: fmt.Sprintf("unknown task status %q", string(status)),
		}
	}
	return nil
}

// tagWithTask attaches task context to an error raised while waiting.
//
// The service does not echo the task ID back on a result query, so without this
// a caller would be handed a failure with no way to tell which task it belongs
// to.
func tagWithTask(err error, taskID, requestID string) error {
	var apiError *APIError
	if errors.As(err, &apiError) {
		apiError.TaskID = taskID
		if apiError.RequestID == "" {
			apiError.RequestID = requestID
		}
		return err
	}

	// Already carries the identifier; wrapping would only make the caller peel
	// one more layer to read the same value.
	var exhausted *PollingExhaustedError
	if errors.As(err, &exhausted) {
		return err
	}

	// Everything else has nowhere to put the identifier, and Solve created the
	// task internally — so without this the caller is left paying for a result
	// they can no longer reach.
	return &WaitInterruptedError{TaskID: taskID, RequestID: requestID, Err: err}
}

// firstNonEmpty returns the first non-empty string, or "".
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
