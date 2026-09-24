package ezcapsolver

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Defaults shared by every language SDK. Overriding any of them is a per-client
// decision made with an [Option]; the ones not overridden keep these values.
const (
	// DefaultAsyncBaseURL serves asynchronous tasks and balance queries.
	DefaultAsyncBaseURL = "https://api.ez-captcha.com"

	// DefaultSyncBaseURL serves synchronous tasks. The service splits the two
	// deployments, so they cannot share one host.
	DefaultSyncBaseURL = "https://sync.ez-captcha.com"

	// DefaultClientKeyEnv is read when no client key is supplied explicitly.
	DefaultClientKeyEnv = "EZCAPTCHA_API_KEY"

	// DefaultTimeout bounds one call to an asynchronous endpoint or the balance
	// endpoint. Those are millisecond-scale enqueue and lookup operations, so a
	// short bound keeps a network fault distinguishable from a slow service.
	DefaultTimeout = 30 * time.Second

	// DefaultSyncTimeout bounds one call to the synchronous task endpoint.
	//
	// A synchronous call blocks until the worker answers, and the service grants
	// its slowest synchronous types a 180-second worker deadline. Sharing
	// DefaultTimeout would abort a call that has already been billed, and a call
	// that times out never receives the task ID it would take to recover the
	// result. Matching 180 seconds exactly would leave no margin for the network
	// and the service's own overhead, so this adds one minute of headroom.
	DefaultSyncTimeout = 240 * time.Second

	// DefaultPollInterval separates two result queries.
	DefaultPollInterval = 3 * time.Second

	// DefaultMaxPollAttempts caps the number of result queries, which with
	// DefaultPollInterval puts the wait ceiling at two and a half minutes.
	DefaultMaxPollAttempts = 50
)

// LevelTrace logs full request and response bodies, with credentials redacted.
//
// slog has no trace level, so this sits one step below [slog.LevelDebug]. Enable
// it with a handler option: slog.HandlerOptions{Level: ezcapsolver.LevelTrace}.
const LevelTrace = slog.Level(-8)

// PollingConfig controls how an asynchronous task is waited on.
type PollingConfig struct {
	// Interval is the delay applied before every result query, the first one
	// included. A task that was just created is still queued for a worker, so
	// querying immediately after creation almost always reports processing.
	Interval time.Duration
	// MaxAttempts is the number of result queries allowed before
	// [PollingExhaustedError] is returned.
	MaxAttempts int
}

// DefaultPolling returns the polling settings a client starts with.
func DefaultPolling() PollingConfig {
	return PollingConfig{Interval: DefaultPollInterval, MaxAttempts: DefaultMaxPollAttempts}
}

// validate rejects settings that would make waiting meaningless.
func (p PollingConfig) validate() error {
	if p.Interval <= 0 {
		return fmt.Errorf("%w: polling interval must be greater than zero", ErrConfig)
	}
	if p.MaxAttempts <= 0 {
		return fmt.Errorf("%w: maximum polling attempts must be greater than zero", ErrConfig)
	}
	return nil
}

// ClientConfig is the effective configuration of a client.
//
// Build one with [Option] values passed to [NewClient] and read it back with
// [EzCapSolverClient.Config]. The client key and the SDK proxy are deliberately
// unexported: they are credentials, and [ClientConfig.String] replaces them so
// that printing a config can never leak them.
type ClientConfig struct {
	clientKey string
	proxy     string

	// AsyncBaseURL serves asynchronous tasks and balance queries.
	AsyncBaseURL string
	// SyncBaseURL serves synchronous tasks.
	SyncBaseURL string
	// Timeout bounds one asynchronous or balance request.
	Timeout time.Duration
	// SyncTimeout bounds one synchronous task request.
	SyncTimeout time.Duration
	// Polling controls how an asynchronous task is waited on.
	Polling PollingConfig
	// AppID is the optional developer application identifier. Nil means unset,
	// which is different from zero.
	AppID *int
	// UserAgent is sent with every request.
	UserAgent string
}

// defaultConfig returns the configuration a client starts from.
func defaultConfig() ClientConfig {
	return ClientConfig{
		AsyncBaseURL: DefaultAsyncBaseURL,
		SyncBaseURL:  DefaultSyncBaseURL,
		Timeout:      DefaultTimeout,
		SyncTimeout:  DefaultSyncTimeout,
		Polling:      DefaultPolling(),
		UserAgent:    defaultUserAgent,
	}
}

// String renders the configuration with the credentials replaced.
//
// Implementing Stringer is what makes this safe: fmt prints unexported struct
// fields too, so %v and %+v would otherwise expose the client key.
func (c ClientConfig) String() string {
	appID := "<nil>"
	if c.AppID != nil {
		appID = fmt.Sprint(*c.AppID)
	}
	return fmt.Sprintf(
		"ClientConfig{ClientKey:%s Proxy:%s AsyncBaseURL:%s SyncBaseURL:%s "+
			"Timeout:%s SyncTimeout:%s Polling:{Interval:%s MaxAttempts:%d} AppID:%s UserAgent:%s}",
		redactedOrEmpty(c.clientKey), redactedOrEmpty(c.proxy),
		c.AsyncBaseURL, c.SyncBaseURL, c.Timeout, c.SyncTimeout,
		c.Polling.Interval, c.Polling.MaxAttempts, appID, c.UserAgent,
	)
}

