package ezcapsolver

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"
)

// marshalToMap serializes a task and decodes it back into a map, which is what
// the assertions below compare against. Comparing maps rather than byte strings
// keeps the tests about field names and values, not about key ordering.
func marshalToMap(t *testing.T, value any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshalling failed: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("the result is not a JSON object: %v; got %s", err, encoded)
	}
	return decoded
}

// keysOf returns a map's keys, sorted.
func keysOf(fields map[string]any) []string {
	names := make([]string, 0, len(fields))
	for name := range fields {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// TestRequestModelFieldNames pins the wire name of every field of all 16 request
// models.
//
// A misspelled key is silently accepted by the service — it lands in the
// pass-through map and the worker never sees the parameter — so this table is
// the only thing standing between a typo and a task that fails for no visible
// reason.
func TestRequestModelFieldNames(t *testing.T) {
	cases := []struct {
		name  string
		task  any
		want  map[string]any
		about string
	}{
		{
			name: "RecaptchaV2Task",
			task: RecaptchaV2Task{
				WebsiteURL:   "https://example.com",
				WebsiteKey:   "site-key",
				IsInvisible:  true,
				Sa:           "anchor",
				S:            "s-value",
				WebsiteTitle: "Example",
				Proxy:        "http://user:pass@host:8080",
			},
			want: map[string]any{
				"websiteURL":   "https://example.com",
				"websiteKey":   "site-key",
				"isInvisible":  true,
				"sa":           "anchor",
				"s":            "s-value",
				"websiteTitle": "Example",
				"proxy":        "http://user:pass@host:8080",
			},
		},
		{
			name: "RecaptchaV3Task",
			task: RecaptchaV3Task{
				WebsiteURL:   "https://example.com",
				WebsiteKey:   "site-key",
				IsInvisible:  Ptr(true),
				PageAction:   "login",
				WebsiteTitle: "Example",
				CheckField:   "check",
				Proxy:        "http://user:pass@host:8080",
			},
			want: map[string]any{
				"websiteURL":   "https://example.com",
				"websiteKey":   "site-key",
				"isInvisible":  true,
				"pageAction":   "login",
				"websiteTitle": "Example",
				"checkField":   "check",
				"proxy":        "http://user:pass@host:8080",
			},
		},
		{
			name: "RecaptchaV2ClassificationTask",
			task: RecaptchaV2ClassificationTask{Image: "base64", Question: "crosswalk", Size: 4},
			want: map[string]any{"image": "base64", "question": "crosswalk", "size": float64(4)},
		},
		{
			name: "FunCaptchaTask",
			task: FunCaptchaTask{
				WebsiteURL:     "https://example.com",
				WebsiteKey:     "public-key",
				Data:           `{"blob":"x"}`,
				APIJSSubdomain: "client-api",
				Proxy:          "http://host:8080:user:pass",
				Cn:             true,
			},
			want: map[string]any{
				"websiteURL":               "https://example.com",
				"websiteKey":               "public-key",
				"data":                     `{"blob":"x"}`,
				"funcaptchaApiJSSubdomain": "client-api",
				"proxy":                    "http://host:8080:user:pass",
				"cn":                       true,
			},
		},
		{
			name: "FunCaptchaClassificationTask",
			task: FunCaptchaClassificationTask{Image: "base64", Question: "rotate"},
			want: map[string]any{"image": "base64", "question": "rotate"},
		},
		{
			name: "HCaptchaTask",
			task: HCaptchaTask{
				WebsiteURL: "https://example.com",
				WebsiteKey: "site-key",
				Lang:       "en-US",
				Proxy:      "http://user:pass@host:8080",
				Invisible:  true,
				RqData:     "rq",
			},
			want: map[string]any{
				"websiteURL": "https://example.com",
				"websiteKey": "site-key",
				"lang":       "en-US",
				"proxy":      "http://user:pass@host:8080",
				"invisible":  true,
				"rqdata":     "rq",
			},
			about: "invisible carries no `is` prefix, and rqdata is all lower-case unlike Cloudflare's rqData",
		},
		{
			name: "HCaptchaClassificationTask",
			task: HCaptchaClassificationTask{
				Image:    "base64",
				Images:   []string{"a", "b"},
				Anchors:  []string{"c"},
				Question: "pick",
				Module:   "module",
			},
			want: map[string]any{
				"image":    "base64",
				"images":   []any{"a", "b"},
				"anchors":  []any{"c"},
				"question": "pick",
				"module":   "module",
			},
		},
		{
			name: "PerimeterXTask",
			task: PerimeterXTask{WebsiteKey: "app-id", Invisible: true},
			want: map[string]any{"websiteKey": "app-id", "invisible": true},
		},
		{
			name: "AkamaiWebTask",
			task: AkamaiWebTask{
				PageURL:      "https://example.com",
				V3URL:        "https://example.com/v3.js",
				Ua:           "Mozilla/5.0",
				Lang:         "en",
				Index:        2,
				Abck:         "abck-value",
				Bmsz:         "bmsz-value",
				ScriptBase64: "c2NyaXB0",
				EncodeData:   "state",
			},
			want: map[string]any{
				"pageUrl":       "https://example.com",
				"v3Url":         "https://example.com/v3.js",
				"ua":            "Mozilla/5.0",
				"lang":          "en",
				"index":         float64(2),
				"abck":          "abck-value",
				"bmsz":          "bmsz-value",
				"script_base64": "c2NyaXB0",
				"encodeData":    "state",
			},
			about: "abck and bmsz are lower-case, script_base64 is snake, encodeData is camel",
		},
		{
			name: "AkamaiSBSDTask",
			task: AkamaiSBSDTask{
				PageURL:      "https://example.com",
				SbsdURL:      "https://example.com/sbsd",
				BmSo:         "bm-so",
				Ua:           "Mozilla/5.0",
				Lang:         "en",
				ScriptBase64: "c2NyaXB0",
			},
			want: map[string]any{
				"pageUrl":       "https://example.com",
				"sbsdUrl":       "https://example.com/sbsd",
				"bmSo":          "bm-so",
				"ua":            "Mozilla/5.0",
				"lang":          "en",
				"script_base64": "c2NyaXB0",
			},
		},
		{
			name: "Cloudflare5sTask",
			task: Cloudflare5sTask{
				WebsiteURL: "https://example.com",
				Proxy:      "http://user:pass@host:8080",
				RqData:     map[string]any{"key": "value"},
			},
			want: map[string]any{
				"websiteURL": "https://example.com",
				"proxy":      "http://user:pass@host:8080",
				"rqData":     map[string]any{"key": "value"},
			},
			about: "rqData goes out as an object; the service stringifies it itself",
		},
		{
			name: "CloudflareTurnstileTask",
			task: CloudflareTurnstileTask{
				WebsiteURL: "https://example.com",
				WebsiteKey: "site-key",
				Proxy:      "http://user:pass@host:8080",
				RqData:     map[string]any{"key": "value"},
			},
			want: map[string]any{
				"websiteURL": "https://example.com",
				"websiteKey": "site-key",
				"proxy":      "http://user:pass@host:8080",
				"rqData":     map[string]any{"key": "value"},
			},
		},
		{
			name: "DataDomeTask",
			task: DataDomeTask{
				HtmlB64:   "PGh0bWw+",
				Step:      DataDomeStepTwo,
				Image:     "base64",
				Referer:   "https://example.com",
				ParentURL: "https://example.com/parent",
				Equipment: "desktop",
			},
			want: map[string]any{
				"html_b64":   "PGh0bWw+",
				"step":       "2",
				"image":      "base64",
				"referer":    "https://example.com",
				"parent_url": "https://example.com/parent",
				"equipment":  "desktop",
			},
			about: "the whole model is snake_case, and step is a string, not a number",
		},
		{
			name: "DataDomeTagsTask",
			task: DataDomeTagsTask{
				Ddk:     "ddjskey",
				JsType:  DataDomeJsTypeLe,
				Cid:     "session",
				Bpc:     3,
				Referer: "https://example.com",
				Ua:      "Mozilla/5.0",
				Fields:  map[string]any{"tags_url": "https://example.com/tags.js"},
			},
			want: map[string]any{
				"ddk":     "ddjskey",
				"jstype":  "le",
				"cid":     "session",
				"bpc":     float64(3),
				"referer": "https://example.com",
				"ua":      "Mozilla/5.0",
				"fields":  map[string]any{"tags_url": "https://example.com/tags.js"},
			},
			about: "this model is camelCase, unlike DataDomeTask",
		},
		{
			name: "IncapsulaTask",
			task: IncapsulaTask{
				Script:         "var reese84;",
				ScriptURL:      "https://example.com/sensor.js",
				PageURL:        "https://example.com",
				AcceptLanguage: "ja-JP,ja;q=0.9",
				Ua:             "Mozilla/5.0",
				Proxy:          "http://user:pass@host:8080",
				Pow:            "pow-token",
			},
			want: map[string]any{
				"script":         "var reese84;",
				"scriptUrl":      "https://example.com/sensor.js",
				"pageUrl":        "https://example.com",
				"acceptLanguage": "ja-JP,ja;q=0.9",
				"ua":             "Mozilla/5.0",
				"proxy":          "http://user:pass@host:8080",
				"pow":            "pow-token",
			},
		},
		{
			name: "TLSForwardTask",
			task: TLSForwardTask{
				TLSType:      "chrome",
				Proxy:        "http://user:pass@host:8080",
				Method:       TLSMethodPOST,
				URL:          "https://example.com/api",
				Headers:      map[string]any{"accept": "*/*"},
				HeadersOrder: "accept,user-agent",
				Cookies:      map[string]any{"session": "abc"},
				Body:         "payload",
				BodyRaw:      true,
			},
			want: map[string]any{
				"tls_type":      "chrome",
				"proxy":         "http://user:pass@host:8080",
				"method":        "POST",
				"url":           "https://example.com/api",
				"headers":       map[string]any{"accept": "*/*"},
				"headers_order": "accept,user-agent",
				"cookies":       map[string]any{"session": "abc"},
				"body":          "payload",
				"body_raw":      true,
			},
		},
	}

	if len(cases) != 16 {
		t.Fatalf("the catalog defines 16 request models, this table covers %d", len(cases))
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got := marshalToMap(t, testCase.task)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Errorf("wire form mismatch (%s)\n got: %v\nwant: %v",
					testCase.about, got, testCase.want)
			}
		})
	}
}

