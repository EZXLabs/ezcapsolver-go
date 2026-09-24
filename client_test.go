package ezcapsolver

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

// Every HTTP test runs against a local server. None of them touch the real API:
// creating a task is billed, and a test suite that spends money is a test suite
// nobody runs.

// recordedRequest is one exchange the fake service saw.
type recordedRequest struct {
	Path      string
	Body      map[string]any
	UserAgent string
	Header    http.Header
}

// fakeService is an httptest server that records what it was sent.
type fakeService struct {
	server   *httptest.Server
	mu       sync.Mutex
	requests []recordedRequest
}

// newFakeService starts a server routing by path, where each handler returns the
// response body for that call.
func newFakeService(t *testing.T, routes map[string]func(call int) (int, string)) *fakeService {
	t.Helper()
	service := &fakeService{}
	counts := map[string]int{}

	service.server = httptest.NewServer(http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			body, _ := io.ReadAll(request.Body)
			var decoded map[string]any
			_ = json.Unmarshal(body, &decoded)

			service.mu.Lock()
			service.requests = append(service.requests, recordedRequest{
				Path:      request.URL.Path,
				Body:      decoded,
				UserAgent: request.Header.Get("User-Agent"),
				Header:    request.Header.Clone(),
			})
			counts[request.URL.Path]++
			call := counts[request.URL.Path]
			service.mu.Unlock()

			handler, known := routes[request.URL.Path]
			if !known {
				writer.WriteHeader(http.StatusNotFound)
				return
			}
			status, response := handler(call)
			writer.Header().Set("Content-Type", "application/json")
			writer.WriteHeader(status)
			_, _ = io.WriteString(writer, response)
		}))
	t.Cleanup(service.server.Close)
	return service
}

// calls returns the recorded exchanges.
func (s *fakeService) calls() []recordedRequest {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]recordedRequest(nil), s.requests...)
}

// countOf returns how many times a path was called.
func (s *fakeService) countOf(path string) int {
	count := 0
	for _, request := range s.calls() {
		if request.Path == path {
			count++
		}
	}
	return count
}

// newTestClient points a client at the fake service, with a polling interval
// short enough to keep the suite fast.
func newTestClient(t *testing.T, service *fakeService, options ...Option) *EzCapSolverClient {
	t.Helper()
	base := []Option{
		WithClientKey("test-client-key"),
		WithBaseURLs(service.server.URL, service.server.URL),
		WithPolling(PollingConfig{Interval: time.Millisecond, MaxAttempts: 5}),
	}
	client, err := NewClient(append(base, options...)...)
	if err != nil {
		t.Fatalf("building the client failed: %v", err)
	}
	return client
}

// ok is a shorthand for a 200 response with a fixed body.
func ok(body string) func(int) (int, string) {
	return func(int) (int, string) { return http.StatusOK, body }
}

// TestCreateTaskRequestEnvelope checks the body and headers of a task creation.
func TestCreateTaskRequestEnvelope(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask: ok(`{"errorId":0,"taskId":"task-1","requestId":"request-1"}`),
	})
	client := newTestClient(t, service, WithAppID(42))

	created, err := client.CreateTask(t.Context(), TaskTypeRecaptchaV2TaskProxyless,
		&RecaptchaV2Task{WebsiteURL: "https://example.com", WebsiteKey: "site-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created.TaskID != "task-1" || created.RequestID != "request-1" {
		t.Errorf("got %+v", created)
	}

	call := service.calls()[0]
	if call.Body["clientKey"] != "test-client-key" {
		t.Errorf("clientKey: got %v", call.Body["clientKey"])
	}
	// appId is an integer service-side; a string would be rejected.
	if call.Body["appId"] != float64(42) {
		t.Errorf("appId must be a JSON number, got %#v", call.Body["appId"])
	}

	task, isObject := call.Body["task"].(map[string]any)
	if !isObject {
		t.Fatalf("task must be an object, got %#v", call.Body["task"])
	}
	if task["type"] != string(TaskTypeRecaptchaV2TaskProxyless) {
		t.Errorf("type: got %v", task["type"])
	}
	if task["websiteURL"] != "https://example.com" {
		t.Errorf("the parameters must sit next to the type, got %v", task)
	}

	if call.UserAgent != defaultUserAgent {
		t.Errorf("User-Agent: got %q, want %q", call.UserAgent, defaultUserAgent)
	}
	if got := call.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type: got %q; the service answers 415 for anything else", got)
	}
}

