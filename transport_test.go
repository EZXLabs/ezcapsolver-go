package ezcapsolver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestParseAPIResponseShapes covers every response form the API contract
// describes.
//
// The two signals — the HTTP status and the envelope's errorId — disagree often
// enough that reading only one of them is the single most common way to get this
// API wrong. A failed task arrives as HTTP 200, and most business errors arrive
// as HTTP 500.
func TestParseAPIResponseShapes(t *testing.T) {
	type target struct {
		Value string `json:"value"`
	}

	t.Run("a success is decoded", func(t *testing.T) {
		var decoded target
		if err := parseAPIResponse(200, []byte(`{"errorId":0,"value":"ok"}`), &decoded); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if decoded.Value != "ok" {
			t.Errorf("got %q", decoded.Value)
		}
	})

	t.Run("HTTP 200 with errorId 1 is an error", func(t *testing.T) {
		// A failed task looks exactly like this. Reading only the status code
		// would report it as a success and then find an empty solution.
		body := `{"errorId":1,"status":"error","errorCode":"ERROR_WORKER","errorDescription":"failed"}`
		err := parseAPIResponse(200, []byte(body), &target{})

		apiErr := requireAPIError(t, err)
		if apiErr.ErrorCode != "ERROR_WORKER" || apiErr.HTTPStatus != 200 {
			t.Errorf("got %+v", apiErr)
		}
	})

	t.Run("HTTP 500 with an envelope keeps the code", func(t *testing.T) {
		// Most business errors use 500. Treating it as a service fault and
		// retrying would burn quota, or get the key banned.
		body := `{"errorId":1,"errorCode":"ERROR_ZERO_BALANCE","errorDescription":"Inadequate balance"}`
		err := parseAPIResponse(500, []byte(body), &target{})

		apiErr := requireAPIError(t, err)
		if apiErr.ErrorCode != "ERROR_ZERO_BALANCE" || apiErr.HTTPStatus != 500 {
			t.Errorf("got %+v", apiErr)
		}
	})

	t.Run("a validation failure keeps the field paths", func(t *testing.T) {
		body := `{
			"errorId":1,
			"errorCode":"ERROR_REQUEST_PARAMETERS",
			"errorDescription":"The request parameters are incorrect",
			"requestId":"request-1",
			"errors":{"task.websiteKey":"Must not be blank","clientKey":"Invalid clientKey"}
		}`
		err := parseAPIResponse(400, []byte(body), &target{})

		apiErr := requireAPIError(t, err)
		want := map[string]string{
			"task.websiteKey": "Must not be blank",
			"clientKey":       "Invalid clientKey",
		}
		if !reflect.DeepEqual(apiErr.Errors, want) {
			t.Errorf("errors: got %v, want %v", apiErr.Errors, want)
		}
		if apiErr.RequestID != "request-1" {
			t.Errorf("requestId: got %q", apiErr.RequestID)
		}
	})

	t.Run("a non-JSON error body still reports the status", func(t *testing.T) {
		// A gateway or CDN answers with HTML. The status has to stay readable
		// from the same place as a structured error.
		err := parseAPIResponse(503, []byte("<html>Service Unavailable</html>"), &target{})

		apiErr := requireAPIError(t, err)
		if apiErr.HTTPStatus != 503 {
			t.Errorf("status: got %d", apiErr.HTTPStatus)
		}
		if apiErr.ErrorCode != "" {
			t.Errorf("there was no code to report, got %q", apiErr.ErrorCode)
		}
		if !strings.Contains(apiErr.ErrorDescription, "Service Unavailable") {
			t.Errorf("the raw body must survive for diagnosis, got %q", apiErr.ErrorDescription)
		}
	})

	t.Run("a non-JSON success body is a decode failure", func(t *testing.T) {
		err := parseAPIResponse(200, []byte("<html>surprise</html>"), &target{})

		if !errors.Is(err, ErrDecode) {
			t.Fatalf("got %v, want a decode failure", err)
		}
		var decodeErr *UnexpectedResponseError
		if !errors.As(err, &decodeErr) {
			t.Fatalf("got %T", err)
		}
		if !strings.Contains(decodeErr.Body, "surprise") {
			t.Errorf("the raw body must survive, got %q", decodeErr.Body)
		}
	})

	t.Run("an unknown error code is preserved verbatim", func(t *testing.T) {
		// Task failures carry codes produced by workers, which are an open set.
		// Folding an unrecognised one into "unknown error" would throw away the
		// only clue the caller has.
		body := `{"errorId":1,"errorCode":"WORKER_SPECIFIC_FAILURE_42"}`
		err := parseAPIResponse(200, []byte(body), &target{})

		if got := requireAPIError(t, err).ErrorCode; got != "WORKER_SPECIFIC_FAILURE_42" {
			t.Errorf("got %q", got)
		}
	})
}