// GoString covers the %#v verb, which ignores Stringer.
func (c ClientConfig) GoString() string { return c.String() }

// redactedOrEmpty keeps "unset" and "set but hidden" tellable apart in debug
// output without revealing the value.
func redactedOrEmpty(value string) string {
	if value == "" {
		return `""`
	}
	return redactedPlaceholder
}

// validate checks everything that can be checked without a network call, so an
// invalid setting fails at construction rather than on the first request.
func (c ClientConfig) validate() error {
	if strings.TrimSpace(c.clientKey) == "" {
		return fmt.Errorf(
			"%w: client key must not be blank; set it with WithClientKey or the %s environment variable",
			ErrConfig, DefaultClientKeyEnv,
		)
	}
	if !validKeyCharacters(c.clientKey) {
		return fmt.Errorf("%w: client key contains invalid characters", ErrConfig)
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("%w: timeout must be greater than zero", ErrConfig)
	}
	if c.SyncTimeout <= 0 {
		return fmt.Errorf("%w: sync timeout must be greater than zero", ErrConfig)
	}
	if c.AsyncBaseURL == "" || c.SyncBaseURL == "" {
		return fmt.Errorf("%w: base URLs must not be empty", ErrConfig)
	}
	if c.proxy != "" {
		if _, err := url.Parse(c.proxy); err != nil {
			return fmt.Errorf("%w: SDK proxy URL is not parseable: %w", ErrConfig, err)
		}
	}
	return c.Polling.validate()
}

// builder accumulates what [Option] values set before [NewClient] validates it.
type builder struct {
	config     ClientConfig
	httpClient *http.Client
	logger     *slog.Logger
}

// Option customises a client. Options are applied in order, so a later one wins
// over an earlier one, and anything left untouched keeps its default.
type Option func(*builder)

// WithClientKey sets the client key instead of reading [DefaultClientKeyEnv].
func WithClientKey(clientKey string) Option {
	return func(b *builder) { b.config.clientKey = clientKey }
}

// WithTimeout sets the bound on one asynchronous or balance request.
func WithTimeout(timeout time.Duration) Option {
	return func(b *builder) { b.config.Timeout = timeout }
}

// WithSyncTimeout sets the bound on one synchronous task request.
//
// [DefaultSyncTimeout] already clears the service's longest worker deadline of
// 180 seconds with a minute to spare; lower it only if you would rather fail
// fast than wait.
func WithSyncTimeout(timeout time.Duration) Option {
	return func(b *builder) { b.config.SyncTimeout = timeout }
}

// WithPolling sets how asynchronous tasks are waited on.
func WithPolling(polling PollingConfig) Option {
	return func(b *builder) { b.config.Polling = polling }
}

// WithAppID sets the optional developer application identifier.
func WithAppID(appID int) Option {
	return func(b *builder) { b.config.AppID = &appID }
}

// WithProxy routes the SDK's own traffic through a proxy.
//
// This is unrelated to the proxy field on a task, which is what the worker uses
// to reach the protected site.
//
// It is ignored when [WithHTTPClient] supplies a client, since that client
// brings its own transport.
func WithProxy(proxy string) Option {
	return func(b *builder) { b.config.proxy = proxy }
}

// WithUserAgent overrides the User-Agent sent with every request.
func WithUserAgent(userAgent string) Option {
	return func(b *builder) { b.config.UserAgent = userAgent }
}

// WithBaseURLs points the client at different hosts, for a self-hosted
// deployment or an httptest server. An empty string keeps the current value.
func WithBaseURLs(asyncBaseURL, syncBaseURL string) Option {
	return func(b *builder) {
		if asyncBaseURL != "" {
			b.config.AsyncBaseURL = asyncBaseURL
		}
		if syncBaseURL != "" {
			b.config.SyncBaseURL = syncBaseURL
		}
	}
}

// WithHTTPClient supplies the HTTP client to send requests with.
//
// Its Timeout field is left alone and should stay zero: the SDK applies its own
// per-request deadline through the context, and the two budgets differ between
// the asynchronous and synchronous endpoints. A client-wide timeout would cut
// the longer one short.
func WithHTTPClient(client *http.Client) Option {
	return func(b *builder) { b.httpClient = client }
}

// WithLogger sends the SDK's logs to a logger. Without it, logging is discarded.
//
// See [LevelTrace] for full request and response bodies.
func WithLogger(logger *slog.Logger) Option {
	return func(b *builder) { b.logger = logger }
}

// validKeyCharacters reports whether the key is printable ASCII with no
// surrounding whitespace. The key also travels in a request header, and
// net/http refuses to send a header value with control characters, so a key
// that fails here would otherwise fail every request with a transport error.
func validKeyCharacters(key string) bool {
	if key != strings.TrimSpace(key) {
		return false
	}
	for _, r := range key {
		if r < 0x20 || r > 0x7e {
			return false
		}
	}
	return true
}
