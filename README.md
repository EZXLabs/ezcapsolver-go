<div align="center">
  <img src="./assets/ez-captcha-logo.svg" alt="EZCaptchaSolver by EZXLabs" height="88">
  &nbsp;&nbsp;
  <img src="./assets/golang.svg" alt="Go" height="88">
  <h1>EZCaptchaSolver Go SDK</h1>
  <p>
    <a href="https://github.com/EZXLabs/ezcapsolver-go/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/EZXLabs/ezcapsolver-go/actions/workflows/ci.yml/badge.svg"></a>
    <a href="https://pkg.go.dev/github.com/EZXLabs/ezcapsolver-go"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/EZXLabs/ezcapsolver-go.svg"></a>
    <a href="./LICENSE"><img alt="License: Apache-2.0" src="https://img.shields.io/badge/license-Apache--2.0-blue.svg"></a>
    <a href="https://go.dev"><img alt="Go 1.26+" src="https://img.shields.io/badge/go-1.26%2B-00ADD8.svg?logo=go&logoColor=white"></a>
    <a href="https://ezxlabs.com"><img alt="EZXLabs website" src="https://img.shields.io/badge/website-ezxlabs.com-FFDB29?logoColor=black"></a>
  </p>
  <p>
    <a href="https://ezxlabs.com">🌐 Official website</a> &nbsp;·&nbsp;
    <a href="https://docs.ezxlabs.com/docs/captcha/api">📚 EZCaptchaSolver API reference</a> &nbsp;·&nbsp;
    <a href="./examples">🧪 Examples</a> &nbsp;·&nbsp;
    <a href="#-supported-captcha-types">🧩 Captcha Types</a>
  </p>
  <p><b>English</b> &nbsp;·&nbsp; <a href="./README.zh-CN.md">简体中文</a></p>
</div>

---

