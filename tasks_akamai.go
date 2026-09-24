package ezcapsolver

import "encoding/json"

// AkamaiWebTask produces one round of Akamai Web sensor data. It runs through
// the synchronous endpoint.
//
// This is a multi-round flow: feed each round's returned
// [AkamaiWebSolution.Encodedata] into the next round's [AkamaiWebTask.EncodeData]
// and raise [AkamaiWebTask.Index]. Note the casing difference between the two —
// the service spells them differently in each direction.
type AkamaiWebTask struct {
	// PageURL is the URL of the page the Akamai script belongs to. Required.
	PageURL string `json:"pageUrl"`
	// V3URL is the URL of the Akamai v3 script. Required.
	//
	// Most sites change this URL on every request, so it has to be read from
	// the page rather than hard-coded. It is the URL itself, not the script
	// the URL serves.
	V3URL string `json:"v3Url"`
	// Ua is the browser User-Agent. Required.
	Ua string `json:"ua"`
	// Lang is the browser language. Required.
	Lang string `json:"lang"`
	// Index is the current round of the flow, starting at zero. Required.
	Index int `json:"index"`
	// Abck is the current `_abck` cookie value.
	Abck string `json:"abck"`
	// Bmsz is the current `bm_sz` cookie value.
	Bmsz string `json:"bmsz"`
	// ScriptBase64 is the base64-encoded Akamai script.
	ScriptBase64 string `json:"script_base64"`
	// EncodeData is the encoded state the previous round returned. Empty on the
	// first round.
	EncodeData string `json:"encodeData"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t AkamaiWebTask) MarshalJSON() ([]byte, error) {
	type wire AkamaiWebTask
	return marshalWithExtra(wire(t), t.Extra)
}

// AkamaiSBSDTask produces Akamai SBSD sensor data. It runs through the
// synchronous endpoint.
type AkamaiSBSDTask struct {
	// PageURL is the URL of the page the challenge belongs to. Required.
	PageURL string `json:"pageUrl"`
	// SbsdURL is the URL of the SBSD script. Required.
	SbsdURL string `json:"sbsdUrl"`
	// BmSo is the existing `bm_so` or equivalent cookie value. Required.
	BmSo string `json:"bmSo"`
	// Ua is the browser User-Agent. Required.
	Ua string `json:"ua"`
	// Lang is the browser language. Required.
	Lang string `json:"lang"`
	// ScriptBase64 is the base64-encoded SBSD script. Required.
	ScriptBase64 string `json:"script_base64"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t AkamaiSBSDTask) MarshalJSON() ([]byte, error) {
	type wire AkamaiSBSDTask
	return marshalWithExtra(wire(t), t.Extra)
}

var (
	_ json.Marshaler = AkamaiWebTask{}
	_ json.Marshaler = AkamaiSBSDTask{}
)
