// Package ext extends the EzCaptchaSolver task API with workflows: functions
// that run the whole flow around a task type in one call — fetching the
// protected pages, chaining the steps of the task, and submitting the answer.
//
// Each workflow is a function named Solve<X>Workflow that takes the
// [ezcapsolver.EzCapSolverClient] its API calls go through:
//
//	cookie, err := ext.SolveDataDomeWorkflow(ctx, client, &ext.DataDomeChallenge{
//		TargetURL: "https://example.com/signup",
//		UserAgent: userAgent,
//		Proxy:     "http://user:pass@host:port",
//	})
//
// # Platforms
//
// The browser side of a workflow runs on lkrequest, an HTTP client with
// browser TLS fingerprints. It is a native library, embedded for darwin/arm64,
// linux/amd64 and windows/amd64; on other platforms this package does not
// build. Importing the package loads the library when the program starts, and
// a failure to load it panics. The ezcapsolver package does not depend on it,
// so a program that does not import ext neither builds nor loads it.
package ext
