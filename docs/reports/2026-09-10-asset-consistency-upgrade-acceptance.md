# MediaGo Desktop 资产一致性与上下文优化升级验收报告

日期：2026-09-10

负责人主任务：`tsk_07675aea941407b7`

代码仓库：`D:\openai\MediaGo-Drama`

基线分支：`mediago-clean-baseline-20260903`

基线 HEAD：`78739afa269e4f634f4431b98630add38bcdb184`

当前状态：未 commit、未 push；全部改动保留在本地工作树，便于最终人工检查后再决定版本提交。

---

## 1. 结论

本轮升级达到既定 Desktop 基础版目标：在不推翻现有 MediaGo 文档、资产、生成、Provider、Electron/Go 架构的前提下，补齐 Canon Core / Variant / Reference、Shot Manifest、Continuity Resolver、确定性 Prompt Compiler、上下文裁剪、旧项目惰性迁移和真实样本 Production QA。

核心变化不是继续增加提示词，而是把原先由模型反复自由解释的角色、造型、场景、道具和镜头状态，下沉为可持久化、可继承、可锁定、可测试的生产数据。

现有旧项目仍可打开和走 legacy generation；新一致性链只有在 Canon/Shot 同步后才接管对应镜头。项目概览会自动做非破坏式惰性同步，用户不需要手工迁移 Markdown，也不需要理解 Canon、Shot Manifest 等内部概念才能继续使用原来的生成按钮。

---

## 2. 可回滚基线

升级前完整仓库 ZIP：

`D:\openai\MediaGo-Backups\MediaGo-Drama-pre-asset-consistency-20260910-005356.zip`

SHA-256：

`FBE6E06FAD92FA9DE2D89A3D26608572E6A996807FF9D6CE6B9452D89646BEA2`

压缩包目录已经验证可读取。

回滚优先级：

1. 在正式提交前，可直接以当前 Git 基线 `78739afa...` 与未提交工作树做差异审查；不接受本轮升级时可丢弃本轮工作树改动。
2. 如需要完整环境级恢复，使用上述 ZIP，并再次核对 SHA-256。
3. 新增数据库结构为附加表，不要求修改原 Markdown；旧项目数据的恢复不依赖 Canon/Shot 表。

---

## 3. P0：Windows VPN / TUN / localhost 兼容

### 已完成

- Electron sidecar 实际服务和端口探测仍绑定 `127.0.0.1`。
- Electron 向 Renderer 发布的 sidecar origin 改为 `http://localhost:<dynamicPort>`。
- 前端开发环境本地 API fallback 改为 `localhost`。
- AgentBridge / Codex Relay 对 loopback 客户端公布 `localhost`。
- 非 loopback / LAN 地址不被错误重写。
- Electron renderer CSP 同时接受 `localhost`、`127.0.0.1`、`::1`。
- 现有 NO_PROXY loopback 保护继续保留。

### 安全边界

没有把 Go sidecar 监听改成 `0.0.0.0`，因此没有新增局域网/公网暴露。

DeepSeek、MiniMax、DMX、OpenRouter 等第三方 Provider 公网 Base URL 未被改写。

### 验证

- Electron / frontend 网络与安全测试：57 tests PASS。
- Electron compile PASS。
- Workspace production build PASS。
- Go app/acp/settings regression PASS。

备注：代码层已经针对 VPN/TUN 影响 loopback literal IP 的场景完成架构修复；仍建议后续在安装 360、Clash/TUN、不同 Windows 版本的实际测试机矩阵上做 field validation。

---

## 4. Canon Core / Variant / Reference

### 数据模型

新增：

- `canon_assets`
- `canon_references`

Canon Asset 使用一套通用模型表达 character / scene / prop，不把漫剧业务拆成三套重复数据库模型。

Core：`parent_id = NULL`

Variant：`parent_id = Core/Parent Canon ID` + `variant_kind`

Reference：只引用现有 `assets.id`，不复制物理媒体文件。

### Canon 权威规则

- stable Core ID 基于 project/resource/document/section 生成。
- source hash 不变时不增加版本。
- source 内容变化时对应对象单独递增版本。
- `locked` Canon 不被文档惰性同步覆盖。
- storyboard 不作为 Canon 资源，保持 execution layer 与 asset layer 分离。

### H2 Core + H3 Variant 自动物化

负责人最终审查时发现原文档 resource parser 只输出 H2。如果不补这一层，新的 Skill 虽然写出 H3 Variant，但数据库会把 H3 内容一起混入 Core。

该问题已经补齐：

- Character：`### 造型变体：婚礼/工作/...` → `variant_kind=look`
- Scene：`### 场景变体：夜间/...` → `variant_kind=scene`
- Scene：`### 区域：取餐口/...` → `variant_kind=zone`
- Prop：`### 状态变体：断裂/...` → `variant_kind=state`

