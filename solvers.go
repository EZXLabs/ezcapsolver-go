package ezcapsolver

import "context"

// Two convenience methods per task type, one per endpoint. Both take that
// type's request model and return that type's solution model.
//
// SolveX creates the task and polls for its result; SyncSolveX runs the same
// task through the synchronous endpoint, which answers on the request that
// created it. Arguments and result types are identical; the only difference is
// [Solved.TaskID], which the synchronous endpoint never assigns and so leaves
// empty.
//
// Neither name is a default — each says which endpoint it uses, so a caller
// never has to look up what a type is classified as. Which endpoint the service
// documents for a type is on [TaskType.Mode], and that value routes nothing:
// it matters because the service may reject a type on the endpoint it does not
// serve, with ERROR_TASK_TYPE_NOT_ALLOWED on the synchronous side. Such a
// rejection happens before the task is billed, so it costs a round trip rather
// than a task.
//
// Every method name matches its task type constant: TaskTypeX pairs with SolveX
// and SyncSolveX.
//
// For per-call polling settings, or for a task type this release does not model,
// use [EzCapSolverClient.SolveWith], [EzCapSolverClient.CreateSyncTask] or [SolveAs].

// -- ReCaptcha V2 ------------------------------------------------------------

// SolveRecaptchaV2TaskProxyless solves a ReCaptcha V2 challenge.
func (c *EzCapSolverClient) SolveRecaptchaV2TaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV2TaskProxyless, task, c.config.Polling)
}

// SyncSolveRecaptchaV2TaskProxyless solves the same task as
// [EzCapSolverClient.SolveRecaptchaV2TaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV2TaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV2TaskProxyless, task)
}

// SolveRecaptchaV2TaskProxylessS9 solves a ReCaptcha V2 challenge on the
// high-score queue, where the returned token scores at least 0.9.
func (c *EzCapSolverClient) SolveRecaptchaV2TaskProxylessS9(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV2TaskProxylessS9, task, c.config.Polling)
}

// SyncSolveRecaptchaV2TaskProxylessS9 solves the same task as
// [EzCapSolverClient.SolveRecaptchaV2TaskProxylessS9] on the synchronous
// endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV2TaskProxylessS9(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV2TaskProxylessS9, task)
}

// SolveRecaptchaV2STaskProxyless solves a ReCaptcha V2 challenge that carries an
// `s` parameter, which routes it to the high-score IPv4 queue.
//
// The `s` parameter is not actually mandatory for this type.
func (c *EzCapSolverClient) SolveRecaptchaV2STaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV2STaskProxyless, task, c.config.Polling)
}

// SyncSolveRecaptchaV2STaskProxyless solves the same task as
// [EzCapSolverClient.SolveRecaptchaV2STaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV2STaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV2STaskProxyless, task)
}

// SolveRecaptchaV2EnterpriseTaskProxyless solves a ReCaptcha V2 Enterprise
// challenge.
func (c *EzCapSolverClient) SolveRecaptchaV2EnterpriseTaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV2EnterpriseTaskProxyless, task, c.config.Polling)
}

// SyncSolveRecaptchaV2EnterpriseTaskProxyless solves the same task as
// [EzCapSolverClient.SolveRecaptchaV2EnterpriseTaskProxyless] on the synchronous
// endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV2EnterpriseTaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV2EnterpriseTaskProxyless, task)
}

// SolveRecaptchaV2SEnterpriseTaskProxyless solves a ReCaptcha V2 Enterprise
// challenge that carries an `s` parameter.
func (c *EzCapSolverClient) SolveRecaptchaV2SEnterpriseTaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV2SEnterpriseTaskProxyless, task, c.config.Polling)
}

// SyncSolveRecaptchaV2SEnterpriseTaskProxyless solves the same task as
// [EzCapSolverClient.SolveRecaptchaV2SEnterpriseTaskProxyless] on the synchronous
// endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV2SEnterpriseTaskProxyless(
	ctx context.Context, task *RecaptchaV2Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV2SEnterpriseTaskProxyless, task)
}

