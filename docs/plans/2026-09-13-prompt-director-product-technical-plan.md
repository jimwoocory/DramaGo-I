# MediaGo Prompt Director：角色 / 场景 / 道具 / 镜头精细化提示词系统

日期：2026-09-13
负责人任务：`tsk_51ec71149ab89ce9`
项目：`D:\openai\MediaGo-Drama`
性质：产品方案 + 技术影响评估 + 开发任务拆解
边界：本方案阶段不修改现有业务代码，不 commit、不 push、不 publish。

---

## 0. 负责人结论

### 0.1 是否需要动 UI？

**需要，中等规模增量改造，不建议重做主导航。**

现有 UI 已能展示：

- 一致性资产数量；
- Production Shot；
- Shot 的动作 / 机位；
- 单镜关键帧生成入口。

但目前看不到、也不能编辑：

- 角色身份锚点、默认外观、造型 Variant 的结构化提示词；
- 场景空间结构、区域、材质、光线、氛围；
- 道具材质、尺寸、磨损、识别特征、交互方式；
- 镜头构图、景别、镜头高度、焦段、主体优先级、光影、情绪、连续性约束；
- 编译后的 Prompt 为什么这样生成；
- 图片 / 视频 / ComfyUI 等目标之间 Prompt 的差异。

因此必须增加“导演资产卡”和“镜头导演面板”，但第一阶段不增加新的一级导航，优先以 Dialog / Drawer 的方式嵌入现有 ProjectOverview 和 Production Shot 流程。

### 0.2 是否需要改底层？

**需要，但不需要推翻现有 Canon / Shot Manifest / Generation Runtime。**

当前已经存在完整骨架：

`Document Resource → Canon Core / Variant → Shot Manifest → Continuity → Deterministic Compiler → Generation`

这正是 Prompt Director 应该依附的主链。

底层真正缺少的是：

1. Canon 的 `PromptText` 目前只是一段字符串；
2. `SpecJSON` 目前主要保存 `summary/sourceText`，没有导演级 typed schema；
3. Shot 的 `ActionText/CameraText` 仍是自由文本；
4. Compiler 目前是按固定顺序拼接文本，不是结构化语义编译；
5. 只有一个 `CompiledPrompt`，没有 image / video / comfyui 目标适配；
6. 缺少 Prompt section trace / warnings，UI 无法解释“为什么生成成这样”；
7. 现有 LLM Prompt Optimize 会改写文字，它不应成为 Canon / Shot 的权威编译器。

### 0.3 是否需要新增 `.go` 文件？

**建议新增，而且应该新增一个独立的纯逻辑包，但不是重做后端。**

推荐新增：

```text
services/server/internal/service/promptdirector/
  schema.go
  normalize.go
  compiler.go
  targets.go
  compiler_test.go
  targets_test.go
```

这个包不负责数据库、不负责 Provider 请求，只负责：

- 解析结构化 Prompt Profile；
- 规范化字段；
- 按确定性顺序生成 Prompt Sections；
- 输出 image / video / comfyui 等目标表达；
- 输出 negative constraints / warnings / trace。

现有 `shotmanifest/compiler.go` 仍然负责：

- 找 Canon Core / Variant；
- 找 Reference；
- 找 Previous Shot；
- 解析 Continuity；
- 组装当前 Shot 的权威上下文。

然后调用 `promptdirector` 做最后的“导演式编译”。

### 0.4 是否需要新增数据库表？

**MVP 不建议新增表。**

建议给现有两张表新增字段：

```text
canon_assets.prompt_profile_json
shot_manifests.director_json
```

原因：

- Canon、Variant、Shot 本来就已经是正确的权威对象；
- Prompt Director 是这些对象的结构化视觉描述，不是另一套独立资产；
- 使用现有 GORM AutoMigrate 即可向旧数据库增加字段；
- 老项目字段为空时自动回退现有 `PromptText + ActionText + CameraText` 编译链。

