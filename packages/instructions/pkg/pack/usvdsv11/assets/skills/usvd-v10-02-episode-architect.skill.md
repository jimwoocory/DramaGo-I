---
name: usvd-v10-02-episode-architect
description: Convert an independently reviewed complete Story Package into a full-season episode map with a mini dramatic arc and explicit state transitions for every episode.
title: USVDS V11 · Episode Architect
hint:
  workflow: usvds-v11
  stage: episode
  baseline: 0c68d9c
---

# USVDS V10 — Episode Architect

## Default output language

Read [the output language contract](../../references/output-language.md). For this creator, write the complete Episode Architecture reading view **field by field in 简体中文 and natural US English**. Keep each Chinese value immediately beside its English counterpart for all 48 episodes, including hook, goal, problem, conflict, escalation, emotion, reversal, payoff, question, cliffhanger, outcome, entry/exit state, character changes, and continuity changes. Pair the field labels too; preserve IDs, references, status codes, and established character names exactly. English is a faithful US-facing development rendering, not literal Chinese-shaped dialogue or an opportunity to add facts. Keep canonical JSON property names and source story values unchanged.

Read [the document delivery contract](../../references/document-delivery.md). Deliver the complete requested Episode Map in matching bilingual DOCX and HTML files from the same Story Package revision, with all episode numbers and both languages for every narrative field covered. Do not substitute TXT, one short sample, or English-only prose for the full bilingual map when the creator requested the full season.

## Existing R4 bilingual view repair

When the creator supplies an already reviewed Episode Architecture R4 and asks for the missing bilingual presentation, work in **readable-view localization** mode. Read the complete current R4 DOCX/HTML and its visible project ID, artifact ID, revision, and digest; use a matching independent review when provided. Translate all 48 episodes field by field, preserving the Chinese source, episode IDs, chronology, reveal order, state transitions, Canon/Promise/source refs, and all numeric constraints. Do not rebuild story architecture, silently apply a new R5 revision, or claim to have recomputed a canonical JSON digest from the DOCX. Label the result `BILINGUAL_VIEW_OF_R4` with `source_binding: REVIEW_ATTESTED_VIEW` when that is the available source. Keep the R4 digest as a displayed source binding only; a presentation translation does not alter the canonical Story Truth.

For each field, show `中文` and `English` together; identifiers and reference IDs may be repeated unchanged. Render the entire 48-episode view into both DOCX and standalone HTML. Check that EP01–EP48 are continuous and that every narrative field has two substantive values. Compare high-risk values—names, dates/times, capacities, legal/ownership facts, knowledge states, secrets, promises, outcomes, and cliffhangers—between languages. Fix translation drift before delivery. If a Chinese source field is genuinely missing, flag that exact field rather than inventing it.

## Ownership and boundary

Own the allocation of an already established story across episodes: episode entry/exit, local dramatic turns, reveal timing within approved windows, payoff timing, and the episode map. Do not invent or change the story's ending, character core, world rules, major reveal, antagonist identity, or approved story outcome. Those decisions belong to Story Architect and require an SCR to change.

Do not write screenplay scenes, dialogue, shot lists, director intent, asset appearance, or model prompts.

## Required inputs and gate

- Current Story Package draft and its exact revision/digest.
- `story-draft-review` report for that same revision with `PASS_FOR_EPISODE_ARCHITECTURE`, no unresolved mandatory finding, and complete evidence.
- Intended season scope, episode count, duration policy, and applicable user decisions.

If the report points to a different digest, any mandatory story finding remains unresolved, or the full story/climax/resolution is missing, return `BLOCKED`. A self-review, user-provided `APPROVED` label, model confidence, or episode outline cannot substitute for the required independent review.

## Workflow

1. Extract the approved story's causal turns, promises, reveal windows, character/relationship transitions, and ending. Cite the source IDs for each planned episode outcome.
2. Partition the complete season arc into episodes. Every episode must make a causal change to the season story; do not use repeated humiliation/revenge cycles that reset stakes.
3. Give every episode its own mini dramatic arc: a specific goal/problem, active conflict, escalation, emotional turn, reveal/reversal, earned payoff, and consequential unresolved question or cliffhanger. A finale may omit a cliffhanger only when its closing payoff and resolution of the series question are explicit.
4. Record entry and exit states for characters, relationships, knowledge/secrets, injuries, locations, and story-critical objects. A character cannot know a secret before a sourced reveal; consequences cannot disappear between episodes.
5. Validate full requested episode coverage and dependency links. Episode transitions must reconcile: previous exit state = next entry state unless an explicit, sourced transition explains the change.
6. Merge the episode map into a new Story Package draft revision. Preserve the full-story outline and immutable approved facts; attach provenance to all scheduling/allocation decisions. Return any contradiction or needed story change to Story Architect as an SCR, not as an implicit rewrite.

## Required fields per episode

Conform to [`episode-entry.schema.json`](../../contracts/episode-entry.schema.json) and include:

