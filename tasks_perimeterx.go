package ezcapsolver

import "encoding/json"

// PerimeterXTask requests PerimeterX clearance cookies.
//
// The canonical type name is `PerimeterX`. Older clients called it `PxCaptcha`,
// which was never the service's spelling.
type PerimeterXTask struct {
	// WebsiteKey is the PerimeterX application identifier. Required.
	WebsiteKey string `json:"websiteKey"`
	// Invisible reports whether the challenge uses invisible mode. Enabling it
	// changes how the task is priced.
	//
	// The wire name is `invisible`, without the `is` prefix the ReCaptcha types
	// use.
	Invisible bool `json:"invisible"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t PerimeterXTask) MarshalJSON() ([]byte, error) {
	type wire PerimeterXTask
	return marshalWithExtra(wire(t), t.Extra)
}

var _ json.Marshaler = PerimeterXTask{}