// TestCreateTaskRejectsAnEmptyTaskID checks that a creation response without a
// usable identifier is reported rather than returned.
func TestCreateTaskRejectsAnEmptyTaskID(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask: ok(`{"errorId":0,"taskId":"  "}`),
	})
	client := newTestClient(t, service)

	_, err := client.CreateTask(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("got %v, want a decode failure", err)
	}
}

// TestPollingPaths covers the four ways waiting can go.
func TestPollingPaths(t *testing.T) {
	t.Run("waits before the first query", func(t *testing.T) {
		// A task that was just created is still queued, so querying immediately
		// only wastes a request. The first wait is not an optimisation to skip.
		service := newFakeService(t, map[string]func(int) (int, string){
			pathGetTaskResult: ok(`{"errorId":0,"status":"ready","solution":{"token":"x"}}`),
		})
		client := newTestClient(t, service,
			WithPolling(PollingConfig{Interval: 40 * time.Millisecond, MaxAttempts: 3}))

		started := time.Now()
		if _, err := client.WaitForResult(t.Context(), "task-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if elapsed := time.Since(started); elapsed < 40*time.Millisecond {
			t.Errorf("queried after %s, before the first interval elapsed", elapsed)
		}
	})

	t.Run("keeps polling while processing", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathGetTaskResult: func(call int) (int, string) {
				if call < 3 {
					return http.StatusOK, `{"errorId":0,"status":"processing"}`
				}
				return http.StatusOK, `{"errorId":0,"status":"ready","solution":{"token":"x"}}`
			},
		})
		client := newTestClient(t, service)

		result, err := client.WaitForResult(t.Context(), "task-1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.IsReady() {
			t.Errorf("status: got %q", result.Status)
		}
		if got := service.countOf(pathGetTaskResult); got != 3 {
			t.Errorf("expected three queries, got %d", got)
		}
	})

	t.Run("returns as soon as it is ready", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathGetTaskResult: ok(`{"errorId":0,"status":"ready","solution":{"token":"x"}}`),
		})
		client := newTestClient(t, service)

		if _, err := client.WaitForResult(t.Context(), "task-1"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := service.countOf(pathGetTaskResult); got != 1 {
			t.Errorf("expected one query, got %d", got)
		}
	})

	t.Run("gives up once the budget runs out", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathGetTaskResult: ok(`{"errorId":0,"status":"processing"}`),
		})
		client := newTestClient(t, service)

		_, err := client.WaitForResult(t.Context(), "task-1")
		if !errors.Is(err, ErrPollingExhausted) {
			t.Fatalf("got %v, want a polling timeout", err)
		}

		var exhausted *PollingExhaustedError
		if !errors.As(err, &exhausted) {
			t.Fatalf("got %T", err)
		}
		// The task may still finish, so the caller needs its identifier to
		// fetch the result later.
		if exhausted.TaskID != "task-1" || exhausted.Attempts != 5 {
			t.Errorf("got %+v", exhausted)
		}
		if got := service.countOf(pathGetTaskResult); got != 5 {
			t.Errorf("expected five queries, got %d", got)
		}
	})
}

// TestWaitForResultRejectsAnUnknownStatus checks that a status outside the three
// the service defines fails loudly instead of being read as "still processing",
// which would turn a contract break into an unexplained timeout.
func TestWaitForResultRejectsAnUnknownStatus(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetTaskResult: ok(`{"errorId":0,"status":"queued"}`),
	})
	client := newTestClient(t, service)

	_, err := client.WaitForResult(t.Context(), "task-1")
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("got %v, want a decode failure", err)
	}
	if !strings.Contains(err.Error(), "queued") {
		t.Errorf("the offending value must be in the message, got %q", err)
	}
}

