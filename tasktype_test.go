package ezcapsolver

import (
	"reflect"
	"slices"
	"testing"
)

// solverMethods pairs every task type with the convenience method that polls it.
//
// The method name is its task type constant with TaskType swapped for Solve;
// the synchronous-endpoint sibling is the same name behind a Sync prefix.
// Keeping that rule mechanical is what lets a reader go from either name to the
// other without a lookup, and this table is what holds the SDK to it.
var solverMethods = map[TaskType]string{
	TaskTypeRecaptchaV2TaskProxyless:             "SolveRecaptchaV2TaskProxyless",
	TaskTypeRecaptchaV2TaskProxylessS9:           "SolveRecaptchaV2TaskProxylessS9",
	TaskTypeRecaptchaV2STaskProxyless:            "SolveRecaptchaV2STaskProxyless",
	TaskTypeRecaptchaV2EnterpriseTaskProxyless:   "SolveRecaptchaV2EnterpriseTaskProxyless",
	TaskTypeRecaptchaV2SEnterpriseTaskProxyless:  "SolveRecaptchaV2SEnterpriseTaskProxyless",
	TaskTypeRecaptchaV2Classification:            "SolveRecaptchaV2Classification",
	TaskTypeRecaptchaV3TaskProxyless:             "SolveRecaptchaV3TaskProxyless",
	TaskTypeRecaptchaV3TaskProxylessS9:           "SolveRecaptchaV3TaskProxylessS9",
	TaskTypeRecaptchaV3EnterpriseTaskProxyless:   "SolveRecaptchaV3EnterpriseTaskProxyless",
	TaskTypeRecaptchaV3EnterpriseTaskProxylessS9: "SolveRecaptchaV3EnterpriseTaskProxylessS9",
	TaskTypeFuncaptchaTaskProxyless:              "SolveFuncaptchaTaskProxyless",
	TaskTypeFunCaptchaClassification:             "SolveFunCaptchaClassification",
	TaskTypePerimeterX:                           "SolvePerimeterX",
	TaskTypeHCaptcha:                             "SolveHCaptcha",
	TaskTypeHCaptchaClassification:               "SolveHCaptchaClassification",
	TaskTypeAkamaiWEBTaskProxyless:               "SolveAkamaiWEBTaskProxyless",
	TaskTypeAkamaiSBSDTaskProxyless:              "SolveAkamaiSBSDTaskProxyless",
	TaskTypeTLSTask:                              "SolveTLSTask",
	TaskTypeCloudFlare5STask:                     "SolveCloudFlare5STask",
	TaskTypeCloudFlareTurnstileTask:              "SolveCloudFlareTurnstileTask",
	TaskTypeDataDomeTaskProxyless:                "SolveDataDomeTaskProxyless",
	TaskTypeDataDomeTagsTaskProxyless:            "SolveDataDomeTagsTaskProxyless",
	TaskTypeIncapsulaTaskProxyless:               "SolveIncapsulaTaskProxyless",
}

// TestCanonicalTaskTypeSpellings pins the wire value of all 23 types.
//
// The service matches these case-insensitively, so a misspelling still works —
// which is exactly why it would go unnoticed. The SDK is published documentation,
// so it sends the canonical form, irregular capitalisation included.
func TestCanonicalTaskTypeSpellings(t *testing.T) {
	want := []TaskType{
		"ReCaptchaV2TaskProxyless",
		"ReCaptchaV2TaskProxylessS9",
		"ReCaptchaV2STaskProxyless",
		"ReCaptchaV2EnterpriseTaskProxyless",
		"ReCaptchaV2SEnterpriseTaskProxyless",
		"ReCaptchaV2Classification",
		"ReCaptchaV3TaskProxyless",
		"ReCaptchaV3TaskProxylessS9",
		"ReCaptchaV3EnterpriseTaskProxyless",
		// The catalog writes this one with a lower-case c. Every SDK normalises
		// it so the whole ReCaptcha family shares one capitalisation; the service
		// matches case-insensitively. Do not "fix" it back.
		"ReCaptchaV3EnterpriseTaskProxylessS9",
		// Lower-case c, while the classification type below has an upper-case one.
		"FuncaptchaTaskProxyless",
		"FunCaptchaClassification",
		"PerimeterX",
		"HCaptcha",
		"HCaptchaClassification",
		"AkamaiWEBTaskProxyless",
		"AkamaiSBSDTaskProxyless",
		"TlsTask",
		"CloudFlare5STask",
		"CloudFlareTurnstileTask",
		"DataDomeTaskProxyless",
		"DataDomeTagsTaskProxyless",
		"IncapsulaTaskProxyless",
	}

	if !slices.Equal(KnownTaskTypes, want) {
		t.Errorf("known task types drifted\n got: %v\nwant: %v", KnownTaskTypes, want)
	}
	if len(KnownTaskTypes) != 23 {
		t.Errorf("the catalog lists 23 supported types, this release has %d", len(KnownTaskTypes))
	}
}