不建议把新数据直接塞进现有 `SpecJSON`，因为 `SpecJSON` 当前由文档同步自动更新。如果 UI 编辑的数据和 source-sync 数据混在同一个 JSON 中，未来文档变化容易覆盖用户已经人工确认的导演信息。

---

# 1. 当前 MediaGo 已有能力审查

## 1.1 Canon 已经具备正确的资产权威层

当前：

`services/server/internal/domain/canon_models.go`

已有：

- `CanonAssetModel`
- Core / ParentID / VariantKind
- `SpecJSON`
- `PromptText`
- `Status`
- `Version`
- `SourceHash`
- `CanonReferenceModel`
- Reference Role / Priority / Locked

说明 MediaGo 不需要重新创建 Character / Scene / Prop 数据模型。

## 1.2 Canon 已经区分 Core 与 Variant

当前：

`services/server/internal/service/canon/`

已经支持：

- character / scene / prop；
- source document → Canon Core；
- source H3 → Variant；
- locked Canon；
- selected asset → Canon Reference；
- Variant PromptText 独立于 Core PromptText。

这非常适合直接映射：

- Character Core = 身份；
- Character Variant = 穿搭 / 发型 / 年龄态 / 战损；
- Scene Core = 空间主体；
- Scene Variant = 时间 / 天气 / 灯光 / 区域态；
- Prop Core = 物件身份；
- Prop Variant = 使用状态 / 损伤状态 / 开启关闭状态。

## 1.3 Shot Manifest 已经是正确的镜头执行合同

当前：

`services/server/internal/domain/shot_models.go`

已有：

- Sequence；
- Start / End / Duration；
- InheritsFromID；
- ActionText；
- CameraText；
- AudioText；
- StyleProfileID；
- BindingsJSON；
- StateChangesJSON；
- ResolvedStateJSON；
- CompiledPrompt。

因此“镜头导演器”不应另建 Shot 数据，而是补充 `director_json`。

## 1.4 现有 Prompt Compiler 已经具备 Deterministic 基础

当前：

`services/server/internal/service/shotmanifest/compiler.go`

已经能确定性地处理：

- Character Core + Variant；
- Scene Core + Variant；
- Prop Core + Variant；
- Canon References；
- Previous Shot；
- Resolved State；
- Action；
- Camera；
- Audio；
- Style Profile。

这意味着 Prompt Director 应是 **compiler v2**，不是重新开始。

## 1.5 当前 UI 缺口明确

`ProjectOverview.tsx` 当前 Production Shot 卡片只显示：

- 镜头序号；
- 时间；
- ActionText；
- CameraText；
- “生成关键帧”。

一致性资产只显示计数和状态摘要。

这就是产品层最大的缺口：底层已经有权威结构，用户却看不到、改不了，也无法理解最终 Prompt。

---

# 2. 产品定位

产品名称暂定：

**Prompt Director / 提示词导演系统**

它不是“Prompt 美化器”，而是：

> 将角色、场景、道具变成可复用导演资产，并让每个 Production Shot 基于权威资产、连续性和当前镜头意图，确定性编译出目标模型可直接使用的 Prompt。

核心价值：

1. 角色一致性；
2. 造型一致性；
3. 场景一致性；
4. 道具一致性；
5. 镜头语言更专业；
6. Prompt 可解释、可预览、可复现；
7. 同一个镜头可以自动输出不同模型需要的 Prompt；
8. 不要求用户每个镜头手写一大段重复描述。

---

# 3. 与现有 Prompt Optimize 的边界

当前 MediaGo 已有：

`generation_runtime_prompt_optimize.go`

其职责是：

> 用 LLM 将用户的一段自由 Prompt 改写成更好的自由 Prompt。

Prompt Director 与它必须严格分工。

## Prompt Optimize

适用：

- 临时自由生图；
- 非 ShotManifest 请求；
- 用户自己输入的一段 Prompt；
- 没有 Canon 权威约束的探索生成。

特点：

- LLM 改写；
- 不保证相同输入完全相同输出；
- 可以优化措辞；
- 不作为资产真相来源。

## Prompt Director

适用：

