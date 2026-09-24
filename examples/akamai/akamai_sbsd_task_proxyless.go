//go:build ignore

// Produce Akamai SBSD sensor data.
//
// Task type: `AkamaiSBSDTaskProxyless`
//
// Unlike the Web flow, this is a single round.
//
//	go run examples/akamai/akamai_sbsd_task_proxyless.go
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

	solved, err := client.SyncSolveAkamaiSBSDTaskProxyless(ctx, &ezcapsolver.AkamaiSBSDTask{
		PageURL:      "https://example.com",
		SbsdURL:      "https://example.com/path/to/sbsd.js",
		BmSo:         "the bm_so cookie value",
		Ua:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/149.0.0.0 Safari/537.36",
		Lang:         "en-US",
		ScriptBase64: "base64-encoded sbsd script",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Payload:     %.48s...\n", solved.Solution.Payload)
	fmt.Printf("bm_lso_time: %s\n", solved.Solution.BmLsoTime)
}