// TestSolveAttachesTaskContext checks that a failure while waiting can still be
// traced back to its task. The service does not echo the task ID on a result
// query, so the SDK has to add it.
func TestSolveAttachesTaskContext(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask: ok(`{"errorId":0,"taskId":"task-7","requestId":"request-7"}`),
		pathGetTaskResult: func(int) (int, string) {
			return http.StatusInternalServerError,
				`{"errorId":1,"errorCode":"ERROR_TASK_NOT_EXIST","errorDescription":"gone"}`
		},
	})
	client := newTestClient(t, service)

	_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})
	apiErr := requireAPIError(t, err)
	if apiErr.TaskID != "task-7" {
		t.Errorf("taskId: got %q, want task-7", apiErr.TaskID)
	}
	if apiErr.RequestID != "request-7" {
		t.Errorf("requestId: got %q, want request-7", apiErr.RequestID)
	}
}

// TestNoAutomaticRetry checks that one call produces exactly one request.
//
// Creating a task is billed and is not idempotent, and repeating a credential
// error thirty times in a minute gets the key banned, so retrying is a decision
// only the caller can make.
func TestNoAutomaticRetry(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask: func(int) (int, string) {
			return http.StatusInternalServerError,
				`{"errorId":1,"errorCode":"ERROR_ZERO_BALANCE","errorDescription":"Inadequate balance"}`
		},
	})
	client := newTestClient(t, service)

	_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})
	if !requireAPIError(t, err).IsAuthenticationError() {
		t.Error("a balance failure must be flagged so callers stop rather than back off")
	}
	if got := service.countOf(pathCreateTask); got != 1 {
		t.Errorf("expected exactly one request, got %d", got)
	}
}

// TestTypedSolverEndToEnd runs an asynchronous convenience method the whole way
// through and checks that the raw value survives alongside the decoded one.
func TestTypedSolverEndToEnd(t *testing.T) {
	solution := `{"gRecaptchaResponse":"the-token","sec_ch_ua":"ua","user_agent":"agent","new":1}`
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask:    ok(`{"errorId":0,"taskId":"task-1"}`),
		pathGetTaskResult: ok(`{"errorId":0,"status":"ready","solution":` + solution + `}`),
	})
	client := newTestClient(t, service)

	solved, err := client.SolveRecaptchaV2TaskProxyless(t.Context(), &RecaptchaV2Task{
		WebsiteURL: "https://example.com",
		WebsiteKey: "site-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solved.Solution.Token != "the-token" {
		t.Errorf("token: got %q", solved.Solution.Token)
	}
	if solved.TaskID != "task-1" {
		t.Errorf("taskId: got %q", solved.TaskID)
	}
	wantExtra(t, solved.Solution.Extra, map[string]any{"new": float64(1)})
	if string(solved.Raw) != solution {
		t.Errorf("the raw value must survive decoding\n got: %s\nwant: %s", solved.Raw, solution)
	}

	task := service.calls()[0].Body["task"].(map[string]any)
	if task["type"] != string(TaskTypeRecaptchaV2TaskProxyless) {
		t.Errorf("the method must pick its own task type, got %v", task["type"])
	}
}

// TestSyncSolverEndToEnd runs a synchronous convenience method, which answers in
// one request and carries no task ID.
func TestSyncSolverEndToEnd(t *testing.T) {
	solution := `{"status":1,"code":200,"headers":{"a":"b"},"cookies":{},"body":"<html/>"}`
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(`{"errorId":0,"taskId":"sync-task-1","status":"ready","solution":` + solution + `}`),
	})
	client := newTestClient(t, service)

	solved, err := client.SyncSolveTLSTask(t.Context(), &TLSForwardTask{
		TLSType: "chrome",
		Proxy:   "http://user:pass@host:8080",
		Method:  TLSMethodGET,
		URL:     "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solved.Solution.Code != 200 || solved.Solution.Body != "<html/>" {
		t.Errorf("got %+v", solved.Solution)
	}
	// The synchronous endpoint assigns an identifier and returns it alongside
	// the result. Dropping it would throw away the only handle on the task.
	if solved.TaskID != "sync-task-1" {
		t.Errorf("TaskID = %q, want the id the service returned", solved.TaskID)
	}
	if service.countOf(pathCreateSyncTask) != 1 || service.countOf(pathGetTaskResult) != 0 {
		t.Error("a synchronous task must not be polled")
	}
}