// TestAPIErrorMessage checks that everything actionable is on the one line a
// caller is going to log.
func TestAPIErrorMessage(t *testing.T) {
	err := &APIError{
		ErrorCode:        "ERROR_REQUEST_PARAMETERS",
		ErrorDescription: "The request parameters are incorrect",
		HTTPStatus:       400,
		Errors: map[string]string{
			"task.websiteKey": "Must not be blank",
			"clientKey":       "Invalid clientKey",
		},
	}

	// The field paths are sorted, because Go randomises map iteration and an
	// error message that changes between runs is hard to work with.
	want := "ERROR_REQUEST_PARAMETERS: The request parameters are incorrect (HTTP 400); " +
		"clientKey: Invalid clientKey, task.websiteKey: Must not be blank"
	if got := err.Error(); got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}

	bare := &APIError{HTTPStatus: 502}
	if got := bare.Error(); !strings.Contains(got, "HTTP 502") {
		t.Errorf("an error without a code must still report the status, got %q", got)
	}
}

// TestErrorCodeClassification covers the two questions the SDK answers about an
// error code, both of which decide whether a caller should try again.
func TestErrorCodeClassification(t *testing.T) {
	t.Run("the ban-triggering codes are flagged", func(t *testing.T) {
		// Thirty of these within a minute get the key banned for three minutes,
		// so a caller has to stop rather than back off.
		for _, code := range []string{
			"ERROR_KEY_DOES_NOT_EXIST",
			"ERROR_KEY_NOT_AVAILABLE",
			"ERROR_ZERO_BALANCE",
		} {
			if !(&APIError{ErrorCode: code}).IsAuthenticationError() {
				t.Errorf("%s must be flagged as a credential problem", code)
			}
		}
		if (&APIError{ErrorCode: "ERROR_REQUEST_LIMIT"}).IsAuthenticationError() {
			t.Error("rate limiting is not a credential problem")
		}
	})

	t.Run("terminal codes are separated from transient ones", func(t *testing.T) {
		if !(&APIError{ErrorCode: "ERROR_WEBSITE_NOT_ALLOWED"}).IsTerminal() {
			t.Error("a rejected site does not become allowed on a retry")
		}
		for _, code := range []string{
			"ERROR_REQUEST_LIMIT",
			"ERROR_SERVICE_UNAVAILABLE",
			"ERROR_SERVICE_TIMEOUT",
			"ERROR_INTERNAL_SERVER_ERROR",
			"WORKER_SPECIFIC_FAILURE_42",
		} {
			if (&APIError{ErrorCode: code}).IsTerminal() {
				t.Errorf("%s may recover, so it must not be called terminal", code)
			}
		}
	})

	t.Run("only the throttling codes are rate limits", func(t *testing.T) {
		// WaitForResult retries exactly this set, so it has to stay narrower
		// than "everything that is not terminal". A code leaking in would make
		// the loop poll on past a task that has genuinely failed.
		for _, code := range []string{"ERROR_REQUEST_LIMIT", "ERROR_REQUEST_BANNED"} {
			if !(&APIError{ErrorCode: code}).IsRateLimited() {
				t.Errorf("%s is a throttled query", code)
			}
		}
		for _, code := range []string{
			"ERROR_SERVICE_UNAVAILABLE",
			"ERROR_SERVICE_TIMEOUT",
			"ERROR_INTERNAL_SERVER_ERROR",
			"ERROR_TASK_NOT_EXIST",
			"ERROR_ZERO_BALANCE",
			"WORKER_SPECIFIC_FAILURE_42",
			"",
		} {
			if (&APIError{ErrorCode: code}).IsRateLimited() {
				t.Errorf("%q is not a throttled query", code)
			}
		}
	})
}

