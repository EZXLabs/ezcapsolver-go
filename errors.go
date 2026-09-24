package ezcapsolver

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

// Sentinel errors, one per failure layer. Match them with errors.Is; reach for
// the concrete types below with errors.As when you need the details.
//
// The layering is by what the caller can do about the failure, not by where it
// came from: a configuration mistake needs a code change, a transport fault
// needs the network looked at, an API error needs its code inspected.
var (
	// ErrConfig marks an invalid client configuration (E1). Retrying is pointless.
	//
	// It is returned by [NewClient], never by a request: configuration is
	// validated once, at construction, so a bad setting surfaces before any
	// billing happens. There is no dedicated type behind it — the reason is
	// wrapped into the message, so read it with Error rather than a field.
	ErrConfig = errors.New("ezcapsolver: invalid configuration")

	// ErrTransport marks a network-level failure (E2): DNS, TCP, TLS, timeout.
	ErrTransport = errors.New("ezcapsolver: transport failure")

	// ErrAPI marks a structured error reported by the service (E3).
	ErrAPI = errors.New("ezcapsolver: api error")

	// ErrPollingExhausted marks a task that never reached a terminal state (E4).
	ErrPollingExhausted = errors.New("ezcapsolver: polling exhausted")

	// ErrDecode marks a response the SDK could not turn into a model (E5).
	ErrDecode = errors.New("ezcapsolver: decode failure")
)

// Characters of a raw value kept in an error message. The full value stays on
// the error, but an oversized worker response must not bury the rest of the
// message.
const errorPreviewChars = 512

// TransportError reports a failure below the HTTP response (E2): the request
// never produced one.
//
// Whether it is safe to send the request again depends on the operation:
// creating a task is billed and is not idempotent, and a timeout cannot tell
// you whether the service already accepted the task. See [APIError.IsTerminal]
// and the package documentation.
type TransportError struct {
	// Op describes the attempted operation, such as
	// `POST https://api.ez-captcha.com/createTask`. It never contains credentials.
	Op string
	// Err is the underlying error. errors.Is reaches it, so
	// errors.Is(err, context.DeadlineExceeded) works on a timeout.
	Err error
}

func (e *TransportError) Error() string {
	return fmt.Sprintf("ezcapsolver: %s: %v", e.Op, e.Err)
}

func (e *TransportError) Unwrap() error { return e.Err }

// Is reports [ErrTransport] while leaving the wrapped error reachable.
func (e *TransportError) Is(target error) bool { return target == ErrTransport }

// APIError reports a structured error from the service (E3).
//
// It covers both a business error carried by a 200 response and any non-2xx
// status, so [APIError.HTTPStatus] is always readable from one place. The
// service signals failure with errorId, not with the HTTP status: most business
// errors arrive as HTTP 500 and a failed task arrives as HTTP 200.
type APIError struct {
	// ErrorCode is the stable machine-readable code, such as ERROR_ZERO_BALANCE.
	// It is empty when the response carried no error envelope.
	//
	// This is an open set: a code produced by a worker is not in the service's
	// own table, so it is exposed verbatim rather than folded into an enum.
	ErrorCode string
	// ErrorDescription is the human-readable description. When the response was
	// not JSON, it holds a slice of the raw body instead.
	ErrorDescription string
	// HTTPStatus is the status that carried the error. Always set.
	HTTPStatus int
	// Errors maps a field path to its validation message. Populated only on a
	// parameter validation failure.
	Errors map[string]string
	// RequestID is the service-side request tracking identifier, when present.
	RequestID string
	// TaskID is filled in by the SDK for failures after a task was created; the
	// service does not echo it back.
	TaskID string
}

// Error renders the code, description and status on one line, with the field
// paths appended when the failure was a validation error. Callers commonly log
// nothing but the error, so everything actionable has to be in this line.
func (e *APIError) Error() string {
	code := e.ErrorCode
	if code == "" {
		code = "UNKNOWN_API_ERROR"
	}
	description := e.ErrorDescription
	if description == "" {
		description = "The API returned an unspecified error"
	}

	var message strings.Builder
	fmt.Fprintf(&message, "%s: %s (HTTP %d)", code, description, e.HTTPStatus)
	if len(e.Errors) > 0 {
		// Map iteration order is randomised in Go, so sort to keep the message
		// reproducible across runs.
		fields := make([]string, 0, len(e.Errors))
		for field := range e.Errors {
			fields = append(fields, field)
		}
		slices.Sort(fields)
		for index, field := range fields {
			separator := ", "
			if index == 0 {
				separator = "; "
			}
			fmt.Fprintf(&message, "%s%s: %s", separator, field, e.Errors[field])
		}
	}
	return message.String()
}