- Production Shot；
- 正式项目镜头；
- 角色 / 场景 / 道具已经建立 Canon；
- 有连续性要求的图片 / 视频生成。

特点：

- Canon / Variant / Shot 为权威；
- 编译必须 deterministic；
- LLM 只能“建议结构化字段”，不能直接覆盖权威事实；
- 同一输入应得到稳定、可测试的输出。

### 强制规则

当请求带 `shotManifestId` 时：

**默认关闭现有自由 Prompt Optimize。**

如果以后允许 AI 优化，只允许在 Prompt Director 的 section 内做受约束润色，不能删除 Canon Facts、不能改绑定资产、不能改 State Change。

---

# 4. 产品信息架构

## 4.1 角色导演卡 Character Director Card

### Core：身份层

字段建议：

- 年龄段；
- 性别表现；
- 体型；
- 身高体感；
- 脸型；
- 五官识别点；
- 肤色 / 肤质；
- 发色 / 基础发型；
- 固定伤痕 / 痣 / 纹身；
- 气质；
- 身份标签；
- 必须保持项；
- 禁止变化项。

### Variant：造型层

字段建议：

- 造型名称；
- 服装；
- 服装材质；
- 配色；
- 发型变化；
- 妆容；
- 配件；
- 鞋；
- 脏污 / 湿润 / 战损；
- Variant 必须保持项；
- Variant 禁止项。

原则：

**身份不放进 Variant，服装不污染 Core。**

---

## 4.2 场景导演卡 Scene Director Card

Core 字段建议：

- 空间类型；
- 功能；
- 尺度感；
- 空间结构；
- 入口 / 出口；
- 固定建筑特征；
- 材质；
- 主色；
- 固定陈设；
- 前景 / 中景 / 后景层次；
- 必须保持的 landmark；
- 禁止出现内容。

Variant 字段建议：

- 时间；
- 天气；
- 季节；
- 光源；
- 光线方向；
- 色温；
- 雾 / 雨 / 烟 / 灰尘；
- 人流状态；
- 特定区域 Zone；
- 临时陈设；
- Variant 禁止项。

---

## 4.3 道具导演卡 Prop Director Card

Core 字段建议：

- 类别；
- 几何形状；
- 尺寸感；
- 材质；
- 颜色；
- 表面处理；
- 品牌 / 标签 / 编号；
- 固定磨损；
- 独特识别点；
- 必须保持项；
- 禁止替换项。

Variant / State 字段建议：

- 开 / 关；
- 新 / 旧；
- 完整 / 损坏；
- 干 / 湿；
- 是否装有内容物；
- 当前归属角色；
- 当前持握方式；
- 当前摆放位置；
- 临时变化。

---

## 4.4 镜头导演卡 Shot Director

镜头导演卡不重新保存角色 / 场景 / 道具事实，只保存“这一镜怎么拍”。

字段建议：

### Subject Intent

- 主体优先级；
- 人物当前动作；
- 姿势；
- 手势；
- 视线；
- 人物之间关系；
- 与道具交互。

### Composition

- 景别；
- 机位高度；
- 摄影机角度；
- 镜头方向；
- 构图方式；
- 主体画面占比；
- 前景遮挡；
- 景深；
- 焦点位置；
- 镜头焦段语义。

### Lighting / Mood

- 主光源；
- 辅助光；
- 光线方向；
- 色温；
- 明暗反差；
- 情绪；
- 空气感；
- 动态氛围。

### Continuity

只显示并允许显式覆盖当前 Shot 的：

- 当前 Look；
- Scene Variant；
- Prop State；
- Previous Shot；
- State Changes。

### Output Intent

- 图片 / 视频；
- 画幅；
- 关键帧类型；
- 静态动作瞬间 / 动态动作过程；
- 是否需要首尾帧连续。

---

# 5. 推荐数据设计

## 5.1 Canon：新增 `prompt_profile_json`

在 `CanonAssetModel` 增加：

```go
PromptProfileJSON string `gorm:"column:prompt_profile_json;not null;type:text;default:'{}'"`
```

建议 JSON：

