# DramaGo P6 — Real-project Smoke & Stacked Merge Readiness

## Status

**Partially verified, not merge-ready.** The company AgentDock connector is operational again. A real project has been smoke-tested on a copied workspace, but it contains no production artifacts. Full V11 production readiness therefore remains unverified.

### Completed

- Built a no-paid-provider Smoke Test CLI that isolates the original workspace.
- Verified the temporary workspace copy and source-file safety with automated tests.
- Executed live empty-project smoke on the company machine (2026-10-08).
- Inspected GitHub PR #1–#6; each is open and reports mergeable.
- Simulated all six merges against current `develop` using `git merge --ff-only`.
- Ran Go/HTTP/MCP integration tests, Workspace 82/82 tests, and production build on the simulated final merge.
- Added a production-evidence check so an empty project cannot pass full production acceptance.

### Still required

Run this smoke against a **populated** real DramaGo project containing Story/Screenplay Documents, approved Canon assets and at least one ShotManifest whose prompt passes generation preflight. A P5 synthetic E2E fixture is not a substitute for the live-data check.

## Company real-project smoke — 2026-10-08

The default `%APPDATA%/mediago-drama/workspace` database was accessible but contained no registered project.

The actual portable workspace was found at:

```text
C:\Users\Administrator\Desktop\dramago-0.1.0-beta.0-win-x64\data\workspace
```

Registered project:

- ID: `project-fe88dfe86e89d256`
- Name: `测试`
- Status: `active`

**Results from the temporary workspace copy:**

| Check | Result |
| --- | --- |
| Database snapshot opened and integrity checked | PASS |
| Project located in actual workspace | PASS |
| Original project write access used | No |
| Project Documents | 0 |
| Canon sync created / updated | 0 / 0 |
| ShotManifest records | 0 |
| GenerationTasks | 0 |
| Ready ShotManifests | 0 |
| Compiled generation preflights passed | 0 |
| V11 Gate states | 6/6 blocked |
| Backend recommended next workflow | `story` (故事架构) |
| `generationReady` | false |
| `productionEvidenceReady` | false |
| Paid Provider submission | None |

All six blocked Gates are correct for an empty project, rather than defects in the Gate evaluator. The project `work/` directory was confirmed empty. A historical log records deletion of a prior draft named `新剧本`, but that document is not present as current production content.

Strict production-evidence blockers:

- `no project documents`
- `no storyboard documents`
- `no ShotManifest records`
- `no shot passed generation preflight`

**Repeated strict execution (2026-10-08):** the compiled Smoke binary was run against the real portable workspace with both `-require-production-evidence` and `-require-generation-ready`. It exited with code `2`, as designed for incomplete production evidence. The original workspace database and the project's manifest/README had matching SHA-256 hashes before and after execution (`0` changed files). The temporary snapshot was removed, and no Provider was called.

**Conclusion:** the real workspace access/isolation/empty-state workflow tests passed. A populated Story → approval → Episode → Screenplay → review → Storyboard → ShotManifest → generation-preflight run has **not** been verified using this project's real content.

## User-supplied HAVEN screenplay overlay smoke — 2026-10-08

A user-supplied DOCX, `HAVEN_EP01-EP03_Screenplay_Draft_V10_R4_Revision02.docx`, was text-extracted to a **private, temporary Markdown test input** and overlaid into the copied workspace project `project-fe88dfe86e89d256`. No source screenplay body was committed or sent to GitHub.

Source provenance: V10 R4 Revision02, `CREATOR_AUTHORIZED_DRAFT`, explicitly not system-approved or production-ready. This draft is external input tested against the pinned V11 DramaGo adapter, **not** a substitute for a V11-approved Story Package or human approval.

| Check | Result |
| --- | --- |
| Workspace snapshot + source-overlay import | PASS |
| Existing Document sync recognizes screenplay category | PASS |
| Existing V11 artifact tag `usvds:artifact:screenplay` recognized | PASS |
| Scene headers / unique IDs preserved | 15 / 15 |
| EP01 / EP02 / EP03 scene counts | 5 / 5 / 5 |
| Summed source scene duration | 425 seconds |
| Temporary imported Markdown matches extracted input byte-for-byte | PASS |
| Document count in copied project | 1 |
| Screenplay Document ID | `haven-ep01-ep03-revision02-smoke` |
| Human screenplay approval | false (correct) |
| Next recommended workflow | `story` (Story Architect; approved Story Package missing) |
| Canon / ShotManifest / Generation preflight | not available from this screenplay alone |
| `generationReady` / `productionEvidenceReady` | false / false |
| Original workspace DB / project manifest / README hash changes | 0 |
| Original project work-file count after smoke | 0 |
| Temporary smoke snapshot removed | yes |
| Paid provider calls | 0 |

**Outcome:** real source-backed screenplay Document ingestion and V11 rejection gates are verified; this is **not** a production-ready E2E pass. The source draft is missing independent human approval, the V11-authoritative Story Package, approved Canon, storyboard production contracts, and a ready ShotManifest. Do not auto-promote the V10 draft or generate media from it.

## Reusable external screenplay overlay — P6 hardening

P6 now supports a reusable **external-source, draft-only** test mode. It is designed for screenplay inputs received outside the DramaGo workspace and never imports them into the original user project.

