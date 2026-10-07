# DramaGo × USVDS V11 P3

## Goal

Close the human-review loop for USVDS V11 without adding an approval database or a second workflow runtime.

P3 makes Story and Screenplay approvals revision-bound, makes stale approvals visible, and derives the recommended next V11 workflow from current DramaGo artifacts.

## Human Gate ownership

Only these V11 gates are explicit human document approvals:

- `story_approved`
- `screenplay_reviewed`

Other gates stay deterministic and domain-owned:

- Canon readiness -> existing Canon / Variant status
- continuity -> existing ShotManifest / ResolvedState
- storyboard readiness -> existing ShotManifest status
- generation readiness -> existing compiled prompt + ShotManifest preflight

A Skill cannot approve itself.

## Revision-bound approval token

Approval is stored in the existing Document `tags` field and is bound to the resulting document version plus SHA-256 of its content.

Example:

```text
usvds:approval:story_approved:v12:sha256:<64-hex-content-digest>
```

Approval flow:

1. client submits `documentId + expectedVersion`;
2. server re-reads the current DramaGo Document;
3. optimistic version check must match;
4. approval token is calculated for the post-tag-mutation version;
5. existing `set_document_tags` persistence increments the Document version;
6. Gate evaluation only accepts an exact current-version + current-content token.

Any later document edit increments the version and/or changes the digest. The old token remains historical metadata but evaluates as `stale=true`, never as approved.

No new approval table is created.

## API

Read:

```http
GET /api/v1/projects/:projectId/usvds-v11/gates
```

Approve current revision:

```http
POST /api/v1/projects/:projectId/usvds-v11/gates/:gate/approve
{
  "documentId": "...",
  "expectedVersion": 12
}
```

Revoke current revision:

```http
POST /api/v1/projects/:projectId/usvds-v11/gates/:gate/revoke
{
  "documentId": "...",
  "expectedVersion": 13
}
```

Version conflicts return HTTP 409.

## Stable artifact tags

The V11 snapshot compiler now requires authoritative creative artifacts to preserve stable DramaGo document tags:

- Story Package: `usvds:artifact:story-package`
- Episode Architecture: `usvds:artifact:episode-architecture`
- Screenplay: `usvds:artifact:screenplay`

These tags locate artifacts. They do not approve them.

Legacy P1 tags such as `usvds:story-approved` and `usvds:screenplay-reviewed` no longer satisfy a current revision approval. P3 removes them when the corresponding gate is explicitly approved/revoked.

## Next workflow recommendation

The project Gate report returns `nextWorkflow` derived from authoritative artifacts and approvals:

```text
no Story Package
  -> Story Architect

Story exists, not current-version approved
  -> Independent Review / human approval

Story approved, no Episode Architecture
  -> Episode Architect

Episode Architecture exists, no Screenplay
  -> Creator Script Draft

Screenplay exists, not current-version reviewed
  -> Independent Review / human approval

Screenplay reviewed, downstream production not ready
  -> Storyboard Director
```

This is guidance only. It is not persisted as a second workflow state machine.

## UI

Project Overview now shows:

- current human approval document and version;
- current-version approved / waiting approval / approval stale;
- Approve Current Version;
- Re-approve;
- Revoke Approval;
- backend-derived recommended next workflow;
- existing V11 workflow buttons.

Starting a workflow seeds the existing DramaGo Agent with:

- real V11 Skill chip;
- project ID;
- Story Package document ID;
- Episode Architecture document ID;
- Screenplay document ID;
- currently blocked Gate reasons.

## Invariants

P3 still preserves the same single source of truth:

```text
Document
Canon / Variant
ShotManifest / ResolvedState
GenerationTask
Asset
```

There is no USVDS approval table, workflow table, project table, asset table, shot table, job table, or DSH dependency.

## Verification

- USVDS V11 service tests: PASS
- revision approval/stale/revoke tests: PASS
- HTTP approve/revoke handler tests: PASS
- Prompt Pack / Skill Registry tests: PASS
- MCP tests: PASS
- AgentComposer + ProjectOverview: 23/23 PASS
- Workspace production build: PASS
- git diff check: PASS
- V11 snapshot reproducibility: PASS
- aggregate snapshot SHA-256:
  `2545567cca2441e9db2d664b2189fbf2ad48fc08103d0a6f802a2f98bab97c63`
