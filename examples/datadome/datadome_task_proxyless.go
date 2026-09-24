//go:build ignore

// Answer a DataDome challenge.
//
// Task type: `DataDomeTaskProxyless`
//
// This is the interception path: DataDome served a challenge page, and the two
// steps below turn it into a cleared session. Step one returns the challenge
// URL to fetch; step two returns the validation instructions to submit.
//
// Referer is where the SDK cannot help you: the service treats it as optional,
// but the real workflow needs it, and it is what site allow-listing is checked
// against. With an allow-list configured, a referer outside it is rejected with
// ERROR_WEBSITE_NOT_ALLOWED.
//
//	go run examples/datadome/datadome_task_proxyless.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	first, err := client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
		HtmlB64: "base64-encoded challenge page",
		Step:    ezcapsolver.DataDomeStepOne,
		Referer: "https://example.com/",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Challenge: %s\n", first.Solution.URL)

	// Fetch that URL, then send what it returns back as step two.
	second, err := client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
		HtmlB64: "base64-encoded challenge page",
		Step:    ezcapsolver.DataDomeStepTwo,
		Referer: "https://example.com/",
		Image:   "base64-encoded slider image",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Kind:      %s\n", second.Solution.Kind)
	fmt.Printf("Validate:  %s\n", second.Solution.URL)
	fmt.Printf("Body:      %s\n", second.Solution.Body)
}
