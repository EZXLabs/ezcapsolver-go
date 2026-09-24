//go:build ignore

// Obtain PerimeterX clearance cookies.
//
// Task type: `PerimeterX`
//
// The canonical type name is `PerimeterX`. Older clients called it `PxCaptcha`,
// which was never the service's spelling.
//
// The cookies come back as top-level fields rather than nested under a cookies
// object. `_px3` is the one that actually passes the check; send all three.
//
//	go run examples/perimeterx/perimeter_x.go
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

	solved, err := client.SolvePerimeterX(ctx, &ezcapsolver.PerimeterXTask{
		WebsiteKey: "PXxxxxxxxx",
		// Invisible mode is priced differently. The field is `invisible`, with
		// no `is` prefix, unlike the reCAPTCHA types.
		Invisible: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Task ID: %s\n", solved.TaskID)
	fmt.Printf("_px3:    %.48s...\n", solved.Solution.Px3)
	fmt.Printf("_pxvid:  %s\n", solved.Solution.PxVid)
	fmt.Printf("_pxde:   %.48s\n", solved.Solution.Pxde)
}
