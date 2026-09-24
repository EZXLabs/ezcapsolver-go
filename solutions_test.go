package ezcapsolver

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestSolutionModelDecoding decodes one sample per solution model, using the
// shapes the task catalog records.
//
// Each case also checks that a key the model does not declare lands in Extra:
// workers ship faster than the SDK, and a field added on their side has to stay
// reachable without an upgrade.
func TestSolutionModelDecoding(t *testing.T) {
	t.Run("Recaptcha", func(t *testing.T) {
		var solution RecaptchaSolution
		decode(t, `{
			"gRecaptchaResponse": "token",
			"sec_ch_ua": "\"Chromium\";v=\"149\"",
			"user_agent": "Mozilla/5.0",
			"brandNew": {"nested": 1}
		}`, &solution)

		if solution.Token != "token" {
			t.Errorf("token: got %q", solution.Token)
		}
		if solution.SecChUa == "" || solution.UserAgent != "Mozilla/5.0" {
			t.Errorf("headers: got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"brandNew": map[string]any{"nested": float64(1)}})
	})

	t.Run("FunCaptcha", func(t *testing.T) {
		var solution FunCaptchaSolution
		decode(t, `{"token": "arkose-token", "later": true}`, &solution)

		if solution.Token != "arkose-token" {
			t.Errorf("token: got %q", solution.Token)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": true})
	})

	t.Run("CloudflareTurnstile", func(t *testing.T) {
		var solution CloudflareTurnstileSolution
		decode(t, `{"token": "cf-token", "header": {"user-agent": "Ua"}, "later": 1}`, &solution)

		if solution.Token != "cf-token" || solution.Header["user-agent"] != "Ua" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": float64(1)})
	})

	t.Run("Cloudflare5s", func(t *testing.T) {
		var solution Cloudflare5sSolution
		decode(t, `{
			"header": {"user-agent": "Ua"},
			"cookies": {"cf_clearance": "abc"},
			"tlsVersion": "chrome149",
			"body": "<html/>",
			"sToken": "st",
			"later": "x"
		}`, &solution)

		// An earlier implementation flattened this structure into one string,
		// dropping the header and the body outright.
		if solution.Header["user-agent"] != "Ua" || solution.Cookies["cf_clearance"] != "abc" {
			t.Errorf("browser state lost: %+v", solution)
		}
		if solution.TLSVersion != "chrome149" || solution.Body != "<html/>" || solution.SToken != "st" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": "x"})
	})

	t.Run("HCaptcha", func(t *testing.T) {
		var solution HCaptchaSolution
		decode(t, `{"generated_pass_UUID": "uuid", "ua": "Ua", "lang": "en", "later": 1}`, &solution)

		if solution.GeneratedPassUUID != "uuid" || solution.Ua != "Ua" || solution.Lang != "en" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": float64(1)})
	})

	t.Run("PerimeterX", func(t *testing.T) {
		var solution PerimeterXSolution
		decode(t, `{"_px3": "px3", "_pxvid": "vid", "_pxde": "de", "_pxhd": "hd"}`, &solution)

		// The cookies are top-level fields, not nested under a cookies object,
		// and the Go names drop the leading underscore the wire names keep.
		if solution.Px3 != "px3" || solution.PxVid != "vid" || solution.Pxde != "de" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"_pxhd": "hd"})
	})

	t.Run("AkamaiWeb", func(t *testing.T) {
		var solution AkamaiWebSolution
		decode(t, `{"payload": "sensor", "encodedata": "state", "later": 1}`, &solution)

		if solution.Payload != "sensor" || solution.Encodedata != "state" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": float64(1)})
	})

	t.Run("AkamaiSBSD", func(t *testing.T) {
		var solution AkamaiSBSDSolution
		decode(t, `{"payload": "cGF5", "bm_lso_time": "123", "later": 1}`, &solution)

		if solution.Payload != "cGF5" || solution.BmLsoTime != "123" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": float64(1)})
	})

	t.Run("TLSForward", func(t *testing.T) {
		var solution TLSForwardSolution
		decode(t, `{
			"status": 1,
			"code": 200,
			"headers": {"content-type": "text/html"},
			"cookies": {"session": "abc"},
			"body": "<html/>",
			"later": 1
		}`, &solution)

		if solution.Status != 1 || solution.Code != 200 || solution.Body != "<html/>" {
			t.Errorf("got %+v", solution)
		}
		if solution.Headers["content-type"] != "text/html" || solution.Cookies["session"] != "abc" {
			t.Errorf("got %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": float64(1)})
	})

	t.Run("DataDome", func(t *testing.T) {
		var stepOne DataDomeSolution
		decode(t, `{"url": "https://challenge", "later": 1}`, &stepOne)
		if stepOne.URL != "https://challenge" {
			t.Errorf("step one: got %+v", stepOne)
		}
		wantExtra(t, stepOne.Extra, map[string]any{"later": float64(1)})

		// Both steps and both DataDome task types share this one model.
		var stepTwo DataDomeSolution
		decode(t, `{"kind": "slider", "url": "https://check", "body": "b"}`, &stepTwo)
		if stepTwo.Kind != "slider" || stepTwo.Body != "b" {
			t.Errorf("step two: got %+v", stepTwo)
		}
	})

	t.Run("Incapsula", func(t *testing.T) {
		var solution IncapsulaSolution
		decode(t, `{"status": 200, "data": "{\"sensor\":1}", "later": 1}`, &solution)

		if solution.Status != 200 {
			t.Errorf("got %+v", solution)
		}
		// The payload is stringified JSON and has to be submitted verbatim, so
		// the SDK must not parse it a second time.
		if solution.Data != `{"sensor":1}` {
			t.Errorf("data must stay a string: got %q", solution.Data)
		}
		wantExtra(t, solution.Extra, map[string]any{"later": float64(1)})
	})
}

// TestUnconfirmedSolutionsKeepEverything covers the two task types whose shape
// no reliable sample exists for. They declare no business fields, so every key
// the worker sends has to survive in Extra.
func TestUnconfirmedSolutionsKeepEverything(t *testing.T) {
	raw := `{"objects": [3], "angle": 137.5}`
	want := map[string]any{"objects": []any{float64(3)}, "angle": 137.5}

	var funCaptcha FunCaptchaClassificationSolution
	decode(t, raw, &funCaptcha)
	wantExtra(t, funCaptcha.Extra, want)

	var hCaptcha HCaptchaClassificationSolution
	decode(t, raw, &hCaptcha)
	wantExtra(t, hCaptcha.Extra, want)
}

// TestOptionalSolutionFieldsMayBeAbsent checks that a worker omitting a field
// yields a zero value rather than a decoding failure. Only a field the service
// confirmed is always returned may be required.
func TestOptionalSolutionFieldsMayBeAbsent(t *testing.T) {
	var akamai AkamaiSBSDSolution
	decode(t, `{"payload": "p"}`, &akamai)
	if akamai.BmLsoTime != "" {
		t.Errorf("got %q", akamai.BmLsoTime)
	}

	var web AkamaiWebSolution
	decode(t, `{"payload": "p"}`, &web)
	if web.Encodedata != "" {
		t.Errorf("got %q", web.Encodedata)
	}

	var hcaptcha HCaptchaSolution
	decode(t, `{"generated_pass_UUID": "u"}`, &hcaptcha)
	if hcaptcha.Ua != "" || hcaptcha.Lang != "" {
		t.Errorf("got %+v", hcaptcha)
	}

	var recaptcha RecaptchaSolution
	decode(t, `{"gRecaptchaResponse": "token"}`, &recaptcha)
	if recaptcha.SecChUa != "" || recaptcha.UserAgent != "" {
		t.Errorf("got %+v", recaptcha)
	}
}

// TestMissingRequiredFieldFails checks that a core output going missing is
// reported instead of producing a model with an empty token in it. Silently
// returning "" is exactly the failure mode typed models exist to prevent.
func TestMissingRequiredFieldFails(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		target any
	}{
		{"Incapsula without data", `{"status": 200}`, &IncapsulaSolution{}},
		{"Recaptcha without a token", `{"sec_ch_ua": "x"}`, &RecaptchaSolution{}},
		{"PerimeterX without _px3", `{"_pxvid": "vid"}`, &PerimeterXSolution{}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := json.Unmarshal([]byte(testCase.raw), testCase.target); err == nil {
				t.Fatal("decoding must fail when a confirmed field is missing")
			}
		})
	}
}

