//go:build ignore

// Solve a reCAPTCHA v3 Enterprise challenge on the high-score queue.
//
// Task type: `RecaptchaV3EnterpriseTaskProxylessS9`
//
// Note the canonical spelling of this one: it opens with a lower-case c, unlike
// every other V3 type. That is the service's spelling, not a typo, and the SDK
// copies it verbatim.
//
//	go run examples/recaptcha_v3/recaptcha_v3_enterprise_task_proxyless_s9.go
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

	solved, err := client.SolveRecaptchaV3EnterpriseTaskProxylessS9(ctx, &ezcapsolver.RecaptchaV3Task{
		WebsiteURL: "https://example.com",
		WebsiteKey: "6Lc...",
		PageAction: "checkout",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
}