Core Prompt 在同步时会剥离这些 H3 Variant 段，因此“白色婚纱”等造型内容不会污染 Character Core。

验证：只修改“婚礼 Variant”时，Core 仍保持 v1，婚礼 Variant 单独从 v1 升到 v2，工作 Variant 仍为 v1。

### Selected Asset 桥接

现有 `ProjectSelectedAssetModel` 可幂等桥接为 CanonReference：

- character → identity
- scene → scene_master
- prop → prop_master

物理图片继续使用原 MediaAsset。

### HTTP

- `GET /api/v1/projects/{projectId}/canon`
- `POST /api/v1/projects/{projectId}/canon/sync`
- `POST /api/v1/projects/{projectId}/canon/{canonId}/variants`
- `PATCH /api/v1/projects/{projectId}/canon/{canonId}/status`

---

## 5. Shot Manifest / Continuity Resolver

新增 `shot_manifests`。

关键字段包括：

- storyboard document / section / shot key
- sequence
- inherits_from_id
- action / camera / audio
- style_profile_id
- bindings_json
- state_changes_json
- resolved_state_json
- compiled_prompt
- source hash / version / status

### Canon 绑定

Shot 可以明确绑定：

- Character Core
- Character Look Variant
- Scene Core / Variant
- Prop Core / Variant
- reference asset IDs

Variant 会校验必须属于相应 Core，错误绑定直接拒绝。

### 状态继承

Continuity Resolver 采用确定性深度合并：

- 当前镜没有显式变化 → 继承上一镜。
- 只改变嵌套字段 → 只覆盖该字段，不重置兄弟状态。
- `null` → 显式清除该状态。
- 换场景不意味着自动换装。
- locked Character 下的 `identity/core/canon` 修改直接拒绝。

### Binding 继承

负责人最终结构审查时发现，仅继承 ResolvedState 仍不足够：如果 Shot 1 绑定 wedding Variant，而 Shot 2 只写“继续走”，Shot 2 的 Binding 可能退回 Core。

该问题已经补齐：

- 同一 Character Core，当前 Shot 没显式 Variant → 继承上一镜 Character Variant。
- 当前 Shot 显式 Variant → 当前值优先。
- 当前 Scene 为空 → 继承上一镜 Scene Core/Variant。
- 当前 Scene Core 相同但 Variant 为空 → 继承上一镜 Scene Variant。
- matching Prop Variant / reference asset IDs 同理继承。

测试已证明：

- Shot 1 = wedding Look + night Scene。
- Shot 2 只写继续动作，不重复造型/场景 → 仍为 wedding + night。
- Shot 3 显式切 work Look → 角色换到 work；同 Scene Core 仍保留 night Variant。
- 编译 Shot 2 仍含“白色婚纱”，没有退回 Core-only。

---

## 6. Deterministic Prompt Compiler

完整 Provider Prompt 不再要求 LLM 每镜重写固定身份和场景描述。

固定编译顺序：

1. Character Core
2. active Character Variant
3. Scene Core / Variant
4. Prop Core / Variant
5. Resolved continuity state
6. Action
7. Camera
8. Audio
9. Style Profile

CanonReference 同时被编译成 provider-neutral reference metadata，再接回现有 `GenerationReferenceBinding` 和视频参考槽位机制。

同一结构化输入重复编译必须产生相同输出；已有测试做字节级/结构级比较。

`CompileAndPersist` 会将最终结果写回 `shot_manifests.compiled_prompt`。

---

## 7. Generation 主链接入

`GenerationMessageRequest` 新增：

`shotManifestId`

当请求携带 Shot Manifest 时：

- Shot Manifest 成为权威执行合同。
- 服务端重新 Compile。
- 自由文本 Prompt 被 Compiler 结果覆盖。
- 旧 `PromptSupplements` 清空。
- 临时 `ReferenceURLs` 清空。
- ReferenceAssetIDs / ReferenceBindings 使用 Canon Compiler 结果。
- DocumentID / SectionID / DocumentContext 从 Shot 自动回填。
- ResourceType 固定为 storyboard。

因此不能通过前端残留 Prompt、额外图片或旧风格补充词绕过 Canon / Continuity 约束。

没有 compiler、缺 projectId、Shot 不存在等情况均 fail closed，而不是静默退回自由文本生成。

---

## 8. Agent Context / Token 优化

审计确认：现有 ACP 并没有直接把完整 character/scene/prop 文档全部塞入每次 Prompt；已有实现主要通过工作区 @ 资源索引提供引用。因此本轮没有虚构“整篇文档移除”这类不存在的优化。

真实优化点是 storyboard 大项目的资源索引。

原上限：120 条。

新增 storyboard compact 上限：36 条。

