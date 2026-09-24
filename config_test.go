package ezcapsolver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// TestDefaultsMatchTheSpecification pins the values every language SDK shares.
// Drifting on any of them makes cross-language behaviour reports incomparable.
func TestDefaultsMatchTheSpecification(t *testing.T) {
	t.Setenv(DefaultClientKeyEnv, "key-from-the-environment")

	client, err := NewClient()
	if err != nil {
		t.Fatalf("building the client failed: %v", err)
	}
	config := client.Config()

	if config.Timeout != 30*time.Second {
		t.Errorf("timeout: got %s, want 30s", config.Timeout)
	}
	if config.SyncTimeout != 240*time.Second {
		t.Errorf("sync timeout: got %s, want 240s", config.SyncTimeout)
	}
	// The two budgets have to stay apart: the synchronous endpoint blocks until
	// a worker answers, and the service allows some task types three minutes.
	if config.Timeout == config.SyncTimeout {
		t.Error("the two timeouts must not share a value")
	}
	// And it has to clear that three-minute worker deadline, or a worker that
	// uses its full budget gets cut off by the client that is paying for it.
	if config.SyncTimeout <= 180*time.Second {
		t.Errorf("sync timeout %s leaves no margin over the 180s worker deadline", config.SyncTimeout)
	}
	if config.Polling.Interval != 3*time.Second || config.Polling.MaxAttempts != 50 {
		t.Errorf("polling: got %+v, want 3s x 40", config.Polling)
	}
	if config.AsyncBaseURL != "https://api.ez-captcha.com" {
		t.Errorf("async base URL: got %q", config.AsyncBaseURL)
	}
	if config.SyncBaseURL != "https://sync.ez-captcha.com" {
		t.Errorf("sync base URL: got %q", config.SyncBaseURL)
	}
	if config.UserAgent != "ezcapsolver-go/"+Version {
		t.Errorf("user agent: got %q", config.UserAgent)
	}
	if config.AppID != nil {
		t.Errorf("appId must be unset by default, got %v", *config.AppID)
	}
}

