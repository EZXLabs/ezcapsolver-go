//go:build ignore

// Solve a FunCaptcha (Arkose Labs) challenge.
//
// Task type: `FuncaptchaTaskProxyless`
//
// Note the canonical spelling: `Funcaptcha` here, but `FunCaptcha` in the
// classification type below it. That inconsistency is the service's, and the
// SDK reproduces it rather than tidying it up.
//
// This is the one task type whose proxy uses the FUN format —
// protocol://host:port:username:password, with the credentials after the host
// rather than before it. Every other type takes the usual form.
//
//	go run examples/funcaptcha/funcaptcha_task_proxyless.go
//
// Docs: https://docs.ezxlabs.com/docs/captcha/api/funcaptcha
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

	solved, err := client.SolveFuncaptchaTaskProxyless(ctx, &ezcapsolver.FunCaptchaTask{
		WebsiteURL: "https://client-api.arkoselabs.com",
		WebsiteKey: "B7D8911C-5CC8-A9A3-35B0-554ACEE604DA",
		// Some integrations pass a per-session blob taken from the page.
		Data:  `{"blob":"..."}`,
		Proxy: os.Getenv("EZCAPTCHA_PROXY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("Token:   %.64s...\n", solved.Solution.Token)
}