// Is reports [ErrAPI], so errors.Is classifies this without a type assertion.
func (e *APIError) Is(target error) bool { return target == ErrAPI }

// IsAuthenticationError reports whether this is a credential or balance problem.
//
// A caller that retries these is not merely wasting a call: the service counts
// them per key, and thirty within a minute earn a three-minute ban. Stop
// instead of backing off.
func (e *APIError) IsAuthenticationError() bool {
	return slices.Contains(authenticationErrorCodes, e.ErrorCode)
}

// IsTerminal reports whether resending the identical request would produce the
// identical failure.
//
// An unknown code returns false, because a code this release has not seen may
// well be transient; that keeps the SDK from talking a caller out of a retry
// that would have worked. The codes it does not recognise as terminal are not
// promised to be retryable either — the retry policy stays with the caller.
func (e *APIError) IsTerminal() bool {
	return slices.Contains(terminalErrorCodes, e.ErrorCode)
}

// IsRateLimited reports whether the service refused the request for throttling
// rather than for anything about the request itself.
//
// Both codes are transient by construction: a rate limit resets with its window
// and a ban expires on its own. [EzCapSolverClient.WaitForResult] treats one as a
// skipped attempt instead of a failed task, and a caller polling by hand with
// [EzCapSolverClient.GetTaskResult] should do the same.
//
// This is narrower than the negation of [APIError.IsTerminal], which is false
// for every unrecognised code as well — including the worker codes that report
// a genuinely failed task.
func (e *APIError) IsRateLimited() bool {
	return slices.Contains(rateLimitedErrorCodes, e.ErrorCode)
}

// Codes that trigger the service's per-key ban counter. Sourced from the
// @ApiDefenses annotations on the service's AsyncTaskController: /createTask
// bans a key for three minutes after thirty of these within one minute, and
// /getTaskResult for one minute after thirty of the first two.
var authenticationErrorCodes = []string{
	"ERROR_KEY_DOES_NOT_EXIST",
	"ERROR_KEY_NOT_AVAILABLE",
	"ERROR_ZERO_BALANCE",
}

// Codes for which an identical request yields an identical failure. Derived
// from the service's TaskResponseCode enum; the rest are left out because an
// internal error, a rate limit, a ban and the two synchronous worker faults
// (ERROR_SERVICE_UNAVAILABLE, ERROR_SERVICE_TIMEOUT) can all clear on their
// own.
var terminalErrorCodes = []string{
	"ERROR_CONTENT_TYPE_ERROR",
	"ERROR_KEY_DOES_NOT_EXIST",
	"ERROR_KEY_NOT_AVAILABLE",
	"ERROR_NOT_FOUND",
	"ERROR_PACKAGE_NOT_EXIST",
	"ERROR_PACKAGE_TASK_TYPE_NOT_SUPPORTED",
	"ERROR_REQUEST_METHOD",
	"ERROR_REQUEST_PARAMETERS",
	"ERROR_REQUEST_PROXY_MISSING",
	"ERROR_SUBSCRIPTION_EXPIRED",
	"ERROR_TASK_NOT_EXIST",
	"ERROR_TASK_TYPE_NOT_ALLOWED",
	"ERROR_TASK_TYPE_NOT_AVAILABLE",
	"ERROR_TASK_TYPE_NOT_SUPPORTED",
	"ERROR_WEBSITE_NOT_ALLOWED",
	"ERROR_ZERO_BALANCE",
}

// The service's two throttling codes, both HTTP 429. Neither says anything
// about a task: the query is refused before the service looks it up, so the
// task keeps running and the next poll can still find it. /getTaskResult counts
// only the first two authentication codes towards its ban counter, so polling
// through a refusal does not dig the hole deeper.
var rateLimitedErrorCodes = []string{
	"ERROR_REQUEST_LIMIT",
	"ERROR_REQUEST_BANNED",
}

// PollingExhaustedError reports that the polling budget ran out before the task
// reached a terminal state (E4).
//
// The task itself may still be running, and it has already been billed, so hand
// [PollingExhaustedError.TaskID] to [EzCapSolverClient.WaitForResult] rather than
// paying for the same work twice. The service holds a result for five minutes
// after creation; past that the id comes back as ERROR_TASK_NOT_EXIST.
type PollingExhaustedError struct {
	// TaskID identifies the unfinished task.
	TaskID string
	// Attempts is the number of result queries that completed.
	Attempts int
	// Interval is the delay that was applied between queries.
	Interval time.Duration
}

func (e *PollingExhaustedError) Error() string {
	return fmt.Sprintf(
		"ezcapsolver: task %q did not complete after %d polling attempts at %s intervals",
		e.TaskID, e.Attempts, e.Interval,
	)
}

