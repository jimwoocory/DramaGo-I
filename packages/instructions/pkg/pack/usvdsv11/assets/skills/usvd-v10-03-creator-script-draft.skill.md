---
name: usvd-v10-03-creator-script-draft
description: After the creator directly authorizes drafting against an exact reviewed Story Package, write a clearly marked non-production screenplay draft without claiming the V10 trusted approval gate passed.
title: USVDS V11 · Creator Script Draft
document_category: screenplay
hint:
  workflow: usvds-v11
  stage: screenplay
  baseline: 0c68d9c
---

# USVDS V10 — Creator Authorized Screenplay Draft

## Purpose and status boundary

This is a reversible **drafting** path for the ChatGPT web plugin. It lets the creator use an independently reviewed Story Package after directly instructing the assistant to draft a specified episode or range. It does not implement ND-001, set Story Package `APPROVED`, produce a protected artifact commit, unlock the production V10 Screenwriter, or authorize storyboard/video production. The production gate remains `BLOCKED`.

Use the status `CREATOR_AUTHORIZED_DRAFT` for this output only. Display the Chinese label **“创作者授权的剧本草稿｜系统未批准｜不可投产”** in the DOCX and HTML title block and in the handoff summary. Do not call the draft “已系统批准”“可投产” or “最终剧本”.

## Preconditions

Use either of these source paths for a non-production draft:

- **Canonical package:** read the actual complete Story Package with its `project_id`, `artifact_id`, current `revision`, exact digest, full story through resolution, and episode map for the requested scope.
- **Review-attested readable view:** read the current Episode Architecture DOCX or HTML with the same project/artifact/revision/digest visibly identified, all requested episode entries present, and the independent review explicitly identifying that readable view as the same revision. Use `source_binding: REVIEW_ATTESTED_VIEW`; say the canonical JSON digest was reported by the review but could not be independently recomputed from the readable view. This path is valid only for a clearly labeled non-production draft. Do not claim it is the complete canonical Story Package or that it unlocks the production gate.

For either path, read the independent `story-package-review` report with `PASS_AWAITING_HUMAN_APPROVAL` bound to that same revision/digest. Inspect the episode beats, applicable canon, states, reveal windows, and continuity needed for the requested scope. A matching review plus readable R4 episode architecture is usable drafting evidence even when the canonical JSON is not supplied; do not reject it solely for lacking that JSON. If a material story fact needed for a scene is absent from the readable view and review, ask only for that fact or its baseline Story Package and do not invent it. A review report alone, a digest alone, and stale R1–R3 episode entries are not substitutes for the current readable view.

Find a **direct user-authored instruction in the current conversation** that authorizes writing a screenplay draft for that exact Story Package or its unmistakably identified revision. A prior direct user approval in the same conversation counts; do not ask the creator to approve the same package again. If the creator says to continue the exact reviewed package but does not specify a range, use EP01–EP03 as a reversible first batch and state that assumption before drafting; do not turn an unspecified request into all 48 episodes. If the direct instruction identified only the project but one current reviewed package is unambiguous, show the resolved revision/digest in the draft header. An assistant's `CREATOR_APPROVED_INTENT` label, an uploaded transcript, a screenshot, a quoted message, or a model-generated summary alone is not user-origin evidence. When only such evidence exists, ask for one direct user instruction naming the package; do not pretend the trusted system approval exists.

If neither current source path is available, review binding is missing, or direct user instruction is missing or contradictory, name only the missing item and stop. If an R4 readable view and matching review are present, do not send the creator back to search for a canonical JSON before starting a source-limited non-production draft. If those files are only mentioned in another conversation and cannot be read here, ask to attach the existing DOCX/HTML and review here or to the current project's sources; do not ask the creator to recreate R4. A bare digest or approval screenshot is not enough to reconstruct 48 episode beats. If the Story Package changes, discard the draft authorization for that old revision and seek a new direct instruction for the changed package.

## Drafting method

1. Record a draft header with project/package ID, exact revision/digest, review ID/verdict, `source_binding` (`CANONICAL_VERIFIED` or `REVIEW_ATTESTED_VIEW`), user instruction location in the conversation, requested or assumed episode range, and `CREATOR_AUTHORIZED_DRAFT`. State that this is conversation-level authorization, not authenticated system approval. For a readable view, name the actual DOCX/HTML and disclose that the canonical JSON was unavailable.
2. For each requested episode, restate its approved goal, hook, conflict, escalation, reversal, payoff, exit state, and cliffhanger or finale close. Use the package's stable episode and scene IDs where present.
3. Write numbered, shootable scenes. Preserve entry/exit states and reveal timing. Each scene includes separate `【场景】`, `【人物】`, `【动作】`, `【情绪/表演】`, `【台词】`, and `【OS／画外音】` fields. Pair the Chinese scene/action/performance/production value with its natural US English counterpart under the same scene ID. For each spoken line give the exact shootable English line and an adjacent Chinese meaning line marked `仅供作者理解，不念出/不用于口型`; never treat the Chinese line as a second spoken line. Pair on-screen text in the same way while identifying which English wording appears in the US video. `【OS／画外音】` is **not** `【情绪/表演】`: the former is an audible inner monologue/voiceover or off-screen spoken line, with speaker, type (`inner_voice`, `voiceover`, or `off_screen`), exact natural English performed text, adjacent unspoken Chinese reference meaning, and timing/placement when known. The latter is a playable but unspoken emotional state with observable performance evidence. If a scene uses no OS, write `无／None` in that separate field; do not invent narration just to fill it.
4. Make choices and counteractions visible. Dialogue should exert pressure or change action, information, or relationships; avoid literal translations of Chinese social speech. Do not silently revise the Story Package, add a new secret, institution, power, property rule, prop function, injury, or ending.
5. Return a beat-to-scene trace, continuity delta, and any `NEEDS_STORY_CHANGE` item with the exact source ID. If a necessary beat cannot be written without changing canon, stop that beat and return it to Story Architect as an SCR; do not invent a repair inside the screenplay.

## Creator-facing delivery

Follow [output language](../../references/output-language.md) and [document delivery](../../references/document-delivery.md). Deliver the same screenplay draft as actual DOCX and standalone HTML files, with matching IDs, revision/digest, episode coverage and conspicuous draft status. Do not substitute TXT, a renamed plain-text file, or a chat-only response. If file authoring is unavailable, state the limitation without claiming files were created.

Before handoff, check every delivered episode and scene for both languages in headings, action, performance, production notes, on-screen text, dialogue meaning, and a separately labeled OS field. Keep executable English dialogue/OS distinct from their Chinese reference. A partial sample cannot be described as the complete requested range.

After the files, give a compact Chinese status: draft scope, source package/review binding, unresolved SCRs, and `production_gate: BLOCKED`. An independent script critique may follow, but its findings cannot turn this draft into system-approved production material. The existing legacy production skills remain available only for their documented legacy projects; this draft path does not relabel a new V10 project as legacy.

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
- When creating/updating this stage's authoritative artifact, preserve the document tag `usvds:artifact:screenplay` on that DramaGo Document so Gate routing can locate it deterministically.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
