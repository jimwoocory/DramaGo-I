---
name: us-vertical-drama-studio
description: Route ideas, existing outlines, adaptation sources, and scripts through US-facing story development before the retained screenplay and production workflow.
title: USVDS V11 · Studio
hint:
  workflow: usvds-v11
  stage: controller
  baseline: 0c68d9c
---

# US Vertical Drama Studio

Use this as the single entry point for developing an original or adapted US-facing vertical microdrama. Coordinate the specialist skills in this plugin; do not duplicate their detailed craft rules.

## First route: story before scripts

For every new project, start with `usvd-v10-controller`, which classifies the material and invokes the V10 Intake, Story Architect, Episode Architect, and independent Review skills. Its five routes cover a simple idea, a completed outline for diagnosis, a novel/comic or other source for adaptation, a creator-selected adaptation direction, and an existing script needing reverse-story diagnosis. Return the specialist's useful diagnosis or direction comparison to the creator. Do not jump from a logline, translated source, existing outline, or screenplay to a production script. This release's delivery contract is **1.2.6**: if an earlier conversation or document says that screenplay action is Chinese-only or dialogue has no Chinese reference, use the current paired-language rules below when regenerating.

Before episode planning, require a complete Story Package through resolution and an independent `PASS_FOR_EPISODE_ARCHITECTURE` report for its exact revision. The V10 trusted approval runtime is not implemented, so the production Screenwriter and downstream production remain blocked. After full episode architecture and an independent `PASS_AWAITING_HUMAN_APPROVAL` report, a direct creator instruction in the same conversation for the exact package may route to `usvd-v10-03-creator-script-draft`. For this non-production draft, the current review-attested Episode Architecture DOCX/HTML plus its exact matching review is a valid readable source when the canonical JSON is unavailable; disclose that narrower source binding and do not invent omitted canon. An unspecified “continue R4” request defaults to an explicitly stated EP01–EP03 first batch. This creates a clearly labeled, non-production `CREATOR_AUTHORIZED_DRAFT`, never a system `APPROVED` Story Package. Do not ask the creator to repeat an already visible direct approval for that same package; do not use an assistant summary or screenshot alone as approval evidence.

The seven original specialist skills remain available for a legacy project with its previously approved Story Bible, Beat Sheet, and downstream handoff evidence. They may also critique supplied material without implying approval. If such evidence is absent, route the project to V10 story intake rather than fabricate an `APPROVED` status.

## Operating principle

Treat the project as a gated production pipeline. Preserve canon between phases, keep a versioned decision ledger, and never present a downstream deliverable as approved when an upstream gate is missing.

## Intake

Collect or infer only the minimum missing information:

- premise or source material and whether it is original or an adaptation
- target market/audience, rating, tone, and language
- episode count and approximate runtime
- requested deliverable (full package, season plan, one episode, screenplay, audit, or continuity pass)

If a missing choice materially changes the story, ask one focused question before proceeding. Otherwise state a reversible assumption.

## Development document language standard

Unless the creator explicitly requests otherwise, deliver early Project Briefs, direction comparisons, complete story outlines, character plans, and independent reviews in 简体中文. Deliver **every field of the full Episode Map**, every screenplay scene and spoken line, and every creator-facing storyboard, asset, shot, and generation-prompt field in paired Chinese and natural US English. Preserve IDs and factual equivalence. An American target market determines story plausibility; it does not justify English-only early outlines. Follow [`output-language.md`](../../references/output-language.md).

Deliver each requested creator-facing development artifact as two actual downloadable files, DOCX and standalone HTML, with matching content and revision. Do not send TXT or a text dump in place of either file. Keep canonical JSON for internal tracking or a separately requested audit export. Follow [`document-delivery.md`](../../references/document-delivery.md); if the environment cannot make a requested file, state the limitation honestly.

At the project's actual final handoff, or when the creator requests an archive, provide a real downloadable ZIP of the current available DOCX/HTML pairs plus reviews and useful machine-readable assets. Include a manifest with versions, episode scope, SHA-256, bilingual coverage, missing items, and gate status. Follow [`final-package.md`](../../references/final-package.md). A partial ZIP must say `PARTIAL`; packaging never upgrades a draft to system-approved production work.