// TestSyncSolveKeepsTheAssignedTaskID covers the untyped escape hatch, which
// builds its Solved separately from the typed convenience methods.
func TestSyncSolveKeepsTheAssignedTaskID(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(`{"errorId":0,"taskId":"sync-task-2","status":"ready","solution":{"token":"t"}}`),
	})
	client := newTestClient(t, service)

	solved, err := client.SyncSolve(t.Context(), TaskTypeTLSTask, map[string]any{"url": "https://example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solved.TaskID != "sync-task-2" {
		t.Errorf("TaskID = %q, want the id the service returned", solved.TaskID)
	}
}

// TestClassificationSolverEndToEnd checks typed decoding and the original result.
func TestClassificationSolverEndToEnd(t *testing.T) {
	raw := `{"type":"multi","objects":[1,4],"confidence":0.9}`
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(
			`{"errorId":0,"status":"ready","requestId":"r-1","solution":` + raw + `}`),
	})
	client := newTestClient(t, service)

	solved, err := client.SyncSolveRecaptchaV2Classification(t.Context(),
		&RecaptchaV2ClassificationTask{Image: "base64", Question: "crosswalk"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !solved.Solution.IsMulti() || solved.Solution.IsSingle() {
		t.Fatalf("unexpected result kind: %+v", solved.Solution)
	}
	if !reflect.DeepEqual(solved.Solution.Objects, []int{1, 4}) {
		t.Errorf("objects: got %v", solved.Solution.Objects)
	}
	wantExtra(t, solved.Solution.Extra, map[string]any{"confidence": 0.9})
	if string(solved.Raw) != raw || solved.RequestID != "r-1" || solved.TaskID != "" {
		t.Errorf("unexpected task metadata or raw result: %+v", solved)
	}
	if service.countOf(pathCreateSyncTask) != 1 || service.countOf(pathGetTaskResult) != 0 {
		t.Error("a synchronous task must not be polled")
	}
}

// TestSyncEndpointUsesItsOwnTimeout checks that the two timeout budgets stay
// separate.
//
// A synchronous call blocks until the worker answers, and the service allows
// some task types three minutes. Sharing the asynchronous timeout would abort a
// call that has already been billed, with no task ID to recover the result.
func TestSyncEndpointUsesItsOwnTimeout(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: func(int) (int, string) {
			time.Sleep(60 * time.Millisecond)
			return http.StatusOK, `{"errorId":0,"status":"ready","solution":{"payload":"p"}}`
		},
	})
	client := newTestClient(t, service,
		WithTimeout(10*time.Millisecond),
		WithSyncTimeout(5*time.Second),
	)

	if _, err := client.SyncSolveAkamaiSBSDTaskProxyless(
		t.Context(), &AkamaiSBSDTask{PageURL: "https://example.com"},
	); err != nil {
		t.Fatalf("the synchronous call was cut off by the asynchronous timeout: %v", err)
	}
}

