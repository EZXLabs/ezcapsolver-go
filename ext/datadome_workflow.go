package ext

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/EZXLabs/ezcapsolver-go"
	"github.com/EZXLabs/lkrequest-go/lkrequest"
)

// browserRequestTimeout bounds each browser-side request.
const browserRequestTimeout = 30 * time.Second

// SolveDataDomeWorkflow clears the DataDome challenge in front of a page and
// returns the cookie DataDome issues for it.
//
// Only the page address and a browser identity are needed; the whole flow runs
// inside this call:
//
//  1. Fetch the target page. A page DataDome did not intercept ends the run
//     with a nil cookie and a nil error.
//  2. Step one finds the challenge address, and the challenge page is fetched.
//  3. The challenge script is read from that page, or fetched from DataDome's
//     CDN when the page links it.
//  4. Step two asks whether the service already caches the script.
//  5. Step three computes the answer, sending the whole script only when it is
//     not cached.
//  6. The answer is submitted after the delay it asks for, and the cookie in
//     DataDome's response is returned.
//
// Steps one to three are `DataDomeTaskProxyless` tasks on the synchronous
// endpoint, so a challenged page costs three tasks. Browser-side requests go
// through [DataDomeChallenge.Proxy] with the Chrome TLS fingerprint of
// [DataDomeChallenge.UserAgent].
//
// Errors match [ErrDataDomeBlocked] when DataDome banned the IP, [ErrDataDome]
// when the run cannot go on, and [ErrBrowserRequest] when a request to the
// protected site or to DataDome fails. An API call fails with the error the
// client returned, as from
// [ezcapsolver.EzCapSolverClient.SyncSolveDataDomeTaskProxyless]. Canceling
// ctx ends the run.
func SolveDataDomeWorkflow(
	ctx context.Context, client *ezcapsolver.EzCapSolverClient, challenge *DataDomeChallenge,
) (*DataDomeCookie, error) {
	if client == nil || challenge == nil {
		return nil, fmt.Errorf("%w: client and challenge are required", ErrDataDome)
	}
	target, err := parseURL(challenge.TargetURL, "TargetURL")
	if err != nil {
		return nil, err
	}
	browser, err := openBrowser(challenge)
	if err != nil {
		return nil, err
	}
	defer browser.close()

	page, err := browser.get(ctx, target.String(), browser.profile.navigation())
	if err != nil {
		return nil, err
	}
	if isHardBlocked(page) {
		return nil, blocked("the target page")
	}
	if !isChallenged(page) {
		return nil, nil
	}

	run := dataDomeRun{client: client, browser: browser, target: target, userAgent: challenge.UserAgent}
	cookie, err := run.clear(ctx, page)
	if err != nil {
		return nil, err
	}
	return &cookie, nil
}

// dataDomeRun is one run against an intercepted page: the API client, the
// browser, and the page the run started from.
type dataDomeRun struct {
	client    *ezcapsolver.EzCapSolverClient
	browser   *browser
	target    *url.URL
	userAgent string
}

// clear takes an intercepted page through steps one to three and the
// submission, and returns the cookie DataDome issues.
func (r dataDomeRun) clear(ctx context.Context, page string) (DataDomeCookie, error) {
	challengeURL, err := r.locateChallenge(ctx, page)
	if err != nil {
		return DataDomeCookie{}, err
	}
	frameHeaders := r.browser.profile.challengeFrame(r.target, isSlider(challengeURL))
	challengePage, err := r.browser.get(ctx, challengeURL.String(), frameHeaders)
	if err != nil {
		return DataDomeCookie{}, err
	}
	if isHardBlocked(challengePage) {
		return DataDomeCookie{}, blocked("the challenge page")
	}
	loadedAt := time.Now()

	script, err := r.browser.challengeScript(ctx, challengePage, challengeURL)
	if err != nil {
		return DataDomeCookie{}, err
	}
	answer, err := r.answer(ctx, challengePage, challengeURL, script)
	if err != nil {
		return DataDomeCookie{}, err
	}
	submission, err := submissionFrom(answer)
	if err != nil {
		return DataDomeCookie{}, err
	}
	if err := sleep(ctx, submitDelay(answer, submission)-time.Since(loadedAt)); err != nil {
		return DataDomeCookie{}, err
	}
	verdict, err := r.browser.submit(ctx, submission, challengeURL)
	if err != nil {
		return DataDomeCookie{}, err
	}
	return readVerdict(verdict, r.target)
}

// locateChallenge is step one: it asks the service where the challenge of an
// intercepted page lives.
func (r dataDomeRun) locateChallenge(ctx context.Context, page string) (*url.URL, error) {
	solved, err := r.client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
		Step:      ezcapsolver.DataDomeStepOne,
		HtmlB64:   encodeText(page),
		Referer:   r.target.String(),
		UserAgent: r.userAgent,
	})
	if err != nil {
		return nil, err
	}
	if solved.Solution.URL == "" {
		return nil, fmt.Errorf("%w: step 1 returned no challenge url", ErrDataDome)
	}
	return parseURL(solved.Solution.URL, "the step 1 challenge url")
}

