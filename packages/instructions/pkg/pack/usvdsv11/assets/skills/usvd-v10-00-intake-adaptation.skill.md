---
name: usvd-v10-00-intake-adaptation
description: Identify the material the creator actually has, route idea development, outline diagnosis, source adaptation, or script diagnosis, and prepare a traceable US-facing Project Brief without writing the full story.
title: USVDS V11 · Intake / Adaptation
hint:
  workflow: usvds-v11
  stage: intake
  baseline: 0c68d9c
---

# USVDS V10 — Intake and Adaptation

## Default output language

Read [the output language contract](../../references/output-language.md). Write the Project Brief reading view, direction cards, adaptation decisions, explanations, and `next_step_for_user` in 简体中文 by default. Set `output_languages` to include `zh-CN-development`; note `en-US-dialogue` only as the later screenplay dialogue policy. Keep JSON keys, status codes, IDs, and necessary original names unchanged. An American setting never implies an English development document. An explicit creator request may override this default.

Read [the document delivery contract](../../references/document-delivery.md). Record `delivery_formats: ["docx", "html"]` in new Briefs by default. Supply the readable Brief and any direction comparison as matching downloadable DOCX and standalone HTML files; do not deliver TXT. Keep the Brief JSON as an internal canonical artifact or optional audit attachment, never as a replacement for the two readable files.

## Ownership

Own the Project Brief, intake route, source coverage, direction preview, adaptation scope, and unresolved intake decisions. Do not own the full story, independent outline verdict, season outline, episode map, screenplay, market conclusions, or human approval. Read [intake routing](../../references/intake-routing.md) when classifying an ambiguous request or proposing directions.

## Inputs

Accept an original idea or market concept, a completed outline, a novel/comic/film or other source work, a source with the creator's chosen adaptation direction, or an existing screenplay. Record which materials the user supplied and which are references only. For adapted works, ask whether the user owns or is authorized to adapt the supplied material when that is unclear. Do not fetch, quote, or reconstruct unsupplied copyrighted source text.

Collect or infer only reversible defaults for US-facing audience/market, format, intended episode count and runtime, rating/tone, adaptation limits, and requested delivery scope. Default the development-document language to Chinese and later screenplay dialogue to natural US English; ask about language only when the creator requests an exception. Ask one focused question when an unresolved choice changes the core premise or rights scope.

## Route before writing

Classify the work by the strongest supplied artifact and the user's immediate question, using exactly one `intake_route`:

| Supplied material and request | `intake_route` | First deliverable / `next_action` |
| --- | --- | --- |
| A simple idea or market concept with no stable conflict engine | `simple_idea` | 2–3 distinct direction cards; `choose_direction` |
| A completed outline whose fit or quality is uncertain | `completed_outline_audit` | Source-bound outline diagnosis; `audit_outline` |
| A novel, comic, film, or story source with no chosen US direction | `source_adaptation` | Source coverage, adaptation tradeoffs, 2–3 US direction cards; `choose_direction` |
| A source and a direction explicitly chosen by the creator | `source_selected_direction` | Check that the chosen direction has a credible US conflict mechanism; `validate_selected_direction` |
| An existing script whose actual story must be assessed | `existing_script_reverse_story` | Reconstruct the story the script actually tells, then diagnose it; `reverse_story_then_audit` |

If the creator has already explicitly chosen a viable direction for a simple idea, record it as `user_selected` and set `next_action: develop_full_story`. If rights, source scope, or a premise-changing choice is unresolved, use `resolve_material_decision` and state the exact decision. A diagnostic route can be `BRIEF_READY` with `direction_selection.status: not_applicable`; the diagnostic is the next action, not a full-story drafting gate. Never treat an existing outline or script as automatically approved.

## Workflow

