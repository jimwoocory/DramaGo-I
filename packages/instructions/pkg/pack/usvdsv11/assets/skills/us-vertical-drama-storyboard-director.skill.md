---
name: us-vertical-drama-storyboard-director
description: Turn a continuity-cleared US vertical-drama screenplay into an asset-locked shot list and video-generation prompt package.
title: USVDS V11 · Storyboard Director
document_category: storyboard
hint:
  workflow: usvds-v11
  stage: storyboard
  baseline: 0c68d9c
---

# US Vertical Drama Storyboard Director

For a new V10 project, do not create production storyboards while its human approval and screenplay gates remain blocked. This skill serves a documented legacy project with existing reviewed script and continuity evidence, or may critique supplied material without advancing approval. Route new story work to `usvd-v10-controller`.

## Trigger and scope

Use only after `SCRIPT DOCTOR PASS` and `CONTINUITY CLEAR`. Convert an annotated screenplay into an asset-creation package, production-ready shot package, and structured Seedance/MediaGo handoff for the selected video model. This role owns visual asset continuity: character identity, costume state, scene state, and hero props. It uses the director rules in `references/director-execution-contract.md` as a production layer; those rules never override approved story canon or the chosen project visual specification.

## Non-goals

Do not rewrite approved dramatic beats, invent canon, change dialogue outcomes, or generate final media. Escalate any required story change to the owning upstream skill.

## Inputs

Required: continuity-cleared screenplay, approved Beat Sheet and Story Bible, current continuity ledger, target video model, aspect ratio, and maximum shot duration. Optional: approved reference images, existing asset library, model-specific prompt guidance, and production constraints.

## Workflow

