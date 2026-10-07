---
name: usvd-v10-04-review-continuity
description: Diagnose supplied outlines and independently review V10 story drafts or complete episode packages with evidence-bound findings; do not rewrite or approve them as a human.
title: USVDS V11 · Independent Review
hint:
  workflow: usvds-v11
  stage: review
  baseline: 0c68d9c
---

# USVDS V10 — Review and Continuity

## Default output language

Read [the output language contract](../../references/output-language.md). Write diagnostic conclusions, review evidence, required repairs, unassessed scope explanations, and the report reading view in 简体中文 by default. Keep rule IDs, issue types, schema keys, and status codes in their required form. English-only prose in an early Chinese development artifact is an `output_language_mismatch` delivery blocker. For episode maps, scripts, and prompt/shot/asset reading views, a missing Chinese or English counterpart is a `bilingual_coverage_mismatch` delivery blocker. Correct a translation-only readable view without claiming a new story revision or a new independent story review; translation alone does not repair story logic.

Read [the document delivery contract](../../references/document-delivery.md). Deliver the diagnostic or review report as matching DOCX and HTML files by default. Confirm both show the same reviewed artifact ID, revision, findings, and verdict. A TXT file or JSON-only output is not the creator-facing report.

## Modes

Run exactly one mode per invocation:

- `outline-diagnostic`: assess an existing outline or script-derived story summary before a creator selects a V10 direction. This is diagnostic only and cannot unlock story or episode stages.
- `story-draft-review`: review the Story Architect's complete story before episode planning.
- `story-package-review`: review the full Story Package after episode architecture.
- `script-review` and `continuity-project`: reserved for a later V10 implementation; currently return `BLOCKED`.

## Independence and binding

Read the actual artifact being reviewed and its source/brief references. For V10 Story Package modes, identify its `project_id`, `artifact_id`, `revision`, and digest; verify the report is about this exact revision. For `outline-diagnostic`, identify the supplied file, excerpt, or user label as `source_label` and assign an `artifact_id` for this diagnostic target; record the digest and revision if available, or set both to `null` and `binding_status=source_label_only`. State that exact-content binding is unverified in the latter case and cite observable spans. Do not edit the artifact, propose replacement prose as if applied, accept the author's self-score, or inherit findings from a different revision without rechecking them.

This skill can issue a semantic review recommendation only. It cannot authenticate a human, create trusted approval evidence, set `APPROVED`, unlock Screenwriter, or claim audience/commercial validation. V10's trusted approval runtime is blocked by ND-001.

## Story review rubric

For all three story modes, assess relevant items with a stable finding ID, issue type, severity, cited story span/decision ID, evidence, and required action:

1. The chosen direction and its audience promise match the referenced Brief decision. In `outline-diagnostic`, identify the direction the existing material actually takes without claiming the creator selected it.
2. Promise and logline clearly express audience draw, protagonist desire, opposition, and stakes.
3. The protagonist makes consequential choices and has a comprehensible external goal and internal need.
4. The opposing force has credible leverage; conflict can sustain the requested length without repetitive resets.
5. Every major turn follows an earlier cause and changes stakes, cost, power, information, or relationships through protagonist choice, opposing counteraction, and a retained consequence that generates the next pressure. Flag coincidence, withheld information, or a rescue that substitutes for causality.
6. Character motives and relationship leverage are clear, limited to useful complexity, and state changes have causes.
7. The full story has connected beginning, development, escalation, major reversal, crisis, climax, and resolution; climax is set up and the ending pays off the promise.
8. Secrets, promises, reveal windows, source adaptations, and planned character knowledge do not contradict one another.
9. For source adaptation, trace emotional payoff, source power/dependency logic, US replacement mechanism, protagonist option, and cost. Flag `surface_transplant` when renamed people, translated dialogue, substituted places/institutions, or the original plot's unchanged power logic is presented as an American story. The reviewer must cite both source and draft evidence where available; missing source coverage is a separate finding, never a guessed comparison.
10. US-facing social/institutional logic is plausible for the chosen world; uncertainty is marked rather than presented as researched fact. Flag `translation_register` when dialogue, social roles, forms of address, obligations, or institutional behavior visibly preserve a Chinese-language or Chinese-market assumption despite English wording. Cite the specific term/interaction and the US-world mechanism it contradicts; do not reject ordinary bilingual or immigrant-character speech on its own.
11. Exposition does not substitute for visible action and conflict; the story is producible in the requested vertical-drama format.
12. Provenance distinguishes user/source facts, dated market evidence, AI proposals, and approved story decisions.
13. Creator-facing prose follows the Brief and current output-language contract. The US market is not a reason to write an English-only outline. For bilingual episode/script/prompt/shot/asset delivery, check field-by-field pairing, meaning alignment, executable dialogue designation, and one-language model-submission designation.
14. At delivery, the readable report and any requested story artifact exist as matching DOCX and HTML files unless the creator explicitly changed formats. Flag `delivery_format_mismatch` if only TXT, chat prose, or a renamed non-DOCX file is offered. This is a delivery defect, not evidence that the story itself passes or fails.

