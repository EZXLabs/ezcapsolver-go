package ezcapsolver

import "encoding/json"

// Solution models, one per task family. Several task types share one model
// where the worker returns the same shape.
//
// Two rules hold for all of them:
//
//   - every model carries an Extra map, so a field a worker starts returning is
//     preserved rather than dropped;
//   - a field is only declared required when the service has confirmed the
//     worker always returns it. A worker that omits an optional field yields a
//     zero value, not a decoding failure.
//
// The service does not define these shapes — workers do, and workers change
// faster than the SDK. Extra plus [Solved.Raw] is what keeps that from being a
// problem.

// RecaptchaSolution is the token returned by every ReCaptcha V2 and V3 token
// task, nine task types in all.
type RecaptchaSolution struct {
	// Token is the value to submit to the protected site. Always returned.
	Token string `json:"gRecaptchaResponse" ezcapsolver:"required"`
	// SecChUa is the matching Sec-CH-UA request header.
	SecChUa string `json:"sec_ch_ua"`
	// UserAgent is the User-Agent that goes with the token.
	UserAgent string `json:"user_agent"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *RecaptchaSolution) UnmarshalJSON(data []byte) error {
	type wire RecaptchaSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s RecaptchaSolution) MarshalJSON() ([]byte, error) {
	type wire RecaptchaSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// FunCaptchaSolution is the token returned by a FunCaptcha (Arkose Labs) task.
type FunCaptchaSolution struct {
	// Token is the value to submit to the protected site.
	Token string `json:"token" ezcapsolver:"required"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *FunCaptchaSolution) UnmarshalJSON(data []byte) error {
	type wire FunCaptchaSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s FunCaptchaSolution) MarshalJSON() ([]byte, error) {
	type wire FunCaptchaSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// FunCaptchaClassificationSolution is a FunCaptcha classification result.
//
// The shape is not confirmed yet. No reliable sample exists, so this declares no
// fields of its own: everything the worker returns lands in Extra, and
// [Solved.Raw] keeps the untouched JSON. Fields get promoted out of Extra as
// samples confirm them, and code reading them from Extra keeps working until
// then.
//
// Guessing at fields would be worse than declaring none — a wrong model reads
// like a contract.
type FunCaptchaClassificationSolution struct {
	// Extra holds every field the worker returned.
	Extra map[string]any `json:"-"`
}

func (s *FunCaptchaClassificationSolution) UnmarshalJSON(data []byte) error {
	type wire FunCaptchaClassificationSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s FunCaptchaClassificationSolution) MarshalJSON() ([]byte, error) {
	type wire FunCaptchaClassificationSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// HCaptchaClassificationSolution is an HCaptcha classification result.
//
// The shape is not confirmed yet, for the same reason as
// [FunCaptchaClassificationSolution].
type HCaptchaClassificationSolution struct {
	// Extra holds every field the worker returned.
	Extra map[string]any `json:"-"`
}

func (s *HCaptchaClassificationSolution) UnmarshalJSON(data []byte) error {
	type wire HCaptchaClassificationSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s HCaptchaClassificationSolution) MarshalJSON() ([]byte, error) {
	type wire HCaptchaClassificationSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// CloudflareTurnstileSolution is the token returned by a Turnstile task.
type CloudflareTurnstileSolution struct {
	// Token is the value to submit to the protected site.
	Token string `json:"token" ezcapsolver:"required"`
	// Header holds request headers to replay alongside the token.
	Header map[string]string `json:"header"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *CloudflareTurnstileSolution) UnmarshalJSON(data []byte) error {
	type wire CloudflareTurnstileSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s CloudflareTurnstileSolution) MarshalJSON() ([]byte, error) {
	type wire CloudflareTurnstileSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// Cloudflare5sSolution is the browser state a worker ended up with after
// clearing a five-second interstitial.
//
// There is no single token here. Replaying these headers and cookies against the
// protected site is what actually clears the challenge, which is why the whole
// structure is kept rather than reduced to one string.
type Cloudflare5sSolution struct {
	// Header holds request headers to replay.
	Header map[string]string `json:"header"`
	// Cookies holds the clearance cookies the worker obtained.
	Cookies map[string]string `json:"cookies"`
	// TLSVersion is the browser fingerprint the worker used, such as `chrome149`.
	TLSVersion string `json:"tlsVersion"`
	// Body is the challenge page content, empty when the worker captured none.
	Body string `json:"body"`
	// SToken is the Turnstile token embedded in the challenge, empty when the
	// flow did not produce one.
	SToken string `json:"sToken"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *Cloudflare5sSolution) UnmarshalJSON(data []byte) error {
	type wire Cloudflare5sSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s Cloudflare5sSolution) MarshalJSON() ([]byte, error) {
	type wire Cloudflare5sSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// HCaptchaSolution is the pass returned by an HCaptcha task.
type HCaptchaSolution struct {
	// GeneratedPassUUID is the generated HCaptcha pass identifier.
	GeneratedPassUUID string `json:"generated_pass_UUID" ezcapsolver:"required"`
	// Ua is the User-Agent that goes with the pass.
	Ua string `json:"ua"`
	// Lang is the language that goes with the pass.
	Lang string `json:"lang"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *HCaptchaSolution) UnmarshalJSON(data []byte) error {
	type wire HCaptchaSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s HCaptchaSolution) MarshalJSON() ([]byte, error) {
	type wire HCaptchaSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// PerimeterXSolution holds the PerimeterX clearance cookies.
//
// The worker returns them as top-level fields, not nested under a cookies
// object. The Go field names drop the leading underscore the wire names carry;
// the wire names themselves are untouched.
type PerimeterXSolution struct {
	// Px3 is the `_px3` clearance cookie, the value that actually passes the
	// check.
	Px3 string `json:"_px3" ezcapsolver:"required"`
	// PxVid is the `_pxvid` visitor identifier.
	PxVid string `json:"_pxvid"`
	// Pxde is the `_pxde` data-enrichment cookie.
	Pxde string `json:"_pxde"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *PerimeterXSolution) UnmarshalJSON(data []byte) error {
	type wire PerimeterXSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s PerimeterXSolution) MarshalJSON() ([]byte, error) {
	type wire PerimeterXSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// AkamaiWebSolution is one round of the Akamai Web flow.
type AkamaiWebSolution struct {
	// Payload is the sensor data for this round.
	Payload string `json:"payload" ezcapsolver:"required"`
	// Encodedata is the encoded state to feed into the next round as
	// [AkamaiWebTask.EncodeData]. Note the casing difference between the two.
	// The round that ends the flow carries none, so its absence is not a failure.
	Encodedata string `json:"encodedata"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *AkamaiWebSolution) UnmarshalJSON(data []byte) error {
	type wire AkamaiWebSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s AkamaiWebSolution) MarshalJSON() ([]byte, error) {
	type wire AkamaiWebSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// AkamaiSBSDSolution is the payload returned by an Akamai SBSD task.
type AkamaiSBSDSolution struct {
	// Payload is the base64-encoded sensor payload.
	Payload string `json:"payload" ezcapsolver:"required"`
	// BmLsoTime is the `bm_lso_time` produced alongside the payload. Not every
	// response carries one.
	BmLsoTime string `json:"bm_lso_time"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *AkamaiSBSDSolution) UnmarshalJSON(data []byte) error {
	type wire AkamaiSBSDSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s AkamaiSBSDSolution) MarshalJSON() ([]byte, error) {
	type wire AkamaiSBSDSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// TLSForwardSolution is the upstream response a TLS forwarding task fetched.
type TLSForwardSolution struct {
	// Status is the worker's own response status.
	Status int `json:"status"`
	// Code is the upstream HTTP status code.
	Code int `json:"code"`
	// Headers are the upstream response headers.
	Headers map[string]any `json:"headers"`
	// Cookies are the upstream response cookies.
	Cookies map[string]any `json:"cookies"`
	// Body is the upstream response body.
	Body string `json:"body"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *TLSForwardSolution) UnmarshalJSON(data []byte) error {
	type wire TLSForwardSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s TLSForwardSolution) MarshalJSON() ([]byte, error) {
	type wire TLSForwardSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// DataDomeSolution is returned by both DataDome challenge steps and by both
// DataDome task types.
//
// Step one carries the challenge address in URL, step two the validation
// instructions. Earlier workers returned step one as a bare string; it is an
// object now, which is why both steps decode into this one type — and why
// [Solved.Raw] does not narrow along with the model.
type DataDomeSolution struct {
	// Kind is the challenge kind, such as `slider` or `interstitial`.
	Kind string `json:"kind"`
	// URL is the challenge URL on step one and the validation endpoint on step two.
	URL string `json:"url"`
	// Body is the optional validation request body.
	Body string `json:"body"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *DataDomeSolution) UnmarshalJSON(data []byte) error {
	type wire DataDomeSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s DataDomeSolution) MarshalJSON() ([]byte, error) {
	type wire DataDomeSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// IncapsulaSolution wraps the Reese84 payload.
type IncapsulaSolution struct {
	// Status is the worker's inner status code.
	Status int `json:"status"`
	// Data is the JSON string to post to the Incapsula sensor endpoint.
	//
	// It is stringified JSON and has to be submitted exactly as it arrived.
	// Parsing it here would change what the caller has to send, so the SDK
	// leaves it alone.
	Data string `json:"data" ezcapsolver:"required"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

func (s *IncapsulaSolution) UnmarshalJSON(data []byte) error {
	type wire IncapsulaSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s IncapsulaSolution) MarshalJSON() ([]byte, error) {
	type wire IncapsulaSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// ReClassificationSolution is the result of a ReCaptcha V2 image classification
// task. Type identifies the result kind; unknown values are preserved for the
// caller to interpret.
type ReClassificationSolution struct {
	// Type is the worker result type, usually "multi" or "single".
	Type string `json:"type"`
	// HasObject reports whether a single image contains the requested object.
	HasObject bool `json:"hasObject"`
	// Objects holds the zero-based indexes of the cells to select.
	Objects []int `json:"objects"`
	// Extra holds worker fields this release does not declare.
	Extra map[string]any `json:"-"`
}

// IsMulti reports whether the worker returned a multi-cell grid result.
func (s ReClassificationSolution) IsMulti() bool { return s.Type == "multi" }

// IsSingle reports whether the worker returned a single-image result.
func (s ReClassificationSolution) IsSingle() bool { return s.Type == "single" }

func (s *ReClassificationSolution) UnmarshalJSON(data []byte) error {
	type wire ReClassificationSolution
	return unmarshalWithExtra(data, (*wire)(s), &s.Extra)
}

func (s ReClassificationSolution) MarshalJSON() ([]byte, error) {
	type wire ReClassificationSolution
	return marshalWithExtra(wire(s), s.Extra)
}

// cloneRaw copies raw JSON so the result does not alias a decoding buffer the
// caller may reuse.
func cloneRaw(data []byte) json.RawMessage {
	cloned := make(json.RawMessage, len(data))
	copy(cloned, data)
	return cloned
}
