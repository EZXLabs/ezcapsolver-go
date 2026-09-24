//go:build ignore

// Solve a reCAPTCHA v2 challenge that carries an `s` parameter.
//
// Task type: `ReCaptchaV2STaskProxyless`
//
// Some sites put a per-session `s` value in the widget's configuration.
// Supplying it routes the task to the high-score IPv4 queue. Despite the type
// name, the service does not actually require the parameter.
//
//	go run examples/recaptcha_v2/recaptcha_v2_s_task_proxyless.go
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

	solved, err := client.SolveRecaptchaV2STaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
		WebsiteURL: "https://example.com/login",
		WebsiteKey: "6Lc...",
		S:          "the-s-value-from-the-page",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
}
