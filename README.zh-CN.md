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
    <a href="https://ezxlabs.com"><img alt="EZXLabs 官网" src="https://img.shields.io/badge/website-ezxlabs.com-FFDB29?logoColor=black"></a>
  </p>
  <p>
    <a href="https://ezxlabs.com">🌐 官网</a> &nbsp;·&nbsp;
    <a href="https://docs.ezxlabs.com/zh/docs/captcha/api">📚 EZCaptchaSolver API 文档</a> &nbsp;·&nbsp;
    <a href="./examples">🧪 示例</a> &nbsp;·&nbsp;
    <a href="#-支持的验证码类型">🧩 验证码类型</a>
  </p>
  <p><a href="./README.md">English</a> &nbsp;·&nbsp; <b>简体中文</b></p>
</div>

---

EZCaptchaSolver Go SDK 是 [EZXLabs](https://ezxlabs.com) 维护的开源 Go 客户端，用于接入其 CAPTCHA 识别任务 API。它为下列受支持的任务类型提供类型化请求和支持 context 的客户端；该模块还包含不用于解验证码的 TLS 转发任务。了解 SDK 产品系列请查看 [EZCaptchaSolver SDK 产品页](https://ezxlabs.com/zh/products/sdk)，查看 HTTP 请求与响应字段请访问 [EZCaptchaSolver API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api)，Go 调用代码请以[本仓库示例](./examples/README.zh-CN.md)为准。

## 🧩 支持的验证码类型

验证码任务类型分为同步和异步两种：

- 同步：创建任务后阻塞请求，直到任务完成取得任务结果。
- 异步：创建任务成功后返回任务ID，后续可通过任务ID轮询尝试获取任务结果。适合验证码处理时长较久的任务类型。

同一种验证码任务类型能够同时支持同步和异步两种，几乎所有类型都会支持同步方式。然后有部分类型只会支持异步方式。

### reCAPTCHA v2

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `ReCaptchaV2TaskProxyless` | all | [示例](#task-recaptchav2taskproxyless) | reCAPTCHA v2 解决方案 |
| `ReCaptchaV2TaskProxylessS9` | all | [示例](#task-recaptchav2taskproxylesss9) | reCAPTCHA v2，返回分值 ≥ 0.9 的 token |
| `ReCaptchaV2STaskProxyless` | all | [示例](#task-recaptchav2staskproxyless) | reCAPTCHA v2 携带挑战绑定的 `s` 参数 |
| `ReCaptchaV2EnterpriseTaskProxyless` | all | [示例](#task-recaptchav2enterprisetaskproxyless) | reCAPTCHA v2 企业版 |
| `ReCaptchaV2SEnterpriseTaskProxyless` | all | [示例](#task-recaptchav2senterprisetaskproxyless) | reCAPTCHA v2 企业版，并携带 `s` 参数 |
| `ReCaptchaV2Classification` | sync | [示例](#task-recaptchav2classification) | reCAPTCHA v2 图片识别 |

### reCAPTCHA v3

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `ReCaptchaV3TaskProxyless` | all | [示例](#task-recaptchav3taskproxyless) | reCAPTCHA v3 解决方案 |
| `ReCaptchaV3TaskProxylessS9` | all | [示例](#task-recaptchav3taskproxylesss9) | reCAPTCHA v3，返回分值 ≥ 0.9 的 token |
| `ReCaptchaV3EnterpriseTaskProxyless` | all | [示例](#task-recaptchav3enterprisetaskproxyless) | reCAPTCHA v3 企业版 |
| `ReCaptchaV3EnterpriseTaskProxylessS9` | all | [示例](#task-recaptchav3enterprisetaskproxylesss9) | reCAPTCHA v3 企业版 返回分值 ≥ 0.9 的 token |

### FunCaptcha / Arkose Labs

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `FuncaptchaTaskProxyless` | async | [示例](#task-funcaptchataskproxyless) | FunCaptcha / Arkose Labs 解决方案 |
| `FunCaptchaClassification` | sync | [示例](#task-funcaptchaclassification) | FunCaptcha 图片识别 |

### hCaptcha

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `HCaptcha` | async | [示例](#task-hcaptcha) | hCaptcha 解决方案 |
| `HCaptchaClassification` | sync | [示例](#task-hcaptchaclassification) | hCaptcha 图片识别，支持单图与多图 |

### Cloudflare

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `CloudFlare5STask` | async | [示例](#task-cloudflare5stask) | CF 5 秒盾，**必须**传 `Proxy` |
| `CloudFlareTurnstileTask` | async | [示例](#task-cloudflareturnstiletask) | Turnstile，返回 token |

### Akamai

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `AkamaiWEBTaskProxyless` | sync | [示例](#task-akamaiwebtaskproxyless) | Akamai Web 解决方案 |
| `AkamaiSBSDTaskProxyless` | sync | [示例](#task-akamaisbsdtaskproxyless) | Akamai SBSD 解决方案 |

> Akamai Web 是多轮流程：把本轮返回的 `Encodedata` 作为下一轮的 `EncodeData` 传入。这两个拼写在线格式上就是不同的，SDK 原样保留服务端的定义，不做「修正」。

### DataDome

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `DataDomeTaskProxyless` | sync | [示例](#task-datadometaskproxyless) | 拦截后的挑战，分两步，靠 `Step` 区分 |
| `DataDomeTagsTaskProxyless` | sync | [示例](#task-datadometagstaskproxyless) | 正常浏览路径上报指纹 |

### 其他

| 任务类型 | 支持方式 | 示例 | 描述 |
| :-: | :---: | :-: | --- |
| `PerimeterX` | async | [示例](#task-perimeterx) | PerimeterX 放行 cookie |
| `IncapsulaTaskProxyless` | sync | [示例](#task-incapsulataskproxyless) | Incapsula Reese84 载荷 |
| `TlsTask` | sync | [示例](#task-tlstask) | HTTP TLS 转发请求，返回上游响应 |

## 📦 安装

```bash
go get github.com/EZXLabs/ezcapsolver-go
```

需要 Go 1.26+。**零第三方依赖**——只用标准库，不往你的依赖树里塞任何东西。

## 🚀 快速开始

不显式传密钥时，客户端从环境变量 `EZCAPTCHA_API_KEY` 读取。

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

客户端可安全跨 goroutine 共享，并持有连接池——建一个用一整个进程周期即可。

如果是按 key 或按租户建多个客户端，用完记得 `client.Close()`：每个客户端都克隆了自己的连接池，不关的话空闲连接会一直占着文件描述符，直到 transport 自己的 `IdleConnTimeout` 到期。对 `WithHTTPClient` 传入的客户端它什么都不做，且关闭后客户端仍然可用。

## 📖 使用说明

下面每种类型一段。片段都假设已经有一个 `client` 和一个 `ctx`，并省略了 import，以突出调用本身。

### 同步/异步

**每个任务类型都有两个方法**，参数与返回类型完全相同，区别只是走哪个端点：

```go
// 创建 + 轮询
solved, err := client.SolveRecaptchaV2TaskProxyless(ctx, task)
// solved.TaskID 非空

// 同步端点，一次请求拿结果
solved, err = client.SyncSolveRecaptchaV2TaskProxyless(ctx, task)
// solved.TaskID 带回该端点分配的任务 ID
```

方法名就是对应的任务类型常量把 `TaskType` 换成 `Solve`，同步端点的那个再加 `Sync` 前缀。这条对称性一直贯到逃生舱：任意类型用 `Solve` / `SyncSolve`，要解码进自己的结构体用 `SolveAs[T]` / `SyncSolveAs[T]`，四个都返回 `*Solved[T]`。

### 代理格式

接受 `proxy` 的任务类型有两种格式，取决于任务类型：

| 格式 | 形态 | 适用 |
| --- | --- | --- |
| `NORMAL` | `protocol://username:password@host:port` | 除 FunCaptcha 外全部 |
| `FUN` | `protocol://host:port:username:password` | 仅 `FuncaptchaTaskProxyless` |

`protocol` 取 `http`、`https` 或 `socks5`。**用户名和密码都必填**——无认证代理会被服务端判为非法
——且 host 不能是内网地址（`127.0.*`、`192.168.*`、`172.16.*`、`10.0.*`）。字段本身可选时留空没问题，
只有「非空但格式不对」才会被拒。


### reCAPTCHA v2

[reCAPTCHA v2 API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/recaptcha-v2)

前五个类型共用 `RecaptchaV2Task` 与 `RecaptchaSolution`，只有方法名不同。

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

参数与普通 v2 完全一致，走高分队列，返回分值 ≥ 0.9 的 token。

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

携带挑战绑定的 `S` 参数。该参数并非强制，不传时行为与普通 v2 一致。

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

企业版。若站点用了 `data-s` 之外的企业参数，放进 `Extra` 透传即可。

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

企业版，并携带 `S` 参数。

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

[reCAPTCHA v2 图片识别 API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/recaptcha-v2-classification)

直接识别图像宫格，返回格子下标而不是 token。

```go
solved, err := client.SyncSolveRecaptchaV2Classification(
	ctx,
	&ezcapsolver.RecaptchaV2ClassificationTask{
		Image:    imageBase64,
		Question: "/m/0k4j",
		Size:     4, // 1 = 1x1、3 = 3x3、4 = 4x4
	},
)
if err != nil {
	log.Fatal(err)
}

answer := solved.Solution
switch {
case answer.IsMulti():
	fmt.Println("需要点选的格子", answer.Objects)
case answer.IsSingle():
	fmt.Println("是否含目标物体:", answer.HasObject)
default:
	fmt.Println("结果类型:", answer.Type)
	fmt.Println("原始结果:", string(solved.Raw))
}
```

### reCAPTCHA v3

[reCAPTCHA v3 API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/recaptcha-v3)

四个类型共用 `RecaptchaV3Task` 与 `RecaptchaSolution`。`PageAction` 要与页面上 `grecaptcha.execute` 传的 `action` 一致，否则站点侧校验会失败。

> ⚠️ `IsInvisible` 在 `RecaptchaV3Task` 上默认为 **true**，因此它是 `*bool`：留空（nil）时字段不出现，服务端套用自己的默认值，结构体零值因此天然正确。要显式关掉，写 `IsInvisible: ezcapsolver.Ptr(false)` —— 普通 `bool` 做不到这点，`omitzero` 会把 `false` 和「没设置」一样丢掉。

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

[FunCaptcha API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/funcaptcha)

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

> FunCaptcha 是唯一使用 `FUN` 代理格式的类型：`protocol://host:port:username:password`

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

// 该类型的结果形态尚未确认，worker 返回的字段全部落在 Extra 里。
fmt.Println(solved.Solution.Extra)
```

### hCaptcha

[hCaptcha API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/hcaptcha)

<a id="task-hcaptcha"></a>

#### HCaptcha

hCaptcha 的 token 字段是 `GeneratedPassUUID`（线格式 `generated_pass_UUID`），不是 `Token`——这是服务端的命名，SDK 原样保留。

```go
solved, err := client.SolveHCaptcha(ctx, &ezcapsolver.HCaptchaTask{
	WebsiteURL: "https://example.com",
	WebsiteKey: "your-site-key",
	Lang:       "en-US",
	// 站点不显示勾选框时传 true。这个字段恒定发送，所以 false 是一种表态而不是遗漏
	Invisible: false,
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("token:", solved.Solution.GeneratedPassUUID)
```

<a id="task-hcaptchaclassification"></a>

#### HCaptchaClassification

单图用 `Image`，多图用 `Images`，全部字段都是可选的——不同的识别模块需要的输入组合不同。

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

// 形态同样未确认，全部字段在 Extra 里。
fmt.Println(solved.Solution.Extra)
```

### Cloudflare

<a id="task-cloudflare5stask"></a>

#### CloudFlare5STask

[Cloudflare 5S API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/cloudflare-5s)

5 秒盾**必须**传 `Proxy`，而且返回的不是单个 token，而是要回放到目标站点的 header 与放行 cookie：

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
fmt.Println("TLS 指纹:", solved.Solution.TLSVersion)
```

把这些请求头和 cookie 重放到受保护站点才是真正通过挑战的动作——结果是一整套浏览器状态，不是一个 token。

<a id="task-cloudflareturnstiletask"></a>

#### CloudFlareTurnstileTask

[Cloudflare Turnstile API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/turnstile)

Turnstile 的 `Proxy` 是可选的，返回单个 token。

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

[Akamai Web API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/akamai-web)

Akamai Web 是多轮流程：把本轮返回的 `Encodedata` 作为下一轮的 `EncodeData` 传回去。两个拼写在线格式上就是不同的，SDK 原样保留服务端的定义。

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

	fmt.Printf("第 %d 轮 payload = %s\n", index, solved.Solution.Payload)
	encodeData = solved.Solution.Encodedata
}
```

<a id="task-akamaisbsdtaskproxyless"></a>

#### AkamaiSBSDTaskProxyless

[Akamai SBSD API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/akamai-sbsd)

单轮任务，六个字段全部必填。

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

DataDome 拦截后的挑战分两步，两步共用 `DataDomeTask`，靠 `Step` 区分；`DataDomeSolution` 里哪些字段有值取决于当前是哪一步：

```go
// 第一步：从被拦截页面拿到挑战地址。
solved, err := client.SyncSolveDataDomeTaskProxyless(ctx, &ezcapsolver.DataDomeTask{
	HtmlB64: htmlB64,
	Step:    ezcapsolver.DataDomeStepOne,
})
if err != nil {
	log.Fatal(err)
}

fmt.Println("挑战地址:", solved.Solution.URL)

// 第二步换成 ezcapsolver.DataDomeStepTwo，结果里带回校验用的 Body。
```

<a id="task-datadometagstaskproxyless"></a>

#### DataDomeTagsTaskProxyless

正常浏览路径上报指纹，字段名是 camelCase，与上面的 `DataDomeTask` 不同。

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

### 其他

<a id="task-perimeterx"></a>

#### PerimeterX

[PerimeterX API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/perimeterx)

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

[Incapsula API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/incapsula)

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

[TLS 转发 API 文档](https://docs.ezxlabs.com/zh/docs/captcha/api/tls-forward)

这个类型不解验证码，而是借 worker 的 TLS 指纹发一次 HTTP 请求，把上游响应原样带回来：

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

fmt.Printf("HTTP %d，响应体 %d 字节\n", solved.Solution.Code, len(solved.Solution.Body))
```

### 结果结构

结果结构有具名字段，读结果不需要写 `solution["gRecaptchaResponse"]`。每个结构还带一个 `Extra` 映射，接住服务端后续新增的字段：

```go
if value, ok := solved.Solution.Extra["aFieldAddedLater"]; ok {
	fmt.Println(value)
}
```

worker 少返回一个字段不会导致解码失败。原始 JSON 始终在 `solved.Raw` 里；在更底层，「响应里没有 `solution` 字段」和「`solution` 的值是 JSON `null`」也始终可以区分：

```go
result, err := client.GetTaskResult(ctx, taskID)
if err != nil {
	log.Fatal(err)
}
if !result.HasSolution() {
	log.Fatal("任务还没完成")
}

var solution ezcapsolver.RecaptchaSolution
if err := result.DecodeSolution(&solution); err != nil {
	log.Fatal(err)
}
```

解码真的失败时，`*SolutionDecodeError` 会把原始值一并带上——那正是排查问题最需要的东西。

### 额外参数

每个任务模型都有 `Extra` 映射，其中的键值会被 flatten 到任务 JSON 的同一层，所以服务端后续新增的参数不用等 SDK 发版就能用。已声明字段永远优先于同名的 `Extra` 项，落败的那一项会被静默丢弃：

```go
task := &ezcapsolver.HCaptchaTask{
	WebsiteURL: "https://example.com",
	WebsiteKey: "site-key",
	Lang:       "en-US",
	Extra: map[string]any{
		// 本版本未建模的参数，或服务端之后才上线的参数。
		// 这里写已声明字段的名字会被丢弃
		"futureFlag": true,
	},
}
```

`type` 字段由 SDK 控制，不能通过这条路径设置。

### 定制化任务类型

任务类型是**开放集合**。直接把类型名当字符串传，参数用任何能序列化成 JSON 对象的值：

```go
solved, err := client.Solve(ctx, "BrandNewTaskType", map[string]any{
	"websiteURL":     "https://example.com",
	"anyFutureParam": 42,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(string(solved.Raw))

// 同步端点的对应写法：参数与返回类型完全相同，只是走另一个端点
solved, err = client.SyncSolve(ctx, "BrandNewSyncType", map[string]any{"input": "..."})
```

也可以直接解码进你自己的类型：

```go
type myShape struct {
	Token string `json:"token"`
}

solved, err := ezcapsolver.SolveAs[myShape](ctx, client, "BrandNewTaskType", params)

// 同步端点
solved, err = ezcapsolver.SyncSolveAs[myShape](ctx, client, "BrandNewSyncType", params)
```

`CreateTask`、`GetTaskResult`、`WaitForResult`、`CreateSyncTask`、`Solve` 和 `SyncSolve` 都接受 `TaskType`，而它本身就是个字符串类型——底层流程和逃生舱用的是同一套方法。`CreateTask` / `CreateSyncTask` 是两条路各自的底层原语，只在需要原始 `TaskResult` 时才直接用。如果官方提供了新的验证码类型，但 SDK 还未同步更新时，可使用此方案临时代替。

## ⚙️ 配置

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

| 配置项 | 默认值 | 作用范围 |
| --- | :---: | --- |
| `WithClientKey` | `EZCAPTCHA_API_KEY` | 不传时回落到环境变量 |
| `WithTimeout` | 30 秒 | 异步端点请求超时 |
| `WithSyncTimeout` | 240 秒 | 同步端点请求超时 |
| `WithPolling` | 3 秒 × 50 | 结果查询，每个任务最多等 150 秒 |
| `WithProxy` | 无 | **SDK 自身出网**用的代理，与任务参数里的 `Proxy` 无关 |
| `WithBaseURLs` | 生产地址 | 自建部署与集成测试 |
| `WithHTTPClient` | 新建一个 | 自定义 transport 或埋点 |
| `WithLogger` | 丢弃 | 见[日志](#-日志) |

所有可校验的配置都在构造客户端时校验，而不是首次请求时——配错了在扣费之前就会暴露。打印 `ClientConfig` 是安全的：密钥与 proxy 会被替换成 `[REDACTED]`。

两条超时预算是刻意分开的。同步调用会阻塞到 worker 返回，服务端给部分类型三分钟；共用 30 秒会掐断一个**已经扣费**、且没有 task ID 可找回结果的调用。

## ⚠️ 错误处理

按「调用方能做什么」分成五层。用 `errors.Is` 判分类，用 `errors.As` 取细节：

| 哨兵 | 类型 | 含义 |
| --- | --- | --- |
| `ErrConfig` | — | 配置非法。只由 `NewClient` 返回，原因写在消息里 |
| `ErrTransport` | `*TransportError` | 没拿到响应：DNS、TCP、TLS、超时 |
| `ErrAPI` | `*APIError` | 服务端返回了结构化错误 |
| `ErrPollingExhausted` | `*PollingExhaustedError` | 等待预算耗尽；任务可能仍会完成 |
| `ErrDecode` | `*UnexpectedResponseError`、`*SolutionDecodeError` | 响应无法使用 |


不论是哪种失败，判断「有没有一个**已扣费**的任务还能救回来」只需要问一句：

```go
if taskID := ezcapsolver.TaskIDOf(err); taskID != "" {
    // 任务在服务端，结果自创建起保留五分钟。再等一次是免费的，
    // 重新创建任务要再扣一次费。
    result, err := client.WaitForResult(ctx, taskID)
}
```

返回空串表示没有扣过费，也就没有东西需要救。

任务创建之后、等待结果期间发生的失败会被包进 `*WaitInterruptedError`，由它把 `TaskID` 带出来。`Solve` 在内部创建任务，这个包装是 task id 唯一露头的地方——拿它对同一个任务再等一次，别去创建第二个任务重复付费。包装保留了原错误，`errors.Is` 对上表哨兵的判断结果与不包装时完全一致。

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
		// 停手，别重试：这三个码累计会触发服务端封禁。
	case apiErr.IsTerminal():
		// 同样的请求重发结果一样。
	case apiErr.IsRateLimited():
		// 被限流了。这两个码都会自行解除，过一会儿再问。
	}

case errors.Is(err, ezcapsolver.ErrPollingExhausted):
	var pollErr *ezcapsolver.PollingExhaustedError
	errors.As(err, &pollErr)
	// task ID 仍然有效，之后可以用 GetTaskResult 把结果取回来。
	fmt.Println(pollErr.TaskID)
}
```

`ErrorCode` 原样暴露。worker 产生的错误码不在服务端自己那张表里——那是个开放集合——所以 SDK 不会把不认识的码归并成「未知错误」。

**`WaitForResult` 唯一会重试的就是被限流的那次轮询。** `ERROR_REQUEST_LIMIT` 与 `ERROR_REQUEST_BANNED` 拒的是**这次查询**，不是任务：服务端在查任务之前就把请求挡回来了，任务仍在排队、也仍然扣着费。所以轮询循环消耗掉这一次机会后接着问，而不是把一个马上就要拿到的结果丢掉。其余任何 API 错误都是这次轮询的答案，会直接结束等待。注意 `IsRateLimited` 比 `!IsTerminal()` 窄得多——后者对所有不认识的码也为真，其中就包括 worker 报告任务真的失败时用的那些码。

## 📝 日志

SDK 通过 `log/slog` 输出，不注入 logger 时全部丢弃。

| 级别 | 事件 |
| :---: | --- |
| `INFO` | 任务创建、任务完成 |
| `DEBUG` | 每个操作一行，每次轮询一行 |
| `LevelTrace` | 请求体，以及响应的状态码、耗时、大小和响应体 |

slog 没有 trace 级别，所以 `LevelTrace` 定在 `slog.LevelDebug` 下面一档。它是用来排查接不通的集成的：请求体渲染时会把**任意嵌套深度**的 `clientKey` 和 `proxy` 都替换掉，而且 handler 没开这个级别时根本不做渲染——所以不开它零成本。

```go
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: ezcapsolver.LevelTrace,
}))
client, err := ezcapsolver.NewClient(ezcapsolver.WithLogger(logger))
```

## 🧪 可运行示例

[`examples/`](./examples) 下每种任务类型一个可运行文件，文件名取自线格式的任务类型名。先设好 `EZCAPTCHA_API_KEY`，然后在仓库根目录运行：

```bash
export EZCAPTCHA_API_KEY=your-client-key

go run examples/recaptcha_v2/recaptcha_v2_task_proxyless.go
go run examples/cloudflare/cloud_flare_turnstile_task.go
go run examples/tls_forward/tls_task.go

# 讲 SDK 本身而非某个任务类型
go run examples/client_setup.go   # 逐项列出全部客户端配置
go run examples/raw_usage.go      # 底层的创建/轮询工作流
```

每个示例都标明了对应的任务类型，并链到官方接口文档的对应页面。需要代理的示例从 `EZCAPTCHA_PROXY` 读取；示例一律不硬编码凭据。

> 每次运行都会创建真实任务并**计费**，worker 失败也照扣。

## 🛠️ 开发

最低 Go 版本：**1.26**。地板跟随仍在收安全补丁的 Go 版本线；本地版本更低时工具链会自动拉取匹配的一个。

```bash
gofmt -l .
go vet ./...
staticcheck ./...
go test -race -cover ./...
```

示例带 `//go:build ignore`，不参与包编译，需要单独校验：

```bash
find examples -name '*.go' -exec go build -o /dev/null {} \;
```

完整流程见 [CONTRIBUTING.md](./CONTRIBUTING.md)，其中列了新增一个任务类型要改哪几处。

## 📄 许可证

基于 [Apache License 2.0](./LICENSE) 授权。
