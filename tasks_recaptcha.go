package ezcapsolver

import "encoding/json"

// RecaptchaV2Task holds the parameters shared by every ReCaptcha V2 task type:
// the plain, high-score, `s`-carrying and enterprise variants all take these.
// Which variant runs is chosen by the method you call, not by a field here.
type RecaptchaV2Task struct {
	// WebsiteURL is the URL of the page containing the challenge. Required.
	WebsiteURL string `json:"websiteURL"`
	// WebsiteKey is the ReCaptcha site key. Required.
	WebsiteKey string `json:"websiteKey"`
	// IsInvisible reports whether the challenge uses invisible mode. The
	// service defaults it to false, which is this field's zero value, so it is
	// always sent.
	IsInvisible bool `json:"isInvisible"`
	// Sa is the optional security anchor parameter.
	Sa string `json:"sa,omitzero"`
	// S is the optional challenge-bound `s` parameter. Supplying it routes the
	// task to the high-score IPv4 queue.
	S string `json:"s,omitzero"`
	// WebsiteTitle is the optional page title.
	WebsiteTitle string `json:"websiteTitle,omitzero"`
	// Proxy is the optional proxy the worker uses to reach the site, in the
	// `protocol://username:password@host:port` form. It is not validated here.
	Proxy string `json:"proxy,omitzero"`
	// Extra carries parameters this release does not model. Its entries are
	// flattened next to the fields above, and a declared field wins on a
	// collision.
	Extra map[string]any `json:"-"`
}

func (t RecaptchaV2Task) MarshalJSON() ([]byte, error) {
	type wire RecaptchaV2Task
	return marshalWithExtra(wire(t), t.Extra)
}

// RecaptchaV3Task holds the parameters shared by every ReCaptcha V3 task type.
type RecaptchaV3Task struct {
	// WebsiteURL is the URL of the page containing the challenge. Required.
	WebsiteURL string `json:"websiteURL"`
	// WebsiteKey is the ReCaptcha site key. Required.
	WebsiteKey string `json:"websiteKey"`
	// IsInvisible reports whether the challenge uses invisible mode.
	//
	// V3 defaults to true, unlike V2, so this is a pointer: nil omits the field
	// and lets the service apply that default, which keeps the zero value of
	// this struct correct. A plain bool could not express the difference —
	// omitzero drops a false, so the one value worth sending explicitly would
	// be the one that never went out.
	//
	//	IsInvisible: ezcapsolver.Ptr(false) // opt out of invisible mode
	IsInvisible *bool `json:"isInvisible,omitzero"`
	// PageAction is the optional action configured by the protected page.
	PageAction string `json:"pageAction,omitzero"`
	// WebsiteTitle is the optional page title.
	WebsiteTitle string `json:"websiteTitle,omitzero"`
	// CheckField is the optional site-specific check field.
	CheckField string `json:"checkField,omitzero"`
	// Proxy is the optional proxy the worker uses to reach the site.
	Proxy string `json:"proxy,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t RecaptchaV3Task) MarshalJSON() ([]byte, error) {
	type wire RecaptchaV3Task
	return marshalWithExtra(wire(t), t.Extra)
}

// RecaptchaV2ClassificationTask asks a worker to read one ReCaptcha V2 image
// grid. It runs through the synchronous endpoint and returns its answer inline.
type RecaptchaV2ClassificationTask struct {
	// Image is the base64-encoded challenge image. Required.
	Image string `json:"image"`
	// Question is the object identifier or classification question. Required.
	Question string `json:"question"`
	// Size is the grid layout: 1 for a single tile, 3 for 3x3, 4 for 4x4.
	//
	// The service defaults it to 4. Leaving this zero omits the field and lets
	// that default apply, so the zero value of this struct behaves correctly.
	// The range is not enforced anywhere, here or service-side.
	Size int `json:"size,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t RecaptchaV2ClassificationTask) MarshalJSON() ([]byte, error) {
	type wire RecaptchaV2ClassificationTask
	return marshalWithExtra(wire(t), t.Extra)
}

var (
	_ json.Marshaler = RecaptchaV2Task{}
	_ json.Marshaler = RecaptchaV3Task{}
	_ json.Marshaler = RecaptchaV2ClassificationTask{}
)