// TestNonObjectSolutionFails checks that a solution which is not a JSON object
// is reported rather than quietly decoding into a zero value. The service always
// sends an object here, so anything else is a contract break.
func TestNonObjectSolutionFails(t *testing.T) {
	for _, raw := range []string{`"a bare string"`, `[1, 2]`, `42`, `true`, `null`} {
		var solution RecaptchaSolution
		if err := json.Unmarshal([]byte(raw), &solution); err == nil {
			t.Errorf("decoding %s must fail", raw)
		}
	}
}

// TestSolutionsRoundTrip checks that a decoded solution re-serializes to what it
// came from, extras included, so a solution can be stored and read back.
func TestSolutionsRoundTrip(t *testing.T) {
	original := `{"gRecaptchaResponse":"token","sec_ch_ua":"ua","user_agent":"agent","later":1}`

	var solution RecaptchaSolution
	decode(t, original, &solution)
	encoded, err := json.Marshal(solution)
	if err != nil {
		t.Fatalf("re-serialization failed: %v", err)
	}

	var before, after map[string]any
	mustUnmarshal(t, original, &before)
	mustUnmarshal(t, string(encoded), &after)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("round trip changed the value\nbefore: %v\n after: %v", before, after)
	}
}

// TestClassificationSolution checks direct field access for every result kind.
func TestClassificationSolution(t *testing.T) {
	t.Run("multi selects grid cells", func(t *testing.T) {
		solution := classify(t, `{"type": "multi", "objects": [0, 3, 7]}`)

		if solution.Type != "multi" || !solution.IsMulti() || solution.IsSingle() {
			t.Fatalf("unexpected result kind: %+v", solution)
		}
		if !reflect.DeepEqual(solution.Objects, []int{0, 3, 7}) {
			t.Errorf("objects: got %v", solution.Objects)
		}
		if solution.HasObject {
			t.Error("an omitted hasObject must default to false")
		}
	})

	t.Run("single answers yes or no", func(t *testing.T) {
		for _, hasObject := range []bool{true, false} {
			raw := fmt.Sprintf(`{"type":"single","hasObject":%t}`, hasObject)
			solution := classify(t, raw)
			if solution.Type != "single" || !solution.IsSingle() || solution.IsMulti() {
				t.Fatalf("unexpected result kind: %+v", solution)
			}
			if solution.HasObject != hasObject || len(solution.Objects) != 0 {
				t.Errorf("unexpected single result: %+v", solution)
			}
		}
	})

	t.Run("an unknown discriminant is kept verbatim", func(t *testing.T) {
		// A challenge kind added service-side must not make the SDK fail to
		// decode; the caller should not have to wait for a release.
		raw := `{"type": "brand-new", "whatever": 1}`
		solution := classify(t, raw)

		if solution.Type != "brand-new" || solution.IsMulti() || solution.IsSingle() {
			t.Errorf("unexpected result kind: %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"whatever": float64(1)})
	})

	t.Run("a missing type does not infer the result kind", func(t *testing.T) {
		solution := classify(t, `{"objects":[2]}`)
		if solution.Type != "" || solution.IsMulti() || solution.IsSingle() {
			t.Errorf("unexpected result kind: %+v", solution)
		}
		if !reflect.DeepEqual(solution.Objects, []int{2}) || solution.HasObject {
			t.Errorf("unexpected fields: %+v", solution)
		}
	})

	t.Run("a non-object decode error preserves the raw value", func(t *testing.T) {
		var solution ReClassificationSolution
		err := decodeSolution(json.RawMessage(`[0,3]`), &solution)
		var decodeError *SolutionDecodeError
		if !errors.As(err, &decodeError) || string(decodeError.Raw) != `[0,3]` {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("fields are preserved regardless of type", func(t *testing.T) {
		solution := classify(t, `{"type":"single","objects":[1],"hasObject":true,"confidence":0.9}`)
		if !solution.HasObject || !reflect.DeepEqual(solution.Objects, []int{1}) {
			t.Errorf("unexpected fields: %+v", solution)
		}
		wantExtra(t, solution.Extra, map[string]any{"confidence": 0.9})
	})
}

// TestClassificationRoundTrips checks that unknown kinds and extra fields survive.
func TestClassificationRoundTrips(t *testing.T) {
	raw := `{"type":"brand-new","hasObject":false,"objects":[],"whatever":{"score":1}}`
	solution := classify(t, raw)

	encoded, err := json.Marshal(solution)
	if err != nil {
		t.Fatalf("re-serialization failed: %v", err)
	}
	var before, after map[string]any
	mustUnmarshal(t, raw, &before)
	mustUnmarshal(t, string(encoded), &after)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("round trip changed the value: got %s, want %s", encoded, raw)
	}
}

func TestClassificationExtraCannotOverwriteFields(t *testing.T) {
	solution := ReClassificationSolution{
		Type:    "single",
		Objects: []int{},
		Extra:   map[string]any{"type": "multi", "hasObject": true, "objects": []int{1}, "later": 42},
	}
	encoded, err := json.Marshal(solution)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	mustUnmarshal(t, string(encoded), &payload)
	wantExtra(t, payload, map[string]any{
		"type": "single", "hasObject": false, "objects": []any{}, "later": float64(42),
	})
}

// decode unmarshals raw into target, failing the test on error.
func decode(t *testing.T, raw string, target any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
}

// classify decodes a classification result, failing the test on error.
func classify(t *testing.T, raw string) ReClassificationSolution {
	t.Helper()
	var solution ReClassificationSolution
	decode(t, raw, &solution)
	return solution
}

// mustUnmarshal decodes raw into target, failing the test on error.
func mustUnmarshal(t *testing.T, raw string, target any) {
	t.Helper()
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		t.Fatalf("decoding failed: %v", err)
	}
}

// wantExtra asserts the pass-through map holds exactly want.
func wantExtra(t *testing.T, got, want map[string]any) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("extra fields: got %v, want %v", got, want)
	}
}

// TestSolutionDecodeErrorCarriesTheRawValue checks that a decode failure keeps
// the value that caused it. A failure means the worker returned a shape this
// release does not model, and seeing that shape is the whole point.
func TestSolutionDecodeErrorCarriesTheRawValue(t *testing.T) {
	raw := json.RawMessage(`{"status": 200}`)

	err := decodeSolution(raw, &IncapsulaSolution{})
	if err == nil {
		t.Fatal("expected a decode failure")
	}
	if !errors.Is(err, ErrDecode) {
		t.Errorf("must classify as a decode failure, got %v", err)
	}

	var decodeErr *SolutionDecodeError
	if !errors.As(err, &decodeErr) {
		t.Fatalf("expected a *SolutionDecodeError, got %T", err)
	}
	if string(decodeErr.Raw) != string(raw) {
		t.Errorf("raw: got %s, want %s", decodeErr.Raw, raw)
	}
	if !contains(err.Error(), `{"status": 200}`) {
		t.Errorf("the message must include the raw value, got %q", err.Error())
	}
}

// TestDecodeSolutionReportsAMissingSolution checks the distinction between a
// response with no solution field and one whose solution is null.
func TestDecodeSolutionReportsAMissingSolution(t *testing.T) {
	var unexpected *UnexpectedResponseError
	err := decodeSolution(nil, &RecaptchaSolution{})
	if !errors.As(err, &unexpected) || !contains(unexpected.Reason, "does not contain a solution") {
		t.Errorf("got %v, want a missing-solution report", err)
	}

	// A null solution did reach the SDK; it is the worker's empty answer, not an
	// unfinished task, so it is a decode failure rather than a missing solution.
	err = decodeSolution(json.RawMessage("null"), &RecaptchaSolution{})
	if errors.As(err, &unexpected) {
		t.Error("a null solution must be told apart from an absent one")
	}
}

// contains reports whether text holds substring.
func contains(text, substring string) bool {
	return len(text) >= len(substring) && indexOf(text, substring) >= 0
}

func indexOf(text, substring string) int {
	for index := 0; index+len(substring) <= len(text); index++ {
		if text[index:index+len(substring)] == substring {
			return index
		}
	}
	return -1
}

// TestEverySolutionRoundTrips decodes a sample into each model and serializes it
// back.
//
// This is what catches a wire name that decoding and encoding disagree on: a
// typo in one direction alone would leave the field silently empty on the way
// back out, and only a round trip notices.
func TestEverySolutionRoundTrips(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		target func() any
	}{
		{"Recaptcha", `{"gRecaptchaResponse":"t","sec_ch_ua":"u","user_agent":"a","later":1}`,
			func() any { return &RecaptchaSolution{} }},
		{"FunCaptcha", `{"token":"t","later":1}`,
			func() any { return &FunCaptchaSolution{} }},
		{"FunCaptchaClassification", `{"angle":137.5}`,
			func() any { return &FunCaptchaClassificationSolution{} }},
		{"HCaptchaClassification", `{"index":3}`,
			func() any { return &HCaptchaClassificationSolution{} }},
		{"CloudflareTurnstile", `{"token":"t","header":{"a":"b"},"later":1}`,
			func() any { return &CloudflareTurnstileSolution{} }},
		{"Cloudflare5s", `{"body":"b","cookies":{"c":"d"},"header":{"a":"b"},` +
			`"sToken":"s","tlsVersion":"chrome149","later":1}`,
			func() any { return &Cloudflare5sSolution{} }},
		{"HCaptcha", `{"generated_pass_UUID":"u","lang":"en","ua":"a","later":1}`,
			func() any { return &HCaptchaSolution{} }},
		{"PerimeterX", `{"_px3":"a","_pxde":"c","_pxvid":"b","_pxhd":"d"}`,
			func() any { return &PerimeterXSolution{} }},
		{"AkamaiWeb", `{"encodedata":"e","payload":"p","later":1}`,
			func() any { return &AkamaiWebSolution{} }},
		{"AkamaiSBSD", `{"bm_lso_time":"1","payload":"p","later":1}`,
			func() any { return &AkamaiSBSDSolution{} }},
		{"TLSForward", `{"body":"b","code":200,"cookies":{"c":"d"},` +
			`"headers":{"a":"b"},"status":1,"later":1}`,
			func() any { return &TLSForwardSolution{} }},
		{"DataDome", `{"body":"b","kind":"slider","url":"https://x","later":1}`,
			func() any { return &DataDomeSolution{} }},
		{"Incapsula", `{"data":"{}","status":200,"later":1}`,
			func() any { return &IncapsulaSolution{} }},
		{"Classification", `{"objects":[1,2],"hasObject":false,"type":"multi","later":1}`,
			func() any { return &ReClassificationSolution{} }},
	}

	if len(cases) != 14 {
		t.Fatalf("expected 12 confirmed models and 2 placeholders; got %d", len(cases))
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			solution := testCase.target()
			decode(t, testCase.raw, solution)

			encoded, err := json.Marshal(solution)
			if err != nil {
				t.Fatalf("re-serialization failed: %v", err)
			}

			var before, after map[string]any
			mustUnmarshal(t, testCase.raw, &before)
			mustUnmarshal(t, string(encoded), &after)
			if !reflect.DeepEqual(before, after) {
				t.Errorf("round trip changed the value\nbefore: %v\n after: %v", before, after)
			}
		})
	}
}