1. Identify source kind, source boundary, and what was actually read. Preserve stable locators such as chapter, page, scene, timestamp, or user-provided passage. Mark unavailable or truncated ranges explicitly.
2. Separate source facts, user decisions, market evidence, AI proposals, and assumptions. Market information is evidence with source/date, never story truth by itself.
3. For adaptation, create a decision matrix with `retain`, `cut`, `merge`, and `redesign` candidates. For every material cut or merge, state the dramatic consequence (“this means…”), the affected character/plot functions, and an exact source locator or mark the proposal as unsupported.
4. Identify the source's emotional promise, then map the power, relationship, cost, and causal mechanisms that make it work. For each, state the US mechanism that could produce the same emotional payoff and why it is credible for these characters and this setting. Mark unverified institutional details as uncertainty. A change of names, location labels, or translated dialogue does not pass this gate; if the story still depends on a source-culture mechanism that has no credible US equivalent, return it for redesign.
5. For an unclear premise or unchosen adaptation, offer 2–3 short direction cards. Keep the emotional promise where appropriate, but change the protagonist's strategy, opposing leverage, cost, and US power logic enough that each direction would generate a different chain of choices and reversals. State the tradeoff and ask the creator to choose. Do not infer selection from an AI recommendation.
6. Record a creator-selected direction with a stable `direction_id`, `direction_selection.status: user_selected`, and a source-backed summary. The selected ID must match one entry in `direction_candidates`. If the creator supplied that direction, include one matching entry; this is a record of their direction, not a claim that alternatives were approved. If cards await a choice, use `NEEDS_USER_DECISION` and a null selected ID.
7. Identify high-level character candidates and the story function each serves, but do not finalize the cast or full character arcs. Flag likely recurring characters, one-scene roles, narrative objects, and production risks as estimates for Story Architect review; do not produce visual asset prompts.
8. Produce one Project Brief and adaptation decision table. Give it a stable `artifact_id`, a monotonically increasing `revision`, and `parent_revision: null` for the first revision or the immediate prior revision thereafter. Do not turn the brief or direction cards into a full story, season outline, episode list, or screenplay.

## Output contract

Return a Project Brief conforming to [`project-brief.schema.json`](../../contracts/project-brief.schema.json) and a concise human-readable summary with:

- `project_id`, `artifact_id`, `revision`, `parent_revision`, working title, `source_kind`, `market: US`, output language policy, `delivery_formats`, delivery scope, season/episode scope, and duration policy.
- `intake_route`, `next_action`, and `next_step_for_user`: tell the creator what was recognized, what they will see next, and what decision is needed. This must be understandable without skill names.
- `source_refs[]`: source ID, supplied/observed range, locator, coverage status, and limitations.
- `direction_candidates[]` with stable `direction_id`, audience promise, protagonist strategy, opposing force, cost, and US power logic; `direction_selection` records whether the creator has explicitly selected one. Use an empty candidate array for a pure outline/script diagnosis.
- `adaptation_limits`: elements to preserve, elements open to change, prohibited changes, and any rights/authorization status the user supplied.
- `adaptation_decisions[]`: decision (`retain | cut | merge | redesign | unresolved`), source references whose locators appear in `source_refs[]`, dramatic reason, downstream implications for characters/plot functions, and whether the proposal is user-decided or still an AI proposal.
- `adaptation_mechanism_map[]`: one entry for each relevant `power | relationship | cost | causality` dimension, naming the emotional payoff, source mechanism, proposed US mechanism, US plausibility basis, and decision status. For original ideas with no source, use an empty array.
- `evidence_and_assumptions[]`: statement, provenance (`user | source | market_evidence | ai_proposal`), source/date/locator when applicable; state uncertainty in the statement rather than implying a verified fact.
- `open_decisions[]` and `brief_status: BRIEF_READY | NEEDS_USER_DECISION`.

Do not label an AI proposal or inferred detail as approved canon. `BRIEF_READY` means the next diagnostic or story task has enough traceable input; it does not mean the direction or story passed review. For full-story drafting, require an explicitly creator-selected `direction_selection.selected_direction_id` and a credible US conflict mechanism. A completed outline or script can proceed to diagnosis without a selected direction ID.

## Stop conditions

Stop after delivering the brief and direction preview if needed. Never claim full source coverage when only excerpts were read. Never claim audience or commercial validation based on market research or genre conventions. Never quietly turn a Chinese social dependency into an American one by changing labels.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: ../../contracts/project-brief.schema.json

