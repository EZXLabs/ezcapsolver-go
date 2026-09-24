//go:build ignore

// Solve a reCAPTCHA v3 Enterprise challenge.
//
// Task type: `ReCaptchaV3EnterpriseTaskProxyless`
//
//	go run examples/recaptcha_v3/recaptcha_v3_enterprise_task_proxyless.go
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

	solved, err := client.SolveRecaptchaV3EnterpriseTaskProxyless(ctx, &ezcapsolver.RecaptchaV3Task{
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
