package ext

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestChallengeOutputRedactsTheProxy(t *testing.T) {
	challenge := DataDomeChallenge{
		TargetURL: "https://example.com",
		UserAgent: "Chrome/153.0.0.0",
		Proxy:     "http://user:secret@proxy.example.com:8080",
	}
	for _, verb := range []string{"%v", "%+v", "%#v", "%s"} {
		for _, value := range []any{challenge, &challenge} {
			rendered := fmt.Sprintf(verb, value)
			if strings.Contains(rendered, "secret") || !strings.Contains(rendered, redactedPlaceholder) {
				t.Errorf("%s leaked or dropped the proxy: %s", verb, rendered)
			}
		}
	}

	if rendered := fmt.Sprint(DataDomeChallenge{}); strings.Contains(rendered, redactedPlaceholder) {
		t.Errorf("an unset proxy must read as unset: %s", rendered)
	}
}

func TestAnEmptyAcceptLanguageFallsBackToTheDefault(t *testing.T) {
	if got := (DataDomeChallenge{}).acceptLanguage(); got != defaultAcceptLanguage {
		t.Errorf("acceptLanguage = %q, want %q", got, defaultAcceptLanguage)
	}
	if got := (DataDomeChallenge{AcceptLanguage: "ja"}).acceptLanguage(); got != "ja" {
		t.Errorf("acceptLanguage = %q, want \"ja\"", got)
	}
}

func TestABanIsAlsoAWorkflowFailure(t *testing.T) {
	if !errors.Is(blocked("the target page"), ErrDataDome) {
		t.Error("errors.Is must match ErrDataDome against a ban")
	}
}
