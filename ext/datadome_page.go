package ext

// Reading the pages of a DataDome challenge.
//
// The script hash is the key the service caches challenge scripts under, so
// the extraction and hashing rules here have to match the service's exactly,
// and those of the other language SDKs. The tests pin them to fixed vectors.

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	// hardBlockMarker is text DataDome puts on the page of a visitor it has
	// banned outright.
	hardBlockMarker = "hardblock"
	// challengeMarker is text every page DataDome intercepted carries.
	challengeMarker = "var dd"
	// sliderPath is a path segment of a slider captcha address; the
	// interstitial uses another.
	sliderPath = "/captcha/"
	// bundleClose is what closes the invisible interstitial bundle.
	bundleClose = "})()"
)

// loaderScripts are CDN scripts that bootstrap a challenge but are not its
// bundle.
var loaderScripts = []string{"i.js", "c.js"}

var (
	// scriptElement matches every <script> element: its opening tag, then its body.
	scriptElement = regexp.MustCompile(`(?is)(<script\b[^>]*>)(.*?)</script>`)
	// srcAttribute marks an opening tag as an external script.
	srcAttribute = regexp.MustCompile(`\bsrc\s*=`)
	// banner is the version banner, `/* DataDome ... */` or `/** DataDome ... */`.
	banner = regexp.MustCompile(`/\*+\s*DataDome[^*]*\*/`)
	// interstitialOpening is how the interstitial bundle opens, right after its
	// banner.
	interstitialOpening = regexp.MustCompile(`^\s*;\s*\(function\s*\(\)\s*\{`)
	// bundleStart is where a bundle without a banner starts: the slider or the
	// interstitial form.
	bundleStart = regexp.MustCompile(`!\s*function\s+[A-Za-z_$][\w$]*\s*\(|;\s*\(function\s*\(\)\s*\{`)
	// namedExternalScript is an external challenge script with a descriptive name.
	namedExternalScript = regexp.MustCompile(
		`(?i)<script\b[^>]*?\bsrc\s*=\s*["']((?:https:)?//ct\.captcha-delivery\.com/(?:interstitial|captcha)\.[^"']+?\.js)["']`)
	// hashedExternalScript is an external challenge script named by its content
	// hash, which DataDome switched to in November 2026.
	hashedExternalScript = regexp.MustCompile(
		`(?i)<script\b[^>]*?\bsrc\s*=\s*["']((?:https:)?//ct\.captcha-delivery\.com/[A-Za-z0-9][A-Za-z0-9._-]*?\.js)["']`)
)

// isHardBlocked reports whether DataDome banned the visitor outright.
//
// No challenge clears that; only a different IP helps.
func isHardBlocked(html string) bool {
	return strings.Contains(html, hardBlockMarker)
}

// isChallenged reports whether DataDome intercepted the page with a challenge.
func isChallenged(html string) bool {
	return strings.Contains(html, challengeMarker)
}

// isSlider reports whether a challenge address leads to the slider captcha
// rather than to the invisible interstitial.
func isSlider(challengeURL *url.URL) bool {
	return strings.Contains(challengeURL.Path, sliderPath)
}

// encodeText base64-encodes page or script text the way the service expects it.
func encodeText(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// stripBanner strips the version banner DataDome puts on the first line of its
// script.
//
// The banner changes with every release while the code under it may not, so it
// stays out of the hash. Only a first line that opens and closes a comment
// counts as a banner; anything else is returned unchanged.
func stripBanner(script string) string {
	if !strings.HasPrefix(script, "/*") {
		return script
	}
	firstLine, rest, _ := strings.Cut(script, "\n")
	if strings.Contains(firstLine, "*/") {
		return rest
	}
	return script
}

// scriptHash is the content-addressed identifier of a challenge script: the
// lower-case hex MD5 of the script without its banner.
func scriptHash(script string) string {
	digest := md5.Sum([]byte(stripBanner(script)))
	return hex.EncodeToString(digest[:])
}

// externalScriptURL is the address of the challenge script when the page loads
// it from DataDome's CDN, or empty when the page embeds it.
//
// A descriptively named script wins; otherwise the first hash-named one that is
// not a loader. A protocol-relative address is completed with `https:`.
func externalScriptURL(html string) string {
	if match := namedExternalScript.FindStringSubmatch(html); match != nil {
		return absoluteSrc(match[1])
	}
	for _, match := range hashedExternalScript.FindAllStringSubmatch(html, -1) {
		if !isLoader(match[1]) {
			return absoluteSrc(match[1])
		}
	}
	return ""
}

// isLoader reports whether a CDN script only bootstraps a challenge.
func isLoader(src string) bool {
	return slices.Contains(loaderScripts, strings.ToLower(path.Base(src)))
}

// absoluteSrc completes a protocol-relative script address with `https:`.
func absoluteSrc(src string) string {
	if strings.HasPrefix(src, "//") {
		return "https:" + src
	}
	return src
}

// inlineScript is the challenge script embedded in the page, cut down to the
// part that is the same for every session, or empty when the page embeds none.
//
// The bundle is the longest inline script by far, and it is cut from its
// version banner on. An interstitial bundle also ends at its closing `})()`:
// what follows is page glue that carries session-unique identifiers, and
// hashing those would never hit the cache. A slider bundle keeps everything to
// the end of its script. Without a banner, the bundle runs from where its code
// starts.
func inlineScript(html string) string {
	script := longestInlineScript(html)
	if script == "" {
		return ""
	}
	location := banner.FindStringIndex(script)
	if location == nil {
		start := 0
		if found := bundleStart.FindStringIndex(script); found != nil {
			start = found[0]
		}
		return script[start:]
	}
	bundle := script[location[0]:]
	afterBanner := script[location[1]:]
	if interstitialOpening.MatchString(afterBanner) {
		if end := strings.LastIndex(bundle, bundleClose); end >= 0 {
			return bundle[:end+len(bundleClose)]
		}
	}
	return bundle
}

// longestInlineScript is the longest body among the scripts without a `src`
// attribute; the first one wins a tie.
func longestInlineScript(html string) string {
	longest := ""
	for _, element := range scriptElement.FindAllStringSubmatch(html, -1) {
		openingTag, body := element[1], element[2]
		if srcAttribute.MatchString(openingTag) {
			continue
		}
		if len(body) > len(longest) {
			longest = body
		}
	}
	return longest
}
