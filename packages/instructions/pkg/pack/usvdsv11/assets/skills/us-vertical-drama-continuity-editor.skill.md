---
name: us-vertical-drama-continuity-editor
description: Maintain a gated canon ledger for US vertical-drama episodes before storyboard handoff.
title: USVDS V11 · Continuity Editor
hint:
  workflow: usvds-v11
  stage: continuity
  baseline: 0c68d9c
---
# US Vertical Drama Continuity Editor

For a new V10 project, do not treat this legacy continuity ledger as proof of story approval or a screenplay unlock. This skill serves a documented legacy project or may audit supplied material without advancing its gate. Route new story work to `usvd-v10-controller`.

## Trigger and scope

Use after Script Doctor PASS and before `us-vertical-drama-storyboard-director`. Reconcile the script with the Story Bible and update the canonical episode-to-episode ledger.

## Non-goals

Do not cure a non-PASS doctor gate, rewrite dramatic beats, create new canon, or create the storyboard/generation prompts.

## Inputs

Required: Script Doctor PASS report, screenplay, APPROVED Beat Sheet, Story Bible, and prior continuity ledger. Optional: production asset ledger and prior storyboard notes.

## Workflow

1. Verify the Script Doctor PASS authorization.
2. Compare stated and implied changes against prior canon.
3. Update `references/continuity-ledger-template.md` for canon, knowledge states, injuries/powers/props/look state where relevant, planted/payoff ledger, relationship state, repetition signatures, and unresolved promises.
4. Issue a storyboard packet containing the screenplay, current approved continuity ledger, Story Bible constraint extract, active reveal windows, power/knowledge state, locked facts, explicit state deltas, and the active character/costume/scene/prop asset states.

## Hard rules

- Never convert an unapproved possibility into canon.
- Track who knows what, when they learned it, and what remains concealed.
- Track injuries, powers, props, and look state only where relevant, with carry-forward defaults explicit.
- Flag repetition signatures even when continuity is technically intact.
- If doctor status is not PASS, block storyboard handoff.

## Output contract

Return `Continuity status: CLEAR|BLOCKED`, the full updated continuity ledger, Story Bible constraint extract, active reveal windows, current power/knowledge state, canon delta, knowledge-state delta, planted/payoff delta, unresolved promises, repetition warning, and `Storyboard authorization` only when CLEAR.

## Handoff contract

Only CLEAR with Script Doctor PASS may hand off to `us-vertical-drama-storyboard-director`. The director receives the screenplay plus the full current approved ledger, Story Bible/reveal constraints, power/knowledge baseline, episode delta, and active asset states.

## Failure and rewrite conditions

Block and return upstream when a Script Doctor PASS is absent, a fact conflicts with the Bible/ledger, a material state change is untracked, a promised payoff disappears, or the script repeats a flagged signature without an approved escalation change.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: references/continuity-ledger-template.md

# Continuity Ledger Template

Use separate records so a true world fact is not confused with who knows it, what the audience has seen, or a temporary production state. Every row requires a **Source** (episode, scene, approved bible, or handoff) and **Certainty** (`confirmed | provisional | unknown`).

## Canon facts

| Fact / rule | Current state | Source | Certainty | Change trigger |
|---|---|---|---|---|
| | | | | |

## Character knowledge

| Character | Knows / believes / suspects | Evidence and source | Certainty | Must not know |
|---|---|---|---|---|
| | | | | |

## Audience knowledge

| Audience knows | Reveal episode / source | Certainty | Dramatic consequence |
|---|---|---|---|
| | | | |

## Physical/asset state

| Character / asset | Injury, power, prop, money, costume/look, location, access | Source | Certainty | Carry-forward requirement |
|---|---|---|---|---|
| | | | | |

## Relationships

| Pair / group | Current bond and leverage | Last change | Source | Certainty |
|---|---|---|---|---|
| | | | | |

## Plants/payoffs

| Plant / question | Planted source | Promised payoff window | Payoff evidence | Certainty / owner |
|---|---|---|---|---|
| | | | | |

## Promises

| Audience promise | Opened source | Required delivery / deadline | Current risk | Certainty |
|---|---|---|---|---|
| | | | | |

## Repetition signatures

| Episode | Mechanism | Power / information / relationship / cost | Location / visual transaction | Cooldown status | Source |
|---|---|---|---|---|---|
| | | | | |


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
