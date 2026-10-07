---
name: us-vertical-drama-adapter
description: Adapt a source premise, synopsis, or episode material for a US-facing vertical microdrama while preserving its audience promise and chosen world.
title: USVDS V11 · Adapter
hint:
  workflow: usvds-v11
  stage: intake
  baseline: 0c68d9c
---
# US Adaptation

For a new project, route intake and adaptation through `usvd-v10-controller` and `usvd-v10-00-intake-adaptation` first. This legacy skill can assist an already documented project or critique a supplied draft, but its old `APPROVED` label cannot replace V10's independent review or trusted human gate.

## Trigger and scope

Use for adapting a non-US premise, trope package, outline, or episode material for a US-facing vertical microdrama. Preserve the dramatic promise, emotional engine, and serial propulsion; localize the social logic that makes events believable.

## Non-goals

Do not write a season bible, episode beat sheet, screenplay, or storyboard. Do not erase genre height, sanitize conflict, or substitute a generic US setting for a causally coherent one.

## Inputs

Required: source premise or material, target audience/market, and intended format. Optional: comparable titles, guardrails, desired rating, and existing canon.

## Workflow

1. State the source promise in one sentence and identify which parts are inviolable.
2. Run `references/us-cultural-plausibility-checklist.md`; distinguish direct transfer, functional equivalent, and material needing redesign.
3. Rebuild only the social logic that needs adaptation—institutions, status signals, dependency, money/resources, rules, work/role, romance, or consequences—so it is legible and causally plausible to the US-facing audience within the chosen world. Fantasy, historical, and alternate settings stay in-world unless relocation is explicitly approved.
4. Deliver an Adaptation Brief with retained promise, localized premise, character/status changes, risk flags, and questions for Showrunner.

## Hard rules

- Do not translate Chinese short-drama tropes literally. Test every major behavior for US-facing audience plausibility while preserving the story promise and the approved setting; US-facing does not mean contemporary-USA relocation.
- Do not claim a stereotype, imported title, ritual, wealth cue, or legal consequence will land without causal support.
- Mark uncertainty as a validation risk; never invent research as fact.

## Output contract

Return: `Adaptation status`, `Story promise preserved`, `US plausibility decisions`, `Localized premise`, `Character/status translation`, `Risks`, and `Showrunner handoff questions`.

## Handoff contract

Handoff to Showrunner only with a clearly marked Adaptation Brief. It is input, not canon, until the Showrunner incorporates it in the Story Bible.

## Failure and rewrite conditions

Rewrite when the adaptation relies on literal trope transfer, cannot explain a major character choice through the chosen world's social logic, unnecessarily relocates a setting, loses the story promise, or leaves a material plausibility risk unresolved.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: references/us-cultural-plausibility-checklist.md

# US Cultural Plausibility Checklist

## Scope

This checks **US-facing audience plausibility**, not relocation to contemporary USA. A fantasy, Norse, historical, or alternate world may remain fully in that world. Adapt legibility, causal/social logic, dialogue, stakes, status cues, and institutions to what the audience can understand from the chosen world; do not force US police, courts, jobs, or family structures into fantasy when they do not belong.

## Checks

- Can a US-facing viewer understand the status hierarchy, institution, and consequence from the scene's own evidence rather than imported social shorthand?
- Does authority arise from visible dependency, belief, law, wealth, oath, care, work, magic, rank, or social reputation appropriate to this world?
- Is the character's choice voluntary, coerced, or constrained in a legible way, with an observable cost for refusal?
- Are dialogue register, humor, conflict style, and status cues clear without flattening the setting into contemporary American speech or stereotypes?
- Does the story preserve its fantasy/world-specific institutions while making their rules, enforcement, and social logic comprehensible?
- Are class, race, gender, religion, disability, and regional details specific and researched enough for the claim being made? Flag uncertainty; do not invent factual validation.
- For every transplanted trope, name the retained story promise and the functional equivalent rebuilt for this setting.


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
