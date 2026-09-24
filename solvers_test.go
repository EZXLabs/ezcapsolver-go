package ezcapsolver

import (
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"
)

// anySolution satisfies the required field of every solution model at once, so
// one fake response can serve all 23 convenience methods. A model only decodes
// the keys it declares, so the extra ones are harmless.
const anySolution = `{
	"gRecaptchaResponse": "token",
	"token": "token",
	"generated_pass_UUID": "uuid",
	"_px3": "px3",
	"payload": "payload",
	"data": "{}",
	"type": "multi",
	"objects": [1]
}`

// TestEverySolverRoutesCorrectly drives all 46 convenience methods and checks
// each one reaches the endpoint its name promises, with the right task type.
//
// The prefix is the whole contract: SolveX always creates and polls, SyncSolveX
// always uses the synchronous endpoint, whatever [TaskType.Mode] says about the
// type. Routing by anything else would put a caller back to looking up how a
// type is classified before every call.
func TestEverySolverRoutesCorrectly(t *testing.T) {
	clientType := reflect.TypeOf(&EzCapSolverClient{})

	for _, taskType := range KnownTaskTypes {
		for _, sync := range []bool{false, true} {
			name := solverMethods[taskType]
			if sync {
				name = "Sync" + name
			}
			runSolverRoutingCase(t, clientType, taskType, name, sync)
		}
	}
}

// runSolverRoutingCase drives one convenience method against a fake service and
// checks where it went.
func runSolverRoutingCase(
	t *testing.T, clientType reflect.Type, taskType TaskType, name string, sync bool,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		service := newFakeService(t, map[string]func(int) (int, string){
			pathCreateTask:     ok(`{"errorId":0,"taskId":"task-1"}`),
			pathGetTaskResult:  ok(`{"errorId":0,"status":"ready","solution":` + anySolution + `}`),
			pathCreateSyncTask: ok(`{"errorId":0,"status":"ready","solution":` + anySolution + `}`),
		})
		client := newTestClient(t, service)

		method, exists := clientType.MethodByName(name)
		if !exists {
			t.Fatalf("method %s is missing", name)
		}
		// A zero-valued task is enough: this is about routing, and the SDK
		// deliberately leaves parameter validation to the service.
		task := reflect.New(method.Type.In(2).Elem())
		results := method.Func.Call([]reflect.Value{
			reflect.ValueOf(client),
			reflect.ValueOf(t.Context()),
			task,
		})

		if err, failed := results[1].Interface().(error); failed && err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if results[0].IsNil() {
			t.Fatal("a successful call must return a result")
		}

		wantPath := pathCreateTask
		if sync {
			wantPath = pathCreateSyncTask
		}
		calls := service.calls()
		if len(calls) == 0 || calls[0].Path != wantPath {
			t.Fatalf("%s went to %v, want %s", name, pathsOf(calls), wantPath)
		}

		created, isObject := calls[0].Body["task"].(map[string]any)
		if !isObject {
			t.Fatalf("task must be an object, got %#v", calls[0].Body["task"])
		}
		if created["type"] != string(taskType) {
			t.Errorf("type: got %v, want %s", created["type"], taskType)
		}

		// The synchronous endpoint answers inline, so nothing is polled
		// and there is no task ID to hand back.
		if sync {
			if service.countOf(pathGetTaskResult) != 0 {
				t.Error("a synchronous task must not be polled")
			}
			if id := results[0].Elem().FieldByName("TaskID").String(); id != "" {
				t.Errorf("a synchronous task has no ID, got %q", id)
			}
		}
	})
}

// TestSolvedResultsCarryTheRawValue checks that every convenience method hands
// back the untouched worker JSON alongside the decoded model, so a field the
// model does not declare is still reachable.
func TestSolvedResultsCarryTheRawValue(t *testing.T) {
	clientType := reflect.TypeOf(&EzCapSolverClient{})

	var compacted map[string]any
	mustUnmarshal(t, anySolution, &compacted)
	want, err := json.Marshal(compacted)
	if err != nil {
		t.Fatalf("re-serialization failed: %v", err)
	}

	for _, taskType := range KnownTaskTypes {
		name := solverMethods[taskType]
		t.Run(name, func(t *testing.T) {
			body := `{"errorId":0,"status":"ready","solution":` + string(want) + `}`
			service := newFakeService(t, map[string]func(int) (int, string){
				pathCreateTask:     ok(`{"errorId":0,"taskId":"task-1"}`),
				pathGetTaskResult:  ok(body),
				pathCreateSyncTask: ok(body),
			})
			client := newTestClient(t, service)

			method, _ := clientType.MethodByName(name)
			results := method.Func.Call([]reflect.Value{
				reflect.ValueOf(client),
				reflect.ValueOf(t.Context()),
				reflect.New(method.Type.In(2).Elem()),
			})
			if err, failed := results[1].Interface().(error); failed && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			raw := results[0].Elem().FieldByName("Raw").Bytes()
			if string(raw) != string(want) {
				t.Errorf("raw value\n got: %s\nwant: %s", raw, want)
			}
		})
	}
}

// TestSolverDecodeFailurePreservesTheRawValue checks that when a worker returns a
// shape the model does not fit, the failure carries the shape that caused it.
func TestSolverDecodeFailurePreservesTheRawValue(t *testing.T) {
	// A ReCaptcha result with no token in it: the one field the service
	// confirmed is always present.
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateTask:    ok(`{"errorId":0,"taskId":"task-1"}`),
		pathGetTaskResult: ok(`{"errorId":0,"status":"ready","solution":{"unexpected":"shape"}}`),
	})
	client := newTestClient(t, service)

	_, err := client.SolveRecaptchaV2TaskProxyless(t.Context(), &RecaptchaV2Task{})
	if err == nil {
		t.Fatal("expected a decode failure")
	}

	var decodeErr *SolutionDecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("got %T: %v", err, err)
	}
	if string(decodeErr.Raw) != `{"unexpected":"shape"}` {
		t.Errorf("raw: got %s", decodeErr.Raw)
	}
}

// TestUnconfirmedSolversReturnEverything checks the two task types whose shape
// is not settled: with no declared fields, all of the worker's output has to
// arrive through the pass-through map.
func TestUnconfirmedSolversReturnEverything(t *testing.T) {
	solution := `{"angle":137.5,"index":3}`
	service := newFakeService(t, map[string]func(int) (int, string){
		pathCreateSyncTask: ok(`{"errorId":0,"status":"ready","solution":` + solution + `}`),
	})
	client := newTestClient(t, service)

	want := map[string]any{"angle": 137.5, "index": float64(3)}

	funCaptcha, err := client.SyncSolveFunCaptchaClassification(t.Context(),
		&FunCaptchaClassificationTask{Image: "i", Question: "q"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantExtra(t, funCaptcha.Solution.Extra, want)

	hCaptcha, err := client.SyncSolveHCaptchaClassification(t.Context(),
		&HCaptchaClassificationTask{Image: "i"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantExtra(t, hCaptcha.Solution.Extra, want)
}

// pathsOf lists the paths a set of recorded calls hit.
func pathsOf(calls []recordedRequest) []string {
	paths := make([]string, 0, len(calls))
	for _, call := range calls {
		if !slices.Contains(paths, call.Path) {
			paths = append(paths, call.Path)
		}
	}
	return paths
}
