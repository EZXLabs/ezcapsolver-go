package ezcapsolver

import (
	"encoding/json"
	"net/http"
)

// TLSMethod is an HTTP method a TLS forwarding task can issue.
//
// The service matches these case-insensitively and accepts nothing outside the
// five below.
type TLSMethod string

const (
	TLSMethodGET    TLSMethod = http.MethodGet
	TLSMethodPOST   TLSMethod = http.MethodPost
	TLSMethodPUT    TLSMethod = http.MethodPut
	TLSMethodDELETE TLSMethod = http.MethodDelete
	TLSMethodPATCH  TLSMethod = http.MethodPatch
)

// TLSForwardTask sends one HTTP request from a worker, using that worker's TLS
// fingerprint instead of yours. It runs through the synchronous endpoint and
// returns the upstream response.
//
// Every field name here is snake_case. This is also the one task type the
// service exempts from site allow-listing and block-listing.
type TLSForwardTask struct {
	// TLSType is the worker's TLS fingerprint identifier, such as `chrome`.
	// Required.
	TLSType string `json:"tls_type"`
	// Proxy is the proxy used for the upstream request. Required for this type.
	Proxy string `json:"proxy"`
	// Method is the upstream HTTP method. Required; leaving it empty sends
	// [TLSMethodGET].
	Method TLSMethod `json:"method"`
	// URL is the upstream URL. Required, and it has to start with http or https.
	URL string `json:"url"`
	// Headers are the optional upstream request headers.
	Headers map[string]any `json:"headers,omitzero"`
	// HeadersOrder optionally pins the order headers are sent in, which is part
	// of what a fingerprint check looks at.
	HeadersOrder string `json:"headers_order,omitzero"`
	// Cookies are the optional upstream request cookies.
	Cookies map[string]any `json:"cookies,omitzero"`
	// Body is the optional upstream request body. Any JSON value is accepted:
	// a string, an object, an array.
	Body any `json:"body,omitzero"`
	// BodyRaw reports whether [TLSForwardTask.Body] is already base64-encoded.
	// Always sent, so the zero value states plainly that it is not, rather than
	// leaving the worker to guess.
	BodyRaw bool `json:"body_raw"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t TLSForwardTask) MarshalJSON() ([]byte, error) {
	type wire TLSForwardTask
	normalized := wire(t)
	if normalized.Method == "" {
		normalized.Method = TLSMethodGET
	}
	return marshalWithExtra(normalized, t.Extra)
}

var _ json.Marshaler = TLSForwardTask{}
