# DramaGo P6 — Real-project Smoke & Stacked Merge Readiness

## Status

P6 is **in progress**.

Completed:

- smoke runner implementation;
- source-workspace safety regression;
- GitHub PR #1–#6 dependency audit;
- local fast-forward merge simulation from current `develop`;
- integrated Go/MCP/Workspace regression on the simulated final merge.

Pending:

- execute the smoke runner against at least one **real DramaGo workspace/project** from the company AgentDock host.

The current `AgentDock-Company` connector is returning an internal error, while `AgentDock-company-2` has no MediaGo Drama workspace at its default path. P6 must not substitute CIH/PVS or a synthetic fixture for the required real-project smoke.

## Smoke runner

Command:

```powershell
go run ./services/server/cmd/usvds-v11-smoke -workspace "<workspace-root>" -list
```

Run one project:

```powershell
go run ./services/server/cmd/usvds-v11-smoke \
  -workspace "<workspace-root>" \
  -project "<project-id>" \
  -json
```

Require final generation readiness:

```powershell
go run ./services/server/cmd/usvds-v11-smoke \
  -workspace "<workspace-root>" \
  -project "<project-id>" \
  -require-generation-ready
```

### Safety model

The runner never opens the source workspace through DramaGo's migrating repository layer.

Instead it:

1. resolves the source workspace;
2. copies `app.db` plus WAL/SHM files into a temporary workspace;
3. opens and integrity-checks only the copied DB;
4. reads the selected project's original `project_dir` from the copied DB;
5. copies the selected project's `work/` and lightweight project metadata into the temporary workspace;
6. rewrites `project_dir` only inside the copied DB;
7. performs Canon and ShotManifest synchronization only inside the copy;
8. compiles ready ShotManifests and executes the V11 generation preflight only inside the copy;
9. removes the temporary snapshot unless `-keep-snapshot` is explicitly requested.

A regression test verifies that the original project Markdown remains byte-identical after smoke execution.

## Smoke output

The report includes:

- selected project;
- document count;
- Canon sync summary;
- ShotManifest sync summary;
- V11 Gate report and blockers;
- revision approval projection;
- backend `nextWorkflow`;
- storyboard document count;
- total/ready shot counts;
- generation preflight pass/failure counts;
- existing GenerationTask summary;
- whether `generation_ready` is true.

No paid Provider call is made.

## Stacked PR audit

GitHub state observed on 2026-10-08:

| PR | Head | Base | GitHub mergeable |
| --- | --- | --- | --- |
| #1 | `feature/dramago-usvds-v11-p0` | `develop` | yes |
| #2 | `feature/dramago-usvds-v11-p1` | P0 | yes |
| #3 | `feature/dramago-usvds-v11-p2` | P1 | yes |
| #4 | `feature/dramago-usvds-v11-p3` | P2 | yes |
| #5 | `feature/dramago-media-reliability-p4` | P3 | yes |
| #6 | `feature/dramago-usvds-v11-p5-e2e` | P4 | yes |

Current remote develop:

```text
baceb323 chore: import DramaGo-I B+C baseline
```

A temporary worktree was created from that exact current `origin/develop`, then these heads were applied in order using **`git merge --ff-only`**.

All six fast-forwards succeeded:

```text
develop
  -> P0  2a523d9
  -> P1  6345166
  -> P2  9fa7c11
  -> P3  d7ac0cd
  -> P4  d6c0529
  -> P5  8f7c11f
```

This proves the current stack is a strict ancestor chain with no hidden branch divergence.

## Integrated merge regression

Tests were executed in the temporary worktree **after the full P0→P5 merge simulation**, not merely on individual PR branches.

Passed:

- USVDS V11 embedded pack;
- USVDS V11 service;
- repository;
- generation;
- media;
- HTTP handlers/middleware/routes compile/tests;
- MCP packages;
- Workspace targeted Agent/USVDS/generation tests: **82/82**;
- Workspace production build.

The existing Vite large-chunk warning remains informational and is not introduced by the V11 stack.

## Merge recommendation

Do not merge out of order.

Safe order:

```text
#1 -> #2 -> #3 -> #4 -> #5 -> #6
```

After each merge, GitHub will normally retarget the next stacked PR automatically or its base can be changed to `develop`. Before the final production merge, rerun P6 real-project smoke against a temporary snapshot of the company DramaGo workspace.

P6 should not be marked complete until that real-project smoke is recorded.