// TestUnsetOptionalFieldsAreOmitted checks the rule that keeps a task from
// overwriting a service-side default: an optional field that was never set must
// not appear in the request at all. Sending null would clear the default.
func TestUnsetOptionalFieldsAreOmitted(t *testing.T) {
	fields := marshalToMap(t, RecaptchaV2Task{WebsiteURL: "u", WebsiteKey: "k"})

	want := []string{"isInvisible", "websiteKey", "websiteURL"}
	if got := keysOf(fields); !slices.Equal(got, want) {
		t.Errorf("optional fields leaked into the request: got %v, want %v", got, want)
	}
}

// TestDocumentedDefaultsSurviveTheZeroValue covers the two models whose
// documented default is not Go's zero value.
//
// Omitting the field is what makes the zero value correct: the service then
// applies its own documented default, so a plain struct literal behaves the way
// the catalog says it should.
func TestDocumentedDefaultsSurviveTheZeroValue(t *testing.T) {
	t.Run("ReCaptcha V3 defaults to invisible", func(t *testing.T) {
		fields := marshalToMap(t, RecaptchaV3Task{WebsiteURL: "u", WebsiteKey: "k"})
		if _, present := fields["isInvisible"]; present {
			t.Error("isInvisible was sent as false, overriding the service default of true")
		}
	})

	t.Run("ReCaptcha V2 defaults to visible", func(t *testing.T) {
		fields := marshalToMap(t, RecaptchaV2Task{WebsiteURL: "u", WebsiteKey: "k"})
		if fields["isInvisible"] != false {
			t.Errorf("V2 must always send isInvisible=false, got %v", fields["isInvisible"])
		}
	})

	t.Run("classification grid size defaults to 4", func(t *testing.T) {
		fields := marshalToMap(t, RecaptchaV2ClassificationTask{Image: "i", Question: "q"})
		if _, present := fields["size"]; present {
			t.Error("size was sent as 0, overriding the service default of 4")
		}
	})
}