{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://usvds.invalid/schemas/v10/project-brief.schema.json",
  "title": "USVDS V10 Project Brief",
  "type": "object",
  "required": [
    "schema_version", "project_id", "artifact_id", "revision", "parent_revision",
    "market", "source_kind", "source_refs",
    "intake_route", "next_action", "next_step_for_user", "direction_candidates",
    "direction_selection", "adaptation_limits", "adaptation_decisions",
    "adaptation_mechanism_map", "season_scope", "episode_count",
    "duration_policy", "output_languages", "delivery_scope",
    "evidence_and_assumptions", "open_decisions", "brief_status"
  ],
  "properties": {
    "schema_version": { "const": "usvd-project-brief/v2" },
    "project_id": { "type": "string", "minLength": 1 },
    "artifact_id": { "type": "string", "minLength": 1 },
    "revision": { "type": "integer", "minimum": 1 },
    "parent_revision": { "type": ["integer", "null"], "minimum": 1 },
    "working_title": { "type": ["string", "null"] },
    "market": { "const": "US" },
    "source_kind": { "enum": ["original_idea", "existing_story", "novel", "comic", "film_or_long_form", "screenplay", "market_concept"] },
    "source_refs": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["source_id", "locator", "coverage_status"],
        "properties": {
          "source_id": { "type": "string" },
          "locator": { "type": "string" },
          "coverage_status": { "enum": ["supplied", "observed", "partial", "unavailable"] },
          "notes": { "type": "string" }
        },
        "additionalProperties": false
      }
    },
    "intake_route": {
      "enum": ["simple_idea", "completed_outline_audit", "source_adaptation", "source_selected_direction", "existing_script_reverse_story"]
    },
    "next_action": {
      "enum": ["choose_direction", "audit_outline", "validate_selected_direction", "reverse_story_then_audit", "develop_full_story", "resolve_material_decision"]
    },
    "next_step_for_user": { "type": "string", "minLength": 1 },
    "direction_candidates": {
      "type": "array",
      "maxItems": 3,
      "items": {
        "type": "object",
        "required": ["direction_id", "audience_promise", "protagonist_strategy", "opposing_force", "cost_of_action", "us_power_logic"],
        "properties": {
          "direction_id": { "type": "string", "minLength": 1 },
          "audience_promise": { "type": "string", "minLength": 1 },
          "protagonist_strategy": { "type": "string", "minLength": 1 },
          "opposing_force": { "type": "string", "minLength": 1 },
          "cost_of_action": { "type": "string", "minLength": 1 },
          "us_power_logic": { "type": "string", "minLength": 1 }
        },
        "additionalProperties": false
      }
    },
    "direction_selection": {
      "type": "object",
      "required": ["status", "selected_direction_id", "summary"],
      "properties": {
        "status": { "enum": ["user_selected", "awaiting_user_selection", "not_applicable"] },
        "selected_direction_id": { "type": ["string", "null"] },
        "summary": { "type": ["string", "null"] }
      },
      "additionalProperties": false,
      "allOf": [
        {
          "if": { "properties": { "status": { "const": "user_selected" } } },
          "then": { "properties": { "selected_direction_id": { "type": "string", "minLength": 1 }, "summary": { "type": "string", "minLength": 1 } } },
          "else": { "properties": { "selected_direction_id": { "type": "null" } } }
        }
      ]
    },
    "adaptation_limits": {
      "type": "object",
      "required": ["preserve", "open_to_change", "prohibited_changes", "rights_status"],
      "properties": {
        "preserve": { "type": "array", "items": { "type": "string" } },
        "open_to_change": { "type": "array", "items": { "type": "string" } },
        "prohibited_changes": { "type": "array", "items": { "type": "string" } },
        "rights_status": { "type": "string" }
      },
      "additionalProperties": false
    },
    "adaptation_decisions": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["decision_id", "decision", "source_refs", "dramatic_reason", "downstream_implications", "status"],
        "properties": {
          "decision_id": { "type": "string" },
          "decision": { "enum": ["retain", "cut", "merge", "redesign", "unresolved"] },
          "source_refs": { "type": "array", "items": { "type": "string" } },
          "dramatic_reason": { "type": "string" },
          "downstream_implications": { "type": "string" },
          "status": { "enum": ["user_decided", "ai_proposal", "unresolved"] }
        },
        "additionalProperties": false
      }
    },
    "adaptation_mechanism_map": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["dimension", "emotional_payoff", "source_mechanism", "us_mechanism", "us_plausibility", "status"],
        "properties": {
          "dimension": { "enum": ["power", "relationship", "cost", "causality"] },
          "emotional_payoff": { "type": "string", "minLength": 1 },
          "source_mechanism": { "type": "string", "minLength": 1 },
          "us_mechanism": { "type": "string", "minLength": 1 },
          "us_plausibility": { "type": "string", "minLength": 1 },
          "status": { "enum": ["user_decided", "ai_proposal", "unresolved"] }
        },
        "additionalProperties": false
      }
    },
    "season_scope": { "type": "string", "minLength": 1 },
    "episode_count": { "type": ["integer", "null"], "minimum": 1 },
    "duration_policy": { "type": "string", "minLength": 1 },
    "output_languages": { "type": "array", "description": "Default: zh-CN-development for all creator-facing planning and review documents; en-US-dialogue applies only to later screenplay dialogue. An explicit creator request may override the development language.", "items": { "type": "string" }, "minItems": 1 },
    "delivery_scope": { "enum": ["writing-only", "full-production"] },
    "delivery_formats": { "type": "array", "description": "Creator-facing documents default to both DOCX and standalone HTML. Optional for older Brief revisions; new Briefs should record both unless the creator explicitly requests another format.", "items": { "enum": ["docx", "html"] }, "minItems": 1, "maxItems": 2 },
    "evidence_and_assumptions": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["statement", "provenance"],
        "properties": {
          "statement": { "type": "string" },
          "provenance": { "enum": ["user", "source", "market_evidence", "ai_proposal"] },
          "source_ref": { "type": "string" },
          "date": { "type": "string", "format": "date" }
        },
        "additionalProperties": false
      }
    },
    "open_decisions": { "type": "array", "items": { "type": "string" } },
    "brief_status": { "enum": ["BRIEF_READY", "NEEDS_USER_DECISION"] }
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


