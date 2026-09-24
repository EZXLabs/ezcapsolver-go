package ezcapsolver

import "encoding/json"

// TaskStatus is the state the service reports for a task.
//
// The service defines exactly three, so an unrecognised value is a contract
// break and is reported as [UnexpectedResponseError] rather than being quietly
// treated as "still processing".
type TaskStatus string

const (
	// TaskStatusProcessing means the task is still being worked on.
	TaskStatusProcessing TaskStatus = "processing"
	// TaskStatusReady means the task succeeded and its solution is available.
	TaskStatusReady TaskStatus = "ready"
	// TaskStatusError means the task failed. Such a response also carries a
	// non-zero errorId, so it surfaces as an [APIError] rather than a result.
	TaskStatusError TaskStatus = "error"
)

func (s TaskStatus) String() string { return string(s) }

// IsValid reports whether the status is one the service defines.
func (s TaskStatus) IsValid() bool {
	switch s {
	case TaskStatusProcessing, TaskStatusReady, TaskStatusError:
		return true
	default:
		return false
	}
}

// ResponseMeta is the envelope every API response carries.
type ResponseMeta struct {
	// ErrorID is 0 on success and 1 on failure. It is the only success
	// criterion; the HTTP status is not.
	ErrorID int `json:"errorId"`
	// RequestID is the service-side request tracking identifier.
	RequestID string `json:"requestId,omitzero"`
	// ErrorCode is present only on failure.
	ErrorCode string `json:"errorCode,omitzero"`
	// ErrorDescription is present only on failure.
	ErrorDescription string `json:"errorDescription,omitzero"`
}

// CreateTaskResponse is what /createTask returns.
type CreateTaskResponse struct {
	ResponseMeta
	// TaskID identifies the created task. Use it to query the result.
	TaskID string `json:"taskId"`
}

// TaskResult is what /getTaskResult and /createSyncTask return.
type TaskResult struct {
	ResponseMeta
	// Status is the current task state.
	Status TaskStatus `json:"status"`
	// TaskID is the identifier the synchronous endpoint assigns.
	//
	// /createSyncTask returns one on both the success and the failure path.
	// /getTaskResult does not echo it back, so it is empty there.
	TaskID string `json:"taskId,omitzero"`
	// Solution is the raw solution JSON, exactly as the worker produced it.
	//
	// It is nil when the response carried no solution field, which means the
	// task has not finished. That is different from the four bytes `null`, which
	// mean the worker finished and returned nothing. Use
	// [TaskResult.HasSolution] rather than comparing against nil by hand.
	Solution json.RawMessage `json:"solution,omitzero"`
}

// HasSolution reports whether the response carried a solution field at all,
// a JSON null included.
func (r TaskResult) HasSolution() bool { return len(r.Solution) > 0 }

// IsProcessing reports whether the task is still being worked on.
func (r TaskResult) IsProcessing() bool { return r.Status == TaskStatusProcessing }

// IsReady reports whether the task completed successfully.
func (r TaskResult) IsReady() bool { return r.Status == TaskStatusReady }

// IsError reports whether the task failed.
func (r TaskResult) IsError() bool { return r.Status == TaskStatusError }

// DecodeSolution decodes the raw solution into target, which must be a pointer.
//
// The raw value stays available afterwards, so a decoding failure loses nothing.
// It returns a [*UnexpectedResponseError] when there was no solution field, and
// a [*SolutionDecodeError] carrying the raw value when the value did not fit.
//
// A ready result without a solution is a broken contract rather than a category
// of its own: the task was billed and the answer is gone. Use [TaskResult.HasSolution]
// to tell that apart from a solution that was present and null.
func (r TaskResult) DecodeSolution(target any) error {
	if !r.HasSolution() {
		return &UnexpectedResponseError{Reason: missingSolutionReason}
	}
	if err := json.Unmarshal(r.Solution, target); err != nil {
		return &SolutionDecodeError{Raw: cloneRaw(r.Solution), Err: err}
	}
	return nil
}

// Solved is a finished task: its decoded solution plus the identifiers and the
// raw JSON it came from.
type Solved[T any] struct {
	// Solution is the decoded result.
	Solution T
	// Raw is the untouched solution JSON. Anything the typed model does not
	// declare is still recoverable from here.
	Raw json.RawMessage
	// TaskID is the identifier the service assigned. Both endpoints supply
	// one: the asynchronous path from task creation, the synchronous path
	// alongside the result.
	TaskID string
	// RequestID is the service-side request tracking identifier.
	RequestID string
}

// balanceResponse is internal: the balance reaches callers as a plain number.
type balanceResponse struct {
	ResponseMeta
	// Balance is a JSON number with at most four decimal places, which float64
	// round-trips exactly. It is for display, not for accounting arithmetic.
	//
	// It is a pointer so that a missing field is tellable from a zero balance.
	// The two call for opposite reactions — stop spending versus report a broken
	// response — so they must not look alike.
	Balance *float64 `json:"balance"`
}