// TestEscapeHatch checks that a task type the SDK does not model can still be
// used, which is what keeps a caller from having to wait for a release.
// TestEscapeHatchIsSymmetric checks that the untyped and typed escape hatches
// both exist for either endpoint, and that the two return the same shape.
//
// An asymmetry here would push the choice of endpoint back into the caller's
// result handling: code written against one endpoint would not compile against
// the other, which is the opposite of what having both is for.
func TestEscapeHatchIsSymmetric(t *testing.T) {
	solution := `{"anything":"goes"}`
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask:     ok(`{"errorId":0,"taskId":"task-1"}`),
		pathGetTaskResult:  ok(`{"errorId":0,"status":"ready","requestId":"r-1","solution":` + solution + `}`),
		pathCreateSyncTask: ok(`{"errorId":0,"status":"ready","requestId":"r-1","solution":` + solution + `}`),
	})
	client := newTestClient(t, service)

	polled, err := client.Solve(t.Context(), "BrandNewTaskType", nil)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	synced, err := client.SyncSolve(t.Context(), "BrandNewTaskType", nil)
	if err != nil {
		t.Fatalf("SyncSolve: %v", err)
	}

	// Identical result type, identical payload; only the task ID differs.
	if string(polled.Raw) != string(synced.Raw) {
		t.Errorf("Raw differs: %s vs %s", polled.Raw, synced.Raw)
	}
	if string(polled.Solution) != string(synced.Solution) {
		t.Errorf("Solution differs: %s vs %s", polled.Solution, synced.Solution)
	}
	if polled.RequestID != synced.RequestID {
		t.Errorf("RequestID differs: %q vs %q", polled.RequestID, synced.RequestID)
	}
	if polled.TaskID == "" {
		t.Error("a polled task must carry its task ID")
	}
	if synced.TaskID != "" {
		t.Errorf("the synchronous endpoint assigns no task ID, got %q", synced.TaskID)
	}

	if service.countOf(pathCreateSyncTask) != 1 || service.countOf(pathGetTaskResult) != 1 {
		t.Errorf("each endpoint must be reached once, got %v", pathsOf(service.calls()))
	}

	// The typed escape hatch has to exist for both endpoints too.
	type myShape struct {
		Anything string `json:"anything"`
	}
	typedPolled, err := SolveAs[myShape](t.Context(), client, "BrandNewTaskType", nil)
	if err != nil {
		t.Fatalf("SolveAs: %v", err)
	}
	typedSynced, err := SyncSolveAs[myShape](t.Context(), client, "BrandNewTaskType", nil)
	if err != nil {
		t.Fatalf("SyncSolveAs: %v", err)
	}
	if typedPolled.Solution != typedSynced.Solution {
		t.Errorf("decoded results differ: %+v vs %+v", typedPolled.Solution, typedSynced.Solution)
	}
}

func TestEscapeHatch(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask:    ok(`{"errorId":0,"taskId":"task-1"}`),
		pathGetTaskResult: ok(`{"errorId":0,"status":"ready","solution":{"anything":"goes"}}`),
	})
	client := newTestClient(t, service)

	t.Run("raw", func(t *testing.T) {
		solved, err := client.Solve(t.Context(), "BrandNewTaskType", map[string]any{"custom": 1})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if string(solved.Raw) != `{"anything":"goes"}` {
			t.Errorf("got %s", solved.Raw)
		}

		task := service.calls()[0].Body["task"].(map[string]any)
		if task["type"] != "BrandNewTaskType" || task["custom"] != float64(1) {
			t.Errorf("got %v", task)
		}
	})

	t.Run("decoded into a caller-supplied type", func(t *testing.T) {
		type myShape struct {
			Anything string `json:"anything"`
		}
		solved, err := SolveAs[myShape](t.Context(), client, "BrandNewTaskType", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if solved.Solution.Anything != "goes" {
			t.Errorf("got %+v", solved.Solution)
		}
	})
}

// TestCreateSyncTaskRejectsProcessing checks that the one status the synchronous
// endpoint must never return is reported. There would be no task ID to follow up
// with, so returning it as a result would strand the caller.
func TestCreateSyncTaskRejectsProcessing(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(`{"errorId":0,"status":"processing"}`),
	})
	client := newTestClient(t, service)

	_, err := client.CreateSyncTask(t.Context(), TaskTypeTLSTask, &TLSForwardTask{})
	if !errors.Is(err, ErrDecode) {
		t.Fatalf("got %v, want a decode failure", err)
	}
}

// TestBalance checks the one endpoint that returns a plain number.
func TestBalance(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetBalance: ok(`{"errorId":0,"balance":12.3456}`),
	})
	client := newTestClient(t, service)

	balance, err := client.Balance(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 12.3456 {
		t.Errorf("got %v", balance)
	}
}

