//go:build ignore

// Solve a reCAPTCHA v2 Enterprise challenge.
//
// Task type: `ReCaptchaV2EnterpriseTaskProxyless`
//
// Enterprise widgets take the same parameters as the standard ones, so the
// request model is shared; only the task type differs.
//
//	go run examples/recaptcha_v2/recaptcha_v2_enterprise_task_proxyless.go
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

	solved, err := client.SolveRecaptchaV2EnterpriseTaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
		WebsiteURL: "https://example.com/enterprise",
		WebsiteKey: "6Lc...",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
}