// TestRecaptchaV3CanOptOutOfInvisibleMode guards the reason IsInvisible is a
// pointer. The service defaults V3 to true, so false is the value a caller has
// to be able to state — and a plain bool with omitzero would drop exactly that
// one, leaving no way to reach non-invisible V3 through the typed field.
func TestRecaptchaV3CanOptOutOfInvisibleMode(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		value *bool
		want  any
	}{
		{name: "unset lets the service default apply", value: nil, want: nil},
		{name: "explicit false is sent", value: Ptr(false), want: false},
		{name: "explicit true is sent", value: Ptr(true), want: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fields := marshalToMap(t, RecaptchaV3Task{
				WebsiteURL:  "u",
				WebsiteKey:  "k",
				IsInvisible: testCase.value,
			})
			got, present := fields["isInvisible"]
			if testCase.want == nil {
				if present {
					t.Errorf("isInvisible must be omitted when unset, got %v", got)
				}
				return
			}
			if !present {
				t.Fatal("isInvisible was omitted, so the service default silently wins")
			}
			if got != testCase.want {
				t.Errorf("isInvisible = %v, want %v", got, testCase.want)
			}
		})
	}
}

// TestRequiredEnumsFallBackToTheirDefault covers the fields the service
// validates against a closed set, where Go's empty string would be rejected.
func TestRequiredEnumsFallBackToTheirDefault(t *testing.T) {
	dataDome := marshalToMap(t, DataDomeTask{HtmlB64: "PGh0bWw+"})
	if dataDome["step"] != "1" {
		t.Errorf("step must fall back to %q, got %v", DataDomeStepOne, dataDome["step"])
	}

	tags := marshalToMap(t, DataDomeTagsTask{Ddk: "k", Referer: "https://example.com", Ua: "ua"})
	if tags["jstype"] != "ch" {
		t.Errorf("jstype must fall back to %q, got %v", DataDomeJsTypeCh, tags["jstype"])
	}
	if tags["bpc"] != float64(1) {
		t.Errorf("bpc must fall back to 1, the lowest value the service accepts, got %v", tags["bpc"])
	}

	forward := marshalToMap(t, TLSForwardTask{TLSType: "chrome", Proxy: "p", URL: "https://x"})
	if forward["method"] != "GET" {
		t.Errorf("method must fall back to %q, got %v", TLSMethodGET, forward["method"])
	}
}

