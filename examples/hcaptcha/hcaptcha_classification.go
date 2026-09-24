//go:build ignore

// Read one or more hCaptcha images.
//
// Task type: `HCaptchaClassification`
//
// Every parameter is optional: different classification modules take different
// combinations — a single Image, a set of Images, or Anchors — and the service
// declares no validation for this type at all.
//
// The solution shape is not confirmed yet, so the SDK declares no fields for it.
// Read the answer from Extra, or from Raw.
//
//	go run examples/hcaptcha/hcaptcha_classification.go
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

	solved, err := client.SyncSolveHCaptchaClassification(ctx, &ezcapsolver.HCaptchaClassificationTask{
		Image:    base64.StdEncoding.EncodeToString(image),
		Question: "Please click each image containing a crosswalk",
	})
	if err != nil {
		log.Fatal(err)
	}

	for key, value := range solved.Solution.Extra {
		fmt.Printf("%-12s %v\n", key+":", value)
	}
	fmt.Printf("Raw:         %s\n", solved.Raw)
}
