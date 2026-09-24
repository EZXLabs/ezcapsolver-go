//go:build ignore

// Read a reCAPTCHA v2 image grid.
//
// Task type: `ReCaptchaV2Classification`
//
// This is a synchronous type: the answer comes back from the request that
// created the task, with no polling and no task ID.
//
// The answer exposes Type, HasObject, Objects and Extra directly. IsMulti and
// IsSingle identify the result kind; unknown kinds keep their type and fields.
//
//	go run examples/recaptcha_v2/recaptcha_v2_classification.go
package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	ctx := context.Background()

	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	image, err := os.ReadFile(filepath.Join("examples", "fixtures", "crosswalks3x3.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	solved, err := client.SyncSolveRecaptchaV2Classification(ctx, &ezcapsolver.RecaptchaV2ClassificationTask{
		Image:    base64.StdEncoding.EncodeToString(image),
		Question: "/m/014xcs", // crosswalk
		Size:     3,
	})
	if err != nil {
		log.Fatal(err)
	}

	solution := solved.Solution
	switch {
	case solution.IsMulti():
		fmt.Printf("Click cells: %v\n", solution.Objects)
	case solution.IsSingle():
		fmt.Printf("Contains it: %t\n", solution.HasObject)
	default:
		fmt.Printf("Type: %s\n", solution.Type)
		fmt.Printf("Raw: %s\n", solved.Raw)
	}
}
