package ezcapsolver

import (
	"encoding/json"
	"fmt"
)

// API paths. All three controllers also mount under /solver/v1, but the root
// paths are what existing clients use, so the SDK stays on those.
const (
	pathCreateTask     = "/createTask"
	pathGetTaskResult  = "/getTaskResult"
	pathCreateSyncTask = "/createSyncTask"
	pathGetBalance     = "/getBalance"
)

// createTaskRequest is the body of /createTask and /createSyncTask.
type createTaskRequest struct {
	ClientKey string      `json:"clientKey"`
	AppID     *int        `json:"appId,omitzero"`
	Task      taskPayload `json:"task"`
}

// queryTaskResultRequest is the body of /getTaskResult.
type queryTaskResultRequest struct {
	ClientKey string `json:"clientKey"`
	TaskID    string `json:"taskId"`
}

// queryBalanceRequest is the body of /getBalance.
type queryBalanceRequest struct {
	ClientKey string `json:"clientKey"`
}

// taskPayload injects the task type into the serialized parameters.
//
// The type belongs in the same object as the parameters, but it is not a field
// on any task model: one parameter model backs several task types — the five
// ReCaptcha V2 variants share one — so a type field would either block that
// reuse or let a caller set a type that contradicts the method they called.
type taskPayload struct {
	taskType TaskType
	params   any
}

func (p taskPayload) MarshalJSON() ([]byte, error) {
	fields := map[string]json.RawMessage{}
	if p.params != nil {
		encoded, err := json.Marshal(p.params)
		if err != nil {
			return nil, fmt.Errorf("task parameters are not serializable: %w", err)
		}
		// A nil map or a nil pointer marshals to `null`, which is a valid
		// document but not an object; treat it as no parameters at all.
		if string(encoded) != "null" {
			if err := json.Unmarshal(encoded, &fields); err != nil {
				return nil, fmt.Errorf("task parameters are not a JSON object: %w", err)
			}
		}
	}

	// Written last so the SDK's type always wins over anything in the
	// parameters, including a stray `type` key passed through the escape hatch.
	encodedType, err := json.Marshal(string(p.taskType))
	if err != nil {
		return nil, err
	}
	fields["type"] = encodedType
	return json.Marshal(fields)
}

var _ json.Marshaler = taskPayload{}
