# DramaGo Media Reliability P4

## Goal

Migrate the reliability semantics proven in Xiaoshuren-Media-MCP into DramaGo's existing generation/media path without importing its parallel Job / ProviderExecution / Asset persistence model.

DramaGo remains the only source of truth:

```text
GenerationTask
GenerationTaskAttempt
Asset
Provider adapters
```

No new reliability table is introduced.

## Reused existing DramaGo fields

GenerationTask already owns the fields required by the reliability model:

- `ProviderTaskID`
- `Status`
- `Error`
- `ErrorCode`
- `ErrorType`
- `Retryable`
- attempt history

Asset already owns:

- `ContentHash`
- `MIMEType`
- `SizeBytes`
- source URL and storage state

P4 adds behavior around these existing fields rather than duplicating them.

## 1. Idempotent generation submission

The public generation request accepts:

```json
{
  "idempotencyKey": "16..128 chars"
}
```

The key is scoped to the existing project/conversation and combined with a SHA-256 semantic request hash.

Internal metadata is persisted in existing `GenerationTask.Params`:

- `_mediago_idempotency_key`
- `_mediago_request_hash`

These internal params are stripped from provider payloads and client-facing params.

### Semantics

Same key + same semantic request:

- returns the first existing GenerationTask;
- never launches a second provider submission.

Same key + different semantic request:

- returns HTTP 409 idempotency conflict;
- never launches a provider submission.

The local task ID is deterministic from scope + idempotency key.

Reservation uses a database-level primary-key insert with `ON CONFLICT DO NOTHING`, so correctness does not depend on one process-local mutex.

The UI supplies one fresh key per user submission. Intentional "generate again" gets a new key. Batch submissions assign a different key to every child item.

## 2. Provider submission uncertainty

P4 distinguishes provider failures from ambiguous submission outcomes.

Ambiguous conditions include provider timeouts and retryable/provider 5xx failures occurring across the submission boundary.

### Statuses

```text
unknown
```

The request may have reached the provider, but DramaGo has no provider task ID. Automatic retry is prohibited because it could create a duplicate remote generation.

```text
reconciling
```

The submission outcome is uncertain, but a ProviderTaskID exists. The normal polling path may query that ID and recover the task to running/completed/failed.

Both statuses preserve:

- `errorCode=provider_submission_unknown`
- `errorType=provider_unknown`
- raw provider error detail

Generic retry is disabled for these statuses. Explicit retry returns conflict until provider state is reconciled.

A `reconciling` task with a ProviderTaskID is included in the existing background polling path. An `unknown` task without a ProviderTaskID is not blindly resubmitted or converted into a normal retry timeout.

## 3. Error normalization

Existing DramaGo failure mapping remains authoritative.

Provider errors continue to retain:

- normalized error code;
- normalized error type/failure reason;
- retryable flag;
- raw provider detail for diagnostics.

P4 does not create a separate Media-MCP error model.

## 4. Remote media ingestion

Generated remote media still enters the existing `MediaAssets.SaveRemoteAssetWithOptions` -> `assets` path.

Before bytes are persisted, P4 now enforces:

1. HTTPS by default;
2. no URL credentials;
3. localhost/private/link-local/metadata/reserved address rejection;
4. the same URL policy for every redirect;
5. redirect cap;
6. declared Content-Length cap;
7. streamed byte cap;
8. MIME is derived/validated from actual media signature;
9. declared media kind must match actual bytes;
10. unsupported/spoofed payloads are rejected;
11. existing SHA-256 content hash is calculated before Asset persistence.

Trusted local test/integration sources may explicitly set `AllowUnsafeLocalSource`; this flag is not exposed as a public HTTP request field.

## 5. UI visibility

Workspace generation status now recognizes:

- `unknown` -> 提交状态未知
- `reconciling` -> 核对中

They stay non-terminal/pending instead of being rewritten to ordinary failure merely because raw error detail is present.

Video tasks in these states retain the existing "检查" action:

- unknown without ProviderTaskID returns the same unresolved task;
- reconciling with ProviderTaskID polls the provider and can recover.

## Boundaries not imported from Xiaoshuren-Media-MCP

P4 deliberately does not import or create:

- Job
- ProviderExecution
- IdempotencyRecord
- Quote
- Outbox
- second Asset model
- separate reconciliation database

The useful semantics are mapped onto DramaGo's current data model.

## Verification

- GenerationTask repository tests: PASS
- database atomic idempotent reservation test: PASS
- cross-service GenerationTask idempotency test: PASS
- duplicate provider submission E2E-style service test: PASS
- idempotency conflict test: PASS
- ambiguous submission / retry-block test: PASS
- reconciling + ProviderTaskID recovery poll test: PASS
- media remote URL/signature/MIME tests: PASS
- Generation service tests: PASS
- Media service tests: PASS
- MCP tests: PASS
- HTTP/app compile tests: PASS
- Workspace reliability/UI targeted tests: 76/76 PASS
- Workspace production build: PASS
- git diff check: PASS
