# 示例

[English](./README.md) · [简体中文](./README.zh-CN.md)

每种任务类型一个可运行文件，文件名取自线格式的任务类型名，与 Rust、Python 两个 SDK
的示例逐个对应。

每个文件都带 `//go:build ignore`，因此不会进入 `go build ./...`，但可以直接运行：

```bash
export EZCAPTCHA_API_KEY=your-key
go run examples/recaptcha_v2/recaptcha_v2_task_proxyless.go
```

请在仓库根目录运行——分类类示例要从 `examples/fixtures/` 读图片。

示例一律不硬编码密钥；需要代理的从 `EZCAPTCHA_PROXY` 读。

> 每次运行都会创建真实任务并**计费**，worker 失败也照扣。reCAPTCHA v2 与 hCaptcha
> 两个示例指向厂商自己的 demo 页面，可以直接跑通；其余示例里的 site key 是占位值，
> 需要你替换。

## SDK 本身

| 示例 | 讲什么 |
| --- | --- |
| [`client_setup.go`](./client_setup.go) | 全部配置项，以及两条超时预算为什么要分开 |
| [`concurrency.go`](./concurrency.go) | 一个客户端跨 goroutine 复用，并限制并发数 |
| [`raw_usage.go`](./raw_usage.go) | 手动轮询，以及调用 SDK 未建模的任务类型 |
| [`logging_and_errors.go`](./logging_and_errors.go) | 注入日志，以及区分五层错误 |

## reCAPTCHA v2

[文档](https://docs.ezxlabs.com/zh/docs/captcha/api/recaptcha-v2)

| 示例 | 任务类型 |
| --- | --- |
| [`recaptcha_v2/recaptcha_v2_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_task_proxyless.go) | `ReCaptchaV2TaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_task_proxyless_s9.go`](./recaptcha_v2/recaptcha_v2_task_proxyless_s9.go) | `ReCaptchaV2TaskProxylessS9` |
| [`recaptcha_v2/recaptcha_v2_s_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_s_task_proxyless.go) | `ReCaptchaV2STaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_enterprise_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_enterprise_task_proxyless.go) | `ReCaptchaV2EnterpriseTaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_s_enterprise_task_proxyless.go`](./recaptcha_v2/recaptcha_v2_s_enterprise_task_proxyless.go) | `ReCaptchaV2SEnterpriseTaskProxyless` |
| [`recaptcha_v2/recaptcha_v2_classification.go`](./recaptcha_v2/recaptcha_v2_classification.go) | `ReCaptchaV2Classification` |

## reCAPTCHA v3

[文档](https://docs.ezxlabs.com/zh/docs/captcha/api/recaptcha-v3)

| 示例 | 任务类型 |
| --- | --- |
| [`recaptcha_v3/recaptcha_v3_task_proxyless.go`](./recaptcha_v3/recaptcha_v3_task_proxyless.go) | `ReCaptchaV3TaskProxyless` |
| [`recaptcha_v3/recaptcha_v3_task_proxyless_s9.go`](./recaptcha_v3/recaptcha_v3_task_proxyless_s9.go) | `ReCaptchaV3TaskProxylessS9` |
| [`recaptcha_v3/recaptcha_v3_enterprise_task_proxyless.go`](./recaptcha_v3/recaptcha_v3_enterprise_task_proxyless.go) | `ReCaptchaV3EnterpriseTaskProxyless` |
| [`recaptcha_v3/recaptcha_v3_enterprise_task_proxyless_s9.go`](./recaptcha_v3/recaptcha_v3_enterprise_task_proxyless_s9.go) | `RecaptchaV3EnterpriseTaskProxylessS9` |

## FunCaptcha / Arkose Labs

[文档](https://docs.ezxlabs.com/zh/docs/captcha/api/funcaptcha)

| 示例 | 任务类型 |
| --- | --- |
| [`funcaptcha/funcaptcha_task_proxyless.go`](./funcaptcha/funcaptcha_task_proxyless.go) | `FuncaptchaTaskProxyless` |
| [`funcaptcha/funcaptcha_classification.go`](./funcaptcha/funcaptcha_classification.go) | `FunCaptchaClassification` |

## hCaptcha

[文档](https://docs.ezxlabs.com/zh/docs/captcha/api/hcaptcha)

| 示例 | 任务类型 |
| --- | --- |
| [`hcaptcha/hcaptcha.go`](./hcaptcha/hcaptcha.go) | `HCaptcha` |
| [`hcaptcha/hcaptcha_classification.go`](./hcaptcha/hcaptcha_classification.go) | `HCaptchaClassification` |

## Cloudflare

| 示例 | 任务类型 |
| --- | --- |
| [`cloudflare/cloud_flare_5s_task.go`](./cloudflare/cloud_flare_5s_task.go) | `CloudFlare5STask` |
| [`cloudflare/cloud_flare_turnstile_task.go`](./cloudflare/cloud_flare_turnstile_task.go) | `CloudFlareTurnstileTask` |

## Akamai

| 示例 | 任务类型 |
| --- | --- |
| [`akamai/akamai_web_task_proxyless.go`](./akamai/akamai_web_task_proxyless.go) | `AkamaiWEBTaskProxyless` |
| [`akamai/akamai_sbsd_task_proxyless.go`](./akamai/akamai_sbsd_task_proxyless.go) | `AkamaiSBSDTaskProxyless` |

## DataDome

| 示例 | 任务类型 |
| --- | --- |
| [`datadome/datadome_task_proxyless.go`](./datadome/datadome_task_proxyless.go) | `DataDomeTaskProxyless` |
| [`datadome/datadome_tags_task_proxyless.go`](./datadome/datadome_tags_task_proxyless.go) | `DataDomeTagsTaskProxyless` |

## 其它防护系统

| 示例 | 任务类型 |
| --- | --- |
| [`perimeterx/perimeter_x.go`](./perimeterx/perimeter_x.go) | `PerimeterX` |
| [`incapsula/incapsula_task_proxyless.go`](./incapsula/incapsula_task_proxyless.go) | `IncapsulaTaskProxyless` |
| [`tls_forward/tls_task.go`](./tls_forward/tls_task.go) | `TlsTask` |

## 图片素材

`fixtures/` 放分类类示例要用的图片，与 Rust、Python 两个 SDK 用的是同三个文件，
方便跨语言对照。