```json
{
  "schemaVersion": "prompt-director.canon.v1",
  "resourceType": "character",
  "identity": {},
  "appearance": {},
  "spatial": {},
  "material": {},
  "constraints": {
    "mustKeep": [],
    "avoid": []
  },
  "promptHints": {
    "master": "",
    "shot": "",
    "motion": ""
  }
}
```

不同 ResourceType 只使用对应字段。

### 为什么不直接改 `SpecJSON`

因为现有 `SpecJSON` 是从文档自动同步得到的：

```text
summary
sourceText
```

如果把人工导演数据塞进去，文档重新 Sync 时有覆盖风险。

`prompt_profile_json` 单独存在，可以做到：

- 原始文档继续自动同步；
- Prompt Director 的人工确认结果不被覆盖；
- 老项目为空时正常 fallback。

---

## 5.2 Shot：新增 `director_json`

在 `ShotManifestModel` 增加：

```go
DirectorJSON string `gorm:"column:director_json;not null;type:text;default:'{}'"`
```

示例：

```json
{
  "schemaVersion": "prompt-director.shot.v1",
  "subject": {
    "primary": ["canon-character-a"],
    "gaze": "看向右前方",
    "pose": "半蹲",
    "interaction": "双手握住擀面杖"
  },
  "composition": {
    "shotSize": "medium_close_up",
    "cameraAngle": "over_shoulder",
    "cameraHeight": "chest",
    "lensIntent": "50mm-natural",
    "framing": "subject-left-third",
    "depthOfField": "shallow"
  },
  "lighting": {
    "key": "冷白顶灯",
    "accent": "红色状态灯",
    "contrast": "medium_high"
  },
  "mood": ["警惕", "压迫"],
  "constraints": {
    "mustKeep": [],
    "avoid": []
  }
}
```

### 重要规则

`director_json` 不允许重复保存：

- Character identity；
- Character Look 的完整事实；
- Scene Master 的完整事实；
- Prop Master 的完整事实。

这些必须从 Canon 读取。

否则未来会重新出现“每个镜头各写一套角色”的一致性问题。

---

# 6. Prompt Compiler v2

## 6.1 两阶段编译

推荐从当前“一次拼字符串”升级成两阶段。

### Stage A：Canonical Prompt Graph

先编译成模型无关结构：

```go
type PromptSection struct {
    Key      string
    Label    string
    Text     string
    Priority int
    Source   string
}

type CanonicalPrompt struct {
    Sections  []PromptSection
    Negatives []string
    Warnings  []CompileWarning
}
```

建议顺序：

1. Output / media intent；
2. Character identity；
3. Active Character Variant；
4. Current action / pose / gaze；
5. Composition / camera；
6. Scene Core；
7. Scene Variant；
8. Prop Core / State / interaction；
9. Lighting / mood；
10. Continuity state；
11. Style Profile；
12. Constraints / negative。

### Stage B：Target Adapter

根据实际生成目标转换：

- `image.generic`
- `image.gpt-image`
- `image.flux-comfyui`
- `video.generic`
- `video.seedance-like`

第一阶段不需要为每个 Provider 建完全不同的数据库，只需要 Target Adapter。

---

# 7. Target Adapter 规则

## 7.1 GPT Image / Image 2 类

特点：

- 自然语言理解好；
- 可以使用完整句；
- 适合把人物、动作、镜头、场景、光线组织成连贯视觉描述；
- Negative 不一定作为独立参数传递，需要根据 route capability 决定折入正文。

输出偏“导演描述”。

## 7.2 Flux / ComfyUI 类

特点：

- 需要更短、更标签化；
- 应减少解释性句子；
- 主体、服装、构图、光线按优先级压缩；
- 如果工作流支持 negative prompt，则单独输出。

输出偏“视觉标签 + 权重语义”。

## 7.3 视频模型

额外强调：

- 起始状态；
- 动作过程；
- 结束状态；
- 摄影机运动；
- 主体轨迹；
- 不允许瞬移 / 换装 / 无依据道具增减；
- Previous Shot continuity。