Treat `surface_transplant`, `translation_register` where it changes story logic, and `causal_break` as mandatory blockers in V10 Story Package modes. A finding must describe the actual failed mechanism and evidence, rather than infer a problem from names, ethnicity, genre, or an automated keyword count. Local style issues that do not affect story logic may be minor.

Semantic quality must cite evidence. JSON field presence, word counts, a score, or a keyword scan cannot establish causality, hook strength, plausibility, or emotional payoff. Record structural/machine-check results separately from semantic findings.

## Mode-specific decision

### `outline-diagnostic`

Read the supplied outline or script and, if needed, extract a brief reverse story summary: promise, protagonist strategy, opposing leverage, key turns, and ending actually present in the material. Diagnose what can be retained, which turns fail causally or culturally, and which source gaps prevent judgment. Do not silently rewrite the material or imply a selected V10 direction. Return `DIAGNOSTIC_COMPLETE` when evidence supports a useful assessment, `BLOCKED` when the provided material cannot support one, or `REJECT` when the supplied direction's premise cannot be repaired within creator limits. Never return a PASS verdict in this mode. A complete diagnostic still requires creator direction selection and a canonical Story Package before episode architecture.

### `story-draft-review`

Require a complete outline through resolution, a Brief-bound selected direction, story engine, main character/relationship architecture, and relevant source coverage. If any mandatory story logic item fails, return `REWRITE` or `BLOCKED`; do not pass an outline because its headings are populated. Only a clean report on a digest-bound revision may return `PASS_FOR_EPISODE_ARCHITECTURE`.

### `story-package-review`

Require the complete intended episode map and its consistency with the reviewed story draft and selected direction. Check every episode for goal, hook/problem, conflict, escalation, emotional beat, reveal/reversal, payoff, unresolved question/cliffhanger or finale close, outcome, entry/exit states, continuity changes, and source/canon/promise references. Check season coverage and adjacent state transitions. On clean review of a digest-bound revision, return `PASS_AWAITING_HUMAN_APPROVAL`; this is a recommendation to present the exact package for human review, not an approval event.

## Output contract

Return a Review Report conforming to [`review-report.schema.json`](../../contracts/review-report.schema.json), containing `review_id`, project/artifact reference, `binding_status`, revision/digest when known, diagnostic `source_label` when applicable, mode/scope, input references, reviewer role (`ai_reviewer`, never `human`), rubric version, structural results, `findings[]` with stable IDs, `issue_type`, evidence and actions, verdict, and `unassessed_scopes`.

Allowed story verdicts: `DIAGNOSTIC_COMPLETE`, `PASS_FOR_EPISODE_ARCHITECTURE`, `PASS_AWAITING_HUMAN_APPROVAL`, `REWRITE`, `REJECT`, `BLOCKED`, subject to mode restrictions above. Never output a human `APPROVED` decision. Stop after the report; do not rewrite the artifact or execute a downstream stage.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: ../../contracts/review-report.schema.json