// Is reports [ErrPollingExhausted].
func (e *PollingExhaustedError) Is(target error) bool { return target == ErrPollingExhausted }

// WaitInterruptedError reports a task that was created and billed, but whose
// result never arrived: the wait broke off on a dropped connection, a gateway,
// a cancelled context, or a response that did not match the contract.
//
// [EzCapSolverClient.Solve] creates the task internally, so this error is the only
// place its identifier appears. Recover by waiting on the same task again with
// [EzCapSolverClient.WaitForResult] rather than creating a second one: the service
// holds a result for five minutes after creation, and a new task is billed
// again.
//
// It wraps the underlying failure rather than replacing it, so the original
// classification still works — errors.Is against [ErrTransport] or [ErrDecode]
// answers the same as it would have without the wrapper.
//
// Errors that carry the identifier themselves are not wrapped: a business
// failure still arrives as a [*APIError] with TaskID set, and an exhausted
// budget as a [*PollingExhaustedError].
type WaitInterruptedError struct {
	// TaskID identifies the task that was created and billed.
	TaskID string
	// RequestID is the correlation id of the creating request, when the service
	// supplied one.
	RequestID string
	// Err is the failure that interrupted the wait.
	Err error
}

func (e *WaitInterruptedError) Error() string {
	return fmt.Sprintf(
		"ezcapsolver: task %q was created but waiting for its result failed: %v", e.TaskID, e.Err,
	)
}

// Unwrap exposes the underlying failure, which is what keeps errors.Is against
// the original sentinel working through the wrapper.
func (e *WaitInterruptedError) Unwrap() error { return e.Err }

// TaskIDOf returns the identifier of a task that was created and billed, when
// err left one behind, and "" when it did not.
//
// Creating a task is what costs money, so a failure after that point leaves a
// result worth recovering: wait on this identifier again with
// [EzCapSolverClient.WaitForResult] instead of creating a second task. The service
// holds a result for five minutes after creation.
//
// An empty string means nothing was billed — the failure happened before or
// during task creation, so there is nothing to recover.
//
// Which error type carries the identifier is an implementation detail; this is
// the one place to ask.
//
//	if taskID := ezcapsolver.TaskIDOf(err); taskID != "" {
//		result, err := client.WaitForResult(ctx, taskID)
//	}
func TaskIDOf(err error) string {
	var interrupted *WaitInterruptedError
	if errors.As(err, &interrupted) {
		return interrupted.TaskID
	}
	var exhausted *PollingExhaustedError
	if errors.As(err, &exhausted) {
		return exhausted.TaskID
	}
	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError.TaskID
	}
	return ""
}

// missingSolutionReason describes a ready result that carried no solution field
// at all, which is not the same as a null solution. Both entry points into
// solution decoding report it, so the wording lives here rather than twice.
const missingSolutionReason = "ready task result does not contain a solution"

// UnexpectedResponseError reports a response the SDK could not use (E5): it did
// not parse, or it parsed but broke the API contract.
type UnexpectedResponseError struct {
	// Reason states what was wrong with the response.
	Reason string
	// Body is a slice of the raw response body, often the only thing left to
	// diagnose with.
	Body string
}

func (e *UnexpectedResponseError) Error() string {
	if e.Body == "" {
		return "ezcapsolver: unexpected API response: " + e.Reason
	}
	return fmt.Sprintf("ezcapsolver: unexpected API response: %s; body: %s", e.Reason, e.Body)
}

// Is reports [ErrDecode].
func (e *UnexpectedResponseError) Is(target error) bool { return target == ErrDecode }

// SolutionDecodeError reports a raw solution that did not fit its model (E5).
//
// The raw value travels with the error. A decode failure means the worker
// returned a shape this release does not model, and that shape is precisely
// what is needed to diagnose it.
type SolutionDecodeError struct {
	// Raw is the complete solution JSON, untruncated.
	Raw json.RawMessage
	// Err is the underlying decoding error.
	Err error
}

func (e *SolutionDecodeError) Error() string {
	return fmt.Sprintf(
		"ezcapsolver: failed to decode task solution: %v; raw solution: %s",
		e.Err, preview(e.Raw, errorPreviewChars),
	)
}

func (e *SolutionDecodeError) Unwrap() error { return e.Err }

// Is reports [ErrDecode] while leaving the wrapped error reachable.
func (e *SolutionDecodeError) Is(target error) bool { return target == ErrDecode }

// preview truncates a body to maxChars, marking the cut with an ellipsis.
func preview(body []byte, maxChars int) string {
	text := string(body)
	if len(text) <= maxChars {
		return text
	}
	// Cut on a rune boundary so the preview stays valid UTF-8.
	runes := []rune(text)
	if len(runes) <= maxChars {
		return text
	}
	return string(runes[:maxChars]) + "..."
}