// SolveRecaptchaV2Classification reads one ReCaptcha V2 image grid.
//
// Inspect the result with [ReClassificationSolution.IsMulti] or
// [ReClassificationSolution.IsSingle], then read its fields directly.
func (c *EzCapSolverClient) SolveRecaptchaV2Classification(
	ctx context.Context, task *RecaptchaV2ClassificationTask,
) (*Solved[ReClassificationSolution], error) {
	return solveTyped[ReClassificationSolution](
		ctx, c, TaskTypeRecaptchaV2Classification, task, c.config.Polling)
}

// SyncSolveRecaptchaV2Classification solves the same task as
// [EzCapSolverClient.SolveRecaptchaV2Classification] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV2Classification(
	ctx context.Context, task *RecaptchaV2ClassificationTask,
) (*Solved[ReClassificationSolution], error) {
	return solveSyncTyped[ReClassificationSolution](ctx, c, TaskTypeRecaptchaV2Classification, task)
}

// -- ReCaptcha V3 ------------------------------------------------------------

// SolveRecaptchaV3TaskProxyless solves a ReCaptcha V3 challenge.
func (c *EzCapSolverClient) SolveRecaptchaV3TaskProxyless(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV3TaskProxyless, task, c.config.Polling)
}

// SyncSolveRecaptchaV3TaskProxyless solves the same task as
// [EzCapSolverClient.SolveRecaptchaV3TaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV3TaskProxyless(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV3TaskProxyless, task)
}

// SolveRecaptchaV3TaskProxylessS9 solves a ReCaptcha V3 challenge on the
// high-score queue.
func (c *EzCapSolverClient) SolveRecaptchaV3TaskProxylessS9(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV3TaskProxylessS9, task, c.config.Polling)
}

// SyncSolveRecaptchaV3TaskProxylessS9 solves the same task as
// [EzCapSolverClient.SolveRecaptchaV3TaskProxylessS9] on the synchronous
// endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV3TaskProxylessS9(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV3TaskProxylessS9, task)
}

// SolveRecaptchaV3EnterpriseTaskProxyless solves a ReCaptcha V3 Enterprise
// challenge.
func (c *EzCapSolverClient) SolveRecaptchaV3EnterpriseTaskProxyless(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV3EnterpriseTaskProxyless, task, c.config.Polling)
}

// SyncSolveRecaptchaV3EnterpriseTaskProxyless solves the same task as
// [EzCapSolverClient.SolveRecaptchaV3EnterpriseTaskProxyless] on the synchronous
// endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV3EnterpriseTaskProxyless(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV3EnterpriseTaskProxyless, task)
}

// SolveRecaptchaV3EnterpriseTaskProxylessS9 solves a ReCaptcha V3 Enterprise
// challenge on the high-score queue.
func (c *EzCapSolverClient) SolveRecaptchaV3EnterpriseTaskProxylessS9(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveTyped[RecaptchaSolution](
		ctx, c, TaskTypeRecaptchaV3EnterpriseTaskProxylessS9, task, c.config.Polling)
}

// SyncSolveRecaptchaV3EnterpriseTaskProxylessS9 solves the same task as
// [EzCapSolverClient.SolveRecaptchaV3EnterpriseTaskProxylessS9] on the
// synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveRecaptchaV3EnterpriseTaskProxylessS9(
	ctx context.Context, task *RecaptchaV3Task,
) (*Solved[RecaptchaSolution], error) {
	return solveSyncTyped[RecaptchaSolution](ctx, c, TaskTypeRecaptchaV3EnterpriseTaskProxylessS9, task)
}

// -- FunCaptcha --------------------------------------------------------------

