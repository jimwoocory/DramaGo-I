# DramaGo × USVDS V11 P2

## Goal

Expose the pinned USVDS V11 creative workflow through DramaGo's existing Prompt Pack, Skill Registry, and Agent runtime.

This phase does not add another workflow engine or persistence model.

## Pinned source

- repository: `jimwoocory/US-Vertical-Drama-Studio`
- branch: `v11`
- commit: `0c68d9cbf525eee6f2a98d3098294cdedbb29e77`
- plugin integration surface: `1.2.6`
- compiled DramaGo pack: `usvds-v11`

V9/V10 directories are migration history. The V11 branch/plugin surface above is the reviewed source.

## Runtime path

```text
Project Overview
    ↓ choose V11 workflow
DramaGo Agent Composer
    ↓ real Skill chip
DramaGo Skill Registry
    ↓ MCP load_skill
Read-only usvds-v11 Prompt Pack
    ↓
Existing DramaGo Agent runtime
    ↓
Documents / Canon / ShotManifest / GenerationTask
```

MCP remains an external/tool interface. DSH is not a runtime dependency.

## Primary workflow entries

| Entry | V11 Skill |
| --- | --- |
| V11 Controller | `usvd-v10-controller` |
| Intake / Adaptation | `usvd-v10-00-intake-adaptation` |
| Story Architect | `usvd-v10-01-story-architect` |
| Episode Architect | `usvd-v10-02-episode-architect` |
| Creator Script Draft | `usvd-v10-03-creator-script-draft` |
| Independent Review | `usvd-v10-04-review-continuity` |
| Continuity | `us-vertical-drama-continuity-editor` |
| Storyboard Director | `us-vertical-drama-storyboard-director` |

All 14 skills in the 1.2.6 plugin surface are available in the Skill Registry. The table above is only the primary product navigation.

## Snapshot compilation

Run:

```bash
pnpm sync:usvds-v11
```

or:

```bash
node scripts/sync-usvds-v11-pack.mjs
```

The generator:

1. refuses a source commit other than the pinned V11 commit;
2. refuses a plugin version other than 1.2.6;
3. reads the actual skill `name:` from frontmatter;
4. resolves direct reference/contract dependencies, including unique cross-skill references;
5. inlines those dependencies so DramaGo's single-file Skill Registry has a self-contained runtime artifact;
6. injects DramaGo workflow metadata and executable document-category hints;
7. appends the DramaGo ownership contract;
8. writes provenance to `SOURCE.json`.

Generated output lives at:

```text
packages/instructions/pkg/pack/usvdsv11/assets/
```

Do not hand-edit generated files. Change the source V11 repo or generator and re-run the sync.

## Ownership contract

A V11 Skill may create or update creative artifacts, but it does not own persistent production state.

- story/development/screenplay/storyboard artifacts → existing DramaGo Documents
- identity/look/scene/prop truth → existing Canon / Variant
- shot execution + continuity → existing ShotManifest / ResolvedState
- model submission + task state → existing GenerationTask
- generated media → existing Asset

A legacy `APPROVED` or `PASS` phrase inside a Skill cannot mutate DramaGo Gate state by itself. Human/system Gate tags continue to use the existing DramaGo path.

## Read-only pack behavior

`usvds-v11` is seeded as a shipped, read-only default pack beside `builtin`.

It can be enabled/disabled through existing pack behavior, but its source entries are not edited in place. Fork/copy mechanisms remain the route for user customization.

## Update rule

When upstream V11 changes:

1. review the new USVDS commit;
2. update the pinned commit/version in the sync generator and Go package together;
3. regenerate the snapshot;
4. inspect generated diff, especially references/contracts;
5. run Prompt Pack, Skill Registry, Agent Composer and Project Overview tests;
6. update the DramaGo gate/workflow catalog only when upstream stage semantics changed.

Never point DramaGo at a moving branch head at runtime.
