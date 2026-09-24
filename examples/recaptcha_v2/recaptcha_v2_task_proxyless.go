//go:build ignore

// Solve a reCAPTCHA v2 challenge.
//
// Task type: `ReCaptchaV2TaskProxyless`
//
// The site below is Google's own reCAPTCHA demo page, so this example runs as
// written once EZCAPTCHA_API_KEY is set:
//
//	go run examples/recaptcha_v2/recaptcha_v2_task_proxyless.go
//
// Docs: https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v2
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	ctx := context.Background()

	// With no key passed, the client reads EZCAPTCHA_API_KEY from the environment.
	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	solved, err := client.SolveRecaptchaV2TaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
		WebsiteURL: "https://www.google.com/recaptcha/api2/demo",
		WebsiteKey: "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Request ID: %s\n", orDash(solved.RequestID))
	fmt.Printf("Task ID:    %s\n", orDash(solved.TaskID))

	solution := solved.Solution
	fmt.Printf("User-Agent: %s\n", solution.UserAgent)
	fmt.Printf("Token:      %.64s...\n", solution.Token)

	balance, err := client.Balance(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Balance:    %v\n", balance)
}

func orDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
