//go:build ignore

// Clear a Cloudflare five-second interstitial.
//
// Task type: `CloudFlare5STask`
//
// There is no token here. What comes back is the browser state the worker ended
// up with — headers, clearance cookies and the fingerprint it used — and
// replaying those against the site is what actually clears the challenge.
//
// A proxy is mandatory for this type: the clearance is bound to the IP that
// obtained it, so it has to be the same IP you then browse from.
//
//	EZCAPTCHA_PROXY=http://user:pass@host:port \
//	  go run examples/cloudflare/cloud_flare_5s_task.go
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

	solved, err := client.SolveCloudFlare5STask(ctx, &ezcapsolver.Cloudflare5sTask{
		WebsiteURL: "https://example.com",
		Proxy:      os.Getenv("EZCAPTCHA_PROXY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID:     %s\n", solved.TaskID)
	fmt.Printf("Fingerprint: %s\n", solved.Solution.TLSVersion)
	fmt.Printf("Cookies:     %v\n", solved.Solution.Cookies)
	fmt.Printf("Headers:     %v\n", solved.Solution.Header)
}
