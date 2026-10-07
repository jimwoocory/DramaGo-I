---
name: usvd-v10-01-story-architect
description: Develop and revise the complete story architecture for a US-facing vertical drama before episode-level planning or screenplay writing.
title: USVDS V11 · Story Architect
hint:
  workflow: usvds-v11
  stage: story
  baseline: 0c68d9c
---

# USVDS V10 — Story Architect

## Default output language

Read [the output language contract](../../references/output-language.md). Write the complete story outline, character and relationship arcs, major-turn explanations, Story Package reading view, and human-readable JSON values in 简体中文 by default. Retain schema keys, IDs, source locators, and necessary established English names. Do not write an English outline merely because the target story world is American. An explicit creator language request overrides this default.

Read [the document delivery contract](../../references/document-delivery.md). Export the complete Chinese Story Package reading view as matching DOCX and HTML files from the same revision; do not hand over TXT. The canonical Story Package JSON remains the internal source of truth, not the creator's only readable deliverable.

## Ownership and boundary

Own story decisions: the promise, protagonist and opposing force, causal story engine, character and relationship architecture, complete beginning-to-ending story, season arc, planned reveals/promises, and narrative asset requirements. Do not write episode-by-episode maps, screenplay scenes, visual designs, camera direction, image/video prompts, or claim human approval.

The single project Story Truth is the current Story Package JSON revision. Markdown is a reading view only. Every major decision must carry provenance; user/source facts and approved decisions must remain distinguishable from proposals and assumptions.

## Required inputs

- A `BRIEF_READY` Project Brief with source coverage and adaptation boundaries.
- The Brief's `direction_selection.status=user_selected` and nonempty `selected_direction_id`. Bind the Story Package to that ID; a candidate or AI proposal is not a creator decision. If the current intake route is an outline audit, diagnose it first and enter full-story construction only after a direction is selected.
- All supplied source material or excerpts within the stated scope, with stable locators where available.
- For revision, the exact current Story Package and any Story Change Request (SCR), including affected dependencies and unresolved review findings.

If scope, source, or a material decision is missing, ask a focused question or mark the uncertainty. Do not fabricate missing source events, US institutions, market findings, character history, or approval status.


## Workflow

1. **Selected direction and Story Promise.** Read the creator's selected direction from the exact Brief revision. Preserve its intended audience payoff, central conflict and limits. State why the audience keeps watching, the protagonist's strongest desire, the central obstacle, emotional experience, core payoff/satisfaction, and central suspense question. Make the promise specific enough to reject superficially similar settings that use a different conflict.
2. **Logline.** Write one concise sentence expressing who wants what, against what force, and at what stakes. If that causal proposition is unclear, mark `STORY_DRAFT_BLOCKED` and resolve it before outlining.
3. **Story Engine and US-world mechanism.** Define protagonist, opposing force and credible leverage, external goal, internal need, conflict engine, escalation mechanism, stakes, time/pressure source, reversal sources, and renewable episode-hook sources. Explain who holds power in the US setting, what concretely constrains the protagonist, which available action they choose, what it costs, and how opposition can respond. Explain why the engine can sustain the requested run without repeating or resetting conflict. Do not present an unfamiliar US legal, workplace, family, medical, financial, or social rule as fact without evidence; mark assumptions that require verification.
4. **Adaptation mechanism.** For source adaptation, compare each story-driving source mechanism with the selected US replacement: retain the emotional payoff, name the source's actual power/dependency logic, explain the new credible US leverage, protagonist option, cost, and plausibility basis. Expand the Brief's `adaptation_mechanism_map` without changing its `decision_status`; record source locators and decision provenance. Do not promote an `ai_proposal` or `unresolved` mechanism into a creator decision. A translated line, English name, US city, or substituted institution does not constitute a replacement. If the source's decisive pressure still requires the original social system, redesign the conflict before proceeding. For an original idea, use an empty mapping and still justify the US-world mechanism.
5. **Character and relationship architecture.** For each core character record identity, visible goal, hidden need, fear, secret, weakness, strength/leverage, protagonist relationship, conflict function, and intended transformation. Map only relationships that cause meaningful alliance, opposition, dependency, leverage, betrayal potential, or emotional bond. Define per-character states and valid transitions from story evidence; do not impose a universal growth ladder.
6. **Full-Series Story Outline.** Write connected narrative prose that explains the whole story from Beginning through Development, Escalation, Major Reversal, Crisis, Climax, and Resolution. Each major turn must arise from an earlier cause and record the protagonist's consequential choice, the opposing response, the cost or persistent state change, and the pressure that forces the next turn. Identify meaningful setups/payoffs and the ending's consequences. If the protagonist waits while coincidences, secrets or rescuers resolve the story, rebuild the turn. A title list, three-act labels, detached bullets, or a few-sentence concept is not a completed outline.
7. **Season/Arc plan.** Define the season question, major turns, reveal windows, escalation, finale outcome, unresolved promises, and which developments belong to later seasons, if any. Give each `major_turns[]` entry a choice, counteraction, consequence and next pressure tied to the outline; for the terminal turn, explain how the chain closes. This is macro architecture, not an episode list.
8. **Planned state, secrets, promises, and narrative assets.** Record what is true, who knows what and when, relationship/character state transitions, setup and payoff windows, and what story-critical people/places/props/vehicles must exist. Describe asset function only; appearance and generation prompts belong downstream.
9. **Provenance and revision.** Assign stable IDs to important decisions and provenance (`user | source | market_evidence | ai_proposal | approved_story`). Cite exact source locators for adapted facts. Do not silently edit an approved revision: submit a patch/SCR and identify affected episodes, characters, assets, canon, and downstream outputs.

