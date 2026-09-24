//go:build ignore

// Read a FunCaptcha image.
//
// Task type: `FunCaptchaClassification`
//
// This is a synchronous type, and its solution shape is not confirmed yet: no
// reliable sample exists, so the SDK declares no fields for it rather than
// guessing. Everything the worker returns arrives in Extra, and the untouched
// JSON is in Raw.
//
// Once the shape is confirmed, fields get promoted out of Extra one at a time —
// and code reading them from Extra keeps working in the meantime.
//
//	go run examples/funcaptcha/funcaptcha_classification.go
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

	image, err := os.ReadFile(filepath.Join("examples", "fixtures", "crosswalks1x1.jpg"))
	if err != nil {
		log.Fatal(err)
	}

	solved, err := client.SyncSolveFunCaptchaClassification(ctx, &ezcapsolver.FunCaptchaClassificationTask{
		Image:    base64.StdEncoding.EncodeToString(image),
		Question: "Pick the image that is the right way up",
	})
	if err != nil {
		log.Fatal(err)
	}

	for key, value := range solved.Solution.Extra {
		fmt.Printf("%-12s %v\n", key+":", value)
	}
	fmt.Printf("Raw:         %s\n", solved.Raw)
}
