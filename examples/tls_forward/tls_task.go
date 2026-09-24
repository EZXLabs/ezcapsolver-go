//go:build ignore

// Forward one HTTP request through a worker's TLS fingerprint.
//
// Task type: `TlsTask`
//
// This one is not a captcha solver. The worker makes the request for you using
// its own TLS fingerprint, so a site that fingerprints the handshake sees a real
// browser rather than a Go client. The upstream response comes back whole:
// status, headers, cookies and body.
//
// Note the canonical type name is `TlsTask`, not `TLSTask`.
//
//	EZCAPTCHA_PROXY=http://user:pass@host:port go run examples/tls_forward/tls_task.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	solved, err := client.SyncSolveTLSTask(ctx, &ezcapsolver.TLSForwardTask{
		TLSType: "chrome146",
		Proxy:   os.Getenv("EZCAPTCHA_PROXY"),
		Method:  ezcapsolver.TLSMethodGET,
		URL:     "https://postman-echo.com/get",
		Headers: map[string]any{
			"accept":     "application/json",
			"user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/149.0.0.0",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Upstream status: %d\n", solved.Solution.Code)
	fmt.Printf("Headers:         %v\n", solved.Solution.Headers)
	fmt.Printf("Cookies:         %v\n", solved.Solution.Cookies)
	fmt.Printf("Body:            %.200s\n", solved.Solution.Body)
}
