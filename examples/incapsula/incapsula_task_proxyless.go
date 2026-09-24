//go:build ignore

// Produce an Incapsula Reese84 sensor payload.
//
// Task type: `IncapsulaTaskProxyless`
//
// The result's Data is stringified JSON, and it has to be posted to the reese84
// endpoint exactly as it arrived. The SDK deliberately does not parse it:
// re-encoding would change the bytes, and the bytes are what is being checked.
//
// Supplying a proxy changes what the task does — with one, the worker submits
// the payload itself and returns the cookie; without one, it only generates the
// payload. Note that this type never satisfies a plan's mandatory-proxy
// requirement either way.
//
//	go run examples/incapsula/incapsula_task_proxyless.go
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

	solved, err := client.SyncSolveIncapsulaTaskProxyless(ctx, &ezcapsolver.IncapsulaTask{
		Script:         "the full reese84 sensor script source",
		ScriptURL:      "https://example.com/path/to/sensor.js",
		PageURL:        "https://example.com",
		AcceptLanguage: "ja-JP,ja;q=0.9,en;q=0.8",
		// Only Chrome 147, 148 and 149 are supported; the major version is all
		// the service reads out of this.
		Ua:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/149.0.0.0 Safari/537.36",
		Proxy: os.Getenv("EZCAPTCHA_PROXY"),
		// Only the sites with PoW challenges enabled need this one.
		Pow: os.Getenv("EZCAPTCHA_INCAPSULA_POW"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status: %d\n", solved.Solution.Status)
	fmt.Printf("Data:   %.64s...\n", solved.Solution.Data)
}
