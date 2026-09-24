//go:build ignore

// Solve a reCAPTCHA v3 challenge on the high-score queue.
//
// Task type: `ReCaptchaV3TaskProxylessS9`
//
// Use this when the site rejects tokens that score below 0.9.
//
//	go run examples/recaptcha_v3/recaptcha_v3_task_proxyless_s9.go
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

	solved, err := client.SolveRecaptchaV3TaskProxylessS9(ctx, &ezcapsolver.RecaptchaV3Task{
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
