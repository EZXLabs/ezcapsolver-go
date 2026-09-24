package ezcapsolver

import "encoding/json"

// HCaptchaTask requests an HCaptcha pass.
type HCaptchaTask struct {
	// WebsiteURL is the URL of the page containing the challenge. Required,
	// though the service checks only that it is non-blank, not that it parses.
	WebsiteURL string `json:"websiteURL"`
	// WebsiteKey is the HCaptcha site key. Required.
	WebsiteKey string `json:"websiteKey"`
	// Lang is the browser language. Required. Only `en-US` is supported at the
	// moment.
	Lang string `json:"lang"`
	// Proxy is the optional worker proxy.
	Proxy string `json:"proxy,omitzero"`
	// Invisible reports whether the challenge runs without a visible checkbox.
	// Required: true when the site shows no HCaptcha checkbox, false when it
	// does. Always sent, so the zero value means a visible checkbox rather than
	// an unset field.
	Invisible bool `json:"invisible"`
	// RqData is the optional `rqdata` value, required by the sites that publish
	// one.
	//
	// The wire name is all lowercase here, unlike [Cloudflare5sTask.RqData] and
	// [CloudflareTurnstileTask.RqData], whose equivalent field is `rqData` and
	// carries an object rather than a string. The two are not interchangeable.
	RqData string `json:"rqdata,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t HCaptchaTask) MarshalJSON() ([]byte, error) {
	type wire HCaptchaTask
	return marshalWithExtra(wire(t), t.Extra)
}

// HCaptchaClassificationTask asks a worker to read one or more HCaptcha images.
// It runs through the synchronous endpoint.
//
// Every field is optional: different classification modules take different
// input combinations, and the service declares no validation at all for this
// type. The solution shape is not confirmed yet; see
// [HCaptchaClassificationSolution].
type HCaptchaClassificationTask struct {
	// Image is a single base64-encoded image.
	Image string `json:"image,omitzero"`
	// Images is a collection of base64-encoded images.
	Images []string `json:"images,omitzero"`
	// Anchors is a collection of anchor images.
	Anchors []string `json:"anchors,omitzero"`
	// Question is the classification question.
	Question string `json:"question,omitzero"`
	// Module selects the classification module.
	Module string `json:"module,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t HCaptchaClassificationTask) MarshalJSON() ([]byte, error) {
	type wire HCaptchaClassificationTask
	return marshalWithExtra(wire(t), t.Extra)
}

var (
	_ json.Marshaler = HCaptchaTask{}
	_ json.Marshaler = HCaptchaClassificationTask{}
)
