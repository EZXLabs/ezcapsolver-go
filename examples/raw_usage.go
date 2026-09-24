//go:build ignore

// Drive the workflow by hand, and reach task types the SDK does not model.
//
// The convenience methods create a task and wait for it. When you want the
// polling schedule under your own control, CreateTask and GetTaskResult are the
// two halves they are built from.
//
// Solve goes one step further: it takes a bare task type string and a map keyed
// by wire names, so a type the service ships today is usable without waiting for
// an SDK release. SolveAs does the same but decodes into a struct of your own.
//
//	go run examples/raw_usage.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/EZXLabs/ezcapsolver-go"
)

const (
	pollInterval = 3 * time.Second
	maxAttempts  = 24
)

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	manualPolling(ctx, client)
	unknownTaskType(ctx, client)
}

// manualPolling creates a task and then polls it on its own schedule.
func manualPolling(ctx context.Context, client *ezcapsolver.EzCapSolverClient) {
	created, err := client.CreateTask(ctx, ezcapsolver.TaskTypeRecaptchaV2TaskProxyless,
		&ezcapsolver.RecaptchaV2Task{
			WebsiteURL: "https://www.google.com/recaptcha/api2/demo",
			WebsiteKey: "6Le-wvkSAAAAAPBMRTvw0Q4Muexq9bi0DJwx_mJ-",
		})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Created:  %s\n", created.TaskID)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		// A task that was just enqueued is almost always still processing, so
		// wait before the first query rather than after it.
		time.Sleep(pollInterval)

		result, err := client.GetTaskResult(ctx, created.TaskID)
		if err != nil {
			log.Fatal(err)
		}
		if result.IsReady() {
			fmt.Printf("Solution: %s\n", result.Solution)
			return
		}
		fmt.Printf("Status:   %s (%d/%d)\n", result.Status, attempt, maxAttempts)
	}

	fmt.Printf("Gave up after %d attempts; the task was still billed.\n", maxAttempts)
}

// unknownTaskType solves a task type this release has no model for.
func unknownTaskType(ctx context.Context, client *ezcapsolver.EzCapSolverClient) {
	// Map keys are wire names and are sent verbatim; the result comes back
	// undecoded.
	solved, err := client.Solve(ctx, "BrandNewTaskType", map[string]any{
		"websiteURL":   "https://example.com",
		"someNewField": 42,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Task ID:  %s\n", solved.TaskID)
	fmt.Printf("Raw:      %s\n", solved.Raw)

	// A synchronous type has to say so — the SDK cannot infer that from a
	// string it has never seen.
	result, err := client.CreateSyncTask(ctx, "BrandNewSyncType", map[string]any{"input": "..."})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Raw:      %s\n", result.Solution)

	// Or decode straight into a type of your own.
	type myShape struct {
		Token string `json:"token"`
	}
	typed, err := ezcapsolver.SolveAs[myShape](ctx, client, "BrandNewTaskType", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Token:    %s\n", typed.Solution.Token)

	// Types the SDK does know are constants, so a known type never has to be
	// spelled out as a string.
	fmt.Printf("Known:    %s\n", ezcapsolver.TaskTypeRecaptchaV2TaskProxyless)
}
