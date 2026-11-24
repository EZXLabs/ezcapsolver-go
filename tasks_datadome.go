package ezcapsolver

import "encoding/json"

// DataDomeStep selects which step of the DataDome challenge flow to run.
//
// It is a string on the wire, not a number.
type DataDomeStep string

const (
	// DataDomeStepOne finds the challenge address of an intercepted page.
	DataDomeStepOne DataDomeStep = "1"
	// DataDomeStepTwo asks whether the service already caches a challenge
	// script, by [DataDomeTask.Hash].
	DataDomeStepTwo DataDomeStep = "2"
	// DataDomeStepThree computes the answer to submit from the challenge page
	// and its script.
	DataDomeStepThree DataDomeStep = "3"
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
	// HtmlB64 is the base64-encoded page: the intercepted page on step one, the
	// challenge page on step three. Step two leaves it empty.
	HtmlB64 string `json:"html_b64,omitzero"`
	// Step is the step of the workflow. Required; the service accepts only "1",
	// "2" and "3". Leaving it empty sends [DataDomeStepOne].
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
	// Hash is, on step two, the hash of the challenge script to look up in the
	// service cache.
	Hash string `json:"hash,omitzero"`
	// UserAgent is the optional browser User-Agent the challenge is solved for.
	UserAgent string `json:"user_agent,omitzero"`
	// ExternalJsHash is, on step three, the hash of the challenge script, when
	// step two reported it cached.
	ExternalJsHash string `json:"external_js_hash,omitzero"`
	// ExternalJsBase64 is, on step three, the base64-encoded challenge script,
	// when step two reported it not cached.
	ExternalJsBase64 string `json:"external_js_base64,omitzero"`
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
