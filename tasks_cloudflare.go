package ezcapsolver

import "encoding/json"

// Cloudflare5sTask clears a Cloudflare five-second interstitial.
type Cloudflare5sTask struct {
	// WebsiteURL is the URL protected by the challenge. Required.
	WebsiteURL string `json:"websiteURL"`
	// Proxy is the worker proxy. Required for this type, unlike most others.
	//
	// Format is `protocol://username:password@host:port` with protocol one of
	// http, https or socks5. Both credentials are required — the service rejects
	// an unauthenticated proxy — and the host may not be a private address.
	Proxy string `json:"proxy"`
	// RqData is optional challenge request data.
	//
	// Send it as an object. The service stringifies it itself when forwarding
	// to the worker, so pre-encoding it to a string produces double encoding.
	RqData map[string]any `json:"rqData,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t Cloudflare5sTask) MarshalJSON() ([]byte, error) {
	type wire Cloudflare5sTask
	return marshalWithExtra(wire(t), t.Extra)
}

// CloudflareTurnstileTask requests a Cloudflare Turnstile token.
type CloudflareTurnstileTask struct {
	// WebsiteURL is the URL of the page containing the widget. Required.
	WebsiteURL string `json:"websiteURL"`
	// WebsiteKey is the Turnstile site key. Required.
	WebsiteKey string `json:"websiteKey"`
	// Proxy is the optional worker proxy.
	Proxy string `json:"proxy,omitzero"`
	// RqData is optional Turnstile metadata, sent as an object for the same
	// reason as [Cloudflare5sTask.RqData].
	RqData map[string]any `json:"rqData,omitzero"`
	// Extra carries parameters this release does not model.
	Extra map[string]any `json:"-"`
}

func (t CloudflareTurnstileTask) MarshalJSON() ([]byte, error) {
	type wire CloudflareTurnstileTask
	return marshalWithExtra(wire(t), t.Extra)
}

var (
	_ json.Marshaler = Cloudflare5sTask{}
	_ json.Marshaler = CloudflareTurnstileTask{}
)
