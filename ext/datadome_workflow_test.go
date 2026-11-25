package ext_test

// The DataDome workflow end to end, against one local server standing in for
// the protected site, DataDome, and the EzCaptchaSolver API. The browser side
// runs on the real lkrequest library.

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/EZXLabs/ezcapsolver-go"
	"github.com/EZXLabs/ezcapsolver-go/ext"
)

const (
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
		"(KHTML, like Gecko) Chrome/153.0.0.0 Safari/537.36"
	interceptedPage    = "<html><script>var dd={'rt':'c','cid':'abc'}</script></html>"
	sliderScript       = "/** DataDome v1.34.0 96eb15b */\n!function A(){}();\nvar g=1;"
	interstitialScript = "/* DataDome v1.33.1 d5110df */\n;(function(){var A={};})()"
	rawCookie          = "datadome=abc123; Max-Age=31536000; Domain=.example.com; Path=/; Secure; SameSite=Lax"
	cookieVerdict      = `{"view":"redirect","cookie":"` + rawCookie + `"}`
	formContentType    = "application/x-www-form-urlencoded; charset=UTF-8"
)

// wantCookie is the cookie every passing run returns.
var wantCookie = ext.DataDomeCookie{
	Name: "datadome", Value: "abc123", Domain: ".example.com", Raw: rawCookie,
}

// seenRequest is what the server recorded of one browser-side request.
type seenRequest struct {
	header http.Header
	body   string
}

// world is the local server, and everything it was sent.
type world struct {
	t      *testing.T
	server *httptest.Server
	mux    *http.ServeMux

	mu    sync.Mutex
	seen  map[string][]seenRequest
	tasks []map[string]any
}

// newWorld starts an empty server.
func newWorld(t *testing.T) *world {
	t.Helper()
	w := &world{t: t, mux: http.NewServeMux(), seen: map[string][]seenRequest{}}
	w.server = httptest.NewServer(w.mux)
	t.Cleanup(w.server.Close)
	return w
}

// page serves a fixed body on a route pattern such as `GET /signup`, and
// records every request to it.
func (w *world) page(pattern string, status int, body string) {
	w.mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
		sent, _ := io.ReadAll(request.Body)
		w.mu.Lock()
		w.seen[pattern] = append(w.seen[pattern], seenRequest{request.Header.Clone(), string(sent)})
		w.mu.Unlock()
		writer.WriteHeader(status)
		_, _ = io.WriteString(writer, body)
	})
}