## Semantic readiness checks

Before handoff, explicitly assess:

- protagonist goal and agency; credible antagonist/opposing leverage;
- sustainable causal conflict and visible escalation;
- relationship logic and character motives;
- complete climax and resolution supported by setup;
- reversals caused by character action/evidence rather than coincidence;
- consequence retention rather than conflict resets;
- clarity, shootability, exposition load, and unnecessary character complexity;
- US-facing social/institutional plausibility without inventing research;
- selected direction consistency and a mechanism-level adaptation rather than names, setting labels or literal English substitution;
- for each major turn, an observable choice → counteraction → cost/state change → next pressure chain.

These are semantic review questions, not deterministic facts. Do not mark the story approved based on a score, field presence, model confidence, or self-review. Return `STORY_DRAFT_READY` only for independent review by `usvd-v10-04-review-continuity` in `story-draft-review` mode.

## Output contract

Produce a Story Package draft conforming to [`story-package.schema.json`](../../contracts/story-package.schema.json), including the Brief-bound `selected_direction_id`, `adaptation_mechanism_map`, `story_promise`, `logline`, `story_engine`, a complete prose `full_story_outline`, `characters`, `relationships`, `state_definitions`, `season_arc`, `canon_facts`, `secrets`, `promises`, `narrative_asset_requirements`, `source_refs`, and `provenance`. Set a new monotonically increasing `revision`, parent reference, and draft status. Include an accompanying Markdown reading view that repeats the package revision and digest when supplied by the caller; never treat that view as an editable second source of truth.

End with `STORY_DRAFT_READY` or `STORY_DRAFT_BLOCKED`, unresolved decisions, source coverage, and the single next action. Never hand directly to Episode Architect without an independent story-draft review.

## Stop conditions

Do not create episode titles as a substitute for the full outline. Do not write dialogue or screenplay. Do not convert market popularity into story truth. Do not mark a human gate `APPROVED`; the trusted approval runtime is not implemented, so Screenwriter execution remains blocked in V10.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: ../../contracts/story-package.schema.json