// TestDataDomeTagsFieldsIsNeverNull guards a Go-specific trap: a nil map
// marshals to JSON null, and the service rejects null for this parameter while
// accepting an empty object.
func TestDataDomeTagsFieldsIsNeverNull(t *testing.T) {
	fields := marshalToMap(t, DataDomeTagsTask{Ddk: "k", Referer: "https://example.com", Ua: "ua"})

	value, present := fields["fields"]
	if !present {
		t.Fatal("fields must always be sent")
	}
	if _, isObject := value.(map[string]any); !isObject {
		t.Errorf("fields must be an object, got %#v", value)
	}
}

// TestOptionalContainersDistinguishUnsetFromEmpty checks the reason these
// fields use omitzero rather than omitempty.
//
// omitempty would drop an explicitly empty map, slice or string, collapsing
// "I did not set this" and "I set this to nothing" into the same request. Some
// of these parameters mean different things in those two cases.
func TestOptionalContainersDistinguishUnsetFromEmpty(t *testing.T) {
	unset := marshalToMap(t, TLSForwardTask{TLSType: "chrome", Proxy: "p", URL: "https://x"})
	for _, field := range []string{"headers", "cookies", "body", "headers_order"} {
		if _, present := unset[field]; present {
			t.Errorf("%s must be omitted when it was never set", field)
		}
	}

	explicit := marshalToMap(t, TLSForwardTask{
		TLSType: "chrome",
		Proxy:   "p",
		URL:     "https://x",
		Headers: map[string]any{},
		Cookies: map[string]any{},
		// An empty body is a legitimate value, not an absent one.
		Body: "",
	})
	for _, field := range []string{"headers", "cookies", "body"} {
		if _, present := explicit[field]; !present {
			t.Errorf("%s was set explicitly and must be sent", field)
		}
	}
	if explicit["body"] != "" {
		t.Errorf("body: got %#v, want an empty string", explicit["body"])
	}

	images := marshalToMap(t, HCaptchaClassificationTask{Images: []string{}})
	if _, present := images["images"]; !present {
		t.Error("an explicitly empty image set must be sent")
	}
}

