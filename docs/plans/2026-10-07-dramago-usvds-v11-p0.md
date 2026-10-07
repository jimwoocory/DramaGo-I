# DramaGo × USVDS V11 P0

## Decision

DramaGo-I is the product host and the only owner of persistent production truth.

USVDS is pinned to the current reviewed baseline:

- branch: `v11`
- commit: `0c68d9cbf525eee6f2a98d3098294cdedbb29e77`
- ChatGPT plugin integration surface: `plugins/us-vertical-drama-studio@1.2.6`

V9/V10 material is migration history only.

## Existing DramaGo truth

Do not create parallel USVDS persistence for these concepts:

| USVDS concern | DramaGo owner |
| --- | --- |
| project/workspace | existing Project / Workspace |
| story and screenplay documents | existing Document |
| character/scene/prop identity | CanonAsset |
| look/state variant | CanonAsset.ParentID / Variant |
| physical reference media | Asset + CanonReference |
| storyboard execution contract | ShotManifest |
| continuity state | ShotManifest.ResolvedStateJSON |
| deterministic model prompt | ShotManifest.CompiledPrompt |
| generation execution | GenerationTask |
| generated outputs | GenerationTaskAsset -> Asset |
| external agent access | packages/mcp |

## V11 integration shape

V11 enters DramaGo through three surfaces.

### Native domain behavior

- Gate status visible in DramaGo.
- Canon/Variant approval requirements.
- Shot execution readiness.
- Continuity readiness.
- Generation readiness.

These must reference current DramaGo records. They do not own new copies.

### Workflow / skill behavior

The V11 creative roles remain prompt/workflow packages:

- adapter
- showrunner
- episode architect
- screenwriter
- script doctor
- continuity editor
- storyboard director

They may produce/update DramaGo Documents and proposed structured changes through existing approval paths.

### Deterministic validator / compiler behavior

Rules that can be checked without an LLM belong in DramaGo code, close to the owning domain:

- required V11 binding completeness
- approval/gate prerequisites
- timeline/duration closure
- Canon/Variant binding validity
- continuity state availability
- compiled prompt readiness
- generation preconditions

Existing `shotmanifest` compiler and `productionqa` remain authoritative implementations where they already cover a rule.

## Media MCP migration

The separate Xiaoshuren-Media-MCP repository is not a second product backend.

### Reuse inside DramaGo

Port only missing reliability semantics where tests prove value:

- provider submission idempotency
- provider unknown/reconciliation state
- safe remote media ingestion
- MIME/magic/hash verification
- server-side credential boundary
- retryability/error normalization

Map these onto existing `GenerationTask`, `Asset`, Provider and repository/service code.

### Keep as optional remote boundary

Only if a remote deployment is required, Media MCP can remain a stateless/edge adapter that invokes DramaGo business actions. It must not own canonical Job/Asset/Quote state.

### Freeze

Do not expand the independent PostgreSQL Job/Asset/ProviderExecution model as a second source of truth.

## DSH boundary

DSH is optional compatibility only.

- May load/call a DramaGo/USVDS workflow.
- May expose a UI adapter.
- Must not own project, approval, asset, shot, generation or model-routing truth for DramaGo.

No new DramaGo core behavior should depend on DSH packages.

## P0

1. Pin V11 baseline in code.
2. Add a non-persistent V11 gate/ownership adapter.
3. Reuse existing DramaGo IDs and statuses.
4. Add deterministic tests for gate blocking/readiness.
5. Keep MCP external.
6. Run server and MCP regression tests.

## P1

- Expose V11 gate state through existing DramaGo service/API/UI.
- Map V11 skill outputs to Documents and approval actions.
- Add visible gate dashboard/edit affordances without duplicate records.
- Port selected Media MCP reliability semantics to GenerationTask/Asset.

## P2

- Full V11 workflow orchestration from story development through generation.
- Remote MCP surface calls the same business action layer used by the UI.
- DSH/Claude/ChatGPT become interchangeable clients/adapters.
