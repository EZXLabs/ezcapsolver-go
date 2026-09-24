package ezcapsolver

import "encoding/json"

// IncapsulaTask produces an Incapsula Reese84 sensor payload. It runs through
// the synchronous endpoint.
//
// The service declares no validation for this type at all. Its own comments
// call most of these fields mandatory, but nothing enforces it, so every field
// is optional here: rejecting a request the service would have accepted is the
// worse failure.
type IncapsulaTask struct {
	// Script is the full source of the Reese84 sensor script.
	Script string `json:"script"`
	// ScriptURL is the URL the script was served from, which has to match where
	// it actually came from.
	ScriptURL string `json:"scriptUrl"`
	// PageURL is the URL of the page running the sensor script.
	PageURL string `json:"pageUrl"`
	// AcceptLanguage is the browser's optional Accept-Language header, such as
	// `ja-JP,ja;q=0.9,en;q=0.8`.
	AcceptLanguage string `json:"acceptLanguage,omitzero"`
	// Ua is the full browser User-Agent. The service reads only the Chrome major
	// version from it and supports 147, 148 and 149.
	Ua string `json:"ua"`
	// Proxy is optional and changes what the task does: supplied, the worker
	// generates the payload and submits it, returning the cookie; omitted, it
	// only generates the payload.
	//
	// Note that this type never counts as having a proxy for the purposes of a
	// plan's mandatory-proxy requirement, even when this is set.
	Proxy string `json:"proxy,omitzero"`
	// Pow is the optional proof-of-work data, required by the sites that have
	// PoW challenges enabled.
	Pow string `json:"pow,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t IncapsulaTask) MarshalJSON() ([]byte, error) {
	type wire IncapsulaTask
	return marshalWithExtra(wire(t), t.Extra)
}

var _ json.Marshaler = IncapsulaTask{}
