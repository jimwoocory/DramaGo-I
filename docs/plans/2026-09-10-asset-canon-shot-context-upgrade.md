# MediaGo Desktop：资产一致性、分镜状态与 Context 优化升级方案

日期：2026-09-10
负责人任务：`tsk_07675aea941407b7`
基线：`78739afa269e4f634f4431b98630add38bcdb184`

## 1. 目标

本次升级只针对 MediaGo Desktop 基础版，保持现有 React/Vite/Electron + Go + SQLite/GORM + generation/provider 架构不变，在现有 document resource、selected asset、mention/reference、storyboard/timeline 能力之上补齐：

1. Canon Asset：把角色、场景、道具从“重复描述的文本”升级为可引用、可版本化、可锁定的对象。
2. Variant：区分 Character Core 与 Look Variant、Scene Root 与 Scene Variant、Prop Core 与 Prop Variant。
3. Shot Manifest：分镜成为最终执行清单，显式绑定 Canon/Variant/Reference，并记录状态变化和继承关系。
4. Continuity Resolver：未发生明确变化时，角色造型、场景、道具和临时状态从上一镜继承。
5. Deterministic Prompt Compiler：固定资产描述、锁定约束、参考图和通用视频规则由代码编译，不再要求 LLM 每镜重复重写。
6. Context Budget：Agent 只拿当前任务需要的 Canon/Shot 摘要，减少整篇角色册/场景册/道具册/分镜重复进入上下文。
7. P0 Loopback：服务监听仍绑定 127.0.0.1；Electron/Agent 访问本地 sidecar/relay 时优先发布 localhost origin，兼容 Windows VPN/TUN/透明代理环境。
8. 旧项目兼容：没有 Canon/Shot Manifest 的旧项目仍可打开、生成；首次使用新功能时惰性同步，不破坏原 Markdown。

## 2. 当前已验证基础

### 2.1 文档资源

`services/server/internal/service/document/document_resources.go` 已将 `character/scene/prop/storyboard` 文档按 section 解析成 `WorkspaceDocumentResourceRecord`，资源 ID 为：

`<category>:<documentID>:<sectionID>`

该 ID 可作为 Canon 的稳定来源键。

### 2.2 物理资产

`ProjectSelectedAssetModel` 已支持：

- ProjectID
- ResourceType
- ResourceID
- ResourceTitle
- AssetID
- SourceDocumentID
- SourceTaskID

这意味着“某角色/场景/道具 section 当前选中哪张生成图”已经有现成数据，不重复造 selected asset 机制。

### 2.3 视频参考引用

`GenerationReferenceBinding` + mention resolver 已能把 `mention://document/section` 和 selected media asset 转成视频模型的 `@图片N` / `@音频N` 参考槽位。

因此 Canon 只需要提供权威绑定和版本，不需要重写 provider reference 层。

### 2.4 Storyboard / Timeline

当前已有：

- `StoryboardVideoReel`
- `EpisodeRecord`
- `TimelineClipRecord`

但它们只保存标题、Prompt、媒体 URL 等，没有角色造型/场景/道具/状态继承。

## 3. 根因

当前链路：

`剧本 → 角色/场景/道具自由文本 → 分镜自由文本 → Prompt 自由文本 → 生成`

问题：

- 每个 Skill 都可以重新解释上游事实。
- storyboard-writer 强制每镜自包含，导致角色/场景/风格大量重复。
- 同一角色每次重写外貌 = 每次重新采样 identity。
- 没有状态继承，因此“半脱外套”“拿花束”“婚纱”等临时状态容易下一镜重置。
- 场景 section 被当作多个独立场景，缺少 Root Scene/Zone 关系。
- 文档、selected asset、generation task 有关联，但没有“当前权威版本”的 Canon 层。

## 4. 权威层级

冲突时必须使用固定优先级：

1. User-approved locked Canon
2. Approved Canon Variant
3. Script explicit fact
4. Previous Shot resolved state
5. Current shot explicit change
6. Model inference

生成结果不得反向修改锁定 Canon。

## 5. 数据模型

### 5.1 CanonAsset

单表表达 Core/Variant，避免为 character/scene/prop 建三套表。

字段建议：

- ID
- ProjectID
- ResourceType: character / scene / prop
- ResourceID: 原 document section ID
- SourceDocumentID
- ParentID: 空表示 Core；非空表示 Variant
- VariantKind: look / time / lighting / damage / prop_state / custom
- Name
- SpecJSON: 结构化稳定属性
- PromptText: 供确定性 Prompt Compiler 使用的精简权威描述
- Status: draft / approved / locked / deprecated
- Version
- SourceHash
- CreatedAt / UpdatedAt

例：

`char_hansanhe` = Core
`char_hansanhe/look/default/v1` = Variant

### 5.2 CanonReference

将 Canon/Variant 绑定到现有 MediaAsset：

- ID
- ProjectID
- CanonAssetID
- AssetID
- Role: identity / front / three_quarter / full_body / scene_master / scene_left / scene_right / prop_master / custom
- Priority
- Locked

不复制图片文件，只引用现有 `assets`。

### 5.3 ShotManifest

一条记录对应 storyboard section 内的一个可执行镜头；第一阶段允许 group 级 Manifest，后续可细分到组内 shot。

字段：

- ID
- ProjectID
- DocumentID
- SectionID
- ShotKey
- Sequence
- InheritsFromID
- ActionText
- CameraText
- AudioText
- StyleProfileID
- BindingsJSON
- ResolvedStateJSON
- StateChangesJSON
- CompiledPrompt
- SourceHash
- Version
- Status

