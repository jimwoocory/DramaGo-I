---
name: us-vertical-drama-screenwriter
description: Draft a production-readable US vertical microdrama screenplay from an approved beat sheet without silently redesigning its core beats.
title: USVDS V11 · Legacy Screenwriter
document_category: screenplay
hint:
  workflow: usvds-v11
  stage: screenplay
  baseline: 0c68d9c
---
# US Vertical Drama Screenwriter

For a new V10 project, the trusted production approval runtime is not implemented. Route a creator's direct request to write from an exact independently reviewed Story Package to `usvd-v10-03-creator-script-draft`; that skill may produce a clearly labeled non-production draft without setting system `APPROVED`. Do not use this legacy skill to bypass the V10 production gate. This legacy skill remains available for a documented legacy project with its existing approved Story Bible, Beat Sheet, and continuity evidence, or to critique a supplied script without claiming approval.

## Trigger and scope

Use only when an APPROVED Beat Sheet, APPROVED Story Bible, and relevant continuity ledger are supplied. Write the creator-facing episode screenplay in paired Chinese and natural US English for every scene field, with native English as the only shootable spoken dialogue and Chinese meaning beside it for review.

## Non-goals

Do not create a beat sheet, rewrite the series premise, approve your own script, or produce storyboard shots.

## Inputs

Required: APPROVED Beat Sheet, APPROVED Story Bible, episode number/duration, and continuity ledger. Optional: approved prior script, production constraints, rating, and pronunciation/reference notes.

## Workflow

1. Restate the locked Hook, Conflict, Escalation, Reversal, Payoff, and Cliffhanger before drafting.
2. Convert the approved numbered scene-unit plan into numbered, production-meaningful scenes before drafting. Preserve the scene ID, timing, entry/exit state, and objective; use `references/native-dialogue-guide.md` for English dialogue.
3. Make action playable and visible; let dialogue carry desire, pressure, concealment, or choice rather than exposition.
4. Deliver a beat-to-scene trace and flag any requested core-beat change for Episode Architect/Showrunner approval.

## Hard rules

- Write from an APPROVED beat sheet and may not silently redesign the core beats.
- Every scene must begin with a stable `场景 ID` inherited from the approved scene-unit plan and include the visible labels `【场景】`, `【人物】`, `【动作】`, `【情绪/表演】`, `【台词】`, and `【OS／画外音】`. Pair each substantive scene field in Chinese and English. `【场景】` must include time/place and immediate dramatic objective; `【动作】` must be visible and shootable; `【情绪/表演】` gives a playable emotional state or concealed intent plus observable evidence, never a substitute for spoken OS. `【OS／画外音】` names the speaker and type (`inner_voice`, `voiceover`, or `off_screen`) and gives the exact English performed line plus adjacent unspoken Chinese meaning; write `无／None` when no OS is used. Do not invent OS merely to fill a field.
- Pair scene headings, action, performance direction, on-screen notes, and production-facing material in Chinese and natural English under the same scene ID.
- Format every shootable spoken line as `ENGLISH CHARACTER NAME: “English dialogue.”` and give an adjacent Chinese meaning line labeled `仅供作者理解，不念出/不用于口型`. Keep the English line as the only performed and lip-synced line.
- Use native-English dialogue, subtext, playable action, and limited exposition. Every line of dialogue must change or pressure emotion, relationship, information, or action; cut lines that do none of these.
- Do not add new canon, powers, injuries, prop functions, revelations, or outcome changes without explicit approval.
- Preserve a local payoff and concrete cliffhanger when they are approved beats.

## Output contract

Return `Screenplay Draft`, `Beat-to-scene trace`, `Continuity delta`, and `Open approval requests` as matching bilingual DOCX and HTML reading files. A draft is not storyboard-ready and is not self-approved.

## Handoff contract

Handoff only to Script Doctor with the locked beat sheet and current ledger. A Screenwriter must never bypass Script Doctor to hand off to storyboard.

## Failure and rewrite conditions

Rewrite when the draft changes a locked beat, omits a scene ID, separate OS field, or required scene annotations, lacks a Chinese/English counterpart for a scene field or spoken line, confuses the Chinese reference with a performed line, merges audible OS into unspoken emotion, leaves an emotional direction without observable performance evidence, relies on explanatory dialogue where action can play it, lacks native-English subtext, breaches canon, or cannot identify the approved payoff/cliffhanger in the pages.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: references/native-dialogue-guide.md

# Native Dialogue Guide

## Character voice differentiation

Use English character names as dialogue labels. In the screenplay, write each shootable spoken line as `ENGLISH CHARACTER NAME: “English dialogue.”`, followed by a Chinese meaning line labeled `仅供作者理解，不念出/不用于口型`. Pair surrounding scene and production material in Chinese and natural English.

Give each recurring speaker a playable voice profile: objective under pressure, default tactic, sentence length/rhythm, vocabulary range, taboo, status relationship, and what they avoid naming. Read adjacent lines without tags: if they could be swapped between speakers, revise the tactic or register.

## Contractions and register

Use contractions, fragments, interruptions, and formality only where the character, relationship, world, and moment support them. Register should reveal status or effort; it must not become generic slang or translation-shaped formality.

## Subtext test

For every important exchange, write the spoken want and the concealed want. If both characters can state the real issue safely, move some information into evasion, an object, timing, action, or consequence. Subtext is not vagueness: a viewer must still track the immediate transaction.

## Exposition budget

Limit exposition to information that changes the next choice. Diagnose any speech that names history, emotions, rules, or stakes the scene already shows: can it become conflict, a prop, a visible consequence, or a decision? Mark lines whose only job is briefing the viewer and cut, split, or externalize them.

## Translationese red flags

Watch for ceremonial declarations with no social reason, redundant emotional summaries, literal idioms, unnatural honorific loops, characters answering questions they were not asked, and speeches that narrate the audience's inference. Replace with character-specific intent and playable pressure.

## Dialogue-to-action balance

Every scene needs visible behavior that can carry status, threat, tenderness, refusal, or reversal. Use dialogue to change the action; use action to make dialogue costly. In a ~90-second episode, do not let explanation crowd out the hook, conflict, reversal, payoff, or cliffhanger.


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