## Script language standard

Unless the user explicitly requests otherwise, write creator-facing episode scripts in paired Chinese and natural US English:

- Pair scene headings, action, performance direction, on-screen notes, beat traces, and production notes field by field in Chinese and English.
- Format each shootable spoken line as an English character name with its natural English dialogue, followed by a Chinese meaning line labeled `仅供作者理解，不念出/不用于口型`. The Chinese line is not a second performed line.
- For prompts, pair Chinese and English master, local, asset, and negative prompts by prompt ID; mark the one language version submitted to the chosen model and the other as a review translation. Preserve existing machine import fields and do not concatenate both versions into a single prompt input.
- Keep `【OS／画外音】` separate from `【情绪/表演】` in each screenplay scene. Emotion is unspoken playable direction; OS is an audible inner voice, voiceover, or off-screen line. When present, give speaker/type, exact English performed line and adjacent unspoken Chinese meaning. When absent, write `无／None`; do not invent an inner monologue.

## Screenplay annotation standard

Unless the user asks for a treatment or outline only, every screenplay scene must visibly label its production inputs. Use stable labels: `【场景】`, `【人物】`, `【动作】`, `【情绪/表演】`, `【台词】`, and `【OS／画外音】`. Pair the substantive value of each field in Chinese and English.

- `【场景】` names the precise dramatic location, time, and scene objective.
- `【人物】` names the characters present; character names in speech are English.
- `【动作】` describes only visible, shootable behavior and changes in the scene.
- `【情绪/表演】` states the playable emotional state or concealed intention and the observable performance evidence that directs performance; it is not audible OS.
- `【台词】` contains the shootable English dialogue in `ENGLISH CHARACTER NAME: “English dialogue.”` format plus the adjacent Chinese reference meaning clearly marked as unspoken.
- `【OS／画外音】` separately identifies `inner_voice`, `voiceover`, or `off_screen` speech with the performed English line and Chinese reference; `无／None` is valid when there is no such line.

For storyboard-ready work, split the screenplay into numbered scene units before shot design. Preserve the approved scene ID, timing, entry/exit state, objective, visible action, observable performance evidence, and dialogue purpose. Do not substitute unlabelled prose for these fields.

## Route the work

For a documented legacy project only, run the smallest complete sequence that satisfies the request:

1. **Adaptation** — use `us-vertical-drama-adapter` for non-US source material or when cultural plausibility is uncertain. Produce an approved Adaptation Brief.
2. **Series canon** — use `us-vertical-drama-showrunner` to create the Story Bible, arc ladder, character engines, reveal ledger, escalation plan, and anti-repetition controls. Require `APPROVED Story Bible`.
3. **Episode architecture** — use `us-vertical-drama-episode-architect` to create episode beats with hook, conflict, escalation, reversal, payoff, and cliffhanger gates. Require `APPROVED Beat Sheet`.
4. **Screenplay** — use `us-vertical-drama-screenwriter` only from an approved Beat Sheet. Draft native, shootable vertical-drama scenes within the requested runtime.
5. **Independent review** — use `us-vertical-drama-script-doctor` to score the draft against the rubric. Do not call it ready unless the result is `PASS`; route failed items back to the responsible phase.
6. **Continuity** — use `us-vertical-drama-continuity-editor` to update the continuity ledger and verify names, rules, props, injuries, time, reveals, and handoffs. Require `CLEAR` before final delivery.
7. **Storyboard and generation packet** — use `us-vertical-drama-storyboard-director` only after `SCRIPT DOCTOR PASS` and `CONTINUITY CLEAR`. It locks character, costume, scene, and prop assets, then creates per-shot video-generation prompts.

For a full-series request, plan the series first, then expand episodes in batches sized to the user's requested review cadence. Do not fabricate all episodes when the user asked for a plan or sample.

## State and handoffs

Maintain these statuses explicitly:

`BRIEF` → `BIBLE APPROVED` → `BEATS APPROVED` → `SCRIPT DRAFT` → `SCRIPT DOCTOR PASS` → `CONTINUITY CLEAR` → `STORYBOARD PACKAGE APPROVED`.

Every handoff includes:

- current canon/version
- deliverable and episode range
- unresolved decisions
- risks or deliberate deviations
- next gate and acceptance condition

A specialist may flag a canon problem but may not silently rewrite another phase's approved decisions. Send changes back to the owning phase and record the revision.

## Quality controls

Keep US plausibility, commercial readability, vertical framing, short-episode rhythm, and emotional clarity visible in every relevant deliverable. Enforce escalation through changed stakes, costs, information, or relationships; reject repetition that only swaps locations or insults. Keep cliffhangers specific and causally earned.

When the user requests only critique, audit, or one specialist task, do not force the full pipeline. Return the requested artifact plus the relevant gate status and the shortest next step.

## Final response format

Lead with the requested creative result. Then provide a compact production status:

- gate status
- canon assumptions
- unresolved decisions
- recommended next handoff

Use the specialist reference files only when their phase is active.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: ../../references/document-delivery.md

# V10 creator-facing document delivery

Unless the creator explicitly changes the format, deliver each requested planning or review artifact as **two actual files** derived from the same current artifact revision:

1. An editable `.docx` for review, comments, and revision.
2. A standalone UTF-8 `.html` with the same substantive content for browser reading.

This applies to the Project Brief, direction comparison, adaptation diagnosis, complete story outline, episode map, screenplay draft, storyboard/asset pack, generation-prompt pack, and independent review report. Give each file a clear Chinese title, project identifier, artifact identifier, revision, and scope. The filenames may use stable Latin IDs. Early development and review documents default to Chinese; episode maps, scripts, and prompt/shot/asset reading views follow [output-language.md](output-language.md) and contain paired Chinese and English content. Keep section order, story facts, episode coverage, findings, language pairs, and gate status aligned between DOCX and HTML. If either file changes after review, regenerate the other from the exact same approved or reviewed source revision. Do not silently translate one format differently or label mismatched files as the same revision.

Do **not** use `.txt` as the creator-facing deliverable, attach a `.txt` instead of a requested file, or paste an unformatted text dump and call it a DOCX or HTML. Machine-readable JSON can remain an internal canonical artifact and may be supplied separately when needed for audit; it does not replace the two readable files.

For DOCX, use actual document export or a document authoring tool. Make headings, paragraphs, tables, and page breaks readable; inspect the rendered pages when the environment provides a renderer. For HTML, create a real `.html` file with UTF-8 encoding, semantic headings, accessible tables where needed, and readable desktop/mobile layout. It should open without external scripts, login, or remote assets. Do not rename plain text with a `.docx` extension or send HTML markup in a `.txt` file.

Before handoff, confirm both files exist and correspond to the same source revision. Provide two direct download links labeled `DOCX` and `HTML`, plus a brief Chinese summary and gate status. If the current environment cannot actually create one format, state the specific limitation; do not claim a file was produced or substitute TXT without the creator's permission.

When the creator asks for the complete project's assets or the workflow reaches its actual final handoff, also follow [final-package.md](final-package.md). The ZIP is an additional download; do not replace the individual DOCX/HTML files with it.


### Bundled: ../../references/final-package.md

# Creator asset ZIP handoff

At the actual end of the requested workflow, or whenever the creator asks for a project archive, collect the **current effective revision** of each artifact available in this conversation/project and hand the creator one downloadable ZIP. The ZIP is a convenience copy; continue to provide each stage's DOCX and HTML links. Do not claim assets from another chat or inaccessible storage have been included.

## Contents

- Matching DOCX and standalone HTML for the Project Brief, selected direction/adaptation analysis, complete Story Package, 48-episode map, each delivered screenplay episode/range, asset/shot/storyboard pack, generation-prompt pack, and creator-readable review reports, insofar as each stage actually exists.
- Available canonical JSON, independent review JSON, continuity ledger, import rows, and workbench JSON as supplementary machine-readable assets. JSON never replaces DOCX/HTML.
- `manifest.json` listing every member's relative path, artifact ID, revision, episode scope, source digest or review binding when known, SHA-256 of the actual file bytes, language coverage (`zh-CN`, `en-US`, or paired), and status. Include `project_id`, generation time, package status, production gate, missing expected assets, and a concise note that a readable view digest has not been independently recomputed when applicable.