// SolveFuncaptchaTaskProxyless solves a FunCaptcha (Arkose Labs) challenge.
//
// This is the one task type whose proxy uses the `FUN` format; see
// [FunCaptchaTask.Proxy].
func (c *EzCapSolverClient) SolveFuncaptchaTaskProxyless(
	ctx context.Context, task *FunCaptchaTask,
) (*Solved[FunCaptchaSolution], error) {
	return solveTyped[FunCaptchaSolution](
		ctx, c, TaskTypeFuncaptchaTaskProxyless, task, c.config.Polling)
}

// SyncSolveFuncaptchaTaskProxyless solves the same task as
// [EzCapSolverClient.SolveFuncaptchaTaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveFuncaptchaTaskProxyless(
	ctx context.Context, task *FunCaptchaTask,
) (*Solved[FunCaptchaSolution], error) {
	return solveSyncTyped[FunCaptchaSolution](ctx, c, TaskTypeFuncaptchaTaskProxyless, task)
}

// SolveFunCaptchaClassification reads one FunCaptcha image.
//
// The solution shape is not confirmed yet, so everything the worker returns
// arrives in [FunCaptchaClassificationSolution.Extra] and in [Solved.Raw].
func (c *EzCapSolverClient) SolveFunCaptchaClassification(
	ctx context.Context, task *FunCaptchaClassificationTask,
) (*Solved[FunCaptchaClassificationSolution], error) {
	return solveTyped[FunCaptchaClassificationSolution](
		ctx, c, TaskTypeFunCaptchaClassification, task, c.config.Polling)
}

// SyncSolveFunCaptchaClassification solves the same task as
// [EzCapSolverClient.SolveFunCaptchaClassification] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveFunCaptchaClassification(
	ctx context.Context, task *FunCaptchaClassificationTask,
) (*Solved[FunCaptchaClassificationSolution], error) {
	return solveSyncTyped[FunCaptchaClassificationSolution](ctx, c, TaskTypeFunCaptchaClassification, task)
}

// -- HCaptcha ----------------------------------------------------------------

// SolveHCaptcha solves an HCaptcha challenge.
func (c *EzCapSolverClient) SolveHCaptcha(
	ctx context.Context, task *HCaptchaTask,
) (*Solved[HCaptchaSolution], error) {
	return solveTyped[HCaptchaSolution](
		ctx, c, TaskTypeHCaptcha, task, c.config.Polling)
}

// SyncSolveHCaptcha solves the same task as [EzCapSolverClient.SolveHCaptcha] on
// the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveHCaptcha(
	ctx context.Context, task *HCaptchaTask,
) (*Solved[HCaptchaSolution], error) {
	return solveSyncTyped[HCaptchaSolution](ctx, c, TaskTypeHCaptcha, task)
}

// SolveHCaptchaClassification reads one or more HCaptcha images.
//
// The solution shape is not confirmed yet, so everything the worker returns
// arrives in [HCaptchaClassificationSolution.Extra] and in [Solved.Raw].
func (c *EzCapSolverClient) SolveHCaptchaClassification(
	ctx context.Context, task *HCaptchaClassificationTask,
) (*Solved[HCaptchaClassificationSolution], error) {
	return solveTyped[HCaptchaClassificationSolution](
		ctx, c, TaskTypeHCaptchaClassification, task, c.config.Polling)
}

// SyncSolveHCaptchaClassification solves the same task as
// [EzCapSolverClient.SolveHCaptchaClassification] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveHCaptchaClassification(
	ctx context.Context, task *HCaptchaClassificationTask,
) (*Solved[HCaptchaClassificationSolution], error) {
	return solveSyncTyped[HCaptchaClassificationSolution](ctx, c, TaskTypeHCaptchaClassification, task)
}

// -- PerimeterX --------------------------------------------------------------

// SolvePerimeterX obtains PerimeterX clearance cookies.
func (c *EzCapSolverClient) SolvePerimeterX(
	ctx context.Context, task *PerimeterXTask,
) (*Solved[PerimeterXSolution], error) {
	return solveTyped[PerimeterXSolution](
		ctx, c, TaskTypePerimeterX, task, c.config.Polling)
}

