//go:build ignore

// Solve a reCAPTCHA v3 challenge.
//
// Task type: `ReCaptchaV3TaskProxyless`
//
// V3 runs invisibly and scores the visitor rather than asking anything, so
// pageAction has to match what the page uses — a mismatch lowers the score.
//
//	go run examples/recaptcha_v3/recaptcha_v3_task_proxyless.go
//
// Docs: https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v3
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

	solved, err := client.SolveRecaptchaV3TaskProxyless(ctx, &ezcapsolver.RecaptchaV3Task{
		WebsiteURL: "https://example.com",
		WebsiteKey: "6Lc...",
		PageAction: "login",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
}
