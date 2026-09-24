//go:build ignore

// Obtain a Cloudflare Turnstile token.
//
// Task type: `CloudFlareTurnstileTask`
//
// Along with the token comes a set of headers to replay: submitting the token
// from a different fingerprint is what usually gets it rejected.
//
// RqData, when a site uses it, is sent as an object. The service stringifies it
// itself on the way to the worker, so encoding it yourself sends it twice over.
//
//	go run examples/cloudflare/cloud_flare_turnstile_task.go
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

	solved, err := client.SolveCloudFlareTurnstileTask(ctx, &ezcapsolver.CloudflareTurnstileTask{
		WebsiteURL: "https://example.com",
		WebsiteKey: "0x4AAA...",
		Proxy:      os.Getenv("EZCAPTCHA_PROXY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
	fmt.Printf("Headers: %v\n", solved.Solution.Header)
}