图片 Prompt 与视频 Prompt 不能只是同一段文字。

---

# 8. Reference Policy 保持现有主链

不重做现有 Reference 机制。

继续使用当前优先级：

1. Character Identity；
2. Active Look；
3. Previous Shot；
4. Scene Variant；
5. Scene Master / Prop 按需要。

Prompt Director 只负责告诉 Compiler “哪些语义必须稳定”，不负责绕过现有 Provider Reference Resolver。

这样能避免：

- 同一个参考图被重复送入；
- Prompt Director 自己再造一套 reference slot；
- ComfyUI / GPT Image / 视频 Provider 接口分叉失控。

---

# 9. UI 产品方案

## 9.1 ProjectOverview：一致性资产卡升级

现有“一致性资产”区域保留。

新增：

- `打开导演资产`；
- Character / Scene / Prop 三类数量；
- 已完成 Prompt Profile 数；
- 缺少视觉锚点数量；
- 有冲突数量。

点击后打开：

**Prompt Director Asset Dialog**

不新增一级导航。

---

## 9.2 Prompt Director Asset Dialog

左侧：

- 角色；
- 场景；
- 道具；
- Core；
- Variant。

中间：结构化字段表单。

右侧：

- 当前参考图；
- Master Prompt 预览；
- Shot Prompt 预览；
- Motion Prompt 预览；
- Must Keep；
- Avoid；
- 数据来源：Source / AI Suggestion / User Confirmed。

操作：

- AI 完善；
- 恢复来源文本；
- 保存；
- 锁定；
- 添加 Variant；
- 绑定参考图。

---

## 9.3 Production Shot 卡升级

现有卡片不应该塞满字段。

增加两个按钮：

- `镜头详情`
- `生成关键帧`

“镜头详情”打开 Shot Director Dialog。

---

## 9.4 Shot Director Dialog

建议 Tabs：

### 镜头理解

- Action；
- Subject；
- Composition；
- Camera；
- Lighting；
- Mood。

### 资产绑定

- Character Core / Look；
- Scene Core / Variant；
- Props；
- Previous Shot；
- State Changes。

### Prompt 预览

目标切换：

- 图片；
- 视频；
- ComfyUI。

展示：

- 各 Prompt Section；
- 最终 Prompt；
- Negative；
- Reference 列表；
- Warning；
- 哪个字段来自哪个 Canon / Variant / Shot。

### 连续性

展示与上一镜的差异：

- Look changed；
- Scene changed；
- Prop state changed；
- inherited；
- explicit override。

---

# 10. API 方案

## 10.1 Canon Prompt Profile

新增：

```text
PUT /api/v1/projects/{projectId}/canon/{canonId}/prompt-profile
```

Body：

```json
{
  "schemaVersion": "prompt-director.canon.v1",
  "profile": {}
}
```

返回完整 Canon Record。

可选新增：

```text
POST /api/v1/projects/{projectId}/canon/{canonId}/prompt-profile/suggest
```

职责：基于 Source Prompt / Spec / 当前 Variant 返回 AI 建议，不自动落库。

必须用户确认后再 PUT。

## 10.2 Shot Director

新增：

```text
PUT /api/v1/projects/{projectId}/shot-manifests/{shotId}/director
```

只修改 `director_json`，不能改 Canon facts。

## 10.3 Compile Preview

扩展现有：

```text
POST /api/v1/projects/{projectId}/shot-manifests/{shotId}/compile
```

支持 query/body：

```text
target=image|video|comfyui
profile=gpt-image|generic-video|flux
persist=true|false
```

返回：

```json
{
  "prompt": "...",
  "negativePrompt": "...",
  "sections": [],
  "references": [],
  "warnings": [],
  "target": "image",
  "profile": "gpt-image"
}
```

UI Preview 使用 `persist=false`。

真正 Generate 时由 Generation Runtime 使用对应 Target 再编译一次，避免用户看到的预览与真实生成错位。

---

# 11. Go 文件级改动评估

## 11.1 必须修改现有 `.go`

### `services/server/internal/domain/canon_models.go`