{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://usvds.invalid/schemas/v10/review-report.schema.json",
  "title": "USVDS V10 Review Report",
  "description": "Review evidence, required actions, conclusions and reading views default to Simplified Chinese unless the creator explicitly requests another development language; machine codes remain unchanged.",
  "type": "object",
  "required": ["review_id", "project_id", "artifact_id", "revision", "digest", "binding_status", "mode", "scope", "input_refs", "reviewer", "independence_statement", "rubric_version", "structural_results", "findings", "verdict", "unassessed_scopes"],
  "properties": {
    "review_id": { "type": "string", "minLength": 1 },
    "project_id": { "type": "string", "minLength": 1 },
    "artifact_id": { "type": "string", "minLength": 1 },
    "revision": { "type": ["integer", "null"], "minimum": 1 },
    "digest": { "type": ["string", "null"], "pattern": "^[a-f0-9]{64}$" },
    "binding_status": { "enum": ["verified_digest", "source_label_only"] },
    "source_label": { "type": "string", "minLength": 1, "description": "Required for a diagnostic on unversioned supplied material; identify the exact file, excerpt, or user-supplied label." },
    "mode": { "enum": ["outline-diagnostic", "story-draft-review", "story-package-review", "script-review", "continuity-project"] },
    "scope": { "type": "string", "minLength": 1 },
    "input_refs": { "type": "array", "items": { "type": "string" } },
    "reviewer": { "const": "ai_reviewer" },
    "independence_statement": { "type": "string", "minLength": 1 },
    "rubric_version": { "type": "string", "minLength": 1 },
    "structural_results": { "type": "array", "items": { "type": "object", "required": ["check_id", "result"], "properties": { "check_id": { "type": "string" }, "result": { "enum": ["PASS", "FAIL", "NOT_RUN"] }, "evidence_ref": { "type": "string" } }, "additionalProperties": false } },
    "findings": { "type": "array", "items": { "type": "object", "required": ["finding_id", "class", "issue_type", "severity", "stable_ref", "evidence", "required_action"], "properties": { "finding_id": { "type": "string" }, "class": { "enum": ["structural", "semantic", "craft"] }, "issue_type": { "enum": ["direction_mismatch", "surface_transplant", "translation_register", "output_language_mismatch", "delivery_format_mismatch", "causal_break", "agency_gap", "opposition_gap", "escalation_reset", "character_logic", "payoff_gap", "source_coverage", "provenance_gap", "production_fit", "continuity", "other"] }, "severity": { "enum": ["mandatory", "major", "minor", "note"] }, "stable_ref": { "type": "string" }, "evidence": { "type": "string" }, "required_action": { "type": "string" } }, "additionalProperties": false } },
    "verdict": { "enum": ["DIAGNOSTIC_COMPLETE", "PASS_FOR_EPISODE_ARCHITECTURE", "PASS_AWAITING_HUMAN_APPROVAL", "REWRITE", "REJECT", "BLOCKED"] },
    "unassessed_scopes": { "type": "array", "items": { "type": "string" } },
    "machine_result_ref": { "type": ["string", "null"] }
  },
  "allOf": [
    { "if": { "properties": { "mode": { "const": "outline-diagnostic" } }, "required": ["mode"] }, "then": { "required": ["source_label"], "properties": { "verdict": { "enum": ["DIAGNOSTIC_COMPLETE", "REJECT", "BLOCKED"] } } } },
    { "if": { "properties": { "mode": { "enum": ["story-draft-review", "story-package-review"] } }, "required": ["mode"] }, "then": { "properties": { "revision": { "type": "integer", "minimum": 1 }, "digest": { "type": "string", "pattern": "^[a-f0-9]{64}$" }, "binding_status": { "const": "verified_digest" }, "verdict": { "enum": ["PASS_FOR_EPISODE_ARCHITECTURE", "PASS_AWAITING_HUMAN_APPROVAL", "REWRITE", "REJECT", "BLOCKED"] } } } },
    { "if": { "properties": { "binding_status": { "const": "source_label_only" } }, "required": ["binding_status"] }, "then": { "properties": { "revision": { "type": "null" }, "digest": { "type": "null" } } } },
    { "if": { "properties": { "binding_status": { "const": "verified_digest" } }, "required": ["binding_status"] }, "then": { "properties": { "digest": { "type": "string", "pattern": "^[a-f0-9]{64}$" } } } },
    { "if": { "properties": { "verdict": { "const": "PASS_FOR_EPISODE_ARCHITECTURE" } }, "required": ["verdict"] }, "then": { "properties": { "mode": { "const": "story-draft-review" } } } },
    { "if": { "properties": { "verdict": { "const": "PASS_AWAITING_HUMAN_APPROVAL" } }, "required": ["verdict"] }, "then": { "properties": { "mode": { "const": "story-package-review" } } } }
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
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