// SyncSolvePerimeterX solves the same task as [EzCapSolverClient.SolvePerimeterX]
// on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolvePerimeterX(
	ctx context.Context, task *PerimeterXTask,
) (*Solved[PerimeterXSolution], error) {
	return solveSyncTyped[PerimeterXSolution](ctx, c, TaskTypePerimeterX, task)
}

// -- Cloudflare --------------------------------------------------------------

// SolveCloudFlare5STask clears a Cloudflare five-second interstitial.
func (c *EzCapSolverClient) SolveCloudFlare5STask(
	ctx context.Context, task *Cloudflare5sTask,
) (*Solved[Cloudflare5sSolution], error) {
	return solveTyped[Cloudflare5sSolution](
		ctx, c, TaskTypeCloudFlare5STask, task, c.config.Polling)
}

// SyncSolveCloudFlare5STask solves the same task as
// [EzCapSolverClient.SolveCloudFlare5STask] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveCloudFlare5STask(
	ctx context.Context, task *Cloudflare5sTask,
) (*Solved[Cloudflare5sSolution], error) {
	return solveSyncTyped[Cloudflare5sSolution](ctx, c, TaskTypeCloudFlare5STask, task)
}

// SolveCloudFlareTurnstileTask obtains a Cloudflare Turnstile token.
func (c *EzCapSolverClient) SolveCloudFlareTurnstileTask(
	ctx context.Context, task *CloudflareTurnstileTask,
) (*Solved[CloudflareTurnstileSolution], error) {
	return solveTyped[CloudflareTurnstileSolution](
		ctx, c, TaskTypeCloudFlareTurnstileTask, task, c.config.Polling)
}

// SyncSolveCloudFlareTurnstileTask solves the same task as
// [EzCapSolverClient.SolveCloudFlareTurnstileTask] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveCloudFlareTurnstileTask(
	ctx context.Context, task *CloudflareTurnstileTask,
) (*Solved[CloudflareTurnstileSolution], error) {
	return solveSyncTyped[CloudflareTurnstileSolution](ctx, c, TaskTypeCloudFlareTurnstileTask, task)
}

// -- Akamai ------------------------------------------------------------------

// SolveAkamaiWEBTaskProxyless produces one round of Akamai Web sensor data.
//
// This is a multi-round flow; see [AkamaiWebTask].
func (c *EzCapSolverClient) SolveAkamaiWEBTaskProxyless(
	ctx context.Context, task *AkamaiWebTask,
) (*Solved[AkamaiWebSolution], error) {
	return solveTyped[AkamaiWebSolution](
		ctx, c, TaskTypeAkamaiWEBTaskProxyless, task, c.config.Polling)
}

// SyncSolveAkamaiWEBTaskProxyless solves the same task as
// [EzCapSolverClient.SolveAkamaiWEBTaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveAkamaiWEBTaskProxyless(
	ctx context.Context, task *AkamaiWebTask,
) (*Solved[AkamaiWebSolution], error) {
	return solveSyncTyped[AkamaiWebSolution](ctx, c, TaskTypeAkamaiWEBTaskProxyless, task)
}

// SolveAkamaiSBSDTaskProxyless produces Akamai SBSD sensor data.
func (c *EzCapSolverClient) SolveAkamaiSBSDTaskProxyless(
	ctx context.Context, task *AkamaiSBSDTask,
) (*Solved[AkamaiSBSDSolution], error) {
	return solveTyped[AkamaiSBSDSolution](
		ctx, c, TaskTypeAkamaiSBSDTaskProxyless, task, c.config.Polling)
}