增加：

- `PromptProfileJSON`

风险：低。

### `services/server/internal/domain/shot_models.go`

增加：

- `DirectorJSON`

风险：低。

### `services/server/internal/repository/canon_repo.go`

增加：

- `UpdatePromptProfile`

不要直接允许 UI 任意 map 更新所有 Canon 字段。

风险：低。

### `services/server/internal/repository/shot_manifest_repo.go`

增加：

- `UpdateDirectorJSON`

注意：现有 source-sync `Upsert` 不应覆盖 `director_json`。

风险：中。

### `services/server/internal/service/canon/service.go`

增加：

- Prompt Profile validation；
- update method；
- API projection 字段。

风险：中。

### `services/server/internal/service/shotmanifest/service.go`

增加：

- Director profile projection；
- update / validate；
- 保持旧 Shot sync 不覆盖人工 Director data。

风险：中。

### `services/server/internal/service/shotmanifest/compiler.go`

职责由“字符串拼接器”升级为：

- 权威数据 resolver；
- 调用 `promptdirector`；
- 输出 sections / negative / warning / references。

现有 fallback 行为必须保留。

风险：高，是本轮核心。

### `services/server/internal/service/generation/generation_shot_manifest.go`

根据实际 generation route / kind 告诉 compiler：

- image；
- video；
- target profile。

风险：中高。

### `services/server/internal/service/generation/generation_runtime.go`

扩展 `shotManifestCompiler` interface。

风险：中。

### `services/server/internal/http/handlers/canon.go`

新增 profile update / suggest handler。

风险：低。

### `services/server/internal/http/handlers/shot_manifests.go`

新增 Director Update；
Compile 支持 target/profile/preview。

风险：中。

### `services/server/internal/http/routes/routes.go`

注册新 route。

风险：低。

### `services/server/internal/app/wire.go`

如果 Prompt Director 只做纯函数编译，改动很小；
如果 P2 加 AI Suggestion Service，则需要注入 TextCompletion。

风险：低 / 中。

## 11.2 推荐新增 `.go`

```text
services/server/internal/service/promptdirector/schema.go
```

定义：

- CanonPromptProfileV1；
- CharacterProfile；
- SceneProfile；
- PropProfile；
- ShotDirectorV1；
- PromptSection；
- CanonicalPrompt；
- TargetPrompt；
- CompileWarning。

```text
services/server/internal/service/promptdirector/normalize.go
```

负责：

- trim；
- empty compact；
- enum validation；
- schema version；
- deterministic ordering。

```text
services/server/internal/service/promptdirector/compiler.go
```

负责：

- Canon Profile + Shot Director → CanonicalPrompt。

```text
services/server/internal/service/promptdirector/targets.go
```

负责：

- image generic；
- GPT Image；
- video generic；
- Flux / ComfyUI。

测试：

```text
compiler_test.go
targets_test.go
```

### P2 可新增

```text
services/server/internal/service/promptdirector/suggest.go
```

只负责 AI Suggestion，不直接写数据库。

---

# 12. 前端文件级改动评估

## 12.1 修改

```text
apps/workspace/src/pages/ProjectOverview.tsx
```

- 一致性资产入口；
- Production Shot “镜头详情”；
- Prompt Director dialogs wiring。

不要继续把所有表单直接写在这个 2000+ 行文件中。

## 12.2 新增 API 文件

推荐：

```text
apps/workspace/src/domains/workspace/api/prompt-director.ts
```

不要把全部新 API 塞进当前 `continuity.ts`。

## 12.3 新增 UI 目录

推荐：

```text
apps/workspace/src/domains/workspace/components/prompt-director/
  PromptDirectorAssetDialog.tsx
  PromptDirectorShotDialog.tsx
  CanonProfileEditor.tsx
  ShotIntentEditor.tsx
  PromptCompilePreview.tsx
  ReferencePolicyPanel.tsx
  ContinuityDiffPanel.tsx
```

## 12.4 新增纯前端 schema/helper

```text
apps/workspace/src/domains/workspace/lib/prompt-director.ts
```

