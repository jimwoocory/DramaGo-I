---
name: us-vertical-drama-episode-architect
description: Turn an approved US vertical-drama Story Bible into an approval-ready episode beat sheet with retention and continuity gates.
title: USVDS V11 · Legacy Episode Architect
hint:
  workflow: usvds-v11
  stage: episode
  baseline: 0c68d9c
---
# Episode Architect

For a new project, use `usvd-v10-controller` and `usvd-v10-02-episode-architect` only after a complete Story Package and digest-bound independent story review. This legacy skill serves projects with documented prior approvals; its old `APPROVED` label cannot replace V10's independent review or trusted human gate.

## Trigger and scope

Use only with an APPROVED Story Bible and current continuity ledger. Create a single episode beat sheet, including Hook, Conflict, Escalation, Reversal, Payoff, and Cliffhanger.

## Non-goals

Do not draft a screenplay, redesign series canon, introduce unapproved reveals, or create storyboards.

## Inputs

Required: APPROVED Story Bible, target episode number and duration, preceding state, and continuity ledger. Optional: current arc objective, production constraints, and Script Doctor notes.

## Workflow

1. Locate the episode on the season/arc ladder and list the required state change.
2. Fill `references/episode-beat-template.md`: Hook, Conflict, Escalation, Reversal, Payoff, Cliffhanger, canon dependencies, and next-episode question. Then fill `references/scene-unit-template.md` as a numbered scene-unit plan; it is the production contract for place/time, characters, objective, visible action, playable emotion/intent with observable evidence, dialogue purpose, entry/exit state, and timing.
3. Check planted/payoff timing, knowledge states, repetition signature, and the next episode's fresh pressure.
4. For EP01, evaluate all six structural beats as substantive gates: 0-5s Hook; early comprehensible Conflict/objective-obstacle; Escalation that materially raises cost, urgency, or narrows options; Reversal that changes power, information, or plan; at least one local Payoff; and a concrete Cliffhanger/next-episode question. These are evaluation anchors, not rigid screenplay formulas.

## Hard rules

- Every beat sheet must explicitly contain Hook, Conflict, Escalation, Reversal, Payoff, and Cliffhanger fields.
- Every screenplay handoff must include a complete numbered scene-unit plan; it is the source for the Screenwriter's `【场景】`, `【人物】`, `【动作】`, `【情绪/内心】`, and `【台词】` labels. Do not approve a beat sheet whose scene units omit an ID, timing, entry/exit state, objective, visible action, observable performance evidence, or dialogue purpose.
- EP01 cannot be APPROVED unless Hook, Conflict, Escalation, Reversal, Payoff, and Cliffhanger are all substantive rather than placeholder fields.
- A beat sheet cannot promise a payoff that conflicts with the secret/reveal ledger or continuity ledger.
- Do not use a cliffhanger as a substitute for a local payoff.

## Output contract

Return `APPROVED Beat Sheet` or `REWRITE Beat Sheet`, with all required fields, rationale for any deliberate restraint, continuity delta, and the Screenwriter handoff packet.

## Handoff contract

Only an APPROVED Beat Sheet may hand off to Screenwriter. EP01 cannot proceed to screenplay or storyboard without APPROVED Story Bible and APPROVED Beat Sheet.

## Failure and rewrite conditions

Rewrite when any required field is absent or cosmetic; the audience cannot identify an early objective/obstacle; EP01 lacks material Escalation or a real Reversal; EP01 lacks a local Payoff or concrete next-episode question; the turn repeats a recent signature; or canon dependencies are unresolved.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: references/episode-beat-template.md

# Episode Beat Template

- Episode / duration / arc location / status: DRAFT | APPROVED | REWRITE
- Incoming canon and knowledge state
- **Hook:** immediate attention event or pressure
- **Conflict:** objective versus obstacle
- **Escalation:** cost, urgency, or narrowing options
- **Reversal:** credible change in power, information, or plan
- **Payoff:** local answer, win, loss, or emotional delivery
- **Cliffhanger:** concrete destabilizing turn
- **Next-episode question:** a specific question created by the cliffhanger
- Plant/payoff dependencies / continuity delta / repetition signature

For EP01 evaluate: 0-5s hook target; early comprehensible objective/obstacle; at least one local payoff; and a concrete next-episode question. These are evaluation anchors, not rigid screenplay formulas.


### Bundled: references/scene-unit-template.md

# Numbered Scene-Unit Template

Use one row or block per scene before screenplay drafting. This is a production handoff, not optional planning prose.

`SCENE-01` — **start–end seconds:**; **beat ID(s):**; **time / precise place:**; **characters present:**; **entry state:**; **immediate dramatic objective:**; **visible, shootable action:**; **playable emotion / concealed intent:**; **observable performance evidence:**; **dialogue purpose:** (emotion, relationship, information, and/or action); **exit state:**; **continuity / asset notes:**.

The next scene must inherit the prior exit state unless the approved beat sheet specifies a transition. Do not leave a cell implicit when it affects performance, prop possession, wardrobe, location, knowledge, injury, or screen direction.


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