// TestErrorLayering checks that each error type answers to its own sentinel, so
// errors.Is classifies a failure without a type assertion.
func TestErrorLayering(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		sentinel error
	}{
		{"config", fmt.Errorf("%w: bad", ErrConfig), ErrConfig},
		{"transport", &TransportError{Op: "POST /x", Err: errors.New("dial")}, ErrTransport},
		{"api", &APIError{HTTPStatus: 500}, ErrAPI},
		{"polling", &PollingExhaustedError{TaskID: "t"}, ErrPollingExhausted},
		{"unexpected response", &UnexpectedResponseError{Reason: "x"}, ErrDecode},
		{"solution decode", &SolutionDecodeError{Err: errors.New("x")}, ErrDecode},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if !errors.Is(testCase.err, testCase.sentinel) {
				t.Errorf("%v does not match its sentinel", testCase.err)
			}
			for _, other := range cases {
				if other.sentinel == testCase.sentinel {
					continue
				}
				if errors.Is(testCase.err, other.sentinel) {
					t.Errorf("%v must not match %v", testCase.err, other.sentinel)
				}
			}
		})
	}

	t.Run("a transport error keeps the cause reachable", func(t *testing.T) {
		cause := errors.New("connection reset")
		err := error(&TransportError{Op: "POST /x", Err: cause})
		if !errors.Is(err, cause) {
			t.Error("the wrapped cause must stay reachable for further classification")
		}
	})
}

// TestRedactionRemovesCredentialsAtEveryDepth checks that a credential cannot
// reach the logs however deeply it is nested.
func TestRedactionRemovesCredentialsAtEveryDepth(t *testing.T) {
	var document any
	mustUnmarshal(t, `{
		"clientKey": "super-secret-key",
		"task": {
			"type": "TlsTask",
			"proxy": "http://user:hunter2@127.0.0.1:8080",
			"rounds": [{"proxy": "socks5://user:hunter2@host:1080"}]
		}
	}`, &document)

	rendered, err := json.Marshal(redact(document))
	if err != nil {
		t.Fatalf("re-serialization failed: %v", err)
	}

	text := string(rendered)
	if strings.Contains(text, "super-secret-key") || strings.Contains(text, "hunter2") {
		t.Fatalf("a credential survived redaction: %s", text)
	}
	for _, expected := range []string{
		`"clientKey":"[REDACTED]"`,
		`"proxy":"[REDACTED]"`,
		`"type":"TlsTask"`,
	} {
		if !strings.Contains(text, expected) {
			t.Errorf("expected %s in %s", expected, text)
		}
	}
	// Redaction must reach inside arrays too, not just nested objects.
	if strings.Count(text, redactedPlaceholder) != 3 {
		t.Errorf("expected three redactions, got %s", text)
	}
}

// TestEndpointJoinsPaths checks URL assembly tolerates a slash on either side,
// so a custom base URL does not have to be written a particular way.
func TestEndpointJoinsPaths(t *testing.T) {
	cases := []struct{ base, path, want string }{
		{"https://api.example.com", "/createTask", "https://api.example.com/createTask"},
		{"https://api.example.com/", "/createTask", "https://api.example.com/createTask"},
		{"https://api.example.com/", "createTask", "https://api.example.com/createTask"},
		{"http://127.0.0.1:8080", "/getBalance", "http://127.0.0.1:8080/getBalance"},
	}

	for _, testCase := range cases {
		if got := endpoint(testCase.base, testCase.path); got != testCase.want {
			t.Errorf("endpoint(%q, %q) = %q, want %q",
				testCase.base, testCase.path, got, testCase.want)
		}
	}
}