负责：

- TS interfaces；
- empty v1 profile；
- UI field labels；
- schema parsing；
- fallback display。

---

# 13. AI 自动生成 / 完善机制

为了达到类似 AI 追光的使用体验，最终必须有 AI 辅助，但 AI 不能直接成为 Canon 真相。

流程：

`Source Document / PromptText → AI Suggestion → Structured Diff → 用户确认 → PromptProfileJSON`

### 角色

AI 从角色正文提取：

- 稳定身份；
- 可变造型；
- must keep；
- avoid。

如果正文把“婚纱”混在角色身份里，Suggestion 应建议拆为 Look Variant，而不是直接写入 Core。

### 场景

AI 提取：

- 空间；
- 固定 landmark；
- 材质；
- 时间 / 灯光 Variant；
- Zone。

### 道具

AI 提取：

- 物件身份；
- 识别细节；
- 状态变化。

### Shot

AI 只建议：

- 景别；
- 机位；
- 构图；
- 光线；
- 情绪；
- 动作展开。

不允许 AI：

- 创建剧本没有的新角色；
- 改已有 Canon identity；
- 擅自换装；
- 擅自增加关键道具；
- 擅自改剧情顺序。

---

# 14. 开发阶段与任务拆解

## P0：合同冻结 / 无 UI

目标：先冻结 schema 和 compiler contract。

任务：

1. `promptdirector` schema v1；
2. Canon PromptProfileJSON；
3. Shot DirectorJSON；
4. fallback compatibility tests；
5. JSON normalize / validation。

验收：

- 旧 app.db 自动迁移；
- 空 profile 与当前版本生成结果等价；
- source sync 不覆盖人工 profile/director。

---

## P1：Compiler v2

目标：不依赖新 UI，先让底层真正成立。

任务：

1. Canon profile parser；
2. Shot director parser；
3. CanonicalPrompt sections；
4. image target；
5. video target；
6. comfyui target；
7. warnings / negative；
8. legacy fallback；
9. generation integration。

验收：

- 同一数据重复编译字节级稳定；
- Core / Variant 不串；
- previous shot 正确继承；
- image/video 输出不同；
- 没有 profile 的旧项目仍走旧 prompt 语义。

---

## P2：Asset Director UI

目标：角色 / 场景 / 道具可视化维护。

任务：

1. Asset Dialog；
2. Character profile editor；
3. Scene profile editor；
4. Prop profile editor；
5. Variant editor；
6. Reference panel；
7. Master / Shot / Motion preview；
8. Save / Lock。

验收：

- 用户不用看 JSON；
- 编辑 Canon Core 不会污染 Variant；
- 编辑 Variant 不改 Core；
- Reference 可追溯。

---

## P3：Shot Director UI

目标：把 Production Shot 从“动作文本卡片”升级成可编辑镜头导演卡。

任务：

1. Shot Detail 入口；
2. Subject / Composition / Lighting / Mood；
3. Canon bindings；
4. continuity diff；
5. Prompt section preview；
6. target selector；
7. generate from preview。

验收：

- UI Preview 与真实 Generation 使用同一 compiler；
- 用户能知道 Prompt 每一段来源；
- 修改镜头语言不修改 Canon。

---

## P4：AI Suggestion

目标：达到“AI 追光式”的低门槛体验。

任务：

1. Canon AI 完善；
2. Variant 拆分建议；
3. Shot 镜头语言建议；
4. diff review；
5. accept/reject；
6. source provenance。

验收：

- AI 永远先建议、不自动覆盖 locked fields；
- 不得引入未绑定主体；
- Suggestion 失败不影响 deterministic compiler。

---

## P5：真实生成验收

用已有真实项目与 Golden Project 做：

1. 角色母图；
2. 默认 Look；
3. 造型 Variant；
4. Scene Master；
5. Scene Variant；
6. Prop Master；
7. 连续 6 镜；
8. Image target；
9. Video target；
10. ComfyUI target（若已配置）。

要求：