// TestClientKeyResolution checks the order the key is looked up in.
func TestClientKeyResolution(t *testing.T) {
	t.Run("the environment is the fallback", func(t *testing.T) {
		t.Setenv(DefaultClientKeyEnv, "key-from-the-environment")

		client, err := NewClient()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.clientKey != "key-from-the-environment" {
			t.Errorf("got %q", client.clientKey)
		}
	})

	t.Run("an explicit key wins", func(t *testing.T) {
		t.Setenv(DefaultClientKeyEnv, "key-from-the-environment")

		client, err := NewClient(WithClientKey("explicit-key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.clientKey != "explicit-key" {
			t.Errorf("got %q", client.clientKey)
		}
	})

	t.Run("a blank key fails at construction", func(t *testing.T) {
		// Failing here rather than on the first request means a misconfigured
		// deployment shows up before anything is billed.
		t.Setenv(DefaultClientKeyEnv, "")

		_, err := NewClient(WithClientKey("   "))
		if !errors.Is(err, ErrConfig) {
			t.Fatalf("got %v, want a configuration error", err)
		}
		// The message has to say how to fix it.
		if !strings.Contains(err.Error(), DefaultClientKeyEnv) {
			t.Errorf("got %q", err)
		}
	})

	t.Run("a key with characters a request cannot carry fails at construction", func(t *testing.T) {
		for _, key := range []string{"key\nX-Injected: 1", " padded ", "k\u00e9y", "key\x00"} {
			_, err := NewClient(WithClientKey(key))
			if !errors.Is(err, ErrConfig) {
				t.Fatalf("%q: got %v, want a configuration error", key, err)
			}
			if !strings.Contains(err.Error(), "client key contains invalid characters") {
				t.Errorf("%q: got %q", key, err)
			}
		}
	})
}

// TestInvalidSettingsAreRejected checks that everything checkable is checked at
// construction.
func TestInvalidSettingsAreRejected(t *testing.T) {
	cases := []struct {
		name   string
		option Option
	}{
		{"zero timeout", WithTimeout(0)},
		{"negative timeout", WithTimeout(-time.Second)},
		{"zero sync timeout", WithSyncTimeout(0)},
		{"zero poll interval", WithPolling(PollingConfig{Interval: 0, MaxAttempts: 5})},
		{"zero attempts", WithPolling(PollingConfig{Interval: time.Second, MaxAttempts: 0})},
		{"unparsable proxy", WithProxy("://not a url")},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewClient(WithClientKey("key"), testCase.option)
			if !errors.Is(err, ErrConfig) {
				t.Fatalf("got %v, want a configuration error", err)
			}
		})
	}

	t.Run("empty base URLs keep the defaults", func(t *testing.T) {
		// WithBaseURLs treats an empty string as "leave this one alone", so
		// overriding only the asynchronous host is a single call.
		client, err := NewClient(WithClientKey("key"), WithBaseURLs("", ""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.Config().AsyncBaseURL != DefaultAsyncBaseURL {
			t.Errorf("got %q", client.Config().AsyncBaseURL)
		}
	})
}

// TestOptionsOverrideIndividually checks that setting one option leaves the
// others at their defaults.
func TestOptionsOverrideIndividually(t *testing.T) {
	client, err := NewClient(
		WithClientKey("key"),
		WithSyncTimeout(300*time.Second),
		WithUserAgent("my-app/2.0"),
		WithAppID(7),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	config := client.Config()
	if config.SyncTimeout != 300*time.Second {
		t.Errorf("sync timeout: got %s", config.SyncTimeout)
	}
	if config.UserAgent != "my-app/2.0" {
		t.Errorf("user agent: got %q", config.UserAgent)
	}
	if config.AppID == nil || *config.AppID != 7 {
		t.Errorf("appId: got %v", config.AppID)
	}
	if config.Timeout != DefaultTimeout {
		t.Errorf("an untouched setting must keep its default, got %s", config.Timeout)
	}

	t.Run("a later option wins", func(t *testing.T) {
		client, err := NewClient(WithClientKey("key"), WithTimeout(time.Second), WithTimeout(time.Minute))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.Config().Timeout != time.Minute {
			t.Errorf("got %s", client.Config().Timeout)
		}
	})
}

// TestConfigDebugOutputRedactsCredentials checks that printing a config cannot
// leak the key or the proxy.
//
// This matters more in Go than elsewhere: fmt prints unexported struct fields
// too, so hiding them is not enough on its own — the Stringer is what does it.
func TestConfigDebugOutputRedactsCredentials(t *testing.T) {
	client, err := NewClient(
		WithClientKey("super-secret-key"),
		WithProxy("http://user:hunter2@proxy.example.com:8080"),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	config := client.Config()

	// Every verb a caller might reach for, including the ones that normally
	// bypass a Stringer.
	for _, rendered := range []string{
		fmt.Sprint(config),
		fmt.Sprintf("%v", config),
		fmt.Sprintf("%+v", config),
		fmt.Sprintf("%#v", config),
		//lint:ignore S1025 the point is to check this verb, not to render the value
		fmt.Sprintf("%s", config),
		config.String(),
	} {
		if strings.Contains(rendered, "super-secret-key") {
			t.Errorf("the client key leaked: %s", rendered)
		}
		if strings.Contains(rendered, "hunter2") || strings.Contains(rendered, "proxy.example.com") {
			t.Errorf("the proxy leaked: %s", rendered)
		}
		if !strings.Contains(rendered, redactedPlaceholder) {
			t.Errorf("expected a redaction marker in %s", rendered)
		}
	}

	t.Run("an unset credential is distinguishable from a hidden one", func(t *testing.T) {
		client, err := NewClient(WithClientKey("key"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.Contains(client.Config().String(), `Proxy:""`) {
			t.Errorf("got %s", client.Config())
		}
	})
}

// TestWithHTTPClientIsUsed checks that an injected client actually carries the
// requests, which is how a caller plugs in their own transport or instrumentation.
func TestWithHTTPClientIsUsed(t *testing.T) {
	service := newFakeService(t, map[string]func(int) (int, string){
		pathGetBalance: ok(`{"errorId":0,"balance":5}`),
	})

	used := false
	custom := &http.Client{Transport: roundTripperFunc(
		func(request *http.Request) (*http.Response, error) {
			used = true
			return http.DefaultTransport.RoundTrip(request)
		})}

	client := newTestClient(t, service, WithHTTPClient(custom))
	if _, err := client.Balance(t.Context()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !used {
		t.Error("the injected HTTP client was bypassed")
	}
}

// roundTripperFunc adapts a function to http.RoundTripper.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

// TestVersionMatchesUserAgent guards the format the service uses to identify
// callers, and that the constant was not left empty.
func TestVersionMatchesUserAgent(t *testing.T) {
	if Version == "" {
		t.Fatal("Version must be set")
	}
	if want := "ezcapsolver-go/" + Version; defaultUserAgent != want {
		t.Errorf("got %q, want %q", defaultUserAgent, want)
	}
	if strings.HasPrefix(Version, "v") {
		t.Errorf("Version must not carry a leading v, got %q", Version)
	}
}