// SyncSolveAkamaiSBSDTaskProxyless solves the same task as
// [EzCapSolverClient.SolveAkamaiSBSDTaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveAkamaiSBSDTaskProxyless(
	ctx context.Context, task *AkamaiSBSDTask,
) (*Solved[AkamaiSBSDSolution], error) {
	return solveSyncTyped[AkamaiSBSDSolution](ctx, c, TaskTypeAkamaiSBSDTaskProxyless, task)
}

// -- DataDome ----------------------------------------------------------------

// SolveDataDomeTaskProxyless answers a DataDome challenge.
//
// Both steps of the flow return a [DataDomeSolution]; which fields are filled in
// depends on [DataDomeTask.Step].
func (c *EzCapSolverClient) SolveDataDomeTaskProxyless(
	ctx context.Context, task *DataDomeTask,
) (*Solved[DataDomeSolution], error) {
	return solveTyped[DataDomeSolution](
		ctx, c, TaskTypeDataDomeTaskProxyless, task, c.config.Polling)
}

// SyncSolveDataDomeTaskProxyless solves the same task as
// [EzCapSolverClient.SolveDataDomeTaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveDataDomeTaskProxyless(
	ctx context.Context, task *DataDomeTask,
) (*Solved[DataDomeSolution], error) {
	return solveSyncTyped[DataDomeSolution](ctx, c, TaskTypeDataDomeTaskProxyless, task)
}

// SolveDataDomeTagsTaskProxyless reports a fingerprint to DataDome on the normal
// browsing path.
func (c *EzCapSolverClient) SolveDataDomeTagsTaskProxyless(
	ctx context.Context, task *DataDomeTagsTask,
) (*Solved[DataDomeSolution], error) {
	return solveTyped[DataDomeSolution](
		ctx, c, TaskTypeDataDomeTagsTaskProxyless, task, c.config.Polling)
}

// SyncSolveDataDomeTagsTaskProxyless solves the same task as
// [EzCapSolverClient.SolveDataDomeTagsTaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveDataDomeTagsTaskProxyless(
	ctx context.Context, task *DataDomeTagsTask,
) (*Solved[DataDomeSolution], error) {
	return solveSyncTyped[DataDomeSolution](ctx, c, TaskTypeDataDomeTagsTaskProxyless, task)
}

// -- Incapsula ---------------------------------------------------------------

// SolveIncapsulaTaskProxyless produces an Incapsula Reese84 payload.
func (c *EzCapSolverClient) SolveIncapsulaTaskProxyless(
	ctx context.Context, task *IncapsulaTask,
) (*Solved[IncapsulaSolution], error) {
	return solveTyped[IncapsulaSolution](
		ctx, c, TaskTypeIncapsulaTaskProxyless, task, c.config.Polling)
}

// SyncSolveIncapsulaTaskProxyless solves the same task as
// [EzCapSolverClient.SolveIncapsulaTaskProxyless] on the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveIncapsulaTaskProxyless(
	ctx context.Context, task *IncapsulaTask,
) (*Solved[IncapsulaSolution], error) {
	return solveSyncTyped[IncapsulaSolution](ctx, c, TaskTypeIncapsulaTaskProxyless, task)
}

// -- TLS forwarding ----------------------------------------------------------

// SolveTLSTask forwards one HTTP request through a worker's TLS fingerprint and
// returns the upstream response.
func (c *EzCapSolverClient) SolveTLSTask(
	ctx context.Context, task *TLSForwardTask,
) (*Solved[TLSForwardSolution], error) {
	return solveTyped[TLSForwardSolution](
		ctx, c, TaskTypeTLSTask, task, c.config.Polling)
}

// SyncSolveTLSTask solves the same task as [EzCapSolverClient.SolveTLSTask] on
// the synchronous endpoint.
func (c *EzCapSolverClient) SyncSolveTLSTask(
	ctx context.Context, task *TLSForwardTask,
) (*Solved[TLSForwardSolution], error) {
	return solveSyncTyped[TLSForwardSolution](ctx, c, TaskTypeTLSTask, task)
}
