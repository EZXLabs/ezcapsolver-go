package ext

import (
	"errors"
	"net/url"
	"testing"
)

const windowsChrome153 = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36"

// headerValue looks a header up by name, reporting whether it is present.
func headerValue(headers []header, name string) (string, bool) {
	for _, h := range headers {
		if h.name == name {
			return h.value, true
		}
	}
	return "", false
}

// mustProfile builds a profile or fails the test.
func mustProfile(t *testing.T, userAgent string) browserProfile {
	t.Helper()
	profile, err := newBrowserProfile(userAgent, "en-US,en;q=0.9")
	if err != nil {
		t.Fatalf("building the profile failed: %v", err)
	}
	return profile
}

// mustParse parses an address or fails the test.
func mustParse(t *testing.T, address string) *url.URL {
	t.Helper()
	parsed, err := url.Parse(address)
	if err != nil {
		t.Fatalf("parsing %q failed: %v", address, err)
	}
	return parsed
}

func TestTheChromeVersionComesFromTheUserAgent(t *testing.T) {
	for _, testCase := range []struct {
		userAgent string
		want      int
		ok        bool
	}{
		{windowsChrome153, 153, true},
		{"Chrome/99", 99, true},
		{"Mozilla/5.0 Firefox/133.0", 0, false},
		{"Chrome/x", 0, false},
	} {
		got, ok := chromeMajor(testCase.userAgent)
		if got != testCase.want || ok != testCase.ok {
			t.Errorf("chromeMajor(%q) = %d, %t; want %d, %t",
				testCase.userAgent, got, ok, testCase.want, testCase.ok)
		}
	}
}

func TestOnlyVersionsWithAPresetAreAccepted(t *testing.T) {
	for _, testCase := range []struct {
		userAgent, want string
	}{
		{"Chrome/131.0.0.0", "chrome_131"},
		{"Chrome/144.0.0.0", "chrome_144"},
		{windowsChrome153, "chrome_153"},
		{"Chrome/158.0.0.0", "chrome_158"},
	} {
		preset, err := mustProfile(t, testCase.userAgent).tlsPreset()
		if err != nil || preset != testCase.want {
			t.Errorf("%q: preset = %q, %v; want %q", testCase.userAgent, preset, err, testCase.want)
		}
	}

	for _, unsupported := range []string{"Chrome/120.0.0.0", "Chrome/143.0.0.0", "Chrome/159.0.0.0"} {
		if _, err := mustProfile(t, unsupported).tlsPreset(); !errors.Is(err, ErrDataDome) {
			t.Errorf("%q has no preset, got %v", unsupported, err)
		}
	}
	if _, err := newBrowserProfile("Firefox/133.0", "en"); !errors.Is(err, ErrDataDome) {
		t.Errorf("a user agent without a Chrome version must be rejected, got %v", err)
	}
}

func TestClientHintsFollowTheUserAgent(t *testing.T) {
	headers := mustProfile(t, windowsChrome153).navigation()
	for name, want := range map[string]string{
		"sec-ch-ua":          `"Chromium";v="153", "Google Chrome";v="153", "Not/A)Brand";v="99"`,
		"sec-ch-ua-platform": `"Windows"`,
		"user-agent":         windowsChrome153,
		"accept-language":    "en-US,en;q=0.9",
	} {
		if got, _ := headerValue(headers, name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}

	mac := mustProfile(t, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/153.0.0.0")
	if got, _ := headerValue(mac.navigation(), "sec-ch-ua-platform"); got != `"macOS"` {
		t.Errorf("sec-ch-ua-platform = %q, want \"macOS\"", got)
	}
}

func TestEachRequestCarriesTheRefererABrowserWould(t *testing.T) {
	profile := mustProfile(t, windowsChrome153)
	target := mustParse(t, "https://shop.example.com/signup?next=1")
	challenge := mustParse(t, "https://geo.captcha-delivery.com/captcha/?initialCid=x")

	for _, testCase := range []struct {
		name    string
		headers []header
		want    string
	}{
		{"slider frame", profile.challengeFrame(target, true), "https://shop.example.com/signup?next=1"},
		{"interstitial frame", profile.challengeFrame(target, false), "https://shop.example.com"},
		{"script", profile.script(challenge), "https://geo.captcha-delivery.com/"},
		{"submission", profile.submission(challenge, false), challenge.String()},
	} {
		if got, _ := headerValue(testCase.headers, "referer"); got != testCase.want {
			t.Errorf("%s: referer = %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

func TestOnlyAPostSubmissionCarriesAFormContentType(t *testing.T) {
	profile := mustProfile(t, windowsChrome153)
	challenge := mustParse(t, "https://geo.captcha-delivery.com/interstitial/?x=1")

	post := profile.submission(challenge, true)
	if got, _ := headerValue(post, "content-type"); got != formContentType {
		t.Errorf("content-type = %q, want %q", got, formContentType)
	}
	if got, _ := headerValue(post, "origin"); got != challengeOrigin {
		t.Errorf("origin = %q, want %q", got, challengeOrigin)
	}

	get := profile.submission(challenge, false)
	for _, name := range []string{"content-type", "origin"} {
		if _, present := headerValue(get, name); present {
			t.Errorf("a GET submission must not carry %s", name)
		}
	}
}
