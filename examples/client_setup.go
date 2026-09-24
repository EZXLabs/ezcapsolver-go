//go:build ignore

// Build a client, from the shortest form to a fully configured one.
//
// A client is safe to share across goroutines and holds the connection pool, so
// build one for the lifetime of your process rather than one per request.
//
//	go run examples/client_setup.go
package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	// The shortest form: the key comes from EZCAPTCHA_API_KEY.
	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Defaults: %s\n", client.Config())

	configured, err := ezcapsolver.NewClient(
		// An explicit key wins over the environment variable.
		ezcapsolver.WithClientKey(os.Getenv("EZCAPTCHA_API_KEY")),

		// Two separate budgets, and they should stay separate. Asynchronous
		// calls just enqueue and poll, so 30 seconds is already generous.
		// Synchronous calls block until a worker answers, and the service
		// allows some task types three minutes — cutting one short wastes a
		// call that has already been billed and has no task ID to recover.
		ezcapsolver.WithTimeout(30*time.Second),
		ezcapsolver.WithSyncTimeout(240*time.Second),

		// How long to wait for an asynchronous task: 3s x 50 is two and a half minutes.
		ezcapsolver.WithPolling(ezcapsolver.PollingConfig{
			Interval:    3 * time.Second,
			MaxAttempts: 50,
		}),

		// The SDK's own outbound proxy. Unrelated to a task's Proxy field,
		// which is what the worker uses to reach the protected site.
		ezcapsolver.WithProxy(os.Getenv("EZCAPTCHA_PROXY")),

		ezcapsolver.WithUserAgent("my-app/1.0"),
		ezcapsolver.WithAppID(0),
		ezcapsolver.WithLogger(slog.New(slog.NewTextHandler(os.Stderr, nil))),

		// Your own HTTP client, for a custom transport or instrumentation.
		// Leave its Timeout at zero: the SDK sets a deadline per request, and a
		// client-wide timeout would override the longer of the two budgets.
		ezcapsolver.WithHTTPClient(&http.Client{}),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Printing a config is safe: the key and the proxy are replaced.
	fmt.Printf("Configured: %s\n", configured.Config())

	balance, err := configured.Balance(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Balance: %v\n", balance)
}
