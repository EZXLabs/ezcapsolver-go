package ezcapsolver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

// redactedPlaceholder replaces a credential in any human-readable output.
const redactedPlaceholder = "[REDACTED]"

// redactedKeys are the JSON keys whose values are credentials. They are
// replaced at every nesting depth before anything is logged.
var redactedKeys = []string{"clientKey", "proxy"}

// tracePreviewChars caps a body in a trace log. An error envelope fits well
// inside it; a solved token gets truncated, which is the point.
const tracePreviewChars = 256

// clientKeyHeader carries the client key on every request, alongside the
// clientKey field of the body.
const clientKeyHeader = "X-API-Key"

// endpoint joins a base URL and a path, tolerating a slash on either side.
func endpoint(baseURL, path string) string {
	return strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
}

// post sends one request and decodes the response into target.
//
// One call means one HTTP request: this function never retries. Creating a task
// is billed and is not idempotent, and the service bans a key that repeats
// certain errors, so retrying is the caller's decision to make. The one retry
// the SDK performs lives a layer up, in the polling loop, and only for a
// throttled query — see [APIError.IsRateLimited].
func (c *EzCapSolverClient) post(
	ctx context.Context, url string, body any, timeout time.Duration, target any,
) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("%w: request body is not serializable: %w", ErrConfig, err)
	}

	c.logRequest(ctx, url, payload)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return &TransportError{Op: "POST " + url, Err: err}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", c.config.UserAgent)
	request.Header.Set(clientKeyHeader, c.clientKey)

	started := time.Now()
	response, err := c.httpClient.Do(request)
	if err != nil {
		return &TransportError{Op: "POST " + url, Err: err}
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return &TransportError{Op: "reading the response of POST " + url, Err: err}
	}
	c.logResponse(ctx, url, response.StatusCode, time.Since(started), responseBody)

	return parseAPIResponse(response.StatusCode, responseBody, target)
}

// parseAPIResponse turns a status and body into either a decoded result or a
// classified error.
//
// Both of the service's success signals have to be read, and they disagree
// often enough to matter: a failed task arrives as HTTP 200 with errorId 1, and
// most business errors arrive as HTTP 500 rather than as a service fault.
func parseAPIResponse(status int, body []byte, target any) error {
	ok := status >= 200 && status < 300

	var envelope ResponseMeta
	if err := json.Unmarshal(body, &envelope); err != nil {
		// A non-success status whose body is not JSON — an HTML page from a
		// gateway, say — is still a failure from the service, so it is reported
		// with the same shape as one that carried an envelope. That keeps the
		// status code readable from a single place.
		if !ok {
			return statusOnlyError(status, body)
		}
		// This covers both a body that is not JSON at all and one that is valid
		// JSON but not an object — a bare string or array from a proxy, say.
		return &UnexpectedResponseError{
			Reason: fmt.Sprintf("response body is not a JSON object: %v", err),
			Body:   preview(body, errorPreviewChars),
		}
	}

	// errorId is the only success criterion. Treating a non-empty errorCode as a
	// second signal would let a code added on the success side turn a solved
	// task into an error, and every code the service defines already comes with
	// a non-zero errorId.
	if envelope.ErrorID != 0 {
		return &APIError{
			ErrorCode:        envelope.ErrorCode,
			ErrorDescription: envelope.ErrorDescription,
			HTTPStatus:       status,
			Errors:           validationErrors(body),
			RequestID:        envelope.RequestID,
		}
	}
	if !ok {
		return statusOnlyError(status, body)
	}

	if err := json.Unmarshal(body, target); err != nil {
		return &UnexpectedResponseError{
			Reason: fmt.Sprintf("response does not match the expected shape: %v", err),
			Body:   preview(body, errorPreviewChars),
		}
	}
	return nil
}

// validationErrors pulls the field-level messages out of a validation failure.
//
// The key is absent on every other kind of error, and decoding it separately
// keeps [ResponseMeta] to the four fields every response really has.
func validationErrors(body []byte) map[string]string {
	var envelope struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil
	}
	return envelope.Errors
}

// statusOnlyError reports a non-success status whose body carried no envelope,
// keeping a slice of that body for diagnosis.
func statusOnlyError(status int, body []byte) error {
	return &APIError{
		ErrorDescription: preview(body, errorPreviewChars),
		HTTPStatus:       status,
	}
}

// logRequest records an outgoing body with credentials replaced.
//
// Nothing is rendered unless trace logging is on, so this costs nothing at the
// default level.
func (c *EzCapSolverClient) logRequest(ctx context.Context, url string, payload []byte) {
	if !c.logger.Enabled(ctx, LevelTrace) {
		return
	}
	var decoded any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		c.logger.Log(ctx, LevelTrace, "sending API request with an unrenderable body",
			"url", url)
		return
	}
	rendered, err := json.Marshal(redact(decoded))
	if err != nil {
		c.logger.Log(ctx, LevelTrace, "sending API request with an unrenderable body",
			"url", url)
		return
	}
	c.logger.Log(ctx, LevelTrace, "sending API request",
		"url", url,
		"body", preview(rendered, tracePreviewChars),
	)
}

// logResponse records a response's status, duration, size and truncated body.
func (c *EzCapSolverClient) logResponse(
	ctx context.Context, url string, status int, elapsed time.Duration, body []byte,
) {
	if !c.logger.Enabled(ctx, LevelTrace) {
		return
	}
	c.logger.Log(ctx, LevelTrace, "received API response",
		"url", url,
		"status", status,
		"elapsed_ms", elapsed.Milliseconds(),
		"bytes", len(body),
		"body", preview(body, tracePreviewChars),
	)
}

// redact replaces credential values in a decoded JSON document, at any depth.
func redact(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			if slices.Contains(redactedKeys, key) {
				redacted[key] = redactedPlaceholder
				continue
			}
			redacted[key] = redact(item)
		}
		return redacted
	case []any:
		redacted := make([]any, len(typed))
		for index, item := range typed {
			redacted[index] = redact(item)
		}
		return redacted
	default:
		return value
	}
}
