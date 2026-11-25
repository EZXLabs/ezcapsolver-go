//go:build ignore

// Answer a DataDome challenge step by step.
//
// Task type: `DataDomeTaskProxyless`
//
// The three steps share one request and one solution model, and all run
// through the synchronous endpoint. Between them you fetch the pages yourself:
//
//  1. Step one turns the intercepted page into the challenge address in URL.
//  2. Fetch that page and its challenge script, then ask step two whether the
//     service already caches the script, by its hash.
//  3. Step three returns the submission in URL and Body, and how long to wait
//     before sending it in SubmitAfterMs. Submitting it from the challenge
//     page brings back the cookie.
//
// Referer is what the service checks a site allow-list against, so leave it
// set even though the parameter is not formally required. With an allow-list
// configured, a referer outside it is rejected with ERROR_WEBSITE_NOT_ALLOWED.
//
//	go run examples/datadome/datadome_task_proxyless.go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/EZXLabs/ezcapsolver-go"
)

const (
	// userAgent is the browser the pages were fetched with.
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36"
	// interceptedHTMLB64 is the page DataDome intercepted, captured and
	// base64-encoded.
	interceptedHTMLB64 = "base64_encoded_intercepted_page"
	// challengeHTMLB64 is the challenge page step one points at, captured and
	// base64-encoded.
	challengeHTMLB64 = "base64_encoded_challenge_page"
	// scriptHash is the MD5 of the challenge script without its first-line
	// version banner.
	scriptHash = "md5_of_the_challenge_script"
	// scriptB64 is the challenge script itself, base64-encoded.
	scriptB64 = "base64_encoded_challenge_script"
)

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient(ezcapsolver.WithClientKey("YOUR_API_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	// Step one: find out where the challenge lives.
	first, err := client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
		Step:      ezcapsolver.DataDomeStepOne,
		HtmlB64:   interceptedHTMLB64,
		Referer:   "https://example.com/",
		UserAgent: userAgent,
	})
	if err != nil {
		log.Fatal(err)
	}
	challengeURL := first.Solution.URL
	fmt.Printf("Challenge:  %s\n", challengeURL)

	// Fetch that page and its script, then ask whether the service caches it.
	second, err := client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
		Step: ezcapsolver.DataDomeStepTwo,
		Hash: scriptHash,
	})
	if err != nil {
		log.Fatal(err)
	}
	cached := second.Solution.Status
	fmt.Printf("Cached:     %t\n", cached)

	// Step three: the submission. A cached script only needs its hash.
	task := &ezcapsolver.DataDomeTask{
		Step:      ezcapsolver.DataDomeStepThree,
		HtmlB64:   challengeHTMLB64,
		Referer:   challengeURL,
		UserAgent: userAgent,
	}
	if cached {
		task.ExternalJsHash = scriptHash
	} else {
		task.ExternalJsBase64 = scriptB64
	}
	third, err := client.SyncSolveDataDomeTaskProxyless(ctx, task)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Kind:       %s\n", third.Solution.Kind)
	fmt.Printf("Submit to:  %s\n", third.Solution.URL)
	fmt.Printf("Body:       %s\n", third.Solution.Body)
	fmt.Printf("Wait (ms):  %d\n", third.Solution.SubmitAfterMs)
}