// TestRequiredBooleansSurviveTheirZeroValue guards the trap that omitzero sets
// for a bool: its zero value is false, so tagging one of these with omitzero
// would drop the very value the caller meant to state, exactly as it once did
// for RecaptchaV3Task.IsInvisible.
//
// Both fields below are required and carry meaning when false, so they have to
// reach the service either way.
func TestRequiredBooleansSurviveTheirZeroValue(t *testing.T) {
	hcaptcha := marshalToMap(t, HCaptchaTask{WebsiteURL: "u", WebsiteKey: "k", Lang: "en-US"})
	if value, present := hcaptcha["invisible"]; !present || value != false {
		t.Errorf("invisible must be sent as false when unset, got %#v (present=%v)", value, present)
	}

	forward := marshalToMap(t, TLSForwardTask{TLSType: "chrome", Proxy: "p", URL: "https://x"})
	if value, present := forward["body_raw"]; !present || value != false {
		t.Errorf("body_raw must be sent as false when unset, got %#v (present=%v)", value, present)
	}
}

// TestExtraFieldsAreFlattened checks that pass-through parameters sit next to
// the declared ones rather than nested under a key of their own. That is what
// lets a worker parameter the SDK does not model be sent without an upgrade.
func TestExtraFieldsAreFlattened(t *testing.T) {
	fields := marshalToMap(t, FunCaptchaTask{
		WebsiteURL: "u",
		WebsiteKey: "k",
		Extra:      map[string]any{"customField": "value", "nested": map[string]any{"a": float64(1)}},
	})

	if fields["customField"] != "value" {
		t.Errorf("extra fields must be flattened to the top level, got %v", fields)
	}
	if _, nested := fields["Extra"]; nested {
		t.Error("the Extra map itself must never be serialized")
	}
	if !reflect.DeepEqual(fields["nested"], map[string]any{"a": float64(1)}) {
		t.Errorf("nested extra values must survive intact, got %v", fields["nested"])
	}
}

// TestDeclaredFieldsBeatExtra checks that an accidental duplicate in
// caller-supplied data cannot replace a modelled field.
func TestDeclaredFieldsBeatExtra(t *testing.T) {
	fields := marshalToMap(t, RecaptchaV2Task{
		WebsiteURL: "https://real.example.com",
		WebsiteKey: "k",
		Extra:      map[string]any{"websiteURL": "https://spoofed.example.com"},
	})

	if fields["websiteURL"] != "https://real.example.com" {
		t.Errorf("the declared field must win, got %v", fields["websiteURL"])
	}
}

// TestTaskPayloadInjectsTheType checks that the task type reaches the wire
// without being a field on any model, and that it cannot be overridden.
func TestTaskPayloadInjectsTheType(t *testing.T) {
	payload := taskPayload{
		taskType: TaskTypeRecaptchaV2TaskProxyless,
		params:   &RecaptchaV2Task{WebsiteURL: "u", WebsiteKey: "k"},
	}
	fields := marshalToMap(t, payload)

	if fields["type"] != string(TaskTypeRecaptchaV2TaskProxyless) {
		t.Errorf("the type must be injected into the task object, got %v", fields["type"])
	}
	if fields["websiteURL"] != "u" {
		t.Errorf("the parameters must sit alongside the type, got %v", fields)
	}

	t.Run("the SDK type wins over one in the parameters", func(t *testing.T) {
		spoofed := taskPayload{
			taskType: TaskTypeHCaptcha,
			params:   map[string]any{"type": "SomethingElse", "websiteKey": "k"},
		}
		if got := marshalToMap(t, spoofed)["type"]; got != string(TaskTypeHCaptcha) {
			t.Errorf("got %v, want %v", got, TaskTypeHCaptcha)
		}
	})

	t.Run("a task type with no parameters still serializes", func(t *testing.T) {
		bare := taskPayload{taskType: TaskTypePerimeterX}
		if got := marshalToMap(t, bare); !reflect.DeepEqual(got, map[string]any{
			"type": string(TaskTypePerimeterX),
		}) {
			t.Errorf("got %v", got)
		}
	})
}
