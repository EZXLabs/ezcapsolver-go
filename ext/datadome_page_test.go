package ext

import (
	"net/url"
	"strings"
	"testing"
)

// The vectors here are the ones the Rust and Python SDKs pin, so all three
// hash the same script to the same cache key.
const (
	bannerLine       = "/* DataDome is a cyberfraud solution to detect bot activity @license https://datadome.co v1.33.1 d5110df */"
	interstitialBody = ";(function(){var A={};})();"
)

func TestOnlyAClosedFirstLineCommentIsABanner(t *testing.T) {
	for _, testCase := range []struct {
		name, script, want string
	}{
		{"banner", bannerLine + "\n" + interstitialBody, interstitialBody},
		{"no banner", interstitialBody, interstitialBody},
		{"code first", "var x=1;\n" + interstitialBody, "var x=1;\n" + interstitialBody},
		{"unclosed comment", "/* unclosed\nvar x=1;", "/* unclosed\nvar x=1;"},
		{"banner only", bannerLine, ""},
	} {
		if got := stripBanner(testCase.script); got != testCase.want {
			t.Errorf("%s: stripBanner = %q, want %q", testCase.name, got, testCase.want)
		}
	}
}

func TestTheHashIsTheMD5OfTheScriptWithoutItsBanner(t *testing.T) {
	// RFC 1321 test vector.
	if got := scriptHash(""); got != "d41d8cd98f00b204e9800998ecf8427e" {
		t.Errorf("scriptHash(\"\") = %s", got)
	}
	if scriptHash(bannerLine+"\n"+interstitialBody) != scriptHash(interstitialBody) {
		t.Error("the banner must not change the hash")
	}
}

func TestAnInterstitialBundleEndsBeforeTheSessionGlue(t *testing.T) {
	html := "<html><head>" +
		`<script src="https://ct.captcha-delivery.com/interstitial.1.33.0.x.js"></script>` +
		"<script>var small=1;</script>" +
		"<script>var ddm={cid:'session-unique'};\n" + bannerLine + "\n" + interstitialBody + "\n" +
		"var glue='initialCid=session-unique';</script>" +
		"</head></html>"

	want := strings.TrimSuffix(bannerLine+"\n"+interstitialBody, ";")
	if got := inlineScript(html); got != want {
		t.Errorf("inlineScript = %q, want %q", got, want)
	}
	if got := externalScriptURL(html); got != "https://ct.captcha-delivery.com/interstitial.1.33.0.x.js" {
		t.Errorf("externalScriptURL = %q", got)
	}
}

func TestASliderBundleKeepsItsStaticGlue(t *testing.T) {
	html := "<script>var ddm={};\n/** DataDome v1.34.0 96eb15b */\n!function A(){}();\nvar g=1;</script>"
	want := "/** DataDome v1.34.0 96eb15b */\n!function A(){}();\nvar g=1;"
	if got := inlineScript(html); got != want {
		t.Errorf("inlineScript = %q, want %q", got, want)
	}
}

func TestABundleWithoutABannerStartsWhereItsCodeDoes(t *testing.T) {
	for _, testCase := range []struct {
		html, want string
	}{
		{"<script>var a=1;</script>", "var a=1;"},
		{"<script>var a=1;\n!function B(){}();</script>", "!function B(){}();"},
		{`<script src="x.js"></script>`, ""},
		{"<p>no scripts</p>", ""},
	} {
		if got := inlineScript(testCase.html); got != testCase.want {
			t.Errorf("inlineScript(%q) = %q, want %q", testCase.html, got, testCase.want)
		}
	}
}

func TestAHashNamedExternalScriptIsFoundPastTheLoaders(t *testing.T) {
	html := `<script src="https://ct.captcha-delivery.com/i.js"></script>` +
		`<script src="//ct.captcha-delivery.com/5d512141a346e3bf.js"></script>`
	if got := externalScriptURL(html); got != "https://ct.captcha-delivery.com/5d512141a346e3bf.js" {
		t.Errorf("externalScriptURL = %q", got)
	}
	if got := externalScriptURL(`<script src="https://ct.captcha-delivery.com/c.js"></script>`); got != "" {
		t.Errorf("a loader alone is not the challenge script, got %q", got)
	}
}

func TestPagesAreClassifiedByTheirMarkers(t *testing.T) {
	if !isHardBlocked("<p>hardblock</p>") {
		t.Error("a hardblock page must be recognized")
	}
	if !isChallenged("<script>var dd={'cid':'x'}</script>") {
		t.Error("an intercepted page must be recognized")
	}
	if isChallenged("<html>welcome</html>") {
		t.Error("a plain page is not intercepted")
	}

	slider, _ := url.Parse("https://geo.captcha-delivery.com/captcha/?initialCid=x")
	interstitial, _ := url.Parse("https://geo.captcha-delivery.com/interstitial/?initialCid=x")
	if !isSlider(slider) || isSlider(interstitial) {
		t.Error("the path tells the slider from the interstitial")
	}
}