// TestContextCancellationStopsPolling checks that a cancelled context ends the
// wait rather than running out the full budget.
func TestContextCancellationStopsPolling(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetTaskResult: ok(`{"errorId":0,"status":"processing"}`),
	})
	client := newTestClient(t, service,
		WithPolling(PollingConfig{Interval: time.Second, MaxAttempts: 100}))

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	started := time.Now()
	_, err := client.WaitForResult(ctx, "task-1")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want a cancellation", err)
	}
	if !errors.Is(err, ErrTransport) {
		t.Errorf("a cancellation belongs to the transport layer, got %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Errorf("cancellation was not honoured promptly: %s", elapsed)
	}
}

// TestMissingSolutionIsReported checks the difference between a ready result
// with no solution field and one whose solution is null.
func TestMissingSolutionIsReported(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(`{"errorId":0,"status":"ready"}`),
	})
	client := newTestClient(t, service)

	// Assert down to Reason: this arrives as the general UnexpectedResponseError,
	// so matching the sentinel alone would let any contract violation pass.
	_, err := client.SyncSolveIncapsulaTaskProxyless(t.Context(), &IncapsulaTask{})
	var unexpected *UnexpectedResponseError
	if !errors.As(err, &unexpected) || !contains(unexpected.Reason, "does not contain a solution") {
		t.Fatalf("got %v, want a missing-solution report", err)
	}

	result, err := client.CreateSyncTask(t.Context(), TaskTypeIncapsulaTaskProxyless, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HasSolution() {
		t.Error("an absent solution field must not read as present")
	}
}

// TestNullSolutionIsNotMissing checks that a worker's empty answer stays
// tellable apart from an unfinished task.
func TestNullSolutionIsNotMissing(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(`{"errorId":0,"status":"ready","solution":null}`),
	})
	client := newTestClient(t, service)

	result, err := client.CreateSyncTask(t.Context(), TaskTypeIncapsulaTaskProxyless, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.HasSolution() {
		t.Error("a null solution is present; the worker answered, it just answered nothing")
	}
	if string(result.Solution) != "null" {
		t.Errorf("got %s", result.Solution)
	}
}

// TestTransportFailureIsClassified checks that a connection failure is reported
// as a transport error with the cause still reachable.
func TestTransportFailureIsClassified(t *testing.T) {
	client, err := NewClient(
		WithClientKey("test-client-key"),
		// Port 0 is not connectable, which is the point.
		WithBaseURLs("http://127.0.0.1:0", "http://127.0.0.1:0"),
	)
	if err != nil {
		t.Fatalf("building the client failed: %v", err)
	}

	_, err = client.Balance(t.Context())
	if !errors.Is(err, ErrTransport) {
		t.Fatalf("got %v, want a transport failure", err)
	}
	var transportErr *TransportError
	if !errors.As(err, &transportErr) {
		t.Fatalf("got %T", err)
	}
	if !strings.Contains(transportErr.Op, pathGetBalance) {
		t.Errorf("the operation must be identifiable, got %q", transportErr.Op)
	}
}

// TestClientIsSafeForConcurrentUse checks that one client can be shared, which
// is how it is meant to be used: it holds the connection pool.
func TestClientIsSafeForConcurrentUse(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetBalance: ok(`{"errorId":0,"balance":1}`),
	})
	client := newTestClient(t, service)

	var group sync.WaitGroup
	for range 20 {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := client.Balance(t.Context()); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	group.Wait()

	if got := service.countOf(pathGetBalance); got != 20 {
		t.Errorf("expected 20 requests, got %d", got)
	}
}

// TestBalanceRejectsAMissingValue keeps "no balance field" from arriving as a
// zero balance. The two call for opposite reactions, so they must not look alike.
func TestBalanceRejectsAMissingValue(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetBalance: ok(`{"errorId":0}`),
	})
	client := newTestClient(t, service)

	if _, err := client.Balance(t.Context()); !errors.Is(err, ErrDecode) {
		t.Errorf("a missing balance must be reported, got %v", err)
	}
}

// TestBalanceReturnsARealZero is the other half: zero is a legitimate value.
func TestBalanceReturnsARealZero(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetBalance: ok(`{"errorId":0,"balance":0}`),
	})
	client := newTestClient(t, service)

	balance, err := client.Balance(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if balance != 0 {
		t.Errorf("balance = %v, want 0", balance)
	}
}

