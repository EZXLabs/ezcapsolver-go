//go:build ignore

// Solve several tasks at once with one shared client.
//
// One client, many goroutines: that is the intended usage. The client holds the
// connection pool, so building one per task would open a new pool each time.
//
// The SDK does not limit concurrency. Your plan does, so a semaphore is usually
// worth having.
//
//	go run examples/concurrency.go
package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/EZXLabs/ezcapsolver-go"
)

// maxInFlight caps how many tasks are open at once.
const maxInFlight = 3

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	sites := []struct{ url, key string }{
		{"https://www.google.com/recaptcha/api2/demo", "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-"},
		{"https://www.google.com/recaptcha/api2/demo", "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-"},
		{"https://www.google.com/recaptcha/api2/demo", "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-"},
	}

	semaphore := make(chan struct{}, maxInFlight)
	var group sync.WaitGroup

	for _, site := range sites {
		group.Add(1)
		go func() {
			defer group.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			solved, err := client.SolveRecaptchaV2TaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
				WebsiteURL: site.url,
				WebsiteKey: site.key,
			})
			if err != nil {
				// One task failing says nothing about the others, so this
				// reports and moves on rather than tearing the batch down.
				fmt.Printf("%-46s failed: %v\n", site.url, err)
				return
			}
			fmt.Printf("%-46s %.32s...\n", site.url, solved.Solution.Token)
		}()
	}

	group.Wait()
}