```powershell
go run ./services/server/cmd/usvds-v11-smoke \
  -workspace "<workspace-root>" \
  -project "<project-id>" \
  -overlay-screenplay "<private-unapproved-screenplay.md>" \
  -json
```

The source must be a regular Markdown file under 2 MiB with `category: screenplay` and `usvds:artifact:screenplay` in YAML frontmatter. Approval tags are rejected, and the overlay is written exclusively into the temporary project snapshot. The report records SHA-256, byte count and `approvalProhibited=true`, **not** the screenplay text or source file path.

Even if the copied workspace has other production artifacts, `productionEvidenceReady` is forced false when a user-supplied screenplay overlay is used. This prevents a draft from accidentally satisfying real-production acceptance criteria.

**Verification against the user-supplied HAVEN EP01–EP03 V10 R4 Revision02:**

| Check | Result |
| --- | --- |
| Source-extracted scene IDs | 15 unique IDs, five per episode |
| Sum of source scene timing budgets | 425 seconds |
| Reusable overlay smoke runner | PASS |
| Imported screenplay Documents | 1 |
| Recognized Document ID | `haven-ep01-ep03-revision02-smoke` |
| Overlay source SHA-256 matched | yes |
| Human-approved Gate count | 0 |
| Backend nextWorkflow | `story` |
| Generation / production evidence readiness | false / false |
| Draft-only overlay blocker | present |
| Original workspace file SHA-256 changes | 0 |
| Temporary snapshot removed | yes |
| Paid Provider submissions | 0 |

No complete digest-bound `USVDS-V10-STORY-PACKAGE-HAVEN · R4` source was found in the checked available material. The user's screenplay references the R4 digest but is explicitly `CREATOR_AUTHORIZED_DRAFT`, not system-approved. Earlier Haven-related development packages are not interchangeable with that exact canonical revision. **Do not synthesize or approve canonical Story Truth from the three screenplay episodes.**

## Smoke runner usage

Run from the repository root with Go 1.25 or a compatible toolchain.

List projects from an isolated database snapshot:

```powershell
go run ./services/server/cmd/usvds-v11-smoke -workspace "<workspace-root>" -list
```

Evaluate one project (no paid Provider call):

```powershell
go run ./services/server/cmd/usvds-v11-smoke -workspace "<workspace-root>" -project "<project-id>" -json
```

Strict data-evidence check, which exits nonzero when required artifacts are missing:

```powershell
go run ./services/server/cmd/usvds-v11-smoke -workspace "<workspace-root>" -project "<project-id>" -require-production-evidence
```

Full readiness check, which additionally requires the `generation_ready` Gate:

```powershell
go run ./services/server/cmd/usvds-v11-smoke -workspace "<workspace-root>" -project "<project-id>" -require-production-evidence -require-generation-ready
```

### Isolation and safety

The runner copies `app.db` and any `-wal`/`-shm` companions, then integrity-checks only the copy. It reads the selected project's directory from the copied DB; copies its `work/` and lightweight metadata to a temporary workspace; rewrites `project_dir` **only in that copy**; and runs existing Document/Canon/ShotManifest synchronization, compilation and V11 Gate projection inside that workspace.

It rejects symlinks/non-regular work files to avoid following pointers outside the selected project. Temporary data is deleted unless `-keep-snapshot` is specified. No creative data is submitted to an external model, and no provider cost is incurred.

Automated tests validate source Markdown byte preservation, proper next-workflow selection and production-evidence classification. Strict production evidence requires nonzero Document, Storyboard, ShotManifest and successful generation-preflight counts.

## PR stack audit

GitHub metadata observed on 2026-10-08:

| PR | Branch | Base | State |
| --- | --- | --- | --- |
| #1 / P0 | `feature/dramago-usvds-v11-p0` | `develop` | open, mergeable |
| #2 / P1 | `feature/dramago-usvds-v11-p1` | P0 | open, mergeable |
| #3 / P2 | `feature/dramago-usvds-v11-p2` | P1 | open, mergeable |
| #4 / P3 | `feature/dramago-usvds-v11-p3` | P2 | open, mergeable |
| #5 / P4 | `feature/dramago-media-reliability-p4` | P3 | open, mergeable |
| #6 / P5 | `feature/dramago-usvds-v11-p5-e2e` | P4 | open, mergeable |

Current `origin/develop` at audit time: `baceb323`.

A disposable Git worktree fast-forwarded successfully through every layer, confirming a strict ancestor chain:

```text
develop@baceb323
  → P0@2a523d9
  → P1@6345166
  → P2@9fa7c11
  → P3@d7ac0cd
  → P4@d6c0529
  → P5@8f7c11f
```

Tests on the fully simulated merged tree:

- Embedded USVDS V11 Skill Pack: PASS
- USVDS V11 service: PASS
- Repository / Generation / Media: PASS
- HTTP handlers/middleware and MCP: PASS
- Workspace targeted Agent/USVDS/Generation tests: **82/82 PASS**
- Workspace production build: PASS

The existing large-Vite-chunk warning is non-blocking.

**Recommended merge order:** #1 → #2 → #3 → #4 → #5 → #6. Before actual production merge, obtain the populated-project smoke evidence; then confirm CI checks and merge in order. Do not label P6 fully passed solely because the code and empty-project tests pass.