`episode_id`, `goal`, `opening_hook`, `immediate_problem`, `conflict`, `escalation`, `emotional_beat`, `reveal_or_reversal`, `payoff`, `unresolved_question`, `cliffhanger`, `outcome`, `entry_state`, `exit_state`, `character_changes`, `continuity_changes`, `canon_refs`, `promise_refs`, `source_refs`, and `finale`.

Write each field as meaningful, causally connected content. “Hero is humiliated,” “hero fights back,” or a title plus one sentence is not episode architecture. Episode entries are dramatic plans, not scripts: they do not contain numbered screenplay scenes, performed dialogue, or OS/voiceover lines. In the DOCX/HTML title block and handoff, explicitly identify this deliverable as **“48 集分集架构／Episode Architecture，不是逐集剧本”**. If the creator asks where the scenes, bilingual dialogue, or OS are, route that request to `usvd-v10-03-creator-script-draft` for the authorized episode range rather than regenerate the map and imply those elements are present.

## Output and next gate

Return the revised Story Package draft, complete Episode Map, coverage/transition report, unresolved items, and `STORY_PACKAGE_REVIEW_REQUIRED`. The next step is independent `usvd-v10-04-review-continuity` in `story-package-review` mode. Only that review may determine whether the exact complete package is ready to be presented for human approval. This skill cannot approve it or authorize Screenwriter execution.

For readable-view localization of an already reviewed R4, return only the paired bilingual DOCX/HTML, a compact field-completeness and meaning-alignment report, and the unchanged underlying R4 review/gate status. A translation-only view does not need a new story-content review or a new digest. If the English rendering changes a story fact, correct it before delivery; if the Chinese Story Truth itself needs to change, open an SCR and resume the normal revision/review route.

## Stop conditions

Stop if an episode requires an unapproved story decision, if a state transition cannot be reconciled, if the requested season is not fully covered, or if any episode lacks a local arc. Report a stable issue reference and route the issue to Story Architect; do not fill the gap by improvising canon.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: ../../contracts/episode-entry.schema.json

{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://usvds.invalid/schemas/v10/episode-entry.schema.json",
  "title": "USVDS V10 Episode Architecture Entry",
  "description": "Human-readable episode goals, hooks, conflicts, reversals and payoffs default to Simplified Chinese. Keep schema keys, IDs and necessary established names unchanged.",
  "type": "object",
  "required": ["episode_id", "goal", "opening_hook", "immediate_problem", "conflict", "escalation", "emotional_beat", "reveal_or_reversal", "payoff", "unresolved_question", "cliffhanger", "outcome", "entry_state", "exit_state", "character_changes", "continuity_changes", "canon_refs", "promise_refs", "source_refs", "finale"],
  "properties": {
    "episode_id": { "type": "string", "minLength": 1 },
    "goal": { "type": "string", "minLength": 1 },
    "opening_hook": { "type": "string", "minLength": 1 },
    "immediate_problem": { "type": "string", "minLength": 1 },
    "conflict": { "type": "string", "minLength": 1 },
    "escalation": { "type": "string", "minLength": 1 },
    "emotional_beat": { "type": "string", "minLength": 1 },
    "reveal_or_reversal": { "type": "string", "minLength": 1 },
    "payoff": { "type": "string", "minLength": 1 },
    "unresolved_question": { "type": ["string", "null"] },
    "cliffhanger": { "type": ["string", "null"] },
    "outcome": { "type": "string", "minLength": 1 },
    "entry_state": { "type": "object" },
    "exit_state": { "type": "object" },
    "character_changes": { "type": "array", "items": { "type": "object", "required": ["character_id", "from_state", "to_state", "cause_ref"], "properties": { "character_id": { "type": "string" }, "from_state": { "type": "string" }, "to_state": { "type": "string" }, "cause_ref": { "type": "string" } }, "additionalProperties": false } },
    "continuity_changes": { "type": "array", "items": { "type": "object", "required": ["subject_id", "state_kind", "from", "to", "cause_ref"], "properties": { "subject_id": { "type": "string" }, "state_kind": { "enum": ["knowledge", "relationship", "injury", "prop", "location", "promise"] }, "from": {}, "to": {}, "cause_ref": { "type": "string" } }, "additionalProperties": false } },
    "canon_refs": { "type": "array", "items": { "type": "string" } },
    "promise_refs": { "type": "array", "items": { "type": "string" } },
    "source_refs": { "type": "array", "items": { "type": "string" } },
    "finale": { "type": "boolean" },
    "closing_payoff": { "type": ["string", "null"] },
    "series_question_resolution": { "type": ["string", "null"] }
  },
  "allOf": [
    {
      "if": { "properties": { "finale": { "const": true } }, "required": ["finale"] },
      "then": { "required": ["closing_payoff", "series_question_resolution"], "properties": { "closing_payoff": { "type": "string", "minLength": 1 }, "series_question_resolution": { "type": "string", "minLength": 1 } } },
      "else": { "properties": { "unresolved_question": { "type": "string", "minLength": 1 }, "cliffhanger": { "type": "string", "minLength": 1 } } }
    }
  ],
  "additionalProperties": false
}


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
- When creating/updating this stage's authoritative artifact, preserve the document tag `usvds:artifact:episode-architecture` on that DramaGo Document so Gate routing can locate it deterministically.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
