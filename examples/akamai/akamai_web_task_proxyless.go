//go:build ignore

// Produce Akamai Web sensor data.
//
// Task type: `AkamaiWEBTaskProxyless`
//
// This is a multi-round flow, which is what the loop below is for: each round
// returns Encodedata, and the next round sends it back as EncodeData with the
// index raised. Note the casing difference between the two — the service spells
// the field differently in each direction, and getting it wrong silently
// restarts the flow.
//
// The round that ends the flow returns no Encodedata.
//
//	go run examples/akamai/akamai_web_task_proxyless.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/EZXLabs/ezcapsolver-go"
)

const rounds = 3

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	task := &ezcapsolver.AkamaiWebTask{
		PageURL: "https://example.com",
		// Most sites change this URL on every request, so read it from the page
		// rather than hard-coding it.
		V3URL:        "https://example.com/v3/...",
		Ua:           "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/149.0.0.0 Safari/537.36",
		Lang:         "en-US",
		Abck:         "the _abck cookie value",
		Bmsz:         "the bm_sz cookie value",
		ScriptBase64: "base64-encoded akamai script",
	}

	for round := range rounds {
		task.Index = round

		solved, err := client.SyncSolveAkamaiWEBTaskProxyless(ctx, task)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Round %d payload: %.48s...\n", round, solved.Solution.Payload)

		// POST the payload to the site here, then feed the state forward.
		if solved.Solution.Encodedata == "" {
			fmt.Println("The flow is complete.")
			break
		}
		task.EncodeData = solved.Solution.Encodedata
	}
}