// answer runs steps two and three: it asks whether the service caches the
// script, then for the answer, sending the script itself only when it is not
// cached.
func (r dataDomeRun) answer(
	ctx context.Context, challengePage string, challengeURL *url.URL, script string,
) (ezcapsolver.DataDomeSolution, error) {
	hash := scriptHash(script)
	cached, err := r.scriptCached(ctx, hash)
	if err != nil {
		return ezcapsolver.DataDomeSolution{}, err
	}
	task := &ezcapsolver.DataDomeTask{
		Step:      ezcapsolver.DataDomeStepThree,
		HtmlB64:   encodeText(challengePage),
		Referer:   challengeURL.String(),
		UserAgent: r.userAgent,
	}
	if cached {
		task.ExternalJsHash = hash
	} else {
		task.ExternalJsBase64 = encodeText(script)
	}
	solved, err := r.client.SyncSolveDataDomeTaskProxyless(ctx, task)
	if err != nil {
		return ezcapsolver.DataDomeSolution{}, err
	}
	return solved.Solution, nil
}

// scriptCached is step two: whether the service already caches the script with
// this hash.
//
// An answer without a status counts as not cached, which only means the script
// is sent in full.
func (r dataDomeRun) scriptCached(ctx context.Context, hash string) (bool, error) {
	solved, err := r.client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
		Step: ezcapsolver.DataDomeStepTwo,
		Hash: hash,
	})
	if err != nil {
		return false, err
	}
	return solved.Solution.Status, nil
}

// browser is the browser side of a run: one cookie-keeping session with a
// Chrome TLS fingerprint, and the identity its requests present.
type browser struct {
	client  *lkrequest.Client
	session *lkrequest.Session
	profile browserProfile
}

// openBrowser opens the session a run goes through. The caller closes it.
func openBrowser(challenge *DataDomeChallenge) (*browser, error) {
	profile, err := newBrowserProfile(challenge.UserAgent, challenge.acceptLanguage())
	if err != nil {
		return nil, err
	}
	preset, err := profile.tlsPreset()
	if err != nil {
		return nil, err
	}
	client, err := lkrequest.NewClientBuilder().
		SetPreset(preset).
		SetTimeoutTotal(uint64(browserRequestTimeout.Milliseconds())).
		Build()
	if err != nil {
		return nil, fmt.Errorf("%w: building the %s browser: %w", ErrDataDome, preset, err)
	}
	builder := lkrequest.NewSessionBuilder(client)
	if challenge.Proxy != "" {
		builder = builder.SetProxy(challenge.Proxy)
	}
	session, err := builder.Build()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("%w: Proxy is not usable: %w", ErrDataDome, err)
	}
	return &browser{client: client, session: session, profile: profile}, nil
}

// close releases the session and the client behind it.
func (b *browser) close() {
	b.session.Close()
	b.client.Close()
}

// get sends a GET and returns the response body.
//
// The status is not checked: DataDome answers an intercepted page with an
// error status, and the body is what tells the cases apart.
func (b *browser) get(ctx context.Context, address string, headers []header) (string, error) {
	return send(ctx, b.session.Get(address), headers, "GET "+address)
}

// post sends a form POST and returns the response body.
func (b *browser) post(ctx context.Context, address string, headers []header, body string) (string, error) {
	request := b.session.Post(address).SetBodyBytes([]byte(body))
	return send(ctx, request, headers, "POST "+address)
}

// challengeScript is the challenge script: fetched from the CDN when the page
// links it, read from the page otherwise.
func (b *browser) challengeScript(ctx context.Context, page string, challengeURL *url.URL) (string, error) {
	src := externalScriptURL(page)
	if src == "" {
		script := inlineScript(page)
		if script == "" {
			return "", fmt.Errorf("%w: the challenge page carries no challenge script", ErrDataDome)
		}
		return script, nil
	}
	script, err := b.get(ctx, src, b.profile.script(challengeURL))
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(script) == "" {
		return "", fmt.Errorf("%w: the external challenge script is empty", ErrDataDome)
	}
	return script, nil
}

// submit hands the answer to DataDome and returns its response body.
func (b *browser) submit(ctx context.Context, s submission, challengeURL *url.URL) (string, error) {
	headers := b.profile.submission(challengeURL, s.isPost())
	if s.isPost() {
		return b.post(ctx, s.url, headers, s.body)
	}
	return b.get(ctx, s.url, headers)
}

// send adds the headers in order, sends the request and returns its body.
//
// operation names the request in an error. Its query is cut off: it carries
// session identifiers and can run long.
func send(ctx context.Context, request *lkrequest.Request, headers []header, operation string) (string, error) {
	for _, h := range headers {
		request = request.AddHeader(h.name, h.value)
	}
	response, err := request.SendWithContext(ctx)
	if err != nil {
		operation, _, _ = strings.Cut(operation, "?")
		return "", fmt.Errorf("%w: %s: %w", ErrBrowserRequest, operation, err)
	}
	defer response.Close()
	return response.String(), nil
}

// sleep waits for the delay, or until ctx ends.
func sleep(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// parseURL parses an address the run has to request, naming it in the error.
func parseURL(address, what string) (*url.URL, error) {
	parsed, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrDataDome, what, err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, fmt.Errorf("%w: %s %q needs a scheme and a host", ErrDataDome, what, address)
	}
	return parsed, nil
}

// blocked is the error for a page DataDome serves to a banned visitor.
func blocked(what string) error {
	return fmt.Errorf("%w: %s is a block page", ErrDataDomeBlocked, what)
}
