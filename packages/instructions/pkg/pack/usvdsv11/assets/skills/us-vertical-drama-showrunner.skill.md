---
name: us-vertical-drama-showrunner
description: Build the canon and serial engine for a 40-100 episode US-facing vertical microdrama from an adaptation brief or original premise.
title: USVDS V11 · Showrunner
hint:
  workflow: usvds-v11
  stage: story
  baseline: 0c68d9c
---
# US Vertical Drama Showrunner

For a new project, route story creation through `usvd-v10-controller` and `usvd-v10-01-story-architect` first. This legacy skill can assist an already documented project or critique a supplied draft, but its old `APPROVED` label cannot replace V10's independent review or trusted human gate.

## Trigger and scope

Use after a US Adaptation Brief or for an original US-facing premise before episode architecture. Create the governing canon and long-form escalation system for a 40-100 episode series.

## Non-goals

Do not draft dialogue, write a screenplay, turn a single episode into shots, or paper over unknown decisions with assumed canon.

## Inputs

Required: approved Adaptation Brief or original premise, intended episode count/running time, and target audience. Optional: characters, world rules, rating, business constraints, and prior canon.

## Workflow

1. Define the audience promise, series question, tone boundary, and final emotional destination.
2. Build the Story Bible using `references/story-bible-template.md`.
3. Build the season/arc ladder using `references/season-arc-template.md` across 40-100 episodes.
4. Specify character engines, secret/reveal ledger, escalation ladder, and anti-repetition controls using `references/anti-repetition-guide.md`.
5. Assign plants, payoff windows, and reveal ownership; identify unresolved decisions.

## Hard rules

- A Story Bible must include character engines, not just biographies.
- The season/arc ladder must escalate stakes, cost, and information; new antagonism alone is not escalation.
- Secret/reveal timing must protect both audience comprehension and future episode turns.
- Use anti-repetition controls: do not recycle the same confrontation, rescue, humiliation, betrayal, or cliffhanger signature without a meaningful state change.

## Output contract

Return an `APPROVED Story Bible` only when complete, plus `Season/arc ladder`, `Character engines`, `Secret/reveal ledger`, `Escalation ladder`, `Anti-repetition controls`, and `Episode Architect brief`.

## Handoff contract

Only an APPROVED Story Bible may hand off to Episode Architect. The architect may propose gaps but may not override canon; changes return to Showrunner for versioned approval.

## Failure and rewrite conditions

Rewrite when the series lacks a sustainable 40-100 episode engine, its reveals have no ledger, escalation merely repeats a scene type, a character has no active engine, or the ending promise contradicts the opening promise.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: references/anti-repetition-guide.md

# Anti-Repetition Guide

## Rolling signature ledger

Maintain a rolling recent-episode window sized in the Story Bible (normally the current major arc, with at least the previous three episodes visible). For each episode record its dominant dramatic transaction and novelty dimensions.

| Episode | Signature / mechanism | Power change | Information change | Relationship change | Cost | Location / visual transaction | Result / escalation |
|---|---|---|---|---|---|---|---|
| | | | | | | | |

**Mechanism** means how the turn happens (trial, trap, bargain, pursuit, confession, ritual, public choice), not merely its label. Track **Power**, **Information**, **Relationship**, **Cost**, and, where visually relevant, **Location / visual transaction** separately.

## Cooldown rule

Do not repeat the same dominant signature inside the configured window unless at least two novelty dimensions change and one is an irreversible cost, knowledge state, or power transfer. After an intentional repeat, put that signature on cooldown until the ledger shows a materially different transaction. Do not spend a cooldown exception on a cosmetic costume, louder dialogue, or a new antagonist performing the same move.

## Escalation test

Before approving an episode, answer: compared with the nearest matching signature, does this episode narrow choices, raise the price, change who holds leverage, expose consequential information, or permanently alter a relationship? If the answer is no, redesign the mechanism or move the episode's function.

