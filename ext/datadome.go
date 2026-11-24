package ext

import (
	"errors"
	"fmt"
)

// defaultAcceptLanguage is sent when a challenge names no Accept-Language.
const defaultAcceptLanguage = "en-US,en;q=0.9"

// redactedPlaceholder stands in for a credential in debug output.
const redactedPlaceholder = "[REDACTED]"

// Errors of the DataDome workflow. An API call it makes fails with the error
// the client returned, so the ezcapsolver sentinels still classify it.
var (
	// ErrDataDome reports a run that cannot go on: an unusable TargetURL,
	// UserAgent or Proxy, a challenge without a script, or a submission
	// DataDome did not accept.
	ErrDataDome = errors.New("ext: datadome workflow failed")

	// ErrDataDomeBlocked reports that DataDome banned the visitor outright. No
	// challenge clears that; retry through a different proxy. It wraps
	// [ErrDataDome], so errors.Is matches both.
	ErrDataDomeBlocked = fmt.Errorf("%w: blocked", ErrDataDome)

	// ErrBrowserRequest reports a request to the protected site or to DataDome
	// that failed. The lkrequest error stays reachable through errors.As.
	ErrBrowserRequest = errors.New("ext: browser request failed")
)

// DataDomeChallenge is a page behind DataDome, and the browser identity to
// clear its challenge with.
//
// Every browser-side request of one run goes out through Proxy with UserAgent.
// Keep using both for the requests that carry the cookie the run returns.
type DataDomeChallenge struct {
	// TargetURL is the page on the protected site. Required.
	//
	// It is fetched first, and the run only goes on when DataDome intercepts it.
	TargetURL string
	// UserAgent is the browser User-Agent sent with every browser-side request.
	// Required.
	//
	// Its Chrome/<major> version also selects the TLS fingerprint, so it must
	// name a Chrome release with a fingerprint preset: 131, or 144 to 158.
	UserAgent string
	// Proxy carries every browser-side request, such as
	// `http://user:pass@host:port`. Empty means a direct connection.
	//
	// Calls to the EzCaptchaSolver API do not use it; they follow
	// [ezcapsolver.WithProxy].
	Proxy string
	// AcceptLanguage is sent with every browser-side request. Empty sends
	// `en-US,en;q=0.9`.
	AcceptLanguage string
}

// String renders the challenge with the proxy replaced, so %v and %+v keep its
// credentials out of logs.
func (c DataDomeChallenge) String() string {
	proxy := `""`
	if c.Proxy != "" {
		proxy = redactedPlaceholder
	}
	return fmt.Sprintf("DataDomeChallenge{TargetURL:%s UserAgent:%s Proxy:%s AcceptLanguage:%s}",
		c.TargetURL, c.UserAgent, proxy, c.AcceptLanguage)
}

// GoString covers the %#v verb, which ignores Stringer.
func (c DataDomeChallenge) GoString() string { return c.String() }

// acceptLanguage is the Accept-Language the run sends.
func (c DataDomeChallenge) acceptLanguage() string {
	if c.AcceptLanguage == "" {
		return defaultAcceptLanguage
	}
	return c.AcceptLanguage
}

// DataDomeCookie is the cookie DataDome issues once its challenge is cleared.
type DataDomeCookie struct {
	// Name is the cookie name, normally `datadome`.
	Name string
	// Value is the cookie value.
	Value string
	// Domain is the domain the cookie applies to: its Domain attribute, or the
	// last two labels of the target host, such as `.example.com`, when DataDome
	// omits it.
	Domain string
	// Raw is the cookie exactly as DataDome returned it, attributes included.
	Raw string
}