// TestRetiredTypesAreAbsent checks that the types the catalog rules out never
// appear.
//
// Four are retired and are never coming back; two are new and not yet stable.
// Both groups are excluded for now, and the second group will be added as a
// non-breaking change once its contract settles.
func TestRetiredTypesAreAbsent(t *testing.T) {
	for _, excluded := range []TaskType{
		"TlsForward2",
		"AkamaiBMPTaskProxyless",
		"KasadaTaskProxyless",
		"KasadaWorkTimeTaskProxyless",
		"Other",
		"ReCaptchaV2SyncTaskProxyless",
		"ReCaptchaV3EnterpriseSyncTaskProxyless",
	} {
		if excluded.IsKnown() {
			t.Errorf("%s must not be modelled", excluded)
		}
		if _, present := solverMethods[excluded]; present {
			t.Errorf("%s must not have a convenience method", excluded)
		}
	}
}

// TestTaskModeSplit checks which endpoint each type runs through. Getting this
// wrong sends a task to an endpoint that rejects it outright.
func TestTaskModeSplit(t *testing.T) {
	syncTypes := []TaskType{
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

	if len(syncTypes) != 9 {
		t.Fatalf("the catalog lists 9 synchronous types, this table has %d", len(syncTypes))
	}
	for _, taskType := range KnownTaskTypes {
		mode, known := taskType.Mode()
		if !known {
			t.Errorf("%s is listed as known but has no mode", taskType)
			continue
		}
		want := ModeAsync
		if slices.Contains(syncTypes, taskType) {
			want = ModeSync
		}
		if mode != want {
			t.Errorf("%s: got %s, want %s", taskType, mode, want)
		}
	}

	t.Run("an unmodelled type reports that it is unknown", func(t *testing.T) {
		// Guessing async for a type the SDK has never seen would send a
		// synchronous task to the wrong endpoint, so the caller is told instead.
		if _, known := TaskType("BrandNewTaskType").Mode(); known {
			t.Error("an unknown type must not claim a mode")
		}
	})
}

// TestEveryTaskTypeHasASolver checks the promise that each task type has two
// convenience methods -- one per endpoint -- taking its request model and
// returning its solution model.
func TestEveryTaskTypeHasASolver(t *testing.T) {
	if len(solverMethods) != len(KnownTaskTypes) {
		t.Fatalf("%d methods for %d task types", len(solverMethods), len(KnownTaskTypes))
	}

	clientType := reflect.TypeOf(&EzCapSolverClient{})
	for _, taskType := range KnownTaskTypes {
		name, mapped := solverMethods[taskType]
		if !mapped {
			t.Errorf("%s has no convenience method", taskType)
			continue
		}

		method, exists := clientType.MethodByName(name)
		if !exists {
			t.Errorf("%s: method %s is missing", taskType, name)
			continue
		}
		// Both endpoints must be reachable, and switching between them must
		// never mean rewriting the call: identical arguments, identical result.
		syncMethod, syncExists := clientType.MethodByName("Sync" + name)
		if !syncExists {
			t.Errorf("%s: method Sync%s is missing", taskType, name)
			continue
		}
		if syncMethod.Type != method.Type {
			t.Errorf("Sync%s has signature %s, want %s", name, syncMethod.Type, method.Type)
		}
		// Receiver, context and task in; result and error out.
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 2 {
			t.Errorf("%s: unexpected signature %s", name, method.Type)
			continue
		}
		if got := method.Type.In(1).String(); got != "context.Context" {
			t.Errorf("%s: first parameter is %s, want context.Context", name, got)
		}
		if method.Type.In(2).Kind() != reflect.Pointer {
			t.Errorf("%s: the task must be passed by pointer, got %s", name, method.Type.In(2))
		}
	}
}

// TestTaskTypeIsAnOpenSet checks the escape hatch: a type the service adds after
// this release still works, without waiting for an SDK update.
func TestTaskTypeIsAnOpenSet(t *testing.T) {
	future := TaskType("BrandNewTaskType")

	if future.IsKnown() {
		t.Error("an unmodelled type must not claim to be known")
	}
	if future.String() != "BrandNewTaskType" {
		t.Errorf("got %q", future)
	}
	if got := marshalToMap(t, taskPayload{taskType: future})["type"]; got != "BrandNewTaskType" {
		t.Errorf("an unmodelled type must still reach the wire, got %v", got)
	}
}
