//go:build ignore

// Report a fingerprint to DataDome and get a cookie back.
//
// Task type: `DataDomeTagsTaskProxyless`
//
// This is the other DataDome path: no challenge has been served yet. You report
// a fingerprint the way the page's own script would, and DataDome answers with a
// cookie. It runs on a different worker from the challenge flow.
//
// Ddk is the site's `window.ddjskey`, read from the inline snippet. Bpc is a
// packet counter: fixed at 1 in `ch` mode, and counting up from 2 in `le` mode.
//
//	go run examples/datadome/datadome_tags_task_proxyless.go
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

	solved, err := client.SyncSolveDataDomeTagsTaskProxyless(ctx, &ezcapsolver.DataDomeTagsTask{
		Ddk:     "the window.ddjskey value",
		JsType:  ezcapsolver.DataDomeJsTypeCh,
		Cid:     "", // Empty is valid here; the service only rejects a null.
		Bpc:     1,
		Referer: "https://example.com/",
		Ua:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) Chrome/149.0.0.0 Safari/537.36",
		Fields:  map[string]any{"tags_url": "https://example.com/tags.js"},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Kind: %s\n", solved.Solution.Kind)
	fmt.Printf("URL:  %s\n", solved.Solution.URL)
	fmt.Printf("Raw:  %s\n", solved.Raw)
}