// TestErrorMessagesAreSelfContained checks that each error type says enough on
// its own, since a caller typically logs nothing but the error.
func TestErrorMessagesAreSelfContained(t *testing.T) {
	cases := []struct {
		err  error
		want []string
	}{
		{fmt.Errorf("%w: timeout must be greater than zero", ErrConfig), []string{"timeout"}},
		{
			&TransportError{Op: "POST https://api.example.com/createTask", Err: errors.New("no route")},
			[]string{"createTask", "no route"},
		},
		{
			&PollingExhaustedError{TaskID: "task-9", Attempts: 40, Interval: 3 * time.Second},
			// The task ID matters most: the task may still finish, and it is
			// what lets the caller come back for the result.
			[]string{"task-9", "40", "3s"},
		},
		{
			&UnexpectedResponseError{Reason: "not JSON", Body: "<html/>"},
			[]string{"not JSON", "<html/>"},
		},
		{
			&SolutionDecodeError{Raw: json.RawMessage(`{"a":1}`), Err: errors.New("type mismatch")},
			[]string{"type mismatch", `{"a":1}`},
		},
	}

	for _, testCase := range cases {
		message := testCase.err.Error()
		for _, fragment := range testCase.want {
			if !strings.Contains(message, fragment) {
				t.Errorf("%T: %q does not mention %q", testCase.err, message, fragment)
			}
		}
	}

	t.Run("an error without a body stays readable", func(t *testing.T) {
		if got := (&UnexpectedResponseError{Reason: "no status"}).Error(); strings.HasSuffix(got, ";") {
			t.Errorf("got %q", got)
		}
	})
}