// TestAWaitFailureCarriesTheBilledTaskID checks that a failure with nowhere of
// its own to put the identifier still carries it out.
//
// Solve creates the task internally, so this error is the only place its id
// appears. Losing it strands a result the caller already paid for: the service
// holds one for five minutes, but only for whoever still knows the id.
func TestAWaitFailureCarriesTheBilledTaskID(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask: ok(`{"errorId":0,"taskId":"task-7","requestId":"request-7"}`),
		// A success envelope with no status at all: a contract break rather than
		// a business error, so nothing upstream attaches the task id.
		pathGetTaskResult: ok(`{"errorId":0}`),
	})
	client := newTestClient(t, service)

	_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})

	var interrupted *WaitInterruptedError
	if !errors.As(err, &interrupted) {
		t.Fatalf("got %v, want a *WaitInterruptedError", err)
	}
	if interrupted.TaskID != "task-7" {
		t.Errorf("taskId: got %q, want task-7", interrupted.TaskID)
	}
	if interrupted.RequestID != "request-7" {
		t.Errorf("requestId: got %q, want request-7", interrupted.RequestID)
	}

	// The wrapper must not cost the caller the original classification.
	if !errors.Is(err, ErrDecode) {
		t.Error("the underlying failure must stay reachable through Unwrap")
	}
}

// TestIdentifiedFailuresAreNotWrappedTwice checks the two error kinds that
// already carry the task id are left alone, so existing matches keep working.
func TestIdentifiedFailuresAreNotWrappedTwice(t *testing.T) {
	t.Run("exhausted budget", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask:    ok(`{"errorId":0,"taskId":"task-7"}`),
			pathGetTaskResult: ok(`{"errorId":0,"status":"processing"}`),
		})
		client := newTestClient(t, service)

		_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})

		var interrupted *WaitInterruptedError
		if errors.As(err, &interrupted) {
			t.Error("PollingExhaustedError already carries the id; wrapping adds a layer")
		}
		var exhausted *PollingExhaustedError
		if !errors.As(err, &exhausted) || exhausted.TaskID != "task-7" {
			t.Errorf("got %v, want a *PollingExhaustedError for task-7", err)
		}
	})

	t.Run("business error", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask: ok(`{"errorId":0,"taskId":"task-7"}`),
			pathGetTaskResult: func(int) (int, string) {
				return http.StatusInternalServerError,
					`{"errorId":1,"errorCode":"ERROR_TASK_NOT_EXIST"}`
			},
		})
		client := newTestClient(t, service)

		_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})

		var interrupted *WaitInterruptedError
		if errors.As(err, &interrupted) {
			t.Error("APIError has its own TaskID field; wrapping breaks existing matches")
		}
		if requireAPIError(t, err).TaskID != "task-7" {
			t.Error("the business error must still be tagged in place")
		}
	})
}

