package ezcapsolver

import "slices"

// TaskType is the value sent in a task's `type` field.
//
// It is a string rather than a closed enum on purpose: a task type the service
// adds after this release still works, without waiting for an SDK update.
//
//	client.Solve(ctx, ezcapsolver.TaskType("BrandNewTaskType"), params)
//
// The wire values below carry the service's own irregularities —
// FuncaptchaTaskProxyless has a lowercase c while FunCaptchaClassification does
// not, and AkamaiWEBTaskProxyless is shouted. The SDK is part of the published
// documentation, so it sends these as the task catalog spells them, and
// TestCanonicalTaskTypeSpellings pins all 23 of them.
//
// One is deliberately normalised, in every language SDK alike: the catalog
// writes the V3 Enterprise S9 type as RecaptchaV3EnterpriseTaskProxylessS9, and
// the SDKs send ReCaptchaV3EnterpriseTaskProxylessS9 so that one capitalisation
// runs through the whole ReCaptcha family. The service matches task types
// case-insensitively, so it reaches the same worker.
//
// The Go identifiers track those spellings, with two deliberate exceptions
// where the wire form fights a Go convention: TlsTask is TaskTypeTLSTask
// because TLS is an initialism, and the ReCaptcha family is spelled Recaptcha
// so that one capitalisation runs through the constants, the convenience
// methods and the request models alike. Neither exception reaches the wire.
type TaskType string

// Task types this release models. Each maps to one request model and one
// solution model; several types share a model where the parameters are the same.
const (
	// Asynchronous types, created with /createTask and then polled.

	TaskTypeRecaptchaV2TaskProxyless             TaskType = "ReCaptchaV2TaskProxyless"
	TaskTypeRecaptchaV2TaskProxylessS9           TaskType = "ReCaptchaV2TaskProxylessS9"
	TaskTypeRecaptchaV2STaskProxyless            TaskType = "ReCaptchaV2STaskProxyless"
	TaskTypeRecaptchaV2EnterpriseTaskProxyless   TaskType = "ReCaptchaV2EnterpriseTaskProxyless"
	TaskTypeRecaptchaV2SEnterpriseTaskProxyless  TaskType = "ReCaptchaV2SEnterpriseTaskProxyless"
	TaskTypeRecaptchaV3TaskProxyless             TaskType = "ReCaptchaV3TaskProxyless"
	TaskTypeRecaptchaV3TaskProxylessS9           TaskType = "ReCaptchaV3TaskProxylessS9"
	TaskTypeRecaptchaV3EnterpriseTaskProxyless   TaskType = "ReCaptchaV3EnterpriseTaskProxyless"
	TaskTypeRecaptchaV3EnterpriseTaskProxylessS9 TaskType = "ReCaptchaV3EnterpriseTaskProxylessS9"
	TaskTypeFuncaptchaTaskProxyless              TaskType = "FuncaptchaTaskProxyless"
	TaskTypePerimeterX                           TaskType = "PerimeterX"
	TaskTypeHCaptcha                             TaskType = "HCaptcha"
	TaskTypeCloudFlare5STask                     TaskType = "CloudFlare5STask"
	TaskTypeCloudFlareTurnstileTask              TaskType = "CloudFlareTurnstileTask"

	// Synchronous types, executed with /createSyncTask, which returns the
	// terminal result inline and no task ID.

	TaskTypeRecaptchaV2Classification TaskType = "ReCaptchaV2Classification"
	TaskTypeFunCaptchaClassification  TaskType = "FunCaptchaClassification"
	TaskTypeHCaptchaClassification    TaskType = "HCaptchaClassification"
	TaskTypeAkamaiWEBTaskProxyless    TaskType = "AkamaiWEBTaskProxyless"
	TaskTypeAkamaiSBSDTaskProxyless   TaskType = "AkamaiSBSDTaskProxyless"
	TaskTypeTLSTask                   TaskType = "TlsTask"
	TaskTypeDataDomeTaskProxyless     TaskType = "DataDomeTaskProxyless"
	TaskTypeDataDomeTagsTaskProxyless TaskType = "DataDomeTagsTaskProxyless"
	TaskTypeIncapsulaTaskProxyless    TaskType = "IncapsulaTaskProxyless"
)

// TaskMode is the endpoint a task type is executed through.
type TaskMode int

const (
	// ModeAsync means the task is created and then polled for its result.
	ModeAsync TaskMode = iota
	// ModeSync means the task returns its result from the request that created it.
	ModeSync
)

func (m TaskMode) String() string {
	if m == ModeSync {
		return "sync"
	}
	return "async"
}

// KnownTaskTypes lists every task type this release models, in the order the
// task catalog documents them.
//
// Being listed here says nothing about availability: whether a type can be used
// depends on service-side configuration and on the caller's plan, which is not
// something the SDK can predict.
var KnownTaskTypes = []TaskType{
	TaskTypeRecaptchaV2TaskProxyless,
	TaskTypeRecaptchaV2TaskProxylessS9,
	TaskTypeRecaptchaV2STaskProxyless,
	TaskTypeRecaptchaV2EnterpriseTaskProxyless,
	TaskTypeRecaptchaV2SEnterpriseTaskProxyless,
	TaskTypeRecaptchaV2Classification,
	TaskTypeRecaptchaV3TaskProxyless,
	TaskTypeRecaptchaV3TaskProxylessS9,
	TaskTypeRecaptchaV3EnterpriseTaskProxyless,
	TaskTypeRecaptchaV3EnterpriseTaskProxylessS9,
	TaskTypeFuncaptchaTaskProxyless,
	TaskTypeFunCaptchaClassification,
	TaskTypePerimeterX,
	TaskTypeHCaptcha,
	TaskTypeHCaptchaClassification,
	TaskTypeAkamaiWEBTaskProxyless,
	TaskTypeAkamaiSBSDTaskProxyless,
	TaskTypeTLSTask,
	TaskTypeCloudFlare5STask,
	TaskTypeCloudFlareTurnstileTask,
	TaskTypeDataDomeTaskProxyless,
	TaskTypeDataDomeTagsTaskProxyless,
	TaskTypeIncapsulaTaskProxyless,
}

// syncTaskTypes is the subset executed through /createSyncTask.
var syncTaskTypes = []TaskType{
	TaskTypeRecaptchaV2Classification,
	TaskTypeFunCaptchaClassification,
	TaskTypeHCaptchaClassification,
	TaskTypeAkamaiWEBTaskProxyless,
	TaskTypeAkamaiSBSDTaskProxyless,
	TaskTypeTLSTask,
	TaskTypeDataDomeTaskProxyless,
	TaskTypeDataDomeTagsTaskProxyless,
	TaskTypeIncapsulaTaskProxyless,
}

func (t TaskType) String() string { return string(t) }

// IsKnown reports whether this release models the type.
func (t TaskType) IsKnown() bool { return slices.Contains(KnownTaskTypes, t) }

// Mode returns the endpoint the service documents for the type, and whether
// this release knows the type at all.
//
// This is informational. Every type has both a SolveX method that polls and a
// SyncSolveX method that uses the synchronous endpoint, and nothing in the SDK
// reads this value to pick between them.
//
// It matters because the service may reject a type on the endpoint it does not
// serve — so this is the mode to follow when there is no reason to prefer the
// other. Such a rejection is refused before billing, so it costs a round trip
// rather than a task.
func (t TaskType) Mode() (TaskMode, bool) {
	if slices.Contains(syncTaskTypes, t) {
		return ModeSync, true
	}
	if t.IsKnown() {
		return ModeAsync, true
	}
	return ModeAsync, false
}
