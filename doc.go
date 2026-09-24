// Package ezcapsolver is the Go SDK for EzCaptchaSolver.
//
// One [EzCapSolverClient] covers every operation. Every task type has two
// convenience methods that take its request model and return its solution
// model: SolveX creates the task and polls for the result, SyncSolveX runs it
// through the synchronous endpoint, which answers on the creating request.
// The name says which endpoint is used, so nothing has to be looked up.
//
// The same pairing covers types this release does not model:
// [EzCapSolverClient.Solve] / [EzCapSolverClient.SyncSolve] take a type name and a
// parameter map, and [SolveAs] / [SyncSolveAs] decode into a caller's struct.
//
//	client, err := ezcapsolver.NewClient()
//	if err != nil {
//		return err
//	}
//	solved, err := client.SolveRecaptchaV2TaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
//		WebsiteURL: "https://example.com",
//		WebsiteKey: "6Lc...",
//	})
//	if err != nil {
//		return err
//	}
//	fmt.Println(solved.Solution.Token)
//
// The client key is read from the EZCAPTCHA_API_KEY environment variable unless
// [WithClientKey] supplies one.
//
// # Errors
//
// Failures are classified into five layers. Use errors.Is against [ErrConfig],
// [ErrTransport], [ErrAPI], [ErrPollingExhausted] or [ErrDecode] to tell them
// apart, and errors.As with [*APIError] to read the service's error code,
// description, HTTP status and field-level validation errors.
//
// A failure raised while waiting for a task that was already created is wrapped
// in a [*WaitInterruptedError], which carries the task identifier out. Solve
// creates the task internally, so that wrapper is the only place the identifier
// appears — recover by waiting on the same task rather than paying for a second
// one. The wrapper keeps the layer classification intact, so errors.Is answers
// exactly as it would without it.
//
// This SDK does not retry a request for the caller. Creating a task is billed
// and is not idempotent, and the service temporarily bans a key that repeats
// certain credential errors, so the retry policy belongs to the caller.
// [APIError.IsTerminal], [APIError.IsAuthenticationError] and
// [APIError.IsRateLimited] provide the facts needed to decide.
//
// The one exception is a throttled poll. ERROR_REQUEST_LIMIT and
// ERROR_REQUEST_BANNED refuse the query rather than the task: the service turns
// the request away before it ever looks the task up, so the task is still queued
// and still billed. [EzCapSolverClient.WaitForResult] spends the attempt and polls
// again instead of discarding a result that was about to arrive. Every other API
// error is the poll's answer and ends the wait.
//
// # Lossless results
//
// Every solution model carries an Extra map holding the worker fields this
// release does not declare, and [Solved.Raw] keeps the untouched JSON. A worker
// that starts returning a new field never loses it, with or without an SDK
// upgrade.
package ezcapsolver
