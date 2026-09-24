package ezcapsolver

import "encoding/json"

// FunCaptchaTask requests a FunCaptcha (Arkose Labs) token.
type FunCaptchaTask struct {
	// WebsiteURL is the URL of the page containing the challenge. Required.
	WebsiteURL string `json:"websiteURL"`
	// WebsiteKey is the FunCaptcha public key. Required.
	WebsiteKey string `json:"websiteKey"`
	// Data is the optional Arkose Labs blob, a JSON string shaped like
	// `{"blob":"..."}`.
	Data string `json:"data,omitzero"`
	// APIJSSubdomain is the optional arkoselabs.com subdomain the site uses.
	APIJSSubdomain string `json:"funcaptchaApiJSSubdomain,omitzero"`
	// Proxy is the optional worker proxy.
	//
	// FunCaptcha is the only task type using the `FUN` proxy format —
	// `protocol://host:port:username:password`, with the credentials appended
	// rather than placed before the host. Every other type takes
	// `protocol://username:password@host:port`.
	//
	// Either way the service requires both a username and a password: an
	// unauthenticated proxy is rejected.
	Proxy string `json:"proxy,omitzero"`

	// Cn reports whether the supplied proxy is inside mainland China.
	Cn bool `json:"cn"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t FunCaptchaTask) MarshalJSON() ([]byte, error) {
	type wire FunCaptchaTask
	return marshalWithExtra(wire(t), t.Extra)
}

// FunCaptchaClassificationTask asks a worker to read one FunCaptcha image.
// It runs through the synchronous endpoint.
//
// The solution shape for this type is not confirmed yet; see
// [FunCaptchaClassificationSolution].
type FunCaptchaClassificationTask struct {
	// Image is the base64-encoded challenge image. Required.
	Image string `json:"image"`
	// Question is the classification question. Required.
	Question string `json:"question"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t FunCaptchaClassificationTask) MarshalJSON() ([]byte, error) {
	type wire FunCaptchaClassificationTask
	return marshalWithExtra(wire(t), t.Extra)
}

var (
	_ json.Marshaler = FunCaptchaTask{}
	_ json.Marshaler = FunCaptchaClassificationTask{}
)
