# DramaGo × USVDS V11 P5 End-to-End Production Flow

## Goal

Turn the V11 integration from a set of callable Skills and Gates into one verifiable DramaGo-native production flow without adding a second workflow engine or production data model.

The production truth remains:

- Document
- Canon / Variant
- ShotManifest / ResolvedState
- GenerationTask
- Asset

## Agent boundary

DramaGo Agent already edits project Markdown directly in the current project work directory.

P5 deliberately does **not** add document-body MCP write tools. The existing boundary remains:

```text
DramaGo Agent
  -> load V11 Skill
  -> edit/create project Markdown
  -> Markdown frontmatter category/tags
  -> existing Document sync
  -> USVDS V11 Gate projection
```

MCP remains a tool/interface layer, not the document persistence owner.

## Deterministic V11 output contract

Primary creative workflows now declare their expected output contract in the server-side `WorkflowDescriptor`.

| Workflow | Artifact | Required tag | Document category |
| --- | --- | --- | --- |
| Story Architect | story-package | `usvds:artifact:story-package` | reference |
| Episode Architect | episode-architecture | `usvds:artifact:episode-architecture` | reference |
| Creator Script Draft | screenplay | `usvds:artifact:screenplay` | screenplay |
| Storyboard Director | storyboard | `usvds:artifact:storyboard` | storyboard |

The same Storyboard tag is compiled into the pinned V11 Storyboard Skill snapshot.

Explicit artifact tags always win over legacy title inference. Title inference remains only as migration compatibility for old projects.

## Artifact adoption / repair

P5 adds a deterministic server action for binding an existing DramaGo Document to a V11 artifact role:

```http
POST /api/v1/projects/:projectId/usvds-v11/artifacts/:artifact/adopt

{
  "documentId": "...",
  "expectedVersion": 4
}
```

Supported artifacts:

- `story-package`
- `episode-architecture`
- `screenplay`
- `storyboard`

The action only updates the existing Document category/tags using existing optimistic versioning.

It does not create an USVDS artifact table.

For screenplay/storyboard it normalizes the existing Document category before applying the artifact tag. Those metadata mutations increment the normal Document version, so prior revision-bound human approval becomes stale automatically.

## Continue V11 Flow

The Project Overview Gate card now exposes one primary CTA whenever the backend reports a next stage:

```text
继续 V11：<nextWorkflow.label>
```

The button does not implement a second client-side workflow state machine.

It uses the backend-derived `nextWorkflow` and seeds the existing DramaGo Agent with:

- the real V11 Skill chip;
- projectId;
- current Story Package documentId;
- current Episode Architecture documentId;
- current Screenplay documentId;
- blocked Gate reasons;
- expected output artifact;
- expected artifact tag;
- expected Document category.

This lets the Skill write the output directly into the existing project and makes the subsequent Document sync deterministic.

## Gate / stale behavior

P3 revision-bound approval remains authoritative.

If the Agent or user edits an approved Story/Screenplay Document:

- the Document version/content digest changes;
- the old approval token becomes stale;
- the corresponding Gate blocks;
- `nextWorkflow` returns review rather than allowing downstream production.

P5 does not bypass or auto-approve this Gate.

## Storyboard to generation

A V11 Storyboard output is bound to the existing DramaGo storyboard Document.

Existing ShotManifest synchronization/compiler remains authoritative for:

- Canon/Variant binding;
- resolved continuity state;
- compiled provider prompt;
- Shot ready/conflict status.

Generation still enters the existing GenerationTask path and must pass the deterministic V11 generation preflight added in P1.

P4 idempotency/unknown-reconciliation/media-ingest reliability remains unchanged.

## Fixture E2E

P5 adds an integration fixture that runs the entire state progression without using a paid provider:

```text
no artifacts
  -> nextWorkflow = Story Architect

Story Document
  -> adopt story-package
  -> nextWorkflow = Review
  -> approve current revision

Episode Architecture Document
  -> adopt episode-architecture
  -> nextWorkflow = Screenplay

Screenplay Document
  -> normalize category=screenplay
  -> adopt screenplay
  -> nextWorkflow = Review
  -> approve current revision

Storyboard Document
  -> normalize category=storyboard
  -> adopt storyboard

Canon approved/locked
ShotManifest ready
ResolvedState present
CompiledPrompt present
  -> all V11 Gates ready
  -> generation preflight PASS
```

This fixture exercises the same ProjectGateService, revision approvals, artifact adoption, ShotManifest state and generation preflight used by the product.

## Real provider smoke boundary

The E2E regression deliberately stops before a paid provider call.

A real-provider smoke continues to use the normal DramaGo generation surface after the final preflight:

```text
ShotManifest ready
  -> existing generation request
  -> GenerationTask
  -> Provider
  -> P4 idempotency / unknown-reconciliation
  -> existing Asset ingestion
```

No V11-specific provider adapter is introduced.

## Verification

- V11 embedded Skill Pack: PASS
- Prompt Pack / Skill Registry: PASS
- artifact adoption tests: PASS
- explicit-tag-over-legacy-title selection: PASS
- revision-bound approval/stale tests: PASS
- USVDS V11 fixture E2E: PASS
- HTTP Gate/adoption handlers: PASS
- P4 repository/generation/media reliability regression: PASS
- MCP packages: PASS
- Workspace Agent/USVDS/generation targeted tests: 82/82 PASS
- Workspace production build: PASS
- git diff check: PASS
- V11 snapshot reproducibility: PASS
- aggregate snapshot SHA-256:
  `ed1f690a2930899f37b32c0511f4b184ff93ae23ae3979ab5b673f2f31276100`

## Invariants

P5 adds no:

- USVDSProject table
- USVDSWorkflow table
- USVDSArtifact table
- second Shot/Job/Generation model
- DSH runtime dependency

DramaGo remains the product host and only production source of truth.
