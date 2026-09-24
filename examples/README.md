# Examples

[English](./README.md) · [简体中文](./README.zh-CN.md)

One runnable file per task type, named after the wire task type so it lines up
with the Rust and Python SDKs file for file.

Every file carries `//go:build ignore`, which keeps them out of `go build ./...`
while leaving them runnable directly:

```bash
export EZCAPTCHA_API_KEY=your-key
go run examples/recaptcha_v2/recaptcha_v2_task_proxyless.go
```

Run them from the repository root — the classification examples read their
images from `examples/fixtures/`.

No example hardcodes a key. Ones that need a proxy read `EZCAPTCHA_PROXY`.

> Every run creates a real task and is billed, whether or not the worker
> succeeds. The reCAPTCHA v2 and hCaptcha examples point at the vendors' own
> demo pages and work as written; the rest carry placeholder site keys for you
> to fill in.

## The SDK itself

| Example | What it covers |
| --- | --- |
| [`client_setup.go`](./client_setup.go) | Every option, and why the two timeout budgets are separate |
| [`concurrency.go`](./concurrency.go) | Sharing one client across goroutines, with a concurrency cap |
| [`raw_usage.go`](./raw_usage.go) | Polling by hand, and reaching a task type the SDK does not model |
| [`logging_and_errors.go`](./logging_and_errors.go) | Injecting a logger, and telling the five failure layers apart |

## reCAPTCHA v2

[Docs](https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v2)

| Example | Task type |
| --- | --- |
| [`recaptcha_v2/recaptcha_v2_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_task_proxyless.go) | `ReCaptchaV2TaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_task_proxyless_s9.go`](./recaptcha_v2/recaptcha_v2_task_proxyless_s9.go) | `ReCaptchaV2TaskProxylessS9` |
| [`recaptcha_v2/recaptcha_v2_s_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_s_task_proxyless.go) | `ReCaptchaV2STaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_enterprise_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_enterprise_task_proxyless.go) | `ReCaptchaV2EnterpriseTaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_s_enterprise_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_s_enterprise_task_proxyless.go) | `ReCaptchaV2SEnterpriseTaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_classification.go`](./recaptcha_v2/recaptcha_v2_classification.go) | `ReCaptchaV2Classification` |

## reCAPTCHA v3

[Docs](https://docs.ezxlabs.com/docs/captcha/api/recaptcha-v3)

| Example | Task type |
| --- | --- |
| [`recaptcha_v3/recaptcha_v3_task_proxyless.go`](./recaptcha_v3/recaptcha_v3_task_proxyless.go) | `ReCaptchaV3TaskProxyless` |
| [`recaptcha_v3/recaptcha_v3_task_proxyless_s9.go`](./recaptcha_v3/recaptcha_v3_task_proxyless_s9.go) | `ReCaptchaV3TaskProxylessS9` |
| [`recaptcha_v3/recaptcha_v3_enterprise_task_proxyless.go`](./recaptcha_v3/recaptcha_v3_enterprise_task_proxyless.go) | `ReCaptchaV3EnterpriseTaskProxyless` |
| [`recaptcha_v3/recaptcha_v3_enterprise_task_proxyless_s9.go`](./recaptcha_v3/recaptcha_v3_enterprise_task_proxyless_s9.go) | `RecaptchaV3EnterpriseTaskProxylessS9` |

## FunCaptcha / Arkose Labs

[Docs](https://docs.ezxlabs.com/docs/captcha/api/funcaptcha)

| Example | Task type |
| --- | --- |
| [`funcaptcha/funcaptcha_task_proxyless.go`](./funcaptcha/funcaptcha_task_proxyless.go) | `FuncaptchaTaskProxyless` |
| [`funcaptcha/funcaptcha_classification.go`](./funcaptcha/funcaptcha_classification.go) | `FunCaptchaClassification` |

## hCaptcha

[Docs](https://docs.ezxlabs.com/docs/captcha/api/hcaptcha)

| Example | Task type |
| --- | --- |
| [`hcaptcha/hcaptcha.go`](./hcaptcha/hcaptcha.go) | `HCaptcha` |
| [`hcaptcha/hcaptcha_classification.go`](./hcaptcha/hcaptcha_classification.go) | `HCaptchaClassification` |

## Cloudflare

| Example | Task type |
| --- | --- |
| [`cloudflare/cloud_flare_5s_task.go`](./cloudflare/cloud_flare_5s_task.go) | `CloudFlare5STask` |
| [`cloudflare/cloud_flare_turnstile_task.go`](./cloudflare/cloud_flare_turnstile_task.go) | `CloudFlareTurnstileTask` |

## Akamai

| Example | Task type |
| --- | --- |
| [`akamai/akamai_web_task_proxyless.go`](./akamai/akamai_web_task_proxyless.go) | `AkamaiWEBTaskProxyless` |
| [`akamai/akamai_sbsd_task_proxyless.go`](./akamai/akamai_sbsd_task_proxyless.go) | `AkamaiSBSDTaskProxyless` |

## DataDome

| Example | Task type |
| --- | --- |
| [`datadome/datadome_task_proxyless.go`](./datadome/datadome_task_proxyless.go) | `DataDomeTaskProxyless` |
| [`datadome/datadome_tags_task_proxyless.go`](./datadome/datadome_tags_task_proxyless.go) | `DataDomeTagsTaskProxyless` |

## Other protection systems

| Example | Task type |
| --- | --- |
| [`perimeterx/perimeter_x.go`](./perimeterx/perimeter_x.go) | `PerimeterX` |
| [`incapsula/incapsula_task_proxyless.go`](./incapsula/incapsula_task_proxyless.go) | `IncapsulaTaskProxyless` |
| [`tls_forward/tls_task.go`](./tls_forward/tls_task.go) | `TlsTask` |

## Fixtures

`fixtures/` holds the images the classification examples read. They are the same
three files the Rust and Python SDKs use, so the three sets are comparable.
