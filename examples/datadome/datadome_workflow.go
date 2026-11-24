//go:build ignore

// Clear a DataDome challenge in one call and get the cookie it issues.
//
// Task type: `DataDomeTaskProxyless`
//
// Steps one to three run inside the call. It fetches the page, locates and
// loads the challenge, reads its script, asks the service for the answer, and
// submits it. Browser-side requests go through the proxy with the Chrome TLS
// fingerprint the user agent names, so keep using both for the requests that
// carry the cookie.
//
// The ext package runs on darwin/arm64, linux/amd64 and windows/amd64.
//
//	EZCAPTCHA_PROXY=http://user:pass@host:port \
//	  go run examples/datadome/datadome_workflow.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/EZXLabs/ezcapsolver-go"
	"github.com/EZXLabs/ezcapsolver-go/ext"
)

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient(ezcapsolver.WithClientKey("YOUR_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	cookie, err := ext.SolveDataDomeWorkflow(ctx, client, &ext.DataDomeChallenge{
		TargetURL: "https://example.com/signup",
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
			"(KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36",
		Proxy: os.Getenv("EZCAPTCHA_PROXY"),
	})
	if err != nil {
		log.Fatal(err)
	}
	if cookie == nil {
		fmt.Println("DataDome did not intercept the page")
		return
	}
	fmt.Printf("Cookie:  %s=%s\n", cookie.Name, cookie.Value)
	fmt.Printf("Domain:  %s\n", cookie.Domain)
}
