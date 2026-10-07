---
name: us-vertical-drama-script-doctor
description: Independently score a US vertical-drama screenplay against its approved beat sheet and issue a binding PASS, REWRITE, or REJECT gate.
title: USVDS V11 · Script Doctor
hint:
  workflow: usvds-v11
  stage: review
  baseline: 0c68d9c
---
# Script Doctor

For a new project, use `usvd-v10-controller` for story diagnosis and review before production. This legacy skill may review a supplied script or a documented legacy project, but a `PASS` here cannot approve the V10 Story Package or bypass its trusted human gate.

## Trigger and scope

Use after Screenwriter submits a draft with its APPROVED Beat Sheet, Story Bible, and continuity ledger. Assess independently; do not become the writer of record.

## Non-goals

Do not silently rewrite pages, replace the beat sheet, waive mandatory failures, or produce storyboards.

## Inputs

Required: Screenplay Draft, APPROVED Beat Sheet, APPROVED Story Bible, and current continuity ledger. Optional: prior doctor notes and production constraints.

## Workflow

1. Confirm the handoff packet is complete and trace every core beat into the draft.
2. Score with `references/script-doctor-rubric.md` for 100 points.
3. Check mandatory-fail conditions before applying numeric score, then complete a scene-ID and dialogue-function audit. Use `references/native-dialogue-guide.md` to assess the target locale, period, class/profession register, character voice, and translationese risk.
4. Issue a precise PASS, REWRITE, or REJECT report with actionable notes by beat/page/scene.

## Hard rules

- This role is independent from writing. Diagnose; do not covertly author a replacement.
- Default gate: >=85 PASS; 75-84 REWRITE; <75 REJECT.
- Mandatory-fail conditions override numeric score: missing required handoff artifact; unapproved core-beat redesign; EP01 missing or materially weak Hook, Conflict, Escalation, Reversal, Payoff, or Cliffhanger; material canon/continuity break; no comprehensible early objective/obstacle in EP01; absent local payoff; or absent concrete next-episode question.
- Check the bilingual script standard: each creator-facing scene/action/performance/production field has paired Chinese and natural English; each spoken line uses `ENGLISH CHARACTER NAME: “English dialogue.”` with adjacent Chinese reference meaning marked unspoken and excluded from lip-sync. Require a separate `【OS／画外音】` field per scene; check English performed OS plus Chinese reference when present, or `无／None` when absent. Never count unspoken `【情绪/表演】` as OS. Flag missing or meaning-divergent pairs.
- Mandatory check for screenplay annotations: every scene must retain a scene ID and label `【场景】`, `【人物】`, `【动作】`, `【情绪/内心】`, and `【台词】`; reject unlabelled prose as storyboard-ready. Reject an emotion/intent direction that has no observable performance evidence. Complete a dialogue-function audit: check that each spoken line changes or pressures emotion, relationship, information, or action.
- For US-facing dialogue, record the target locale/world, period, social register, and character voice evidence. Flag translationese, false-local idiom, or an unsupported regional/class claim; natural English alone is not sufficient evidence of localization.
- A non-PASS script must not hand off to storyboard.

## Output contract

Return `Script Doctor Gate: PASS|REWRITE|REJECT`, 100-point category scorecard, mandatory-fail result, evidence, required rewrite targets, and a storyboard authorization only for PASS.

## Handoff contract

PASS goes to Continuity Editor, then `us-vertical-drama-storyboard-director`. REWRITE goes to Screenwriter with the same locked beat sheet unless upstream approval changes it. REJECT goes to Showrunner/Episode Architect for structural repair. A non-PASS script must not hand off to storyboard.

## Failure and rewrite conditions

REWRITE when score is 75-84 without a mandatory fail. REJECT when score is <75 or a mandatory fail exposes a structural/canon failure. EP01 retention-structure failures return REWRITE when the approved series architecture remains valid, and REJECT when the defect requires changing the Story Bible or approved episode architecture. Do not issue PASS until all mandatory-fail conditions are clear.

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


### Bundled: references/script-doctor-rubric.md

# Script Doctor 100-Point Rubric

| Category | Points | Required evidence |
|---|---:|---|
| Beat-sheet fidelity and causal clarity | 20 | Beat IDs or explicit handoff references; each turn has a causal trigger and consequence. |
| Hook, conflict, payoff, cliffhanger | 20 | Timestamped/located retention anchors, local answer, and concrete next-episode question. |
| Character objective, pressure, reversal | 15 | Each lead choice exposes an objective, pressure, and changed tactic or status. |
| Dialogue and subtext | 15 | Voice-specific samples, playable subtext, and exposition diagnosis. |
| Escalation and anti-repetition | 10 | Signature comparison against the rolling window plus a material state change. |
| Canon and continuity | 10 | Ledger evidence for knowledge, assets, injuries/powers, relationships, and plants/payoffs. |
| Production readability and runtime discipline | 10 | Visible action, practical scope, and beat timing appropriate to target duration. |

## Scoring anchors

Apply these anchors to every category; scale the assigned points proportionally rather than awarding points for field presence.

| Anchor | Standard |
|---|---|
| Excellent | All required evidence is specific, playable, and causally linked; no material contradiction or retention weakness remains. |
| Partial | Evidence exists but one or more beats are generic, weakly motivated, underplayed, or need a targeted rewrite. |
| Fail | Required evidence is absent, contradicted, fake, or cannot be executed without breaking the approved handoff. |

The report must list: score by category, evidence cited, defect, rewrite instruction, and whether the defect affects a mandatory gate.

## Decision bands

- **>=85 PASS:** only if all mandatory gates and retention minimums below are met.
- **75–84 REWRITE:** targeted revision is required; it cannot hand off to storyboard.
- **<75 REJECT:** a new or substantially rebuilt draft is required; it cannot hand off to storyboard.

## Mandatory-fail interaction

Mandatory fail overrides the total score. Record **REJECT** or **REWRITE** (never PASS) for: incomplete handoff; unapproved core-beat redesign; material canon/continuity break; an EP01 missing or materially weak Hook, Conflict, Escalation, Reversal, Payoff, or Cliffhanger; no early comprehensible objective/obstacle; no local payoff; or no concrete next-episode question. A mandatory-fail script cannot receive PASS even when its numeric total is 85 or higher.

## EP01 structural minimums`r`n`r`nFor EP01, each of the six core beats has an independent minimum. Hook must create a legible destabilizing question; Conflict must expose objective versus obstacle; Escalation must materially increase cost/urgency or narrow options; Reversal must change power, information, or plan; Payoff must answer or transform a planted question; Cliffhanger must change the next available choice. A high aggregate score cannot compensate for failure of any one minimum.`r`n`r`n## Retention-quality gate

Hook, payoff, or cliffhanger cannot receive PASS solely because the category total is high. Each must meet its own minimum: the hook presents a legible destabilizing question early; the payoff answers or transforms a question at a cost; the cliffhanger changes the next choice with concrete information, danger, deadline, or irreversible action. Fake withholding, delayed explanation with no changed state, and an unanswered setup mislabeled as payoff fail this minimum.


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
