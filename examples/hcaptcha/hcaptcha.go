//go:build ignore

// Solve an hCaptcha challenge.
//
// Task type: `HCaptcha`
//
// Parameters the SDK does not model go through Extra and reach the worker
// unchanged.
//
//	go run examples/hcaptcha/hcaptcha.go
//
// Docs: https://docs.ezxlabs.com/docs/captcha/api/hcaptcha
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

	solved, err := client.SolveHCaptcha(ctx, &ezcapsolver.HCaptchaTask{
		WebsiteURL: "https://accounts.hcaptcha.com/demo",
		WebsiteKey: "338af34c-7bcb-4c7c-900b-acbec73d7d43",
		Lang:       "en-US",
		Proxy:      os.Getenv("EZCAPTCHA_PROXY"),
		// False because the demo page shows a checkbox. Sites that hide it need
		// true here.
		Invisible: false,
		// Only the sites that publish an rqdata value need this one.
		RqData: os.Getenv("EZCAPTCHA_HCAPTCHA_RQDATA"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Pass:    %s\n", solved.Solution.GeneratedPassUUID)
	fmt.Printf("Ua:      %s\n", solved.Solution.Ua)
}