### Bundled: ../../references/intake-routing.md

# Intake routing and US adaptation decisions

This guide supports `00-intake-adaptation`. It describes decisions at the entrance to V10; it does not replace independent story or outline review.

## Classify by evidence, then by request

Read the supplied artifact before choosing a route. A filename or the creator's label is not proof that an outline is complete. A complete outline follows the major causal turns through a climax and resolution; a treatment that stops at the premise is still an idea. When a full screenplay is supplied for quality assessment, choose `existing_script_reverse_story` even if the user calls it an outline. When source material and a separate US adaptation direction are both supplied, choose `source_selected_direction`.

For `completed_outline_audit`, identify what the existing outline claims and send it to diagnosis. Do not repair it inside Intake. For `existing_script_reverse_story`, first reconstruct the protagonist's actual decisions, opposition, changed states, climax, and ending from the script, with scene locators. The reconstructed story is the object of diagnosis. A polished English script can still have an unworkable story engine.

For `simple_idea` and `source_adaptation`, give the creator 2–3 short directions when the core engine is unsettled. Each card must answer: what emotional result is promised, what the protagonist actively does, who can stop them, what the action costs, and which US social or institutional arrangement creates that pressure. Directions must differ in choices and consequences, not only in occupation, city, or character names. If no credible alternatives can be formed from the available material, state the missing premise decision in `open_decisions`; do not fabricate details.

For `source_selected_direction`, record the user's exact intended direction, then check its engine. Preserve the user's selection as a decision, but mark any proposed mechanism repair as `ai_proposal`. If the selected direction is implausible, explain the specific break and offer a repair for that direction before suggesting a replacement.

## Preserve payoff; rebuild the mechanism

Summarize the source's emotional promise in ordinary language: what the audience expects to feel when the protagonist wins, loses, or uncovers the truth. Identify how four mechanisms deliver it:

| Dimension | Question for the source | Question for the US version |
| --- | --- | --- |
| Power | Who can impose a meaningful constraint, and through what authority or dependence? | Who has credible leverage in this US setting, and what limits it? |
| Relationship | Why do people stay, trust, betray, protect, or obey? | Which specific bond or dependency would make that behavior plausible? |
| Cost | What can the protagonist lose by acting? | What immediate, visible price makes each strategy a consequential choice? |
| Causality | Why does one action produce the next reversal? | Which US facts, rules, incentives, or character choices make that result follow? |

Fill `adaptation_mechanism_map` for all four dimensions when the source supplies enough evidence. If coverage is partial, record the gap and mark the dimension `unresolved`; do not invent the missing source. `us_plausibility` must explain the character or setting logic, and must label any unverified legal, medical, workplace, or institutional premise. A proposed US equivalent must change how the conflict operates where needed while preserving the chosen emotional payoff.

Reject a proposed adaptation direction at this gate when the only changes are English names, US place labels, literal dialogue translation, or generic American decoration while decisive pressure still relies on the original culture's unstated hierarchy or obligations. Name the failed mechanism and the required redesign. Do not smooth it over with fluent prose.

## Direction record and handoff

Assign stable IDs to direction cards. An explicit creator choice sets `direction_selection.status: user_selected` and uses the matching `direction_id`; a model recommendation leaves `awaiting_user_selection` with a null selected ID. A creator-supplied selected direction can be represented by a single card. For an outline or script diagnostic, `not_applicable` and an empty card list are valid.

Use `next_step_for_user` to say exactly what happens next: choose among cards, receive an outline diagnosis, receive a script-derived story diagnosis, check a selected adaptation direction, or proceed to a complete story. Diagnostic handoffs do not imply approval. Full-story drafting follows an explicit direction choice and a plausible conflict mechanism; unresolved adaptation failures return to direction work.

These entry patterns borrow three distinct methods: source selection and a quick creator-facing skeleton; staged handoff of a chosen brief; and comparison of genuinely different conflict engines. V10 retains its own story architecture and independent review stages.


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