1. Extract an asset ledger before making shots. Assign stable IDs to each recurring character, each costume/look, each scene/location, and each hero prop. For every asset that needs to be generated or refreshed, create an `Asset Creation Pack` using `references/asset-ledger-template.md`: approved asset ID/state, image prompt, negative prompt, reference image ID/path if supplied, aspect ratio, model/seed when supplied, and approval status. Do not claim an asset is locked until its creation pack is approved or a supplied reference is approved.
2. Set one project specification before drafting: target model and endpoint, aspect ratio, visual medium, locale, maximum per-video duration, audio capability, and export target. Resolve conflicts in supplied templates here. Do not mix an unapproved 16:9/animation default with a 9:16/live-action project.
3. First group source coverage into independently generatable `VIDEO-*` packages of no more than 15 seconds in V8-compatible mode. Every video package must receive its own `VIDEO-PROMPT-*` master prompt and video-level negative prompt with matched Chinese and natural English reading versions. It describes the complete time sequence, all intended cuts, active assets, camera progression, lighting, sound/voice plan, continuity in/out, and explicit exclusions. Name the selected model-submission language; its VIDEO prompt is the direct submission for normal generation and is never replaced by a list of child shots or a concatenation of both languages.
4. Then split each video package into contiguous, independently adjustable `SHOT-*` micro-shots. A micro-shot normally runs 0.5–3 seconds and carries one visible action or reaction; it may exceed 3 seconds only for uninterrupted dialogue or action, with `【例外原因】`. Each shot has paired Chinese/English local layered prompts for replacement or fine adjustment: `Asset Lock Prompt` (approved immutable anchors), `Shot Delta Prompt` (only this shot's change), and `Negative/Do-not-change`. The local prompt must agree with its parent VIDEO prompt; do not use it as a substitute for the VIDEO master prompt.
5. Apply the relevant sections of `references/director-execution-contract.md`: visual-action clarity, temporal/physical continuity, per-video source coverage, dialogue/voice execution, and the V8-style shot grammar. Do not treat optional camera-style suggestions as mandatory aesthetics.
6. Fill `references/seedance-mediago-output-schema.md` for every shot. It contains the model, aspect ratio, duration, asset reference IDs, dialogue/lip-sync timing, camera, start/end frame references when used, seed when used, and import row. Seedance and MediaGo fields must remain empty-and-marked-optional when the selected model does not support them; never invent model support.
7. Provide a `Prompt Placement Map` that links every prompt to `Scene ID → VIDEO ID → SHOT ID → package in/out → prompt ID → asset IDs → export row`. Prompts never appear as an unlinked appendix. Include `continuity in/out`, screen direction, character entry/exit, and prop handoff when relevant.
8. Emit `production-workbench.json` according to `references/production-workbench-manifest.md`. It must trace each approved asset, Chinese display label, `VIDEO-*`, each micro-shot `SHOT-*`, `PROMPT-*`, generation state, and production task. This is the DSH/MediaGo status source, not an optional appendix.
9. Verify `sum(SHOT duration)` against each scene range and the episode runtime within the user's stated tolerance. If spoken dialogue or OS is present, check it can be performed within the shot duration and supply `dialogue_audio` / lip-sync or voiceover timing as appropriate. Keep `inner_voice`, `voiceover`, and `off_screen` cues distinct from unspoken emotion/performance notes; do not lip-sync voiceover merely because it has words. Check costume, prop possession/damage, wounds, time/weather, scene geography, and transition state between adjacent shots. Flag any conflict rather than silently correcting the screenplay.

## Hard rules

- Do not describe a character generically after the first asset definition. Reuse its `CHAR-*` ID and active `LOOK-*` ID in every applicable shot.
- Treat wardrobe as continuity-critical: define garment, palette/material, footwear, accessories, condition, and episode/scene state. A change requires a visible script-supported transition or approved asset delta.
- Treat hero props as continuity-critical: define `PROP-*` ID, material/form, scale, condition, owner/holder, location, narrative function, and state changes. Reference each applicable prop in the shot prompt.
- Treat locations as continuity-critical: define `SET-*` ID with geography, key set dressing, time, weather, light, and recurring visual anchors. Preserve screen direction and object placement when the script requires it.
- Every shot must visibly label `【镜头编号】`, `【时长】`, `【关联场景】`, `【画面/构图】`, `【角色与服装】`, `【场景】`, `【道具】`, `【动作与情绪】`, `【镜头运动】`, `【连续性进/出】`, `【视频生成提示词】`, and `【负面约束】`.
- Each micro-shot must additionally label `【包内时间】`, `【镜头类型】`, `【生成方式】`, and, when longer than 3 seconds, `【例外原因】`. `【包内时间】` is the exact `0.0s–1.2s` placement inside its parent video package.
- Each video package must additionally label `【项目规格】`, `【视频编号】`, `【中文显示名】`, `【场景与连续状态】`, `【光线】`, `【出场人物】`, `【声音与台词】`, `【视频生成总提示词】`, and `【视频级负面约束】`. `【视频生成总提示词】` is the directly usable VIDEO prompt; child `【视频生成提示词】` is only the local micro-shot prompt. Keep source coverage as `原文单元 → 视频编号 → 镜头编号`; do not expand approved plot material merely to fill time.
- Keep the workbench import labels `【视频编号】`, `【镜头编号】`, `【包内时间】`, `【时长】`, `【视频生成总提示词】`, `【视频级负面约束】`, `【视频生成提示词】`, and `【负面约束】` unchanged in machine-compatible exports. Creator-facing delivery is matching bilingual DOCX and HTML; a legacy Markdown/TXT workbench import is only an optional machine sidecar, never the creator's sole deliverable.
- In every human-readable package and workbench view, use `【视频编号】` and `【镜头编号】` as the labels. The stable IDs (`VIDEO-*`, `SHOT-*`) may appear only as their values or in the JSON/export machine fields; never use English ID prefixes as the visible production heading.
- Spoken dialogue uses the approved English character name and English dialogue, with adjacent Chinese reference meaning in the readable view marked unspoken. Pair scene/action/production notes, human-facing prompt explanations, asset descriptions, shot descriptions, and negative constraints in Chinese and natural English. Keep stable IDs and import keys unchanged. For timing, use the actual English speaking rate; do not time the unspoken Chinese reference as dialogue.
- Never assume a model generates usable audio. When it does not, emit a voice/lip-sync plan with precise in/out times and separate dialogue, ambience, SFX, and music cues rather than claiming an in-model audio result.
- Provide both Chinese and English for each `【视频生成提示词】`, VIDEO master prompt, asset prompt, and negative prompt under the same ID. Keep the selected model's designated one-language submission prompt in the existing import field (Chinese unless the creator or target specifies otherwise); put the other version in the creator reading view or a compatible companion field. Never merge both languages into one model input.
- Do not leave costume, prop, location, prompt placement, scene/episode timing, or approved asset source implicit. A prompt may reference only approved `CHAR-*`, `LOOK-*`, `SET-*`, and `PROP-*` assets; it may not add an unapproved garment, accessory, prop, or visual fact.

## Output contract

Return these seven sections in order:

1. `Asset Creation Pack` — paired Chinese/English asset image-generation prompt and negative prompt, reference/seed/model/aspect-ratio data, and approval state for every asset that needs creation.
2. `Asset Ledger` — `CHAR-*`, `LOOK-*`, `SET-*`, and `PROP-*` definitions plus current state.
3. `Video Packages and Shot List` — one independently generatable video package per source coverage range. Put its complete `VIDEO-PROMPT-*` before its numbered micro-shots; then include the local layered prompt for every `SHOT-*`.
4. `Prompt Placement Map` — `Scene → Video → Shot → time range → prompt ID → asset IDs → export row`.
5. `Seedance / MediaGo Import Rows` — one structured row per shot using the required schema; include JSON or CSV only when the target system accepts it.
6. `Continuity warnings and approval requests` — only unresolved visual conflicts or required upstream decisions.
7. `production-workbench.json` — machine-readable asset, shot, task and review status; follow `references/production-workbench-manifest.md` exactly. Deliver the complete creator-readable package as matching bilingual DOCX and HTML; machine JSON is supplemental.

## Failure and rewrite conditions

Rewrite the package when an asset requiring creation lacks a creation prompt or approval state; when a video package lacks its own direct VIDEO master prompt or video-level negative prompt; when a video package lacks contiguous micro-shots; when a micro-shot lacks source scene, package in/out time, active character/look/set/prop state, local layered generation prompt, negative constraint, import row, prompt placement, or production-workbench record; when a normal micro-shot exceeds 3 seconds without an exception; when scene/episode durations do not reconcile; when dialogue cannot fit its shot; when a costume or prop changes without an approved transition; when shot geography contradicts the screenplay; or when a prompt invents an unapproved visual fact.

---

## DramaGo bundled V11 references

This section is generated from the exact pinned USVDS V11 snapshot. Treat it as supporting reference material; the skill rules above remain authoritative.


### Bundled: references/asset-ledger-template.md

# Visual Asset Ledger Template

## Character

`CHAR-01` — English name; age/presentation; face and hair anchors; body/silhouette; recurring identity markers; approved reference image ID/path if supplied; image prompt; negative prompt; aspect ratio; target image model/seed if supplied; approval state.

## Costume / look

`LOOK-01A` — linked character ID; garment layers; palette/material; footwear; accessories; condition; episode/scene range; transition trigger; image prompt; negative prompt; reference image ID/path if supplied; aspect ratio; target image model/seed if supplied; approval state.

## Scene / set

`SET-01` — place and geography; time/weather/light; key set dressing; visual anchors; screen-direction constraints; episode/scene range; establishing-image prompt; negative prompt; reference image ID/path if supplied; aspect ratio; target image model/seed if supplied; approval state.

## Hero prop

`PROP-01` — form/material/scale; condition; owner or holder; current location; narrative function; state changes; episode/scene range; isolated-image prompt; negative prompt; reference image ID/path if supplied; aspect ratio; target image model/seed if supplied; approval state.

## Shot linkage

`SHOT-01` — source scene; range; `CHAR-*`; `LOOK-*`; `SET-*`; `PROP-*`; asset-lock prompt; shot-delta prompt; negative/do-not-change; prompt ID; continuity in/out; screen direction; entry/exit and prop handoff notes.


### Bundled: references/director-execution-contract.md

# V8-Inspired Director Execution Contract

Apply this reference when the storyboard package is intended for visual generation, voice execution, or MediaGo-style assembly. It complements the asset ledger and import schema; it does not replace them.

## Project specification is the source of truth

Start each episode with one approved specification: visual medium, aspect ratio, target model/endpoint, maximum generated-video duration, locale, dialogue language, and audio path. If a source template contains conflicting defaults, report the conflict and use the approved project specification.

## Video package

Each independently generatable video package contains:

- `【项目规格】` — approved project specification.
- `【视频编号】` — stable `VIDEO-*` ID, Chinese display name, total duration, and source coverage.
- Human-readable headings use `【视频编号】` and `【镜头编号】`; `VIDEO-*` and `SHOT-*` remain the value-level stable IDs used only for tracing, JSON, and exports.
- `【场景与连续状态】` — location, screen geography, positions, orientation, eye line, inherited pose, active assets, and carry-over effects.
- `【光线】` and `【出场人物】` — the active set state and bound `CHAR-*` / `LOOK-*` IDs.
- `【视频生成总提示词】` — one complete Chinese `VIDEO-PROMPT-*` that can be submitted to the selected video model as-is. It sequences the package from first frame to last frame, names every planned cut, binds all active assets, camera/light/audio progression, continuity in/out, and ending state.
- `【视频级负面约束】` — package-level identity, wardrobe, prop, geography, timing, and unwanted-edit exclusions.
- numbered `SHOT-*` micro-shots — each has `【包内时间】` and normally runs 0.5–3 seconds, playing one clear visual action, insert, reveal, or reaction. The video package is the assembly container; the micro-shot is the independently adjustable generation unit.
- `【声音与台词】` — dialogue/voice, ambience, SFX, and music cues, each with time range and execution responsibility.

Use short packages that respect the chosen endpoint. The V8 compatibility target is 15 seconds maximum per independently generated video, not a universal model claim.

Micro-shots must cover the package timeline without gaps or overlap. A micro-shot over 3 seconds requires `【例外原因】` showing why it cannot safely be divided (for example one uninterrupted English line, a continuous physical action, or an endpoint limitation). A meaningful cut, reaction, information reveal, camera change, or prompt change always starts a new micro-shot.

The VIDEO master prompt is mandatory even when every micro-shot has a local prompt. Submit the VIDEO prompt for the normal one-call video generation; use the SHOT prompt only to regenerate or fine-adjust the selected slice. The two levels must describe the same chronology and locked asset state.

## Shot grammar

Every shot makes its visual priority inspectable:

- camera/framing, angle, motivated movement, focus, and depth;
- subject position plus foreground/midground/background when material;
- a visible action, playable expression, and resultant change;
- active asset IDs, including the costume, prop holder/location, and scene state;
- continuity in/out, screen direction, entry/exit, and prop handoff when relevant.

Use a coherent 180-degree axis unless a deliberate, labeled break is needed. Treat shot-style libraries as optional creative choices, never as an instruction to imitate a named filmmaker.

## Language and operator readability

Creator-readable production descriptions, shot/asset fields, and prompt explanations are paired Chinese and natural English. Keep stable IDs such as `VIDEO-005` and `SHOT-005-03` for traceability; retain required `display_name_zh` / `【中文显示名】` import fields and add the matched English display in the DOCX/HTML view. The approved English spoken line remains the only performed line; provide its Chinese meaning as an unspoken reference. Do not change the locked line through translation.

## Voice and dialogue

Use the approved English name and English spoken line exactly once dialogue is locked. Mark `on-camera`, `off-screen`, `voice-over`, or `inner voice`; only on-camera speech receives lip-sync timing. Give each cue a speaker/voice identifier, performance state, start/end time, and language/locale/accent if supplied. Do not infer a voice ID, audio capability, or model feature.

## Asset and continuity audit

The asset ledger remains the source of truth. Every shot binds `CHAR-*`, `LOOK-*`, `SET-*`, and applicable `PROP-*` IDs. Record wardrobe condition, wounds, dirt/damage, prop owner/location/condition, weather/light, and set changes in machine-checkable fields. Use `Asset Lock Prompt + Shot Delta Prompt`: never rely only on prose inheritance from a previous video.

## Traceability and export

Preserve both mappings:

1. source screenplay unit → `VIDEO-*` → time-coded `SHOT-*` micro-shot;
2. `SHOT-*` → `PROMPT-*` → asset IDs → export row → generation result/status.
3. `VIDEO-*` → `VIDEO-PROMPT-*` → target model submission → generation result/status.

No prompt may be a detached appendix. If the selected MediaGo build has no documented import schema, export a review table only and label the native import fields as pending.


### Bundled: references/production-workbench-manifest.md

# Production Workbench Manifest

For every storyboard delivery, emit `production-workbench.json` beside the readable review package. It is the status source for the DSH production workbench and MediaGo handoff; it never replaces the human-readable screenplay, asset ledger, or shot package. Keep the `/v1` schema identifier for backward compatibility, but use the additive `videos` and time-coded micro-shot fields below for the visual workbench.

Legacy Markdown/TXT workbench import remains an optional machine sidecar; it is not the creator-facing deliverable. Deliver the paired Chinese/English shot and prompt package as matching DOCX and HTML. Keep the import labels `【视频编号】`, `【镜头编号】`, `【包内时间】`, `【时长】`, `【视频生成总提示词】`, `【视频级负面约束】`, `【视频生成提示词】`, and `【负面约束】` unchanged in any machine-compatible sidecar. JSON remains the exact interchange format for assets, statuses and task queues.

```json
{
  "schema_version": "us-vertical-drama-workbench/v1",
  "episode_id": "EP-001",
  "episode_display_name_zh": "第 1 集：拍卖厅的火花",
  "assets": [
    { "id": "CHAR-EVE", "display_name_zh": "女主角：伊芙", "kind": "character", "status": "approved" },
    { "id": "LOOK-EVE-NIGHT", "kind": "look", "status": "approved" },
    { "id": "SET-AUCTION", "kind": "set", "status": "approved" },
    { "id": "PROP-BOTTLE", "kind": "prop", "status": "approved" }
  ],
  "videos": [
    {
      "video_id": "VIDEO-001",
      "display_name_zh": "伊芙发现入口异动",
      "source_scene_id": "EP01-SC01",
      "duration_seconds": 8,
      "video_prompt_id": "VIDEO-PROMPT-001",
      "video_master_prompt": "8 秒竖屏视频。深夜拍卖厅，伊芙先感到入口异动，随后抬眼、吊灯闪烁、手中的酒杯出现裂纹，最后她握紧酒杯凝视入口；人物、晚礼服、拍卖厅和酒杯持续一致。镜头依次为手部特写、反应近景、吊灯插入和人物近景；冷白吊灯光，紧张克制。",
      "video_negative_prompt": "不改变角色身份、晚礼服、酒杯、拍卖厅方位或持杯的手；不新增人物；不跳切到其他地点。",
      "status": "ready_to_generate"
    }
  ],
  "shots": [
    {
      "shot_id": "SHOT-001",
      "video_id": "VIDEO-001",
      "display_name_zh": "伊芙回头看向入口",
      "prompt_id": "PROMPT-001",
      "source_scene_id": "EP01-SC01",
      "timeline_in_seconds": 0,
      "timeline_out_seconds": 1.6,
      "duration_seconds": 1.6,
      "shot_type": "反应",
      "generation_mode": "independent",
      "asset_ids": ["CHAR-EVE", "LOOK-EVE-NIGHT", "SET-AUCTION", "PROP-BOTTLE"],
      "asset_lock_prompt": "Approved immutable asset anchors.",
      "shot_delta_prompt": "Only the shot-specific action, composition and motion.",
      "negative_prompt": "Explicit identity, wardrobe, prop and geography exclusions.",
      "status": "ready_to_generate"
    }
  ],
  "tasks": [
    { "id": "TASK-001", "status": "todo", "targets": ["SHOT-001"] }
  ]
}
```

Use `character`, `look`, `set`, or `prop` for asset kinds. Asset status is `draft`, `pending_approval`, `approved`, or `blocked`. Shot status is `draft`, `blocked`, `ready_to_generate`, `generating`, `review_required`, or `approved`. Task status is `todo`, `in_progress`, `blocked`, `review`, or `done`.

A video package may not exceed 15 seconds in V8-compatible mode. Every explicit `videos[]` record requires `video_prompt_id`, `video_master_prompt`, and `video_negative_prompt`. The master prompt field holds the selected one-language, directly usable VIDEO-level submission (Chinese by default for the existing adapter): it covers the entire package timeline, cuts/shot progression, locked assets, camera, light, sound/voice plan, continuity, and exclusions. The creator DOCX/HTML also presents a matched English version and identifies which version is submitted. The child `SHOT-*` prompt is local replacement guidance, not the only video prompt. Its child shots are time-coded micro-shots: each must fit inside its parent video, their ranges must be contiguous with no overlap, and each should be 0.5–3 seconds. A longer micro-shot requires `duration_exception_reason_zh`; a shorter one requires `short_duration_reason_zh` when it is under 0.5 seconds. `shot_type` and `generation_mode` retain the required schema terms. `display_name_zh` is required for every human-visible video and shot; provide English display text in the reading view.

A shot may move to `ready_to_generate` only after every bound asset is `approved`, its local three prompt layers are present, source/trace IDs are present, its parent video has all three VIDEO prompt fields and is valid, and the micro-shot has a valid timeline. Do not turn unresolved items green: retain them as `blocked` with an explicit task or continuity warning.


### Bundled: references/seedance-mediago-output-schema.md

# Seedance / MediaGo Shot Output Schema

Emit two linked records: one direct-generation record per `VIDEO-*`, followed by one local-adjustment record per child `SHOT-*`. Keep the same `scene_id`, IDs, and approved asset IDs in the screenplay, prompt map, and import output. Emit a Markdown table for review; additionally emit JSON or CSV only when the selected MediaGo build specifies an accepted import format.

| Field | Required | Meaning |
|---|---:|---|
| `video_prompt_id`, `video_master_prompt`, `video_negative_prompt` | yes on every VIDEO record | Direct-submission prompt in the selected model language for the full video package, including time sequence/cuts, locked assets, camera, light, sound/voice plan, continuity and exclusions. Pair Chinese/English versions in the creator view under the same ID; submit only the selected version. This is not replaced by child SHOT prompts. |
| `scene_id`, `video_id`, `shot_id`, `prompt_id` | yes | Stable source, video package, time-coded micro-shot, and prompt-placement IDs. Supply Chinese display names beside each human-visible ID. |
| `display_name_zh`, `shot_type` | yes | Chinese operator-facing name and Chinese shot type such as `反应`, `特写插入`, `动作`, or `对白`. |
| `timeline_in_seconds`, `timeline_out_seconds`, `duration_seconds` | yes | Exact placement inside its parent video package and planned micro-shot duration. The ranges are contiguous; a normal micro-shot runs 0.5–3 seconds. |
| `target_model` | yes | Selected target, e.g. `Seedance 2.0`; never assume a capability. |
| `character_reference_ids`, `look_reference_ids`, `scene_reference_id`, `prop_reference_ids` | yes when applicable | Approved `CHAR/LOOK/SET/PROP` IDs and approved image/reference paths. |
| `asset_lock_prompt`, `shot_delta_prompt`, `negative_prompt` | yes | Selected one-language production prompts: immutable approved asset anchors, the micro-shot-specific change, and exclusions. Show paired Chinese/English versions in the creator view. Embedded approved English character names and spoken lines remain exact. |
| `camera`, `continuity_in`, `continuity_out` | yes | Composition/movement plus inter-shot state and screen-direction continuity. |
| `source_coverage`, `asset_state_in`, `asset_state_out` | yes | Source unit mapping plus machine-checkable wardrobe, prop, set, weather/light, and damage state. |
| `dialogue_audio`, `lip_sync_timing` | required when dialogue is spoken | Exact approved English line(s), speaker, and start/end seconds; otherwise `none`. |
| `voice_cues` | yes | Dialogue/voice, ambience, SFX, and music cues with time range and execution responsibility; mark unsupported model features as external. |
| `start_frame_reference`, `end_frame_reference`, `seed`, `duration_exception_reason_zh`, `short_duration_reason_zh` | optional / conditional | Supply only when supported and selected. An exception reason is required outside the normal 0.5–3 second micro-shot range. |
| `export_row`, `generation_status`, `generation_result_id` | yes / pending / optional | Stable row ID, `pending` until generated, and a returned result ID only after the target system provides one. |

## Seedance use

Treat `target_model: Seedance 2.0` as a target label, not proof of a specific duration, reference-image, seed, frame-control, or audio feature. Obtain those limits from the user’s enabled endpoint or current MediaGo adapter; validate against them before export.

## MediaGo handoff

MediaGo receives the Asset Creation Pack first, then approved reference IDs/paths, followed by one direct VIDEO record and its time-coded micro-shot records. If MediaGo has no structured-import adapter, return the same fields as a paired Chinese/English DOCX/HTML review table and preserve the prompt-placement map; do not pretend a generic JSON blob can be imported. The native adapter must map `VIDEO-*`, `VIDEO-PROMPT-*`, `SHOT-*`, `PROMPT-*`, asset IDs, and `export_row` without flattening them away.


---

## DramaGo runtime integration contract

- Baseline: USVDS v11@0c68d9cbf525eee6f2a98d3098294cdedbb29e77, plugin 1.2.6.
- Write creator/development artifacts into the existing DramaGo project Documents; update the current artifact instead of inventing a parallel project store.
- Use DramaGo document categories for executable artifacts: screenplay for screenplay output and storyboard for storyboard/shot output when applicable.
- When creating/updating this stage's authoritative artifact, preserve the document tag `usvds:artifact:storyboard` on that DramaGo Document so Gate routing can locate it deterministically.
- Character/scene/prop identity and variants remain owned by DramaGo Canon/Variant; shot execution and continuity remain owned by ShotManifest/ResolvedState; generation execution remains owned by GenerationTask/Asset.
- Never create shadow USVDS project, asset, shot, approval, job, or generation state.
- Legacy words such as APPROVED/PASS inside this Skill are narrative/workflow guidance only and must not set DramaGo human/system Gate tags by themselves. Gate state changes require the existing DramaGo approval/user action path.
- Before downstream production work, inspect the current DramaGo USVDS V11 Gate state and stop when the required gate is blocked.