排序规则：

1. 用户显式 @ reference：最高优先级，绝不裁掉。
2. 当前用户请求 / 当前 storyboard 正文直接出现的资源标题。
3. 相关 document title。
4. 稳定 fallback。

如果显式引用本身超过 36 条，限制自动扩大，最高仍受旧 120 上限保护。

验证：

- 80 个角色资源：当前镜提到第 79 个、显式 @ 第 80 个 → 两者都保留，总索引压到 36。
- 40 个显式引用 → 40 个全部保留，不强行裁成 36。

---

## 9. Skill 契约升级

### character-writer

- 一名角色一个 H2 Character Core。
- Look/时期/状态变化进入 H3 Variant。
- 不把婚礼/工作/便装拆成“不同人物”。
- 原文没有的年龄、关系、经历等不得补成事实。
- 必要纯视觉缺口只能标记为“可选视觉建议”。

### scene-writer

- 一个物理地点一个 H2 Scene Root。
- 区域/昼夜/状态进入 H3 Variant。
- 禁止无依据室内→室外、绿色→银灰、结构/门窗/设备漂移。

### prop-writer

- 改为 extraction-only 事实整理层。
- 不新增原文没有的拍摄、录制、分享、清洁、加热、破坏、传递等动作。
- 临时状态进入 H3 State Variant。

### storyboard-writer

- 改为 Shot Manifest 的人类可读来源，不是完整 Provider Prompt。
- 使用已有 @ 资源绑定。
- 明确 State Change；没有变化写“继承上一镜”。
- 删除逐镜强制复制伦勃朗光、体积光、8K、HDR10+、120fps、无字幕/无水印等固定模板。
- 全局固定质量词进入 Style Profile / 生成端。
- 不新增原剧本不存在的对白、动作、道具用途、时间跳转和因果关系。
- 加入总时长检查。
- 明确禁止约 12 分钟内容在无说明情况下压成约 4 分钟。

---

## 10. 旧项目惰性迁移

新增：

`POST /api/v1/projects/{projectId}/shot-manifests/sync`

行为：

1. 先同步 Canon。
2. 读取已有 storyboard H2 resources。
3. 已经有 Shot Manifest 的 section → Reused，不覆盖。
4. 没有的 section → 创建 draft Shot。
5. 优先解析已有 `mention://document/section`。
6. 无 mention 时，只做保守的已有 Canon 名称匹配，不随意创造 ID。
7. Shot 文本出现 H3 Variant 名称时绑定该 Variant。
8. 后续 Shot 未重复 Variant 时由 Continuity binding inheritance 继承。
9. 不修改原 Markdown。

因此旧项目仍可继续使用；用户手工调整过的 Shot 不会被再次同步覆盖。

---

## 11. 桌面端 UI / 小白路径

新增 `apps/workspace/src/domains/workspace/api/continuity.ts`。

项目概览加载既有资源后自动：

- sync Canon
- sync Shot Manifest
- 刷新一致性数据

新增“一致性资产”摘要，只展示用户能理解的状态，不暴露 JSON/内部数据结构：

- 角色 Core
- 场景 Core
- 道具 Core
- Variant 数量
- 结构化 Shot
- 继承上一镜数量
- locked 数量
- 已编译数量
- conflict 数量

原操作习惯保持：

- 单镜“生成视频”自动根据 document+section 找到 shotManifestId。
- 批量生成同样自动绑定 shotManifestId。
- 没有 Shot Manifest 时仍走 legacy fallback。
- 用户不需要手工填写 Canon ID / Variant ID / Shot ID。

`shotManifestId` 已贯穿：

MediaGenerationDialog → VideoGenerationDialog → DocumentSectionGenerator → MediaGenerationWorkspace → useGenerationWorkspace → useGenerationSubmit → GenerationMessageRequest。

Manifest-backed 请求允许没有自由文本 Prompt，并主动清理前端残留 reference，最终以服务端 Compiler 为准。

---

## 12. Production QA 与真实项目回归

新增：

`services/server/internal/service/productionqa`

当前确定性检测：

- `duration_compression`
- `scene_canon_conflict`
- `repeated_global_boilerplate`
- `unsupported_prop_usage`

### 真实样本

直接读取：

`project-679e2eab7f3a12cc`

不是只测试人工构造 fixture。

已确认：

- 剧本：本集时长约十二分钟 = 720 秒。
- 分镜：16 个 H2 group，每组 14.8 秒，总计 236.8 秒 = 3.95 分钟。
- 时长比例：约 32.89%。
- Scene Canon：老厂房食堂前厅、绿色老式卷帘门。
- Storyboard 第 1 组：门外、银灰卷帘门。
- Storyboard 中“无字幕 / 无水印 / 无 BGM”等固定词大量逐镜重复。
- Prop 文档的手机使用关系增加“拍 / 录 / 传出”等原剧本没有的动作。

