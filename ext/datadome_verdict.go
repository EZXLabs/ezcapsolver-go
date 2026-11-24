package ext

// Handing the step-three answer to DataDome, and reading the cookie it answers
// with.

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/EZXLabs/ezcapsolver-go"
)

const (
	// defaultGetDelay is how long a GET submission waits when the answer names
	// no delay.
	defaultGetDelay = 20 * time.Second
	// sliderKind is the challenge kind whose answer always goes out as a GET.
	sliderKind = "slider"
	// captchaView is the `view` of a submission response that shows the
	// captcha again.
	captchaView = "captcha"
	// bannedMarker is the query value of `t` that marks a banned visitor.
	bannedMarker = "bv"
	// previewRunes is how much of an unreadable response an error quotes.
	previewRunes = 200
)

// submission is the request that hands the step-three answer to DataDome: a
// form POST when it has a body, a GET otherwise.
type submission struct {
	url  string
	body string
}

// isPost reports whether the answer is posted as a form.
func (s submission) isPost() bool {
	return s.body != ""
}

// submissionFrom reads the submission out of a step-three answer.
//
// A slider answer, or one without a body, goes out as a GET; any other is
// posted.
func submissionFrom(answer ezcapsolver.DataDomeSolution) (submission, error) {
	if answer.URL == "" {
		return submission{}, fmt.Errorf("%w: step 3 returned no submission url", ErrDataDome)
	}
	if strings.Contains(answer.Kind, sliderKind) {
		return submission{url: answer.URL}, nil
	}
	return submission{url: answer.URL, body: answer.Body}, nil
}

// submitDelay is how long to wait before submitting, counted from when the
// challenge page loaded.
//
// The answer's SubmitAfterMs decides; when it is zero, a GET waits
// [defaultGetDelay] and a POST goes out at once.
func submitDelay(answer ezcapsolver.DataDomeSolution, s submission) time.Duration {
	switch {
	case answer.SubmitAfterMs > 0:
		return time.Duration(answer.SubmitAfterMs) * time.Millisecond
	case s.isPost():
		return 0
	default:
		return defaultGetDelay
	}
}

// readVerdict reads DataDome's response to a submission and returns the cookie
// it issued.
//
// A response without a cookie is a failed run: a ban when DataDome shows the
// captcha again with `t=bv`, a rejected answer otherwise.
func readVerdict(body string, target *url.URL) (DataDomeCookie, error) {
	var verdict map[string]any
	if err := json.Unmarshal([]byte(body), &verdict); err != nil {
		return DataDomeCookie{}, fmt.Errorf("%w: the submission response is not JSON: %s",
			ErrDataDome, preview(body))
	}
	cookie := textField(verdict, "cookie")
	if cookie == "" {
		return DataDomeCookie{}, rejection(verdict)
	}
	return parseCookie(cookie, target), nil
}

// rejection explains why a submission brought back no cookie.
func rejection(verdict map[string]any) error {
	view := textField(verdict, "view")
	if view == captchaView && queryValue(textField(verdict, "url"), "t") == bannedMarker {
		return fmt.Errorf("%w: the submission was answered with a ban (t=bv)", ErrDataDomeBlocked)
	}
	return fmt.Errorf("%w: the challenge was not passed (view %q)", ErrDataDome, view)
}

// parseCookie splits DataDome's cookie string into its parts.
//
// The domain falls back to the last two labels of the target host when the
// cookie carries no Domain attribute.
func parseCookie(raw string, target *url.URL) DataDomeCookie {
	attributes := strings.Split(raw, ";")
	name, value, _ := strings.Cut(attributes[0], "=")
	return DataDomeCookie{
		Name:   strings.TrimSpace(name),
		Value:  strings.TrimSpace(value),
		Domain: cookieDomain(attributes[1:], target),
		Raw:    raw,
	}
}

// cookieDomain is the first Domain attribute, or the fallback built from the
// target host.
func cookieDomain(attributes []string, target *url.URL) string {
	for _, attribute := range attributes {
		attribute = strings.TrimSpace(attribute)
		for _, prefix := range []string{"Domain=", "domain="} {
			if domain, found := strings.CutPrefix(attribute, prefix); found {
				return strings.TrimSpace(domain)
			}
		}
	}
	return fallbackDomain(target)
}

// fallbackDomain is `.` plus the last two labels of the target host, such as
// `.example.com`.
func fallbackDomain(target *url.URL) string {
	labels := strings.Split(target.Hostname(), ".")
	return "." + strings.Join(labels[max(len(labels)-2, 0):], ".")
}

// textField is a string field of the verdict, empty when absent or not a
// string.
func textField(verdict map[string]any, name string) string {
	text, _ := verdict[name].(string)
	return text
}

// queryValue is the first value of a query parameter, for an absolute or a
// relative address.
func queryValue(address, name string) string {
	_, query, found := strings.Cut(address, "?")
	if !found {
		return ""
	}
	query, _, _ = strings.Cut(query, "#")
	// A malformed pair is skipped rather than failing the whole lookup.
	values, _ := url.ParseQuery(query)
	return values.Get(name)
}

// preview is the start of a response body, for an error message.
func preview(body string) string {
	runes := []rune(body)
	if len(runes) <= previewRunes {
		return body
	}
	return string(runes[:previewRunes])
}