// api answers each DataDome step with the solution given for it, and records
// every task.
func (w *world) api(solutions map[string]string) {
	w.mux.HandleFunc("POST /createSyncTask", func(writer http.ResponseWriter, request *http.Request) {
		var envelope struct {
			Task map[string]any `json:"task"`
		}
		if err := json.NewDecoder(request.Body).Decode(&envelope); err != nil {
			w.t.Errorf("the API was sent something that is not JSON: %v", err)
		}
		w.mu.Lock()
		w.tasks = append(w.tasks, envelope.Task)
		w.mu.Unlock()

		step, _ := envelope.Task["step"].(string)
		solution, known := solutions[step]
		if !known {
			w.t.Errorf("no answer prepared for step %q", step)
			solution = "{}"
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(writer,
			`{"errorId":0,"taskId":"task-`+step+`","status":"ready","solution":`+solution+`}`)
	})
}

// requests returns what a route was sent.
func (w *world) requests(pattern string) []seenRequest {
	w.mu.Lock()
	defer w.mu.Unlock()
	return slices.Clone(w.seen[pattern])
}

// sentTasks returns every task the API received, in order.
func (w *world) sentTasks() []map[string]any {
	w.mu.Lock()
	defer w.mu.Unlock()
	return slices.Clone(w.tasks)
}

// client is an API client pointed at the server.
func (w *world) client() *ezcapsolver.EzCapSolverClient {
	w.t.Helper()
	client, err := ezcapsolver.NewClient(
		ezcapsolver.WithClientKey("test-client-key"),
		ezcapsolver.WithBaseURLs(w.server.URL, w.server.URL),
	)
	if err != nil {
		w.t.Fatalf("building the client failed: %v", err)
	}
	return client
}

// challenge is a run against the server's /signup page.
func (w *world) challenge() *ext.DataDomeChallenge {
	return &ext.DataDomeChallenge{TargetURL: w.server.URL + "/signup", UserAgent: userAgent}
}

// solve runs the workflow against the server.
func (w *world) solve(ctx context.Context) (*ext.DataDomeCookie, error) {
	return ext.SolveDataDomeWorkflow(ctx, w.client(), w.challenge())
}

// encode is the base64 the service expects.
func encode(text string) string {
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// wantFields fails the test when a task lacks one of the fields, or carries one
// it must not.
func wantFields(t *testing.T, task map[string]any, want map[string]any, absent ...string) {
	t.Helper()
	for name, value := range want {
		if task[name] != value {
			t.Errorf("step %v: %s = %v, want %v", task["step"], name, task[name], value)
		}
	}
	for _, name := range absent {
		if _, present := task[name]; present {
			t.Errorf("step %v must not send %s", task["step"], name)
		}
	}
}

// wantOneRequest fails the test unless a route was hit exactly once, and
// returns that request.
func wantOneRequest(t *testing.T, w *world, pattern string) seenRequest {
	t.Helper()
	requests := w.requests(pattern)
	if len(requests) != 1 {
		t.Fatalf("%s was requested %d times, want once", pattern, len(requests))
	}
	return requests[0]
}

func TestASliderRunSendsAnUncachedScriptInFull(t *testing.T) {
	w := newWorld(t)
	base := w.server.URL
	challengePage := "<html><script>var ddm={};\n" + sliderScript + "</script></html>"
	w.page("GET /signup", http.StatusForbidden, interceptedPage)
	w.page("GET /captcha/{$}", http.StatusOK, challengePage)
	w.page("GET /captcha/check", http.StatusOK, cookieVerdict)
	w.api(map[string]string{
		"1": `{"url":"` + base + `/captcha/?initialCid=abc"}`,
		"2": `{"status":false}`,
		"3": `{"url":"` + base + `/captcha/check?cid=abc","kind":"slider","submit_after_ms":1}`,
	})

	cookie, err := w.solve(t.Context())
	if err != nil {
		t.Fatalf("the run failed: %v", err)
	}
	if cookie == nil || *cookie != wantCookie {
		t.Errorf("cookie = %+v, want %+v", cookie, wantCookie)
	}

	target := wantOneRequest(t, w, "GET /signup")
	if got := target.header.Values("User-Agent"); !slices.Equal(got, []string{userAgent}) {
		t.Errorf("the target page got User-Agent %q, want exactly the challenge's", got)
	}
	frame := wantOneRequest(t, w, "GET /captcha/{$}")
	if got := frame.header.Get("Referer"); got != base+"/signup" {
		t.Errorf("the slider frame got referer %q, want the full page address", got)
	}
	wantOneRequest(t, w, "GET /captcha/check")

	tasks := w.sentTasks()
	if len(tasks) != 3 {
		t.Fatalf("the API got %d tasks, want 3", len(tasks))
	}
	wantFields(t, tasks[0], map[string]any{
		"type": "DataDomeTaskProxyless", "step": "1", "html_b64": encode(interceptedPage),
		"referer": base + "/signup", "user_agent": userAgent,
	})
	wantFields(t, tasks[1], map[string]any{"step": "2"}, "html_b64")
	wantFields(t, tasks[2], map[string]any{
		"step": "3", "html_b64": encode(challengePage),
		"referer": base + "/captcha/?initialCid=abc", "user_agent": userAgent,
		"external_js_base64": encode(sliderScript),
	}, "external_js_hash")
}

func TestAnInterstitialRunSendsOnlyTheHashOfACachedScript(t *testing.T) {
	w := newWorld(t)
	base := w.server.URL
	challengePage := "<html><script>var ddm={};\n" + interstitialScript + ";\nvar glue='cid=abc';</script></html>"
	w.page("GET /signup", http.StatusForbidden, interceptedPage)
	w.page("GET /interstitial/{$}", http.StatusOK, challengePage)
	w.page("POST /interstitial/check", http.StatusOK, cookieVerdict)
	w.api(map[string]string{
		"1": `{"url":"` + base + `/interstitial/?initialCid=abc"}`,
		"2": `{"status":true}`,
		"3": `{"url":"` + base + `/interstitial/check","body":"payload=1"}`,
	})

	cookie, err := w.solve(t.Context())
	if err != nil {
		t.Fatalf("the run failed: %v", err)
	}
	if cookie == nil || *cookie != wantCookie {
		t.Errorf("cookie = %+v, want %+v", cookie, wantCookie)
	}

	frame := wantOneRequest(t, w, "GET /interstitial/{$}")
	if got := frame.header.Get("Referer"); got != base {
		t.Errorf("the interstitial frame got referer %q, want only the page origin", got)
	}
	submitted := wantOneRequest(t, w, "POST /interstitial/check")
	if got := submitted.header.Get("Content-Type"); got != formContentType {
		t.Errorf("the submission content type = %q, want %q", got, formContentType)
	}
	if submitted.body != "payload=1" {
		t.Errorf("the submission body = %q, want the step 3 body", submitted.body)
	}

	// The bundle ends at its closing `})()`, and the banner stays out of the hash.
	digest := md5.Sum([]byte(";(function(){var A={};})()"))
	hash := hex.EncodeToString(digest[:])
	tasks := w.sentTasks()
	if len(tasks) != 3 {
		t.Fatalf("the API got %d tasks, want 3", len(tasks))
	}
	wantFields(t, tasks[1], map[string]any{"step": "2", "hash": hash})
	wantFields(t, tasks[2], map[string]any{"step": "3", "external_js_hash": hash}, "external_js_base64")
}

func TestAPageDataDomeLetsThroughCostsNoTask(t *testing.T) {
	w := newWorld(t)
	w.page("GET /signup", http.StatusOK, "<html>welcome</html>")
	w.api(nil)

	cookie, err := w.solve(t.Context())
	if cookie != nil || err != nil {
		t.Errorf("got %+v, %v; want no cookie and no error", cookie, err)
	}
	if tasks := w.sentTasks(); len(tasks) != 0 {
		t.Errorf("the API got %d tasks, want none", len(tasks))
	}
}

func TestABlockPageIsReportedAsABan(t *testing.T) {
	w := newWorld(t)
	w.page("GET /signup", http.StatusForbidden, "<html><script>var dd={'t':'bv'}</script>hardblock</html>")
	w.api(nil)

	if _, err := w.solve(t.Context()); !errors.Is(err, ext.ErrDataDomeBlocked) {
		t.Errorf("err = %v, want ErrDataDomeBlocked", err)
	}
	if tasks := w.sentTasks(); len(tasks) != 0 {
		t.Errorf("the API got %d tasks, want none", len(tasks))
	}
}

func TestCancelingTheContextEndsTheWaitBeforeSubmitting(t *testing.T) {
	w := newWorld(t)
	base := w.server.URL
	w.page("GET /signup", http.StatusForbidden, interceptedPage)
	w.page("GET /captcha/{$}", http.StatusOK, "<script>"+sliderScript+"</script>")
	w.page("GET /captcha/check", http.StatusOK, cookieVerdict)
	w.api(map[string]string{
		"1": `{"url":"` + base + `/captcha/"}`,
		"2": `{"status":true}`,
		"3": `{"url":"` + base + `/captcha/check","kind":"slider","submit_after_ms":60000}`,
	})

	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()
	if _, err := w.solve(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("err = %v, want the context deadline", err)
	}
	if submitted := w.requests("GET /captcha/check"); len(submitted) != 0 {
		t.Error("the answer must not be submitted once the context ended")
	}
}

func TestAnAPIErrorComesBackAsTheClientReturnedIt(t *testing.T) {
	w := newWorld(t)
	w.page("GET /signup", http.StatusForbidden, interceptedPage)
	w.mux.HandleFunc("POST /createSyncTask", func(writer http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(writer,
			`{"errorId":1,"errorCode":"ERROR_ZERO_BALANCE","errorDescription":"no balance"}`)
	})

	_, err := w.solve(t.Context())
	var apiErr *ezcapsolver.APIError
	if !errors.As(err, &apiErr) || apiErr.ErrorCode != "ERROR_ZERO_BALANCE" {
		t.Errorf("err = %v, want the API error", err)
	}
}

func TestAnUnusableChallengeFailsBeforeAnyRequest(t *testing.T) {
	w := newWorld(t)
	for name, challenge := range map[string]*ext.DataDomeChallenge{
		"nil":              nil,
		"relative target":  {TargetURL: "/signup", UserAgent: userAgent},
		"not Chrome":       {TargetURL: w.server.URL, UserAgent: "Mozilla/5.0 Firefox/133.0"},
		"no Chrome preset": {TargetURL: w.server.URL, UserAgent: "Chrome/120.0.0.0"},
		"bad proxy":        {TargetURL: w.server.URL, UserAgent: userAgent, Proxy: "::not a proxy"},
	} {
		if _, err := ext.SolveDataDomeWorkflow(t.Context(), w.client(), challenge); !errors.Is(err, ext.ErrDataDome) {
			t.Errorf("%s: err = %v, want ErrDataDome", name, err)
		}
	}
}
