package ext

// The browser a challenge run presents itself as.
//
// DataDome compares what each request claims against how it is sent, so every
// header set here mirrors what Chrome sends for that kind of request, and the
// TLS fingerprint follows the Chrome version the user agent names.

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

const (
	// documentAccept is the Accept of a document or frame navigation.
	documentAccept = "text/html,application/xhtml+xml,application/xml;q=0.9," +
		"image/avif,image/webp,image/apng,*/*;q=0.8"
	// acceptEncoding is the Accept-Encoding Chrome sends.
	acceptEncoding = "gzip, deflate, br, zstd"
	// challengeOrigin is the origin of the page a POST submission comes from.
	challengeOrigin = "https://geo.captcha-delivery.com"
	// formContentType is the content type of a POST submission.
	formContentType = "application/x-www-form-urlencoded; charset=UTF-8"
	// chromeToken is where the Chrome version starts in a user agent.
	chromeToken = "Chrome/"
)

// Chrome versions lkrequest has a TLS fingerprint preset for: one older
// release, then a contiguous range.
const (
	legacyChromePreset = 131
	oldestChromePreset = 144
	newestChromePreset = 158
)

// header is one request header. Headers travel as a slice because Chrome sends
// them in a fixed order.
type header struct {
	name, value string
}

// browserProfile is the browser identity every request of one run shares.
type browserProfile struct {
	userAgent      string
	acceptLanguage string
	chromeMajor    int
}

// newBrowserProfile builds the profile, rejecting a user agent that names no
// Chrome version.
func newBrowserProfile(userAgent, acceptLanguage string) (browserProfile, error) {
	major, ok := chromeMajor(userAgent)
	if !ok {
		return browserProfile{}, fmt.Errorf("%w: UserAgent must name a Chrome version, got %q",
			ErrDataDome, userAgent)
	}
	return browserProfile{userAgent: userAgent, acceptLanguage: acceptLanguage, chromeMajor: major}, nil
}

// tlsPreset is the lkrequest preset with the TLS and HTTP/2 fingerprint of the
// Chrome version the user agent names.
func (p browserProfile) tlsPreset() (string, error) {
	major := p.chromeMajor
	if major != legacyChromePreset && (major < oldestChromePreset || major > newestChromePreset) {
		return "", fmt.Errorf("%w: no TLS fingerprint for Chrome %d; use %d, or %d to %d",
			ErrDataDome, major, legacyChromePreset, oldestChromePreset, newestChromePreset)
	}
	return "chrome_" + strconv.Itoa(major), nil
}

// navigation is the headers of the first visit to the protected page.
func (p browserProfile) navigation() []header {
	return append(p.device(),
		header{"accept", documentAccept},
		header{"upgrade-insecure-requests", "1"},
		header{"sec-fetch-site", "none"},
		header{"sec-fetch-mode", "navigate"},
		header{"sec-fetch-user", "?1"},
		header{"sec-fetch-dest", "document"},
	)
}

// challengeFrame is the headers of loading the challenge into its frame.
//
// The slider frame gets the full page address as its referer; the interstitial
// only its origin, as a real visitor's browser sends them.
func (p browserProfile) challengeFrame(target *url.URL, slider bool) []header {
	referer := origin(target)
	if slider {
		referer = target.String()
	}
	return append(p.device(),
		header{"accept", documentAccept},
		header{"upgrade-insecure-requests", "1"},
		header{"sec-fetch-site", "cross-site"},
		header{"sec-fetch-mode", "navigate"},
		header{"sec-fetch-dest", "iframe"},
		header{"referer", referer},
	)
}

// script is the headers of fetching the external challenge script.
//
// It has to look like a script load: the CDN rejects navigation headers.
func (p browserProfile) script(challengeURL *url.URL) []header {
	return append(p.device(),
		header{"accept", "*/*"},
		header{"sec-fetch-site", "same-site"},
		header{"sec-fetch-mode", "no-cors"},
		header{"sec-fetch-dest", "script"},
		header{"referer", origin(challengeURL) + "/"},
	)
}

// submission is the headers of submitting the answer from the challenge frame.
func (p browserProfile) submission(challengeURL *url.URL, post bool) []header {
	headers := p.device()
	if post {
		headers = append(headers,
			header{"content-type", formContentType},
			header{"origin", challengeOrigin},
		)
	}
	return append(headers,
		header{"accept", "*/*"},
		header{"sec-fetch-site", "same-origin"},
		header{"sec-fetch-mode", "cors"},
		header{"sec-fetch-dest", "empty"},
		header{"referer", challengeURL.String()},
	)
}

// device is the headers every request carries: the client hints and the
// identity.
func (p browserProfile) device() []header {
	return []header{
		{"sec-ch-ua", p.secChUa()},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"` + p.platform() + `"`},
		{"user-agent", p.userAgent},
		{"accept-encoding", acceptEncoding},
		{"accept-language", p.acceptLanguage},
	}
}

// secChUa is the Sec-CH-UA brand list matching the Chrome version.
func (p browserProfile) secChUa() string {
	return fmt.Sprintf(`"Chromium";v="%[1]d", "Google Chrome";v="%[1]d", "Not/A)Brand";v="99"`,
		p.chromeMajor)
}

// platform is the Sec-CH-UA-Platform value matching the operating system in
// the user agent.
func (p browserProfile) platform() string {
	switch {
	case strings.Contains(p.userAgent, "Macintosh"):
		return "macOS"
	case strings.Contains(p.userAgent, "Linux"):
		return "Linux"
	default:
		return "Windows"
	}
}

// chromeMajor is the Chrome major version named in a user agent.
func chromeMajor(userAgent string) (int, bool) {
	_, rest, found := strings.Cut(userAgent, chromeToken)
	if !found {
		return 0, false
	}
	end := strings.IndexFunc(rest, func(r rune) bool { return r < '0' || r > '9' })
	if end < 0 {
		end = len(rest)
	}
	major, err := strconv.Atoi(rest[:end])
	return major, err == nil
}

// origin is `scheme://host[:port]` of an address, the form an origin-only
// referer takes.
func origin(address *url.URL) string {
	return address.Scheme + "://" + address.Host
}