## Motif versus lazy repetition

An **intentional motif recurrence** echoes a recognizable image, line, object, ritual, or dilemma while meaning changes because audience knowledge, power, cost, or relationship has changed. **Lazy repetition** recreates a prior suspense transaction and expects the same surprise with no meaningful state change. Mark motif recurrences in the ledger with their altered meaning; reject lazy repeats.


### Bundled: references/season-arc-template.md

# Season/Arc Template

## Configuration

| Parameter | Set for this series | Decision rule |
|---|---|---|
| Episode Count | `[40–100 or other approved count]` | Set before outlining; do not infer an arbitrary block plan. |
| Target Duration | `[for example: ~90 seconds]` | Determines beat density and production scope. |
| Major-Arc Cadence | `[for example: ~12 episodes]` | Choose a cadence that serves the promise, not a fixed universal range. |
| Final-block Allowance | `[shorter / equal / longer]` | State why the ending needs a different length, if it does. |
| Season Count | `[one or more]` | Define handoff between seasons and the promise each season resolves. |

## Arc ladder

Create rows by applying the approved cadence to the configured episode count. Number the actual episode spans only after the parameters above are approved.

| Arc / season | Actual episode span | Promise under pressure | Opposition and power change | Escalation / irreversible cost | Reveal window and payoff | Exit question | Repetition guard |
|---|---|---|---|---|---|---|---|
| `[A]` | `[derived]` | | | | | | |
| `[B]` | `[derived]` | | | | | | |

For every row, name the permanent state change. A louder argument, a larger attacker, or a reset rescue does not count as escalation unless power, knowledge, relationship, cost, or available choices changes.

## 80 × ~90s Golden configuration

For the Golden Norse process fixture, set Episode Count to **80**, Target Duration to **~90 seconds**, and Major-Arc Cadence to **~12 episodes**. A valid derived ladder can use six approximately 12-episode major arcs plus a shorter final resolution block. This is an example configuration, not a required universal partition.


### Bundled: references/story-bible-template.md

# Story Bible Template

## Approval

- Version, owner, date, status: `DRAFT | APPROVED | SUPERSEDED`
- Series question, audience promise, tone/rating boundary, and final emotional destination
- Canon certainty: `confirmed | provisional | unknown`; record an owner and resolution deadline for every unknown.

## Series architecture

| Parameter | Approved value | Rationale |
|---|---|---|
| Episode count | | |
| Target duration | | |
| Season / major-arc Cadence | | |
| Release rhythm and recap tolerance | | |
| Ending destination | | |
| Production constraints | cast, locations, VFX/action limits, budget-sensitive assets, and schedule constraints |

## Promise hierarchy

State the promise at three levels: **series promise** (why return), **season/arc promise** (what this run resolves or complicates), and **episode promise** (today's dramatic transaction). A lower-level twist may not violate a higher-level promise without explicit Showrunner approval.

## Canon

- World rules, institutions, status cues, and consequences appropriate to the chosen world.
- **Reveal windows:** secret, holder, intended episode/arc window, trigger, proof, audience state, consequence, and fallback if delayed.
- Cost and limit rules for money, status, injury, magic/technology, law, access, and time.

## Power map

Who can compel whom, through what leverage, cost, limit, and countermeasure.

## Knowledge map

What each character knows, believes, suspects, and has misread; separately record audience knowledge.

## Character engines

For each continuing character record want, wound, tactic, leverage, blind spot, pressure point, relationship dependency, change pressure, and what they will do when denied. The engine must produce choices, not biography.

## Serial engine

- Season/arc ladder reference and escalation ladder: pressure, cost, irreversible change, and new question.
- Plants/payoffs and unresolved promises with source, certainty, deadline, and payoff owner.
- Rolling anti-repetition window, signature cooldowns, and prohibited lazy repeats.
- Ending destination: final power map, relationship truth, promise payoff, and what remains deliberately open.


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
