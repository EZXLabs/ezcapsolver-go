package ezcapsolver

import "encoding/json"

// DataDomeStep selects which half of the DataDome challenge flow to run.
//
// It is a string on the wire, not a number.
type DataDomeStep string

const (
	// DataDomeStepOne fetches the challenge URL.
	DataDomeStepOne DataDomeStep = "1"
	// DataDomeStepTwo produces the validation instructions.
	DataDomeStepTwo DataDomeStep = "2"
)

// DataDomeJsType selects the DataDome tags JavaScript mode.
type DataDomeJsType string

const (
	// DataDomeJsTypeCh is challenge mode, where the packet counter is fixed at 1.
	DataDomeJsTypeCh DataDomeJsType = "ch"
	// DataDomeJsTypeLe is the legacy or external mode, where the packet counter
	// starts at 2 and increments.
	DataDomeJsTypeLe DataDomeJsType = "le"
)

// DataDomeTask answers a DataDome challenge after an interception. It runs
// through the synchronous endpoint.
//
// Every field name here is snake_case, unlike [DataDomeTagsTask].
type DataDomeTask struct {
	// HtmlB64 is the base64-encoded challenge HTML. Required.
	HtmlB64 string `json:"html_b64"`
	// Step is the step of the workflow. Required; the service accepts only "1"
	// and "2". Leaving it empty sends [DataDomeStepOne].
	Step DataDomeStep `json:"step"`
	// Image is an optional base64-encoded image.
	Image string `json:"image"`
	// Referer is the page or challenge URL.
	//
	// The service declares it optional, but the real workflow needs it, and it
	// is what site allow-listing is checked against: with an allow-list
	// configured, a referer outside it is rejected with
	// ERROR_WEBSITE_NOT_ALLOWED.
	Referer string `json:"referer,omitzero"`
	// ParentURL is the optional parent page URL.
	ParentURL string `json:"parent_url,omitzero"`
	// Equipment is the optional equipment identifier.
	Equipment string `json:"equipment,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t DataDomeTask) MarshalJSON() ([]byte, error) {
	type wire DataDomeTask
	normalized := wire(t)
	if normalized.Step == "" {
		normalized.Step = DataDomeStepOne
	}
	return marshalWithExtra(normalized, t.Extra)
}

// DataDomeTagsTask reports a fingerprint to DataDome on the normal browsing
// path, in exchange for a cookie. It runs through the synchronous endpoint and a
// different worker from [DataDomeTask].
//
// Field names here are flat lowercase, unlike [DataDomeTask]'s snake_case.
type DataDomeTagsTask struct {
	// Ddk is the DataDome JavaScript key, read from the site's inline snippet as
	// `window.ddjskey`. Required.
	Ddk string `json:"ddk"`
	// JsType is the JavaScript mode. Required; the service accepts only "ch"
	// and "le". Leaving it empty sends [DataDomeJsTypeCh].
	JsType DataDomeJsType `json:"jstype"`
	// Cid is the session identifier. Required to be present, but an empty
	// string is valid, so it is always sent.
	Cid string `json:"cid"`
	// Bpc is the one-based packet counter. Required to be at least 1, so
	// leaving it zero sends 1.
	Bpc int64 `json:"bpc"`
	// Referer is the current page URL. Required.
	Referer string `json:"referer"`
	// Ua is the browser User-Agent. Required. The service also accepts
	// `user_agent`, but new integrations should use this one.
	Ua string `json:"ua"`
	// Fields carries external business fields, commonly `{"tags_url": "..."}`.
	//
	// Required to be present but allowed to be empty, so a nil map is sent as
	// `{}`. A JSON null here is rejected.
	Fields map[string]any `json:"fields"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t DataDomeTagsTask) MarshalJSON() ([]byte, error) {
	type wire DataDomeTagsTask
	normalized := wire(t)
	if normalized.JsType == "" {
		normalized.JsType = DataDomeJsTypeCh
	}
	if normalized.Bpc == 0 {
		normalized.Bpc = 1
	}
	// Go serializes a nil map to null, but the service accepts only an object
	// for `fields`: an empty object is fine, a null is rejected.
	if normalized.Fields == nil {
		normalized.Fields = map[string]any{}
	}
	return marshalWithExtra(normalized, t.Extra)
}

var (
	_ json.Marshaler = DataDomeTask{}
	_ json.Marshaler = DataDomeTagsTask{}
)