// TestThrottledPollsAreRetried covers the one API error the polling loop does
// not treat as the task's answer. A throttled query is refused before the
// service ever looks the task up, so the task is still queued — and it has
// already been billed, which is what makes giving up on it expensive.
func TestThrottledPollsAreRetried(t *testing.T) {
	t.Run("a rate limit only costs an attempt", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask: ok(`{"errorId":0,"taskId":"task-9"}`),
			pathGetTaskResult: func(call int) (int, string) {
				if call == 1 {
					return http.StatusTooManyRequests,
						`{"errorId":1,"errorCode":"ERROR_REQUEST_LIMIT"}`
				}
				return http.StatusOK, `{"errorId":0,"status":"ready","solution":{"token":"t"}}`
			},
		})
		client := newTestClient(t, service)

		solved, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})
		if err != nil {
			t.Fatalf("a throttled poll must not fail the task: %v", err)
		}
		if solved.TaskID != "task-9" {
			t.Errorf("taskId: got %q, want task-9", solved.TaskID)
		}
		if got := service.countOf(pathGetTaskResult); got != 2 {
			t.Errorf("polled %d times, want 2", got)
		}
	})

	t.Run("a ban is retried too", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask: ok(`{"errorId":0,"taskId":"task-9"}`),
			pathGetTaskResult: func(call int) (int, string) {
				if call < 3 {
					return http.StatusTooManyRequests,
						`{"errorId":1,"errorCode":"ERROR_REQUEST_BANNED"}`
				}
				return http.StatusOK, `{"errorId":0,"status":"ready","solution":{"token":"t"}}`
			},
		})
		client := newTestClient(t, service)

		if _, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{}); err != nil {
			t.Fatalf("a ban clears on its own, so the wait must continue: %v", err)
		}
	})

	t.Run("throttling still runs out the budget", func(t *testing.T) {
		// Retrying must not turn a five-attempt budget into an endless loop:
		// the result is only held for five minutes, so the wall clock the
		// budget stands for has to keep running.
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask: ok(`{"errorId":0,"taskId":"task-9"}`),
			pathGetTaskResult: func(int) (int, string) {
				return http.StatusTooManyRequests,
					`{"errorId":1,"errorCode":"ERROR_REQUEST_LIMIT"}`
			},
		})
		client := newTestClient(t, service)

		_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})

		var exhausted *PollingExhaustedError
		if !errors.As(err, &exhausted) || exhausted.TaskID != "task-9" {
			t.Fatalf("got %v, want a *PollingExhaustedError for task-9", err)
		}
		if got := service.countOf(pathGetTaskResult); got != 5 {
			t.Errorf("polled %d times, want the configured 5", got)
		}
	})

	t.Run("any other API error is the answer", func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask: ok(`{"errorId":0,"taskId":"task-9"}`),
			pathGetTaskResult: func(int) (int, string) {
				return http.StatusInternalServerError,
					`{"errorId":1,"errorCode":"ERROR_INTERNAL_SERVER_ERROR"}`
			},
		})
		client := newTestClient(t, service)

		_, err := client.Solve(t.Context(), TaskTypeHCaptcha, &HCaptchaTask{})

		if !errors.Is(err, ErrAPI) {
			t.Fatalf("got %v, want the API error to end the wait", err)
		}
		if got := service.countOf(pathGetTaskResult); got != 1 {
			t.Errorf("polled %d times, want the wait to stop on the first", got)
		}
	})
}

// TestTaskIDOfAsksOnce checks that callers can ask one question — "is there a
// billed task to rescue?" — without knowing which error type carries the id.
func TestTaskIDOfAsksOnce(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"interrupted wait", &WaitInterruptedError{TaskID: "task-1", Err: errors.New("reset")}, "task-1"},
		{"exhausted budget", &PollingExhaustedError{TaskID: "task-2"}, "task-2"},
		{"business error", &APIError{TaskID: "task-3"}, "task-3"},
		// Failing before creation means no task and no charge, so there is
		// nothing to recover.
		{"failed before creation", &TransportError{Op: "POST /createTask", Err: errors.New("dial")}, ""},
		{"nil", nil, ""},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := TaskIDOf(testCase.err); got != testCase.want {
				t.Errorf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

// TestCloseOnlyTouchesTheClientItBuilt checks that Close leaves a caller's own
// HTTP client alone: that one may be shared with the rest of their program.
func TestCloseOnlyTouchesTheClientItBuilt(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetBalance: ok(`{"errorId":0,"balance":1.5}`),
	})

	t.Run("its own", func(t *testing.T) {
		client := newTestClient(t, service)
		if !client.ownsHTTP {
			t.Fatal("a client that built its own transport must own it")
		}
		client.Close()

		// Closing idle connections is not a teardown: the client keeps working.
		if _, err := client.Balance(t.Context()); err != nil {
			t.Errorf("the client must stay usable after Close: %v", err)
		}
	})

	t.Run("supplied", func(t *testing.T) {
		client := newTestClient(t, service, WithHTTPClient(&http.Client{}))
		if client.ownsHTTP {
			t.Fatal("a supplied client must not be owned")
		}
		client.Close() // must be a no-op rather than reaching into the caller's pool
	})
}