// TestTaskResultHelpers covers the status predicates and the decoding shortcut
// on a raw result.
func TestTaskResultHelpers(t *testing.T) {
	ready := TaskResult{Status: TaskStatusReady, Solution: json.RawMessage(`{"token":"t"}`)}
	if !ready.IsReady() || ready.IsProcessing() || ready.IsError() {
		t.Errorf("got %+v", ready)
	}
	if TaskStatusError.String() != "error" || !TaskStatusError.IsValid() {
		t.Error("the three documented statuses must be recognised")
	}
	if TaskStatus("queued").IsValid() {
		t.Error("an undocumented status must not be accepted")
	}

	var solution FunCaptchaSolution
	if err := ready.DecodeSolution(&solution); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if solution.Token != "t" {
		t.Errorf("got %+v", solution)
	}

	empty := TaskResult{Status: TaskStatusProcessing}
	var absent *UnexpectedResponseError
	if err := empty.DecodeSolution(&solution); !errors.As(err, &absent) ||
		!contains(absent.Reason, "does not contain a solution") {
		t.Errorf("got %v, want a missing-solution report", err)
	}

	// A decode failure still hands back the value that caused it.
	mismatched := TaskResult{Status: TaskStatusReady, Solution: json.RawMessage(`{"a":1}`)}
	err := mismatched.DecodeSolution(&FunCaptchaSolution{})
	if !errors.Is(err, ErrDecode) {
		t.Errorf("got %v, want a decode failure", err)
	}
}