{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://usvds.invalid/schemas/v10/story-package.schema.json",
  "title": "USVDS V10 Story Package",
  "description": "Canonical project story truth. Human-readable story prose and Markdown reading views default to Simplified Chinese; JSON keys, IDs, status codes and necessary names remain unchanged. Review and human approval are separately bound to this package revision and digest.",
  "type": "object",
  "required": ["schema_version", "project_id", "artifact_id", "revision", "parent_revision", "status", "brief_ref", "selected_direction_id", "source_refs", "adaptation_mechanism_map", "story_promise", "logline", "story_engine", "full_story_outline", "characters", "relationships", "state_definitions", "season_arc", "episodes", "canon_facts", "secrets", "promises", "narrative_asset_requirements", "provenance"],
  "properties": {
    "schema_version": { "const": "usvd-story-package/v2" },
    "project_id": { "type": "string", "minLength": 1 },
    "artifact_id": { "type": "string", "minLength": 1 },
    "revision": { "type": "integer", "minimum": 1 },
    "parent_revision": { "type": ["integer", "null"], "minimum": 1 },
    "status": { "enum": ["DRAFT", "REVIEW_REQUIRED", "HUMAN_APPROVAL_REQUIRED", "APPROVED", "CHANGE_REQUESTED", "SUPERSEDED", "STALE"] },
    "brief_ref": { "$ref": "#/$defs/artifactRef" },
    "selected_direction_id": { "type": "string", "minLength": 1, "description": "Must equal direction_selection.selected_direction_id in the exact referenced BRIEF_READY Project Brief revision; this is a reference, not an independent selection." },
    "source_refs": { "type": "array", "items": { "$ref": "#/$defs/sourceRef" } },
    "adaptation_mechanism_map": { "type": "array", "description": "Empty for original ideas. Source adaptation requires mechanism-level replacement for story-driving power, relationship, cost, and causality, with source and provenance binding.", "items": { "type": "object", "required": ["dimension", "source_ref", "emotional_payoff", "source_mechanism", "us_mechanism", "us_plausibility", "protagonist_choice", "cost_or_consequence", "decision_status", "provenance"], "properties": { "dimension": { "enum": ["power", "relationship", "cost", "causality"] }, "source_ref": { "type": "string", "minLength": 1 }, "emotional_payoff": { "type": "string", "minLength": 1 }, "source_mechanism": { "type": "string", "minLength": 1 }, "us_mechanism": { "type": "string", "minLength": 1 }, "us_plausibility": { "type": "string", "minLength": 1 }, "protagonist_choice": { "type": "string", "minLength": 1 }, "cost_or_consequence": { "type": "string", "minLength": 1 }, "decision_status": { "enum": ["user_decided", "ai_proposal", "unresolved"] }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false } },
    "story_promise": { "type": "object", "required": ["audience_reason_to_continue", "protagonist_desire", "central_obstacle", "emotional_experience", "core_payoff", "central_question"], "properties": { "audience_reason_to_continue": { "type": "string", "minLength": 1 }, "protagonist_desire": { "type": "string", "minLength": 1 }, "central_obstacle": { "type": "string", "minLength": 1 }, "emotional_experience": { "type": "string", "minLength": 1 }, "core_payoff": { "type": "string", "minLength": 1 }, "central_question": { "type": "string", "minLength": 1 } }, "additionalProperties": false },
    "logline": { "type": "object", "required": ["who", "wants", "against", "stakes", "sentence"], "properties": { "who": { "type": "string", "minLength": 1 }, "wants": { "type": "string", "minLength": 1 }, "against": { "type": "string", "minLength": 1 }, "stakes": { "type": "string", "minLength": 1 }, "sentence": { "type": "string", "minLength": 1 } }, "additionalProperties": false },
    "story_engine": { "type": "object", "required": ["protagonist_id", "opposing_force_id", "external_goal", "internal_need", "conflict_engine", "us_power_logic", "opposing_leverage", "protagonist_available_action", "cost_of_action", "escalation_mechanism", "stakes", "pressure_source", "reversal_sources", "episode_hook_sources", "sustainability_explanation"], "properties": { "protagonist_id": { "type": "string" }, "opposing_force_id": { "type": "string" }, "external_goal": { "type": "string", "minLength": 1 }, "internal_need": { "type": "string", "minLength": 1 }, "conflict_engine": { "type": "string", "minLength": 1 }, "us_power_logic": { "type": "string", "minLength": 1 }, "opposing_leverage": { "type": "string", "minLength": 1 }, "protagonist_available_action": { "type": "string", "minLength": 1 }, "cost_of_action": { "type": "string", "minLength": 1 }, "escalation_mechanism": { "type": "string", "minLength": 1 }, "stakes": { "type": "string", "minLength": 1 }, "pressure_source": { "type": "string", "minLength": 1 }, "reversal_sources": { "type": "array", "items": { "type": "string" }, "minItems": 1 }, "episode_hook_sources": { "type": "array", "items": { "type": "string" }, "minItems": 1 }, "sustainability_explanation": { "type": "string", "minLength": 1 } }, "additionalProperties": false },
    "full_story_outline": { "type": "object", "required": ["beginning", "development", "escalation", "major_reversal", "crisis", "climax", "resolution"], "properties": { "beginning": { "$ref": "#/$defs/prose" }, "development": { "$ref": "#/$defs/prose" }, "escalation": { "$ref": "#/$defs/prose" }, "major_reversal": { "$ref": "#/$defs/prose" }, "crisis": { "$ref": "#/$defs/prose" }, "climax": { "$ref": "#/$defs/prose" }, "resolution": { "$ref": "#/$defs/prose" } }, "additionalProperties": false },
    "characters": { "type": "array", "items": { "type": "object", "required": ["id", "identity", "visible_goal", "hidden_need", "fear", "secret_refs", "weakness", "strength", "leverage", "protagonist_relationship", "conflict_function", "transformation", "provenance"], "properties": { "id": { "type": "string", "minLength": 1 }, "identity": { "type": "string", "minLength": 1 }, "visible_goal": { "type": "string", "minLength": 1 }, "hidden_need": { "type": "string", "minLength": 1 }, "fear": { "type": "string", "minLength": 1 }, "secret_refs": { "type": "array", "items": { "type": "string" } }, "weakness": { "type": "string", "minLength": 1 }, "strength": { "type": "string", "minLength": 1 }, "leverage": { "type": "string", "minLength": 1 }, "protagonist_relationship": { "type": "string", "minLength": 1 }, "conflict_function": { "type": "string", "minLength": 1 }, "transformation": { "type": "string", "minLength": 1 }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false } },
    "relationships": { "type": "array", "items": { "type": "object", "required": ["id", "from_character_id", "to_character_id", "alliance", "opposition", "dependency", "leverage", "betrayal_potential", "emotional_bond", "provenance"], "properties": { "id": { "type": "string" }, "from_character_id": { "type": "string" }, "to_character_id": { "type": "string" }, "alliance": { "type": "string" }, "opposition": { "type": "string" }, "dependency": { "type": "string" }, "leverage": { "type": "string" }, "betrayal_potential": { "type": "string" }, "emotional_bond": { "type": "string" }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false } },
    "state_definitions": { "type": "array", "items": { "type": "object", "required": ["character_id", "states", "transitions"], "properties": { "character_id": { "type": "string" }, "states": { "type": "array", "items": { "type": "string" }, "minItems": 1 }, "transitions": { "type": "array", "items": { "type": "object", "required": ["from", "to", "cause_ref"], "properties": { "from": { "type": "string" }, "to": { "type": "string" }, "cause_ref": { "type": "string" } }, "additionalProperties": false } } }, "additionalProperties": false } },
    "season_arc": { "type": "object", "required": ["season_id", "episode_count", "season_question", "major_turns", "reveal_windows", "finale_outcome", "unresolved_promises"], "properties": { "season_id": { "type": "string" }, "episode_count": { "type": "integer", "minimum": 1 }, "season_question": { "type": "string", "minLength": 1 }, "major_turns": { "type": "array", "items": { "$ref": "#/$defs/majorTurn" } }, "reveal_windows": { "type": "array", "items": { "$ref": "#/$defs/plannedTurn" } }, "finale_outcome": { "type": "string", "minLength": 1 }, "unresolved_promises": { "type": "array", "items": { "type": "string" } } }, "additionalProperties": false },
    "episodes": { "type": "array", "items": { "$ref": "episode-entry.schema.json" } },
    "canon_facts": { "type": "array", "items": { "$ref": "#/$defs/provenancedFact" } },
    "secrets": { "type": "array", "items": { "type": "object", "required": ["id", "holder_ids", "known_by", "planned_reveal", "provenance"], "properties": { "id": { "type": "string" }, "holder_ids": { "type": "array", "items": { "type": "string" } }, "known_by": { "type": "array", "items": { "type": "string" } }, "planned_reveal": { "type": ["string", "null"] }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false } },
    "promises": { "type": "array", "items": { "$ref": "#/$defs/provenancedFact" } },
    "narrative_asset_requirements": { "type": "array", "items": { "type": "object", "required": ["id", "kind", "story_function", "first_needed_ref", "provenance"], "properties": { "id": { "type": "string" }, "kind": { "enum": ["character", "costume", "prop", "vehicle", "location", "story_critical_object"] }, "story_function": { "type": "string", "minLength": 1 }, "first_needed_ref": { "type": "string" }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false } },
    "provenance": { "type": "array", "items": { "$ref": "#/$defs/provenance" } },
    "approval_ref": { "$ref": "#/$defs/approvalRef" }
  },
  "$defs": {
    "prose": { "type": "string", "minLength": 1, "description": "Connected narrative prose; completeness and causal sufficiency are semantic review responsibilities, not character-count checks." },
    "artifactRef": { "type": "object", "required": ["artifact_id", "revision", "digest"], "properties": { "artifact_id": { "type": "string" }, "revision": { "type": "integer", "minimum": 1 }, "digest": { "type": "string", "pattern": "^[a-f0-9]{64}$" } }, "additionalProperties": false },
    "sourceRef": { "type": "object", "required": ["source_id", "locator", "coverage_status"], "properties": { "source_id": { "type": "string" }, "locator": { "type": "string" }, "coverage_status": { "enum": ["supplied", "observed", "partial", "unavailable"] } }, "additionalProperties": false },
    "provenance": { "type": "object", "required": ["decision_id", "kind", "statement"], "properties": { "decision_id": { "type": "string" }, "kind": { "enum": ["user", "source", "market_evidence", "ai_proposal", "approved_story"] }, "statement": { "type": "string", "minLength": 1 }, "source_ref": { "type": "string" }, "date": { "type": "string", "format": "date" } }, "additionalProperties": false },
    "provenancedFact": { "type": "object", "required": ["id", "statement", "provenance"], "properties": { "id": { "type": "string" }, "statement": { "type": "string", "minLength": 1 }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false },
    "plannedTurn": { "type": "object", "required": ["id", "summary", "episode_window", "cause_refs", "provenance"], "properties": { "id": { "type": "string" }, "summary": { "type": "string", "minLength": 1 }, "episode_window": { "type": "string" }, "cause_refs": { "type": "array", "items": { "type": "string" } }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false },
    "majorTurn": { "type": "object", "required": ["id", "summary", "episode_window", "cause_refs", "protagonist_choice", "opposing_counteraction", "cost_or_state_change", "next_pressure", "provenance"], "properties": { "id": { "type": "string", "minLength": 1 }, "summary": { "type": "string", "minLength": 1 }, "episode_window": { "type": "string" }, "cause_refs": { "type": "array", "items": { "type": "string" }, "minItems": 1 }, "protagonist_choice": { "type": "string", "minLength": 1 }, "opposing_counteraction": { "type": "string", "minLength": 1 }, "cost_or_state_change": { "type": "string", "minLength": 1 }, "next_pressure": { "type": "string", "minLength": 1 }, "provenance": { "$ref": "#/$defs/provenance" } }, "additionalProperties": false },
    "approvalRef": { "type": "object", "required": ["approval_id", "revision", "digest", "scope"], "properties": { "approval_id": { "type": "string" }, "revision": { "type": "integer", "minimum": 1 }, "digest": { "type": "string", "pattern": "^[a-f0-9]{64}$" }, "scope": { "type": "string" } }, "additionalProperties": false }
  },
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