Use stable folders such as `01-brief/`, `02-story/`, `03-episodes/`, `04-scripts/`, `05-assets-storyboards-prompts/`, `06-reviews/`, and `07-machine/`. Include only files actually available and belonging to the same project. Never silently replace a current R4 file with an older R3, mix revisions inside one DOCX/HTML pair, or include unrelated uploads. If only some stages exist, label the package `PARTIAL` and list missing or blocked stages in the manifest and handoff; `COMPLETE` means all requested deliverables actually exist. Do not call a `CREATOR_AUTHORIZED_DRAFT` screenplay or any downstream asset system-approved or production-ready while ND-001 remains blocked. Packaging itself grants no approval.

For the bilingual stages, each delivered DOCX/HTML pair must include both languages at the required field granularity. An English dialogue line has a Chinese reference translation clearly marked **not spoken / not for lip-sync**. A prompt pair identifies the single model-submission language and its review translation; never concatenate both into a one-language model input field. Keep machine import keys and existing prompt fields compatible with the target schema.

Make a real `.zip` using an available file tool or archive library. The optional `tools/build-delivery-package.py` can package a prepared input manifest in an environment with Python. Verify that the archive exists and provide a direct download link to the author along with a brief Chinese inventory, episode coverage, package status, and production gate. If file creation is unavailable, state the limitation and give the individual links; never claim to have sent a ZIP that does not exist. Do not email or upload the archive to another person or service without a separate direct creator instruction.


### Bundled: ../../references/output-language.md

# V10 output language contract

The US market describes the audience and story world; it does not set the language of the creator's development documents.

Unless the creator explicitly requests another language, write early creator-facing development artifacts in **简体中文**: Project Brief, direction comparison, adaptation decision table, complete story outline, character and relationship arcs, diagnostic report, review findings, and gate summary. For this creator's **Episode Architecture / Episode Map, per-episode screenplay, storyboard, asset creation pack, shot list, and generation-prompt reading views**, use **简体中文 + natural US English in parallel**. Chinese comes first, followed immediately by its English counterpart. Pair every narrative field in every episode, every screenplay scene field and spoken line, and every human-readable image/video prompt and negative constraint. This includes the complete 48-episode R4 re-render and later requested episode ranges. Keep IDs, timing, reference bindings, and status codes identical in both languages. The English is a faithful US-facing rendering, not a literal translation of Chinese-market behavior or a chance to invent new story facts.

For machine-readable JSON, keep contract property names, IDs, enum values, status codes, hashes, and source locators exactly as required by the schema. Keep existing canonical values and import fields in their required language; do not add bilingual text inside a single prompt field if the target accepts only one prompt. In that case keep the designated model-submission prompt intact and place the matched translation beside it in the DOCX/HTML reading view or a clearly named companion field allowed by the target. The bilingual reading view is a presentation layer; it does not change the Story Package digest or production-workbench import schema. Preserve original titles, exact source quotations, and established English character or place names when needed.

At the screenplay stage, the shootable English dialogue, explicit OS/voiceover, and story-world text are the executable US-facing lines. Provide a neighboring Chinese meaning line for the creator, clearly marked as **reference only / not spoken / not for lip-sync**. Pair scene headings, action, performance notes, and production annotations in Chinese and natural English. Keep `【OS／画外音】` separate from unspoken `【情绪/表演】`; an unused OS field says `无／None`, rather than inventing narration. For generation prompts, label which one language version is submitted to the selected model; the other is a review translation, never an instruction to submit both together. If the creator explicitly requests another language policy, record the override in the Brief.

Before delivery, scan every readable artifact. For early development documents, use the creator's Chinese default. For episode maps, screenplays, shot/asset packages, and prompts, verify that every requested unit and substantive field has paired Chinese and English values with no omissions or meaning drift. Check names, dates, durations, property/legal facts, knowledge states, secrets, asset IDs, visual anchors, dialogue intent, and negative constraints. A Chinese-only or English-only field in these bilingual deliverables fails delivery. A line saying `Chinese development documents` while the body is English is also a failed language check.

For a short formatting model, see [the Chinese development sample](../examples/chinese-development-sample.md). It shows Chinese brief, outline, episode and review prose while preserving machine field names and codes.


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