The EZCaptchaSolver Go SDK is an open-source Go client maintained by [EZXLabs](https://ezxlabs.com) for its CAPTCHA recognition task API. It provides typed requests and a context-aware client for the supported task types below; the module also includes a TLS forwarding task that does not solve CAPTCHAs. For the wider SDK family, see the [EZCaptchaSolver SDK product page](https://ezxlabs.com/products/sdk); for HTTP request and response fields, see the [EZCaptchaSolver API reference](https://docs.ezxlabs.com/docs/captcha/api); for Go usage, see the [examples in this repository](./examples/README.md).

## 🧩 Supported Captcha Types

Captcha task types come in a synchronous and an asynchronous form:

- Synchronous: the request blocks after the task is created and returns once the task is done.
- Asynchronous: creating the task returns a task ID, and the result is fetched later by polling that ID. This suits captcha types that take a while to solve.

A captcha type can support both forms at once, and almost every type supports the synchronous one. A few types are asynchronous only.

### reCAPTCHA v2

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `ReCaptchaV2TaskProxyless` | all | [Example](#task-recaptchav2taskproxyless) | reCAPTCHA v2 |
| `ReCaptchaV2TaskProxylessS9` | all | [Example](#task-recaptchav2taskproxylesss9) | reCAPTCHA v2, returns a token scored ≥ 0.9 |
| `ReCaptchaV2STaskProxyless` | all | [Example](#task-recaptchav2staskproxyless) | reCAPTCHA v2 carrying the challenge-bound `s` parameter |
| `ReCaptchaV2EnterpriseTaskProxyless` | all | [Example](#task-recaptchav2enterprisetaskproxyless) | reCAPTCHA v2 Enterprise |
| `ReCaptchaV2SEnterpriseTaskProxyless` | all | [Example](#task-recaptchav2senterprisetaskproxyless) | reCAPTCHA v2 Enterprise, carrying the `s` parameter |
| `ReCaptchaV2Classification` | sync | [Example](#task-recaptchav2classification) | reCAPTCHA v2 image recognition |

### reCAPTCHA v3

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `ReCaptchaV3TaskProxyless` | all | [Example](#task-recaptchav3taskproxyless) | reCAPTCHA v3 |
| `ReCaptchaV3TaskProxylessS9` | all | [Example](#task-recaptchav3taskproxylesss9) | reCAPTCHA v3, returns a token scored ≥ 0.9 |
| `ReCaptchaV3EnterpriseTaskProxyless` | all | [Example](#task-recaptchav3enterprisetaskproxyless) | reCAPTCHA v3 Enterprise |
| `ReCaptchaV3EnterpriseTaskProxylessS9` | all | [Example](#task-recaptchav3enterprisetaskproxylesss9) | reCAPTCHA v3 Enterprise, returns a token scored ≥ 0.9 |

### FunCaptcha / Arkose Labs

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `FuncaptchaTaskProxyless` | async | [Example](#task-funcaptchataskproxyless) | FunCaptcha / Arkose Labs |
| `FunCaptchaClassification` | sync | [Example](#task-funcaptchaclassification) | FunCaptcha image recognition |

### hCaptcha

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `HCaptcha` | async | [Example](#task-hcaptcha) | hCaptcha |
| `HCaptchaClassification` | sync | [Example](#task-hcaptchaclassification) | hCaptcha image recognition, single or multiple images |

### Cloudflare

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `CloudFlare5STask` | async | [Example](#task-cloudflare5stask) | CF five-second interstitial, **requires** a `Proxy` |
| `CloudFlareTurnstileTask` | async | [Example](#task-cloudflareturnstiletask) | Turnstile, returns a token |

### Akamai

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `AkamaiWEBTaskProxyless` | sync | [Example](#task-akamaiwebtaskproxyless) | Akamai Web |
| `AkamaiSBSDTaskProxyless` | sync | [Example](#task-akamaisbsdtaskproxyless) | Akamai SBSD |

> Akamai Web is a multi-round flow: feed the `Encodedata` of one round back as the `EncodeData` of the next. The two spellings genuinely differ on the wire; the SDK keeps the service's definitions as they are rather than "fixing" them.

### DataDome

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `DataDomeTaskProxyless` | sync | [Example](#task-datadometaskproxyless) | The challenge after an interception, in two steps selected by `Step` |
| `DataDomeTagsTaskProxyless` | sync | [Example](#task-datadometagstaskproxyless) | Reports a fingerprint on the normal browsing path |

### Other

| Task type | Modes | Example | Description |
| :-: | :---: | :-: | --- |
| `PerimeterX` | async | [Example](#task-perimeterx) | PerimeterX clearance cookies |
| `IncapsulaTaskProxyless` | sync | [Example](#task-incapsulataskproxyless) | Incapsula Reese84 payload |
| `TlsTask` | sync | [Example](#task-tlstask) | HTTP request forwarded over TLS, returns the upstream response |

## 📦 Installation

```bash
go get github.com/EZXLabs/ezcapsolver-go
```

Requires Go 1.26+. **Zero third-party dependencies** — the standard library only, and nothing added to your dependency tree.

## 🚀 Quick Start

The client reads `EZCAPTCHA_API_KEY` from the environment when no key is passed explicitly.

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/EZXLabs/ezcapsolver-go"
)

func main() {
	client, err := ezcapsolver.NewClient()
	if err != nil {
		log.Fatal(err)
	}

	solved, err := client.SolveRecaptchaV2TaskProxyless(
		context.Background(),
		&ezcapsolver.RecaptchaV2Task{
			WebsiteURL: "https://example.com",
			WebsiteKey: "6Lc_your_site_key",
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("task id:", solved.TaskID)
	fmt.Println("token:  ", solved.Solution.Token)
}
```

The client is safe to share across goroutines and holds a connection pool — build one and use it for the life of the process.

Building clients per key or per tenant instead? Call `client.Close()` when done: each client clones its own pool, and the idle connections otherwise sit on file descriptors until the transport's `IdleConnTimeout` expires. It is a no-op for a client supplied through `WithHTTPClient`, and the client stays usable afterwards.

## 📖 Usage

One section per task type below. The snippets assume a `client` and a `ctx` are already in scope, and leave out the imports to keep the call itself in focus.

### Sync / async

**Every task type has two methods**, taking the same arguments and returning the same type; only the endpoint differs:

```go
// Create, then poll
solved, err := client.SolveRecaptchaV2TaskProxyless(ctx, task)
// solved.TaskID is set

// The synchronous endpoint, answering on the creating request
solved, err = client.SyncSolveRecaptchaV2TaskProxyless(ctx, task)
// solved.TaskID carries the identifier this endpoint assigns
```

A method's name is its task type constant with `TaskType` swapped for `Solve`, plus a `Sync` prefix for the synchronous endpoint. The symmetry runs all the way to the escape hatch: `Solve` / `SyncSolve` for any type, `SolveAs[T]` / `SyncSolveAs[T]` to decode into a struct of your own. All four return a `*Solved[T]`.

### Proxy format

Task types that accept a `proxy` take it in one of two shapes, decided by the task type:

| Format | Shape | Used by |
| --- | --- | --- |
| `NORMAL` | `protocol://username:password@host:port` | every type except FunCaptcha |
| `FUN` | `protocol://host:port:username:password` | `FuncaptchaTaskProxyless` only |

`protocol` is one of `http`, `https` or `socks5`. **Both credentials are required** — the service
rejects an unauthenticated proxy — and the host may not be a private address (`127.0.*`,
`192.168.*`, `172.16.*`, `10.0.*`). Where the field is optional, leaving it empty is fine; it is
only a non-empty malformed value that is rejected.


### reCAPTCHA v2

[reCAPTCHA v2 API reference](https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v2)

The first five types share `RecaptchaV2Task` and `RecaptchaSolution`; only the method name differs.

<a id="task-recaptchav2taskproxyless"></a>

#### ReCaptchaV2TaskProxyless

```go
solved, err := client.SolveRecaptchaV2TaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("task id:", solved.TaskID)
fmt.Println("token:  ", solved.Solution.Token)
```

<a id="task-recaptchav2taskproxylesss9"></a>

#### ReCaptchaV2TaskProxylessS9

Identical parameters to plain v2; runs on the high-score queue and returns a token scored ≥ 0.9.

```go
solved, err := client.SolveRecaptchaV2TaskProxylessS9(ctx, &ezcapsolver.RecaptchaV2Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

<a id="task-recaptchav2staskproxyless"></a>

#### ReCaptchaV2STaskProxyless

Carries the challenge-bound `S` parameter. It is not mandatory; without it the behaviour matches plain v2.

```go
solved, err := client.SolveRecaptchaV2STaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
	S:          "value-from-the-page",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

<a id="task-recaptchav2enterprisetaskproxyless"></a>

#### ReCaptchaV2EnterpriseTaskProxyless

Enterprise. If the site uses enterprise parameters beyond `data-s`, put them in `Extra` and they are forwarded as-is.

```go
solved, err := client.SolveRecaptchaV2EnterpriseTaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

<a id="task-recaptchav2senterprisetaskproxyless"></a>

#### ReCaptchaV2SEnterpriseTaskProxyless

Enterprise, carrying the `S` parameter.

```go
solved, err := client.SolveRecaptchaV2SEnterpriseTaskProxyless(ctx, &ezcapsolver.RecaptchaV2Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
	S:          "value-from-the-page",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

<a id="task-recaptchav2classification"></a>

#### ReCaptchaV2Classification

[reCAPTCHA v2 classification API reference](https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v2-classification)

Recognises the image grid directly and returns tile indices rather than a token.

```go
solved, err := client.SyncSolveRecaptchaV2Classification(
	ctx,
	&ezcapsolver.RecaptchaV2ClassificationTask{
		Image:    imageBase64,
		Question: "/m/0k4j",
		Size:     4, // 1 = 1x1, 3 = 3x3, 4 = 4x4
	},
)
if err != nil {
	log.Fatal(err)
}

answer := solved.Solution
switch {
case answer.IsMulti():
	fmt.Println("tiles to click", answer.Objects)
case answer.IsSingle():
	fmt.Println("contains the object:", answer.HasObject)
default:
	fmt.Println("result type:", answer.Type)
	fmt.Println("raw result: ", string(solved.Raw))
}
```

`ReClassificationSolution` carries `Type`, `HasObject`, `Objects` and `Extra`, mapping the JSON `type`, `hasObject`, `objects` and the pass-through fields. An unknown type is kept verbatim and `Solved.Raw` still holds the original JSON. Missing fields fall back to Go zero values; `IsMulti()` and `IsSingle()` only look at `Type`.

### reCAPTCHA v3

[reCAPTCHA v3 API reference](https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v3)

The four types share `RecaptchaV3Task` and `RecaptchaSolution`. `PageAction` has to match the `action` the page passes to `grecaptcha.execute`, otherwise the site-side check fails.

> ⚠️ `IsInvisible` defaults to **true** on `RecaptchaV3Task`, so it is a `*bool`: leaving it nil omits the field and the service applies that default, which is what makes the zero value of the struct correct. To opt out, state it explicitly with `IsInvisible: ezcapsolver.Ptr(false)` — a plain `bool` could not, since `omitzero` drops a `false` exactly like an unset field.

<a id="task-recaptchav3taskproxyless"></a>

#### ReCaptchaV3TaskProxyless

```go
solved, err := client.SolveRecaptchaV3TaskProxyless(ctx, &ezcapsolver.RecaptchaV3Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
	PageAction: "login",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("task id:", solved.TaskID)
fmt.Println("token:  ", solved.Solution.Token)
```

<a id="task-recaptchav3taskproxylesss9"></a>

#### ReCaptchaV3TaskProxylessS9

```go
solved, err := client.SolveRecaptchaV3TaskProxylessS9(ctx, &ezcapsolver.RecaptchaV3Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
	PageAction: "login",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

<a id="task-recaptchav3enterprisetaskproxyless"></a>

#### ReCaptchaV3EnterpriseTaskProxyless

```go
solved, err := client.SolveRecaptchaV3EnterpriseTaskProxyless(ctx, &ezcapsolver.RecaptchaV3Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
	PageAction: "login",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

<a id="task-recaptchav3enterprisetaskproxylesss9"></a>

#### ReCaptchaV3EnterpriseTaskProxylessS9

```go
solved, err := client.SolveRecaptchaV3EnterpriseTaskProxylessS9(ctx, &ezcapsolver.RecaptchaV3Task{
	WebsiteURL: "https://example.com",
	WebsiteKey: "6Lc_your_site_key",
	PageAction: "login",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

### FunCaptcha / Arkose Labs

[FunCaptcha API reference](https://docs.ezxlabs.com/docs/captcha/api/funcaptcha)

<a id="task-funcaptchataskproxyless"></a>

#### FuncaptchaTaskProxyless

```go
solved, err := client.SolveFuncaptchaTaskProxyless(ctx, &ezcapsolver.FunCaptchaTask{
	WebsiteURL: "https://example.com",
	WebsiteKey: "your-public-key",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

> FunCaptcha is the one task type whose proxy uses the `FUN` format: `protocol://host:port:username:password`, with the credentials after the host rather than before it.

<a id="task-funcaptchaclassification"></a>

#### FunCaptchaClassification

```go
solved, err := client.SyncSolveFunCaptchaClassification(
	ctx,
	&ezcapsolver.FunCaptchaClassificationTask{
		Image:    imageBase64,
		Question: "Pick the animal facing left",
	},
)
if err != nil {
	log.Fatal(err)
}

// The result shape of this type is unconfirmed; every field the worker returns lands in Extra.
fmt.Println(solved.Solution.Extra)
```

### hCaptcha

[hCaptcha API reference](https://docs.ezxlabs.com/docs/captcha/api/hcaptcha)

<a id="task-hcaptcha"></a>

#### HCaptcha

The hCaptcha token field is `GeneratedPassUUID` (`generated_pass_UUID` on the wire), not `Token` — that is the service's naming, and the SDK keeps it.

```go
solved, err := client.SolveHCaptcha(ctx, &ezcapsolver.HCaptchaTask{
	WebsiteURL: "https://example.com",
	WebsiteKey: "your-site-key",
	Lang:       "en-US",
	// True on sites that show no checkbox. Always sent, so false is a
	// statement rather than an omission.
	Invisible: false,
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.GeneratedPassUUID)
```

<a id="task-hcaptchaclassification"></a>

#### HCaptchaClassification

Use `Image` for one image and `Images` for several; every field is optional, because different recognition modules need different combinations of input.

```go
solved, err := client.SyncSolveHCaptchaClassification(
	ctx,
	&ezcapsolver.HCaptchaClassificationTask{
		Images:   images,
		Question: "Please click each image containing a bicycle",
	},
)
if err != nil {
	log.Fatal(err)
}

// The shape is unconfirmed here too; every field is in Extra.
fmt.Println(solved.Solution.Extra)
```

### Cloudflare

<a id="task-cloudflare5stask"></a>

#### CloudFlare5STask

[Cloudflare 5S API reference](https://docs.ezxlabs.com/docs/captcha/api/cloudflare-5s)

The five-second interstitial **requires** a `Proxy`, and returns not a single token but the headers and clearance cookies to replay against the target site:

```go
solved, err := client.SolveCloudFlare5STask(ctx, &ezcapsolver.Cloudflare5sTask{
	WebsiteURL: "https://example.com",
	Proxy:      "http://user:pass@127.0.0.1:8080",
})
if err != nil {
	log.Fatal(err)
}

for name, value := range solved.Solution.Cookies {
	fmt.Printf("%s=%s\n", name, value)
}
fmt.Println("TLS fingerprint:", solved.Solution.TLSVersion)
```

Replaying those headers and cookies against the protected site is what actually clears the challenge — the result is a whole browser state, not a token.

<a id="task-cloudflareturnstiletask"></a>

#### CloudFlareTurnstileTask

[Cloudflare Turnstile API reference](https://docs.ezxlabs.com/docs/captcha/api/turnstile)

Turnstile's `Proxy` is optional, and it returns a single token.

```go
solved, err := client.SolveCloudFlareTurnstileTask(ctx, &ezcapsolver.CloudflareTurnstileTask{
	WebsiteURL: "https://example.com",
	WebsiteKey: "0x4AAA_your_site_key",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.Token)
```

### Akamai

<a id="task-akamaiwebtaskproxyless"></a>

#### AkamaiWEBTaskProxyless

[Akamai Web API reference](https://docs.ezxlabs.com/docs/captcha/api/akamai-web)

Akamai Web is a multi-round flow: feed the `Encodedata` of one round back as the `EncodeData` of the next. The two spellings genuinely differ on the wire; the SDK keeps the service's definitions as they are.

```go
encodeData := ""

for index := 1; index <= 3; index++ {
	solved, err := client.SyncSolveAkamaiWEBTaskProxyless(ctx, &ezcapsolver.AkamaiWebTask{
		PageURL:      "https://example.com",
		Ua:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
		Lang:         "zh-CN",
		Index:        index,
		Abck:         abck,
		Bmsz:         bmsz,
		ScriptBase64: scriptBase64,
		EncodeData:   encodeData,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("round %d payload = %s\n", index, solved.Solution.Payload)
	encodeData = solved.Solution.Encodedata
}
```

<a id="task-akamaisbsdtaskproxyless"></a>

#### AkamaiSBSDTaskProxyless

[Akamai SBSD API reference](https://docs.ezxlabs.com/docs/captcha/api/akamai-sbsd)

A single-round task; all six fields are required.

```go
solved, err := client.SyncSolveAkamaiSBSDTaskProxyless(ctx, &ezcapsolver.AkamaiSBSDTask{
	PageURL:      "https://example.com",
	SbsdURL:      "https://example.com/.well-known/sbsd",
	BmSo:         "value-from-the-page",
	Ua:           "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	Lang:         "zh-CN",
	ScriptBase64: scriptBase64,
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("payload:", solved.Solution.Payload)
```

### DataDome

<a id="task-datadometaskproxyless"></a>

#### DataDomeTaskProxyless

The challenge after a DataDome interception runs in two steps that share `DataDomeTask` and are selected by `Step`; which fields of `DataDomeSolution` carry a value depends on the step:

```go
// Step one: get the challenge address from the intercepted page.
solved, err := client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
	HtmlB64: htmlB64,
	Step:    ezcapsolver.DataDomeStepOne,
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("challenge address:", solved.Solution.URL)

// Step two uses ezcapsolver.DataDomeStepTwo, and the result carries the validation Body.
```

<a id="task-datadometagstaskproxyless"></a>

#### DataDomeTagsTaskProxyless

Reports a fingerprint on the normal browsing path. Its field names are camelCase, unlike `DataDomeTask` above.

```go
solved, err := client.SyncSolveDataDomeTagsTaskProxyless(ctx, &ezcapsolver.DataDomeTagsTask{
	Ddk:     "your-datadome-key",
	Referer: "https://example.com",
	Ua:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
	Bpc:     1,
})
if err != nil {
	log.Fatal(err)
}

fmt.Println(solved.Solution.Extra)
```

### Other

<a id="task-perimeterx"></a>

#### PerimeterX

[PerimeterX API reference](https://docs.ezxlabs.com/docs/captcha/api/perimeterx)

```go
solved, err := client.SolvePerimeterX(ctx, &ezcapsolver.PerimeterXTask{
	WebsiteKey: "PX_your_app_id",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("_px3:  ", solved.Solution.Px3)
fmt.Println("_pxvid:", solved.Solution.PxVid)
```

<a id="task-incapsulataskproxyless"></a>

#### IncapsulaTaskProxyless

[Incapsula API reference](https://docs.ezxlabs.com/docs/captcha/api/incapsula)

```go
solved, err := client.SyncSolveIncapsulaTaskProxyless(ctx, &ezcapsolver.IncapsulaTask{
	Script:         script,
	ScriptURL:      "https://example.com/sensor.js",
	PageURL:        "https://example.com",
	AcceptLanguage: "zh-CN,zh;q=0.9",
	Ua:             "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("reese84:", solved.Solution.Data)
```

<a id="task-tlstask"></a>

#### TlsTask

[TLS forwarding API reference](https://docs.ezxlabs.com/docs/captcha/api/tls-forward)

This type does not solve a captcha. It sends one HTTP request through the worker's TLS fingerprint and brings the upstream response back untouched:

```go
solved, err := client.SyncSolveTLSTask(ctx, &ezcapsolver.TLSForwardTask{
	TLSType: "chrome",
	Proxy:   "http://user:pass@127.0.0.1:8080",
	Method:  ezcapsolver.TLSMethodGET,
	URL:     "https://example.com/api",
})
if err != nil {
	log.Fatal(err)
}

fmt.Printf("HTTP %d, %d bytes of body\n", solved.Solution.Code, len(solved.Solution.Body))
```

### Task result structures

Result structures have named fields, so reading a result never means writing `solution["gRecaptchaResponse"]`. Each one also carries an `Extra` map that catches fields the service adds later:

```go
if value, ok := solved.Solution.Extra["aFieldAddedLater"]; ok {
	fmt.Println(value)
}
```

A field the worker omits does not break decoding. The raw JSON is always in `solved.Raw`, and at a lower level "the response had no `solution` field" stays distinguishable from "`solution` was JSON `null`":

```go
result, err := client.GetTaskResult(ctx, taskID)
if err != nil {
	log.Fatal(err)
}
if !result.HasSolution() {
	log.Fatal("the task has not finished")
}

var solution ezcapsolver.RecaptchaSolution
if err := result.DecodeSolution(&solution); err != nil {
	log.Fatal(err)
}
```

When decoding genuinely fails, `*SolutionDecodeError` carries the original value along — which is exactly what a diagnosis needs.

### Pass-through fields

Every task model has an `Extra` map whose entries are flattened into the task JSON alongside the declared fields, so parameters the service adds later work without an SDK release. A declared field always wins over an `Extra` entry of the same name, and the losing entry is dropped silently:

```go
task := &ezcapsolver.HCaptchaTask{
	WebsiteURL: "https://example.com",
	WebsiteKey: "site-key",
	Lang:       "en-US",
	Extra: map[string]any{
		// A parameter this release does not model, or one the service ships
		// later. Naming a declared field here instead would be dropped.
		"futureFlag": true,
	},
}
```

The `type` field is under the SDK's control and cannot be set this way.

### Custom task types

Task types are an **open set**. Pass the type name as a string and the parameters as any value that serialises to a JSON object:

```go
solved, err := client.Solve(ctx, "BrandNewTaskType", map[string]any{
	"websiteURL":     "https://example.com",
	"anyFutureParam": 42,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(string(solved.Raw))

// The synchronous-endpoint counterpart: same arguments, same result type, other endpoint
solved, err = client.SyncSolve(ctx, "BrandNewSyncType", map[string]any{"input": "..."})
```

Or decode straight into a type of your own:

```go
type myShape struct {
	Token string `json:"token"`
}

solved, err := ezcapsolver.SolveAs[myShape](ctx, client, "BrandNewTaskType", params)

// On the synchronous endpoint
solved, err = ezcapsolver.SyncSolveAs[myShape](ctx, client, "BrandNewSyncType", params)
```

`CreateTask`, `GetTaskResult`, `WaitForResult`, `CreateSyncTask`, `Solve` and `SyncSolve` all take a `TaskType`, which is itself a string type — the low-level flow and the escape hatch use the same methods. `CreateTask` and `CreateSyncTask` are the primitives underneath each path; reach for them only when you need the raw `TaskResult`. Use this when the service ships a new captcha type the SDK has not caught up with yet.

## ⚙️ Configuration

```go
client, err := ezcapsolver.NewClient(
	ezcapsolver.WithClientKey("your-client-key"),
	ezcapsolver.WithTimeout(30*time.Second),
	ezcapsolver.WithSyncTimeout(240*time.Second),
	ezcapsolver.WithPolling(ezcapsolver.PollingConfig{Interval: 3 * time.Second, MaxAttempts: 50}),
	ezcapsolver.WithAppID(42),
	ezcapsolver.WithProxy("http://127.0.0.1:8080"),
	ezcapsolver.WithUserAgent("my-app/1.0"),
)
```

| Setting | Default | Scope |
| --- | :---: | --- |
| `WithClientKey` | `EZCAPTCHA_API_KEY` | Falls back to the environment variable |
| `WithTimeout` | 30 s | Request timeout for the asynchronous endpoint |
| `WithSyncTimeout` | 240 s | Request timeout for the synchronous endpoint |
| `WithPolling` | 3 s × 50 | Result queries, up to 150 s per task |
| `WithProxy` | none | **The SDK's own egress**, unrelated to the `Proxy` inside task parameters |
| `WithBaseURLs` | production | Self-hosted deployments and integration tests |
| `WithHTTPClient` | a new one | Custom transport or instrumentation |
| `WithLogger` | discard | See [Logging](#-logging) |

Everything checkable is checked when the client is built rather than on the first request — a misconfiguration surfaces before anything is billed. Printing a `ClientConfig` is safe: the key and proxy are replaced with `[REDACTED]`.

The two timeout budgets are kept apart deliberately. A synchronous call blocks until the worker returns, and the service allows up to three minutes for some types; sharing the 30-second budget would cut off a call that **has already been billed** and has no task ID to recover the result with.

## ⚠️ Errors

Five layers, split by what a caller can do about them. Use `errors.Is` for the category and `errors.As` for the detail:

| Sentinel | Type | Meaning |
| --- | --- | --- |
| `ErrConfig` | — | Invalid configuration. Only returned by `NewClient`; the reason is in the message |
| `ErrTransport` | `*TransportError` | No response arrived: DNS, TCP, TLS, timeout |
| `ErrAPI` | `*APIError` | The service returned a structured error |
| `ErrPollingExhausted` | `*PollingExhaustedError` | The wait budget ran out; the task may still finish |
| `ErrDecode` | `*UnexpectedResponseError`, `*SolutionDecodeError` | The response is unusable |


Whichever failure it is, one question answers whether a **billed** task is still recoverable:

```go
if taskID := ezcapsolver.TaskIDOf(err); taskID != "" {
    // The task is on the service; its result is held for five minutes after
    // creation. Waiting again is free — creating a second task is billed again.
    result, err := client.WaitForResult(ctx, taskID)
}
```

An empty string means nothing was billed, so there is nothing to recover.

A failure raised while waiting for an already-created task is wrapped in a `*WaitInterruptedError`, which carries `TaskID` out. `Solve` creates the task internally, so that wrapper is the only place the id appears — wait on the same task again instead of paying for a second one. The wrapper unwraps to the original error, so `errors.Is` against the sentinels above answers exactly as it would without it.

```go
var apiErr *ezcapsolver.APIError
switch {
case errors.As(err, &apiErr):
	fmt.Println(apiErr.ErrorCode, apiErr.HTTPStatus, apiErr.TaskID)
	for field, reason := range apiErr.Errors {
		fmt.Printf("  %s: %s\n", field, reason)
	}

	switch {
	case apiErr.IsAuthenticationError():
		// Stop. Do not retry: these three codes accumulate into a server-side ban.
	case apiErr.IsTerminal():
		// The same request will get the same answer.
	case apiErr.IsRateLimited():
		// Throttled. Both codes clear on their own, so ask again later.
	}

case errors.Is(err, ezcapsolver.ErrPollingExhausted):
	var pollErr *ezcapsolver.PollingExhaustedError
	errors.As(err, &pollErr)
	// The task ID is still valid; GetTaskResult can fetch the result later.
	fmt.Println(pollErr.TaskID)
}
```

`ErrorCode` is exposed verbatim. Worker-produced codes are not in the service's own table — that is an open set — so the SDK never collapses an unrecognised code into "unknown error".

**A throttled poll is the one thing `WaitForResult` retries.** `ERROR_REQUEST_LIMIT` and `ERROR_REQUEST_BANNED` refuse the *query*, not the task: the service turns the request away before it ever looks the task up, so the task is still queued and still billed. The loop spends the attempt and polls again rather than discarding a result that was about to arrive. Every other API error is the poll's answer and ends the wait. Note that `IsRateLimited` is narrower than `!IsTerminal()`, which is also true of every unrecognised code — including the worker codes that report a task that genuinely failed.

## 📝 Logging

The SDK logs through `log/slog` and discards everything until a logger is injected.

| Level | Events |
| :---: | --- |
| `INFO` | Task created, task finished |
| `DEBUG` | One line per operation, one per poll |
| `LevelTrace` | Request bodies, and response status, duration, size and body |

slog has no trace level, so `LevelTrace` sits one step below `slog.LevelDebug`. It is for diagnosing an integration that will not connect: request bodies are rendered with `clientKey` and `proxy` replaced at **any nesting depth**, and with the handler's level above it the body is never rendered at all — so leaving it off costs nothing.

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: ezcapsolver.LevelTrace,
}))
client, err := ezcapsolver.NewClient(ezcapsolver.WithLogger(logger))
```

## 🧪 Runnable Examples

[`examples/`](./examples) holds one runnable file per task type, named after the wire task type. Set `EZCAPTCHA_API_KEY`, then run them from the repository root:

```bash
export EZCAPTCHA_API_KEY=your-client-key

go run examples/recaptcha_v2/recaptcha_v2_task_proxyless.go
go run examples/cloudflare/cloud_flare_turnstile_task.go
go run examples/tls_forward/tls_task.go

# About the SDK itself rather than a task type
go run examples/client_setup.go   # every client setting, one by one
go run examples/raw_usage.go      # the low-level create/poll workflow
```

Each example states the task type it covers and links to the matching page of the official API docs. Examples that need a proxy read `EZCAPTCHA_PROXY`; none of them hard-code credentials.

> Every run creates a real task and is billed, whether or not the worker succeeds.

## 🛠️ Development

Minimum Go version: **1.26**. The floor tracks the Go release line still receiving security patches; a lower local version makes the toolchain fetch a matching one automatically.

```bash
gofmt -l .
go vet ./...
staticcheck ./...
go test -race -cover ./...
```

Examples carry `//go:build ignore` and are excluded from the package build, so they need a separate check:

```bash
find examples -name '*.go' -exec go build -o /dev/null {} \;
```

The full process is in [CONTRIBUTING.md](./CONTRIBUTING.md), including every place a new task type has to be added.

## 📄 License

Licensed under the [Apache License 2.0](./LICENSE).