- identity 不漂移；
- 显式换装才换装；
- scene variant 正确继承；
- prop 不无依据复制；
- Previous Shot 只影响短期连续性；
- 生成结果不能反写 Canon。

---

# 15. 开发依赖顺序

必须按下面顺序：

```text
P0 Schema
  ↓
P1 Compiler v2
  ↓
P2 Asset Director UI
  ↓
P3 Shot Director UI
  ↓
P4 AI Suggestion
  ↓
P5 Golden / Legacy / UI Acceptance
```

不要先做漂亮 UI 再补 compiler。

否则 UI 会围绕不稳定的数据合同反复返工。

---

# 16. 不建议的方案

## 16.1 不建议再建 CharacterPrompt / ScenePrompt / PropPrompt 三套表

原因：

- 与 Canon 重复；
- Variant 关系需要再同步一遍；
- Reference 再重复一遍；
- 最终一定发生双真相源。

## 16.2 不建议直接让 LLM 每镜重写完整 Prompt

原因：

- identity 漂移；
- token 重复；
- Prompt 不可复现；
- 同一镜不同时间输出不同事实；
- 无法做自动测试。

## 16.3 不建议把编译后的 Prompt 当可长期手工编辑的主数据

`CompiledPrompt` 是结果，不是真相。

用户要修改：

- 角色 → 改 Canon；
- 造型 → 改 Variant；
- 镜头 → 改 Shot Director；
- 风格 → 改 Style Profile。

然后重新 Compile。

## 16.4 不建议第一阶段新增一级“Prompt 导演”导航

先融入现有项目流。

等功能稳定、资产量变大后，再评估是否提升为一级 Studio 页面。

---

# 17. 风险清单

## 高风险：Compiler 与现有生成链不一致

解决：

- Preview 和 Generate 必须调用同一 Go compiler；
- 前端禁止自行拼 Prompt。

## 高风险：用户编辑数据被 source sync 覆盖

解决：

- 单独 `prompt_profile_json` / `director_json`；
- source sync 不写这两个字段。

## 中风险：Prompt 过长

解决：

- Sections 有 Priority；
- target adapter 做预算；
- locked facts / identity 优先级最高；
- 通用质量词最低。

## 中风险：不同 Provider 参考图数量限制不同

解决：

- Prompt Director 不直接分配 Provider slot；
- 继续交给现有 Reference Resolver / route capability。

## 中风险：AI Suggestion 幻觉

解决：

- suggestion ≠ truth；
- 必须 diff；
- locked field 不覆盖；
- 只能引用现有 Canon / Script facts。

---

# 18. 最终实施建议

负责人推荐：

### 架构判断

**保留现有 MediaGo 主架构，做一次 Prompt Compiler v2 增量升级。**

### 数据库判断

**MVP 不新建表；现有 Canon / Shot 各加一个 JSON 字段。**

### Go 判断

**需要新增 `promptdirector` 包；需要修改现有 Canon、ShotManifest、Generation 集成代码。**

### UI 判断

**必须新增导演资产和镜头详情 UI，但先用 Dialog / Drawer，不重做主导航。**

### AI 判断

**AI 负责“建议结构化 Prompt 数据”，代码负责“确定性编译”，两者不能混为一层。**

### 兼容判断

**旧项目、旧 PromptText、旧 ShotManifest 必须继续工作；新字段为空就完全 fallback。**

最终目标链路：

```text
剧本 / 设定文档
   ↓
Canon Core / Variant
   ↓
Prompt Director Profile
   ↓
Production Shot + Director Intent
   ↓
Continuity Resolver
   ↓
Canonical Prompt Graph
   ↓
Target Adapter
   ├─ GPT Image
   ├─ Generic Image
   ├─ ComfyUI / Flux
   └─ Video
   ↓
现有 Generation Runtime / Reference Resolver / Provider
```

这条路线能最大程度复用当前 MediaGo 已经完成的资产一致性升级，同时把目前“底层能力存在、前端不可控、Prompt 只是拼接文本”的状态升级成真正可用于漫剧规模化生产的导演级提示词系统。