// TestPreviewTruncates checks that an oversized body cannot bury the rest of a
// message, while a short one is passed through untouched.
func TestPreviewTruncates(t *testing.T) {
	if got := preview([]byte("short"), 16); got != "short" {
		t.Errorf("got %q", got)
	}

	long := strings.Repeat("x", 600)
	got := preview([]byte(long), errorPreviewChars)
	if len(got) != errorPreviewChars+3 || !strings.HasSuffix(got, "...") {
		t.Errorf("truncation must be marked, got %d characters ending %q", len(got), got[len(got)-3:])
	}

	// Cutting must land on a rune boundary, or the preview stops being valid UTF-8.
	if got := preview([]byte(strings.Repeat("解", 10)), 4); got != "解解解解..." {
		t.Errorf("got %q", got)
	}
}

// requireAPIError asserts err is a *APIError and returns it.
func requireAPIError(t *testing.T, err error) *APIError {
	t.Helper()
	if err == nil {
		t.Fatal("expected an error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected a *APIError, got %T: %v", err, err)
	}
	return apiErr
}

// TestEveryEndpointSendsTheClientKey covers all four endpoints, through both a
// client-built and a caller-supplied http.Client. The key has to be set per
// request, since a supplied client carries none of the SDK's own headers.
//
// No request may carry X-Request-Id: the gateway assigns the correlation id and
// the SDK only reads it back. One sent from here would compete with the id the
// platform is already tracing on.
func TestEveryEndpointSendsTheClientKey(t *testing.T) {
	for _, supplied := range []bool{false, true} {
		t.Run(fmt.Sprintf("supplied http client %t", supplied), func(t *testing.T) {
			var paths []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				paths = append(paths, r.URL.Path)
				if got := r.Header.Get("X-API-Key"); got != "client-key" {
					t.Errorf("%s: X-API-Key = %q, want the client key", r.URL.Path, got)
				}
				if _, present := r.Header["X-Request-Id"]; present {
					t.Errorf("%s: the SDK sent a correlation id of its own", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				switch r.URL.Path {
				case pathCreateTask:
					fmt.Fprint(w, `{"errorId":0,"taskId":"t-1"}`)
				case pathGetBalance:
					fmt.Fprint(w, `{"errorId":0,"balance":1.5}`)
				default:
					fmt.Fprint(w, `{"errorId":0,"status":"ready","solution":{"token":"tok"}}`)
				}
			}))
			defer server.Close()

			options := []Option{
				WithClientKey("client-key"),
				WithBaseURLs(server.URL, server.URL),
				WithPolling(PollingConfig{Interval: time.Millisecond, MaxAttempts: 3}),
			}
			if supplied {
				options = append(options, WithHTTPClient(&http.Client{}))
			}
			client, err := NewClient(options...)
			if err != nil {
				t.Fatalf("building the client failed: %v", err)
			}
			params := map[string]any{"websiteURL": "https://example.com"}
			if _, err := client.Solve(t.Context(), "CustomTask", params); err != nil {
				t.Fatalf("Solve: %v", err)
			}
			if _, err := client.SyncSolve(t.Context(), "CustomTask", params); err != nil {
				t.Fatalf("SyncSolve: %v", err)
			}
			if _, err := client.Balance(t.Context()); err != nil {
				t.Fatalf("Balance: %v", err)
			}

			want := []string{pathCreateTask, pathGetTaskResult, pathCreateSyncTask, pathGetBalance}
			if !slices.Equal(paths, want) {
				t.Errorf("requests went to %v, want %v", paths, want)
			}
		})
	}
}