真实目录 integration test 已成功检测全部四类问题。

其他历史问题的测试覆盖：

- 服装/状态突然重置 → Continuity Resolver + Binding inheritance。
- 错误 Variant → Variant ownership validation。
- locked identity 被改 → conflict rejection。
- reference 缺少权威绑定 → CanonReference + GenerationReferenceBinding。
- 下游新增对白/行为 → Skill contract regression。
- Token 高、资源索引过大 → compact reference index。

---

## 13. 测试与构建门禁

### Go 后端当前最终代码树

执行：

`go test ./...`

结果：PASS，覆盖 server 全包，包括：

- app / MCP
- document
- generation
- canon
- shotmanifest
- productionqa
- prompt / acp
- repository
- media / settings / selection 等

执行：

`go build ./...`

结果：PASS。

执行：

`git diff --check`

结果：PASS。

### 真实样本

使用：

`MEDIAGO_REGRESSION_SAMPLE_PROJECT=<project-679...>`

执行：

`go test -v ./internal/service/productionqa`

结果：PASS。

### Instruction Pack

- `go test ./packages/instructions/pkg/pack/builtin ./packages/instructions/pkg/official`
- PASS。

### Frontend

新增 manifest-backed submit 定向测试：17/17 PASS。

ProjectOverview 定向测试：30/30 PASS。

全量 frontend 测试最终使用受控并发运行：

- Test Files: **226 passed / 226**
- Tests: **1465 passed / 1465**

说明：第一次高并发全量执行出现 3 个无关测试 5 秒 timeout，以及 2 个 ProjectOverview 旧断言把新增 Canon/Shot sync POST 计入“总 POST 数量”。三项 timeout 单文件复跑全部 PASS；两项 ProjectOverview 测试已修正为只断言 `/generation/batches` 调用次数，随后 30/30 PASS；最终受控并发全量 1465/1465 PASS。

最终再次执行：

`pnpm build`

结果：TypeScript + Vite production build PASS。

Vite 仍有既有的大 chunk warning，属于性能优化提示，不是本轮编译失败。

---

## 14. 剩余非阻塞风险

### 14.1 Source Variant 删除/重命名后的 stale Variant

当前 source H3 Variant 可以新增和独立更新，但尚未显式区分 `source_synced` 与 `manual` Variant，因此删除/重命名 H3 后，不会自动把旧 source Variant 标为 deprecated。

建议后续增加 source ownership / tombstone reconciliation，避免误删用户手工 Variant。

### 14.2 Variant 专属参考图工作流仍可继续增强

数据模型已经允许 Canon Variant 绑定自己的 CanonReference，Compiler 也会读取 Variant Reference；但目前现有 `ProjectSelectedAsset` 的自动桥接天然以 H2 resource/Core 为主。

后续可以在 UI 增加“为某个 Look Variant 定稿参考图”的更直接操作，不影响本轮 Core/Variant/Shot 连续性能力。

### 14.3 Production QA 不是完整语义审稿器

当前自动化覆盖了可可靠确定的时长、固定颜色事实、重复 boilerplate、部分道具新增行为。复杂对白归属、所有时序倒退、跨几十场的语义矛盾仍应由后续 Quality Gate / Agent Reviewer 扩展，而不应假装简单 regex 可以完全解决。

### 14.4 VPN/TUN 仍需要多机器现场矩阵

代码和自动测试已经验证“bind 与 advertised origin 分离”，但 360、不同 VPN/TUN 驱动、Windows 10/11 不同组合仍建议用朋友测试机做 field validation。

---

## 15. 本轮未做

- 未 commit。
- 未 push GitHub。
- 未删除/覆盖原始项目 Markdown。
- 未把 MediaGo 基础版硬编码成 ComicGo 专属业务。
- 未启动 SaaS 改造；SaaS 可行性评估仍在独立任务中。

---

## 16. 负责人验收判定

基于当前自动测试、构建、旧数据迁移、真实样本回归和结构审查，本轮 Desktop 基础版升级满足主任务既定验收条件。

判定：**PASS，可进入人工体验测试 / 打包测试阶段。**

在正式提交版本前，建议优先做以下人工验证：

1. 一台正常 Windows 环境。
2. 一台开启 Clash/TUN/VPN 的 Windows 环境。
3. 一台装有 360 安全卫士/360 浏览器或残留安全策略的测试机。
4. 新建一个包含“同一角色默认装 → 婚礼装 → 连续 3 镜 → 工作装”的小项目，确认项目概览 Variant/Shot 数量与生成结果一致。
5. 打开现有旧项目，确认自动 sync 不改变原 Markdown，原生成入口仍然可用。
