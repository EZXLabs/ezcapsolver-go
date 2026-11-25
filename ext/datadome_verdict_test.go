package ext

import (
	"errors"
	"testing"
	"time"

	"github.com/EZXLabs/ezcapsolver-go"
)

func TestASliderOrBodylessAnswerIsSubmittedAsAGet(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		answer ezcapsolver.DataDomeSolution
		want   submission
	}{
		{"slider", ezcapsolver.DataDomeSolution{URL: "https://dd/check", Kind: "slider", Body: "x=1"},
			submission{url: "https://dd/check"}},
		{"bodyless", ezcapsolver.DataDomeSolution{URL: "https://dd/check"},
			submission{url: "https://dd/check"}},
		{"interstitial", ezcapsolver.DataDomeSolution{URL: "https://dd/i", Kind: "interstitial", Body: "x=1"},
			submission{url: "https://dd/i", body: "x=1"}},
	} {
		got, err := submissionFrom(testCase.answer)
		if err != nil || got != testCase.want {
			t.Errorf("%s: got %+v, %v; want %+v", testCase.name, got, err, testCase.want)
		}
	}

	if _, err := submissionFrom(ezcapsolver.DataDomeSolution{}); !errors.Is(err, ErrDataDome) {
		t.Errorf("an answer without a url cannot be submitted, got %v", err)
	}
}

func TestTheAnswerDecidesTheDelayWithADefaultForGets(t *testing.T) {
	get := submission{url: "u"}
	post := submission{url: "u", body: "x=1"}
	for _, testCase := range []struct {
		name       string
		answer     ezcapsolver.DataDomeSolution
		submission submission
		want       time.Duration
	}{
		{"timed", ezcapsolver.DataDomeSolution{SubmitAfterMs: 1500}, get, 1500 * time.Millisecond},
		{"untimed get", ezcapsolver.DataDomeSolution{}, get, defaultGetDelay},
		{"untimed post", ezcapsolver.DataDomeSolution{}, post, 0},
	} {
		if got := submitDelay(testCase.answer, testCase.submission); got != testCase.want {
			t.Errorf("%s: delay = %s, want %s", testCase.name, got, testCase.want)
		}
	}
}

func TestTheCookieIsSplitIntoItsParts(t *testing.T) {
	target := mustParse(t, "https://shop.example.com/signup")
	raw := "datadome=abc123; Max-Age=31536000; Domain=.example.com; Path=/; Secure"

	cookie, err := readVerdict(`{"view":"redirect","cookie":"`+raw+`"}`, target)
	want := DataDomeCookie{Name: "datadome", Value: "abc123", Domain: ".example.com", Raw: raw}
	if err != nil || cookie != want {
		t.Errorf("got %+v, %v; want %+v", cookie, err, want)
	}

	withoutDomain, err := readVerdict(`{"cookie":"datadome=v; Path=/"}`, target)
	if err != nil || withoutDomain.Domain != ".example.com" {
		t.Errorf("the domain must fall back to the target host, got %+v, %v", withoutDomain, err)
	}
}

func TestAResponseWithoutACookieSaysWhy(t *testing.T) {
	target := mustParse(t, "https://example.com")

	banned := `{"view":"captcha","url":"https://geo.captcha-delivery.com/captcha/?t=bv&cid=x"}`
	if _, err := readVerdict(banned, target); !errors.Is(err, ErrDataDomeBlocked) {
		t.Errorf("t=bv is a ban, got %v", err)
	}

	rejected := `{"view":"captcha","url":"https://geo.captcha-delivery.com/captcha/?t=fe"}`
	_, err := readVerdict(rejected, target)
	if !errors.Is(err, ErrDataDome) || errors.Is(err, ErrDataDomeBlocked) {
		t.Errorf("a rejected answer is not a ban, got %v", err)
	}

	if _, err := readVerdict("<html>oops</html>", target); !errors.Is(err, ErrDataDome) {
		t.Errorf("a response that is not JSON must fail the run, got %v", err)
	}
}
