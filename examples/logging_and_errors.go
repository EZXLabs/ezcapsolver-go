//go:build ignore

// Turn on logging, and tell the five kinds of failure apart.
//
//	go run examples/logging_and_errors.go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	ctx := context.Background()

	// Logging is discarded unless a logger is injected. LevelTrace adds full
	// request and response bodies, with clientKey and proxy replaced at any
	// nesting depth — nothing is even rendered below that level, so leaving it
	// off costs nothing.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: ezcapsolver.LevelTrace,
	}))

	client, err := ezcapsolver.NewClient(ezcapsolver.WithLogger(logger))
	if err != nil {
		// A configuration error is the one failure that surfaces here rather
		// than on a request: everything checkable is checked at construction,
		// so a bad setting shows up before anything is billed.
		log.Fatal(err)
	}

	_, err = client.SolveRecaptchaV2TaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
		WebsiteURL: "https://www.google.com/recaptcha/api2/demo",
		WebsiteKey: "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
	})
	classify(err)
}

// classify shows what each failure layer means and what to do about it.
func classify(err error) {
	if err == nil {
		fmt.Println("Solved.")
		return
	}

	// Ask this first. Whether a billed task is still recoverable matters more
	// than what broke, because it decides whether you pay again — and it is one
	// question regardless of which error type happens to carry the id.
	if taskID := ezcapsolver.TaskIDOf(err); taskID != "" {
		fmt.Printf("Recover:    WaitForResult(%s) — do not create a second task\n", taskID)
	}

	switch {
	case errors.Is(err, ezcapsolver.ErrAPI):
		// The service answered with a structured error. What to do next depends
		// on the code, so the code is what you branch on.
		var apiErr *ezcapsolver.APIError
		errors.As(err, &apiErr)

		fmt.Printf("API error:  %s\n", apiErr.ErrorCode)
		fmt.Printf("HTTP:       %d\n", apiErr.HTTPStatus)
		fmt.Printf("Task ID:    %s\n", apiErr.TaskID)
		for field, reason := range apiErr.Errors {
			fmt.Printf("  %s: %s\n", field, reason)
		}

		switch {
		case apiErr.IsAuthenticationError():
			// Stop. Do not back off and try again: the service counts these per
			// key, and thirty within a minute earn a three-minute ban.
			fmt.Println("Fix the key or top up the balance; retrying will get you banned.")
		case apiErr.IsTerminal():
			fmt.Println("The same request will fail the same way. Change it.")
		case apiErr.IsRateLimited():
			// Throttling refuses the query, not the task. WaitForResult already
			// polls through one of these on its own.
			fmt.Println("Throttled. Both codes clear on their own; ask again later.")
		default:
			fmt.Println("This one may be transient. Backing off is up to you.")
		}

	case errors.Is(err, ezcapsolver.ErrTransport):
		// The request never produced a response. Whether resending is safe
		// depends on what was being done: creating a task is billed and is not
		// idempotent, and a timeout cannot tell you whether the service already
		// accepted it. Treat a create timeout as a terminal failure.
		fmt.Printf("Transport:  %v\n", err)
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Println("Timed out. The task may still have been created, and billed.")
		}

	case errors.Is(err, ezcapsolver.ErrPollingExhausted):
		// The budget ran out, but the task may still finish. Its ID stays
		// valid, so the result can be fetched later with GetTaskResult.
		var pollErr *ezcapsolver.PollingExhaustedError
		errors.As(err, &pollErr)
		fmt.Printf("Gave up on %s after %d attempts\n", pollErr.TaskID, pollErr.Attempts)

	case errors.Is(err, ezcapsolver.ErrDecode):
		// Either the worker returned a shape this release does not model, in
		// which case the raw value travels with the error and is the only thing
		// that explains it, or the response broke the contract some other way —
		// a ready result with no solution field at all, for instance.
		var decodeErr *ezcapsolver.SolutionDecodeError
		if errors.As(err, &decodeErr) {
			fmt.Printf("Undecodable solution: %s\n", decodeErr.Raw)
		} else {
			fmt.Printf("Unusable response: %v\n", err)
		}

	default:
		fmt.Printf("Unclassified: %v\n", err)
	}
}