BindingsJSON 结构：

```json
{
  "characters": [
    {
      "canonId": "char_a",
      "variantId": "char_a_wedding_v1",
      "referenceAssetIds": ["asset_x"]
    }
  ],
  "scene": {
    "canonId": "scene_church",
    "variantId": "scene_church_day_v1"
  },
  "props": [
    {"canonId": "prop_bouquet", "state": "held_by:char_a"}
  ]
}
```

## 6. Continuity Resolver

规则：

1. 当前镜头显式绑定优先。
2. 未绑定但 `inheritsFrom` 存在时复制上一镜 resolved state。
3. `stateChanges` 只覆盖指定键，不重置其余状态。
4. 场景切换不自动换角色服装。
5. 服装变化不自动修改 Character Core identity。
6. 临时状态持续，直到明确清除。
7. 如果锁定 Canon 与镜头描述冲突，标记 conflict，不让生成 silently continue。

## 7. Prompt Compiler

LLM 只负责输出 Shot Manifest 中的“动作/镜头/变化/引用选择”。完整生成 Prompt 由代码编译。

编译顺序：

1. Character Core PromptText
2. Active Character Variant PromptText
3. Scene Core/Variant PromptText
4. Prop PromptText
5. Resolved temporary state
6. ActionText
7. CameraText
8. AudioText
9. Style Profile
10. Provider-specific reference tokens
11. Negative/locked constraints

固定内容只保存一份，编译时展开，不进入每一次 Agent 推理上下文。

## 8. Context Budget

Agent 上下文禁止默认注入整篇角色册/场景册/道具册。

新的任务上下文组成：

- 当前 screenplay scene 摘要
- 当前 storyboard section
- 相关 Canon 摘要
- 上一镜 resolved state
- 可用 Variant 列表
- 必要 reference asset IDs

目标：同一镜头规划时固定资产信息只出现一次；生成 Prompt 的重复文本不计入 Agent 推理 Token。

## 9. Skill 契约调整

### character-writer / scene-writer / prop-writer

职责：创建/更新“设定来源文档”，禁止凭空扩写剧情事实；新增内容必须区分：

- locked facts
- optional visual suggestions
- variants

### storyboard-writer

删除“每镜都必须重复完整角色/场景描述”的硬要求。

改为：

- 必须忠实原剧本顺序。
- 必须为角色、场景、道具使用已有 mention 引用。
- 如果造型/场景/道具状态发生变化，必须显式写 State Change。
- 未发生变化时不重复描述身份信息，由 Shot Manifest 继承。
- 通用光影/8K/HDR/无字幕等改为 Style Profile，不要求每镜重复。

## 10. 兼容策略

旧项目不强制迁移 Markdown。

惰性同步流程：

1. 读取现有 document resources。
2. Canon 表中不存在 source key 时创建 draft Canon。
3. 已有 ProjectSelectedAsset 作为初始 CanonReference。
4. 旧 storyboard section 创建 draft ShotManifest。
5. 无明确 Variant 时绑定 Core/default。
6. 原 Prompt 保留，作为 `legacyPrompt`；新 Compiler 可随时关闭回退旧路径。

Feature flag：

- `assetCanon.enabled`
- `shotManifest.enabled`
- `promptCompiler.enabled`

第一阶段默认只对新建/显式同步项目开启，验收后再默认开启。

## 11. P0 Loopback 兼容

禁止全局替换 127.0.0.1。

保留：

- Go server bind 127.0.0.1
- 端口探测 listen 127.0.0.1
- 安全白名单包含 127.0.0.1 / localhost / ::1

修改：

- Electron sidecar `origin` 对客户端发布 `http://localhost:<port>`
- dev fallback API origin 发布 localhost
- AgentBridge 默认对本地客户端公布 localhost
- Codex relay runtime 使用 localhost bridge base URL
- dev electron URL 改 localhost，并同步信任/导航测试

验收：VPN/TUN 开启情况下，前端→sidecar、Codex→relay、relay→国内第三方 Provider 均可用。

## 12. 分工

项目负责人：

- 方案冻结
- 任务依赖排序
- 接口边界审查
- 合并审查
- 回归/最终验收

子任务 A：P0 Network

- localhost advertised origin
- Electron/Go/前端测试

子任务 B：Canon Core

- GORM models/repository/service
- lazy sync
- selected asset reference bridge

子任务 C：Shot/Compiler

- ShotManifest
- resolver
- deterministic compiler
- generation request integration

子任务 D：Skills/UI/Compatibility

- Skill 契约
- 分镜展示绑定/锁定状态
- legacy fallback
- sample migration

## 13. 验收

必须同时满足：

1. 旧项目可打开，旧生成链仍可用。
2. Character Core 与 Look Variant 分离。
3. Shot 能绑定 character/look/scene/prop。
4. 上一镜状态默认继承；明确变化才更新。
5. locked Canon 冲突会报错/警告，而不是静默覆盖。
6. 生成 Prompt 可由纯代码重复生成相同结果，不依赖 LLM 重写固定描述。
7. storyboard-writer 不再要求 80 个镜头重复同一套全局质量词。
8. `project-679e2eab7f3a12cc` 可检测至少：内外景漂移、卷帘门颜色冲突、服装状态重置、道具自行加戏、分镜时序倒退。
9. Go/前端相关测试通过。
10. 完成回滚说明和风险清单。
