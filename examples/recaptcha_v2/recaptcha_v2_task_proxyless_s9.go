//go:build ignore

// Solve a reCAPTCHA v2 challenge on the high-score queue.
//
// Task type: `ReCaptchaV2TaskProxylessS9`
//
// Same parameters as the plain V2 type; the difference is the queue, which
// returns a token scoring at least 0.9. It costs more and takes longer.
//
//	go run examples/recaptcha_v2/recaptcha_v2_task_proxyless_s9.go
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

	solved, err := client.SolveRecaptchaV2TaskProxylessS9(ctx, &ezcapsolver.RecaptchaV2Task{
		WebsiteURL: "https://www.google.com/recaptcha/api2/demo",
		WebsiteKey: "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
}
