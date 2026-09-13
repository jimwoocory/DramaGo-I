# MediaGo Desktop Release Candidate 验收报告

日期：2026-09-11

项目：`D:\openai\MediaGo-Drama`

负责人任务：`tsk_2b7bce639015ac58`

## 1. 负责人结论

**Release Candidate：PASS（可进入用户/内部最终人工验收，不代表已授权正式发布）。**

当前资产一致性升级、旧项目兼容、Windows RC 构建、Codex Image、ACP Agent、重启恢复、取消恢复、普通 Windows 网络链路和全量自动回归均已通过。

本轮仍保持 **未 commit、未 push、未 publish**。未经用户最终确认，不得推送远端正式分支或发布安装包。

## 2. 基线与候选版本

- Branch：`mediago-clean-baseline-20260903`
- Base HEAD：`78739afa269e4f634f4431b98630add38bcdb184`
- Base subject：`fix(agent): harden project workflow reliability`
- Desktop package version：`0.1.0-beta.0`
- Bundled Codex CLI：`0.151.0`
- Codex ACP：`@agentclientprotocol/codex-acp@1.1.2`
- 图片生成验证 Provider：Codex `image_generation`
- 本轮不再使用即梦作为黄金视觉验收图片 Provider。

## 3. 核心功能验收

### 3.1 Canon / Variant / Shot / Compiler

已验证以下结构化主链：

`Script → Canon Core / Variant → Shot Manifest → Continuity Resolver → Prompt Compiler → Reference Policy → Generation`

关键规则：

- Character Core 保持身份权威，不被 Wedding / Work 等造型 Variant 污染。
- H3 Character Look、Scene Variant / Zone、Prop State Variant 可独立物化。
- Shot 未显式换装时继承上一镜 Look；显式 Work Variant 时替换 Wedding Variant。
- Scene Variant 可连续继承。
- Prop 默认连续继承，显式 `stateChanges` 清除时停止继承。
- Prompt Compiler 按 Canon / Shot 确定性编译，不再让模型逐镜重新猜固定资产。

### 3.2 Reference Policy

当前实际策略已经收敛为：

`Character Identity → Active Look → Previous Shot → Scene Variant → Scene Master / Prop（按需要）`

当 Scene Variant 已有视觉锚点时，不重复塞 Scene Master，减少冗余条件冲突。

上一镜只用于短期连续性，不写回 Canon，不替代长期身份/造型/场景权威。

### 3.3 Codex-only Golden Visual

新隔离 Golden 项目已经使用 Codex Image 完整生成：

- Identity Anchor
- Wedding Look Anchor
- Work Look Anchor
- Scene Master
- Night Scene Anchor
- Shot01–Shot06

Shot02–05 连续使用 Identity + Wedding + Previous Shot + Night；Shot06 显式切换 Work。

独立视觉 QA 对身份、Wedding、Work、Scene、Prop 判定均为 PASS，`release_blockers=[]`。

因此原先 Shot04/05 在单 Identity Reference 下暴露的一致性风险，已经通过多层视觉锚点 + Previous Shot continuity 解决。

## 4. 旧项目兼容

真实旧项目：`project-679e2eab7f3a12cc`

验收使用完整隔离副本，不直接修改原项目。

结果：

- 旧 DB 初始无 `canon_assets / canon_references / shot_manifests`。
- 新版启动后 schema migration 正常创建新表。
- Lazy sync 首次：`Created=16`，最终 Canon=39、Shot=16。
- 第二次 sync：`Created=0 / Reused=16`，幂等性通过。
- 18 个 Markdown 迁移前后 SHA-256 **0 变化**。
- 原有 project / assets / project_selected_assets / generation_tasks 原字段未被迁移流程篡改。
- Legacy generation request 在不传 `shotManifestId` 的情况下仍可用。

### 4.1 迁移期间发现并修复的路径缺陷

旧 `projects.project_dir` 会保存原安装目录绝对路径。复制工作区后，如果继续信任旧绝对路径，新生成素材可能写回旧目录。

已增加 workspace-relative rebase：

- `relative_dir` 是当前 workspace 内合法相对路径，且目标目录存在 → 启动时自动重绑 `project_dir` 到当前 workspace。
- 外部绝对项目路径不强制改写。

修复后真实 Legacy Codex Image：

- 57.9 秒完成
- 0 重试
- 新 PNG 只写入隔离副本
- 原项目目录没有新增文件
- Markdown hash 继续保持 0 变化

## 5. Windows Release Candidate

候选产物：

### NSIS Installer

`apps\workspace\release\jw-drama-0.1.0-beta.0-win-x64.exe`

- Size：471,983,428 bytes
- SHA-256：`AF76F253D9AC5FBE6978352595F3A25EB50ED8BD1B853EA2E1383A07A940E297`

### ZIP

`apps\workspace\release\jw-drama-0.1.0-beta.0-win-x64.zip`

- Size：626,604,223 bytes
- SHA-256：`1BE449B9A4727EC6E92912491C6064EF74BCDEA79A9069AB13499FA0A70E81E1`

### Blockmap

`apps\workspace\release\jw-drama-0.1.0-beta.0-win-x64.exe.blockmap`

- Size：494,677 bytes
- SHA-256：`AEB565ACB944014BF845BC9714707801BBAA585ADA05D91BC45DA77EB046F5BA`

### 包内运行时验证

最终 `win-unpacked` 中已经实际确认：

- Codex CLI = `0.151.0`
- `mediago-server.exe` 存在
- `mediago-document-mcp.exe` 存在
- `mediago-generation-mcp.exe` 存在
- `local-cli.json` 中 `generationClis=[]`，默认不启用 Jimeng 图片 CLI

## 6. 包内 Sidecar / ACP 实测

### Sidecar

最终 RC 包内 `mediago-server.exe`：

- 正确 sidecar token 调 `/api/v1/health` → HTTP 200
- 不带 sidecar token → HTTP 401
- `codex.image-generation` → `available + configured=true`
- Active Agent backend → `codex`

### ACP Agent

真实包内链路已经跑通：

`mediago-server.exe → codex-acp.exe → cmd.exe → packaged codex.exe 0.151 app-server`

最小 Agent 请求：

`只回复 OK，不调用任何工具。`

最终：

- Agent session = `completed`
- chat final answer = `OK`

说明包内 ACP 和 Codex 不是仅“文件存在”，而是已经完成真实 Agent round-trip。

注意：直接单独运行 `codex-acp.exe` 且不提供 `CODEX_PATH` 时，第三方 adapter 会尝试解析 `@openai/codex/bin/codex.js` 并失败；MediaGo 的真实 Runner 会按 `agent.json.codexBin` 注入 `CODEX_PATH`，该真实路径已经验证 PASS。因此这不是 MediaGo 实际运行阻塞，但 `codex-acp.exe` 不能脱离 MediaGo 环境独立启动。

## 7. 稳定性与恢复

### 7.1 服务重启恢复

同一隔离 workspace：

1. 首次 Agent 返回 `FIRST`，session completed，chat 持久化。
2. 强制停止 packaged sidecar。
3. 使用同一 workspace 重启 server。
4. 原 `FIRST` chat/status 可读取。
5. 同一 MediaGo session 再发 `SECOND`。

日志明确出现：

`acp session recap injected`

说明系统使用持久历史 recap 建立新 ACP session，而不是依赖已经死亡的旧进程。

最终聊天顺序保持 FIRST → SECOND。

### 7.2 运行中取消 / 再恢复

- 运行中的 Agent 调 cancel 后立即收口为 `cancelled / Agent 运行已中断`。
- 同一 MediaGo session 再提交任务后恢复为 completed。
- 最终返回 `RECOVERED`。
- 无永久 running、无死锁。

### 7.3 Image orphan recovery

Jimeng 图片链测试阶段发现 provider 调用可能留下无 providerTaskID 的 `running` 孤儿任务。

已增加：

- Jimeng Image 240 秒硬超时
- stale running orphan 在服务启动/worker 恢复时转成 `failed / timeout / retryable=true`

真实 Golden DB 重启验证通过。

此逻辑保留为兼容性安全网；黄金视觉图片已迁移到 Codex Image。

## 8. 网络环境

### 已真实覆盖

当前 Windows 主机：

- Realtek 有线为默认 IPv4 路由
- WinHTTP Proxy：`192.168.8.1:7893`
- bypass 包含 `localhost / 127.0.0.1`
- System Internet ProxyEnable=0

RC packaged sidecar：

- `127.0.0.1` health = 200
- `localhost` health = 200
- Node fetch：127 首次约 30ms，后续 1–2ms；localhost 首次约 21ms，后续 1–3ms

普通 Windows + 当前局域网代理环境：PASS。

### 未真实覆盖

当前机器没有激活：

- Wintun / Clash / Mihomo / WireGuard / OpenVPN / Tailscale 等实际 TUN 进程
- 360 主动防护（`Q360AMPPL` 当前 Stopped）

因此不得把“活跃 VPN/TUN”和“360 主动防护开启”标成已通过。这两项保留为现场/手工矩阵。

## 9. 最终自动回归

基于当前最终代码树重新执行：

- `services/server go test ./... -count=1`：PASS
- `packages/vendor` prepare-agent / prepare-rights / prepare-tool：PASS
- `packages/instructions` 全包：PASS
- Frontend Vitest：**226/226 Test Files PASS**
- Frontend Tests：**1467/1467 PASS**
- `project-679...` productionqa：PASS
- `pnpm -C apps/workspace build`：PASS
- `git diff --check`：PASS

Production build 仅有既有 chunk-size warning，无编译错误。

## 10. 已知限制 / 非阻塞项

1. 活跃 VPN/TUN 实机矩阵未在当前机器覆盖。
2. 360 主动防护开启状态未实测。
3. 打包 GUI 已真实启动到启动登录页，但负责人未知合法启动密码，因此没有绕过认证；“合法登录 → Workspace”保留最终人工 UI gate。
4. `codex-acp.exe` 必须由 MediaGo Runner 注入 `CODEX_PATH` 后运行，不能当独立 CLI 使用。
5. Source H3 Variant 删除/重命名后的 stale source variant 自动 deprecate/reconcile 仍可进一步加强。
6. Production QA 当前是确定性规则型 QA，不等于完整语义审片 Agent。
7. Variant/Scene Reference 后端能力已成立，面向小白的“给某个造型/场景定稿图”UI 仍可继续增强。

## 11. 工作树发布卫生

当前升级仍位于未提交工作树。

特别发现：

`D:\openai\MediaGo-Drama\MediaGO-Drama\`

是一个独立、干净的嵌套 Git clone：

- branch：`main`
- HEAD：`78739afa269e4f634f4431b98630add38bcdb184`
- origin：`https://github.com/jimwoocory/MediaGO-Drama.git`
- 约 118 MB
- 外层仓库当前 **未 ignore** 该目录

负责人没有删除它，因为来源不能百分百确认。

在任何 commit 前必须：

- 明确排除该目录，或由用户确认后再清理；
- 不使用无差别 `git add -A`；
- 使用明确文件列表/路径进行 staged review；
- staged 后再次核 `git diff --cached --check`。

## 12. 回滚方案

### 代码层

当前 Base HEAD 仍是：

`78739afa269e4f634f4431b98630add38bcdb184`

因为本轮没有 commit / push，所以原 Git 基线没有被改写。

### 完整文件级备份

升级前备份：

`D:\openai\MediaGo-Backups\MediaGo-Drama-pre-asset-consistency-20260910-005356.zip`

此前已完成并记录 SHA-256：

`FBE6E06FAD92FA9DE2D89A3D26608572E6A996807FF9D6CE6B9452D89646BEA2`

需要完全回滚时，应优先使用该备份恢复到单独目录后核对，再替换当前工作目录；不要直接用破坏性 clean 命令清理当前工作树，因为当前存在未跟踪的新实现文件和嵌套 clone。

### Desktop 层

本轮 RC 未 publish。正式替换前仍可继续使用现有上一版 desktop；新 RC 只作为候选包存在于本地 release 目录。

## 13. 发布建议

负责人的最终建议：

**允许把当前版本定义为“内部/用户验收 RC”。**

在正式 push / publish 前，仍建议完成：

1. 使用合法启动账号执行一次 `登录页 → Workspace → sidecar ready → 新项目` 的人工 UI smoke。
2. 如果目标用户确实经常启用 Clash/Wintun/TUN 或 360 主动防护，再补对应两台/两种实机矩阵。
3. 清理或明确 ignore 根目录的 `MediaGO-Drama/` 嵌套 clone。
4. 由用户确认后再决定 commit / push / publish。

在这些最终人工项完成之前，本报告的 PASS 含义是 **RC-ready，不是 GA/public-release-ready**。

## 14. 人工 UI 验收补充（2026-09-11 晚间）

用户已在最终 Windows RC `win-unpacked` 中使用合法启动凭据完成登录，Electron 成功进入项目管理界面，`mediago-server.exe` 由 GUI 主进程自动拉起。

首次进入时 UI 显示空项目。排查确认并非数据丢失：

- 旧便携版完整 workspace 仍位于：`D:\openai\MediaGo-Builds\JW-Drama-Unified-20260905\desktop\win-unpacked\data\workspace`
- 旧 workspace 数据完整：3 个 active 项目、110 个 assets、55 个 project_selected_assets、110 个 generation_tasks。
- `project-679e2eab7f3a12cc` 仍有 18 个 Markdown 文档；项目级 55 个 assets、55 个 selected-asset 绑定、110 个 generation tasks。
- 旧 `settings.db` 仍存在并包含 3 条 API key、3 条 app_settings、2 条 generation_preferences、1 个 pack、19 个 pack_entries。

根因是新 RC 位于新的便携目录，默认 `dirname(process.execPath)\data\workspace` 因此是一个新的空 workspace；旧便携版的数据目录不会自动被新目录发现。RC 已具备 `workspace-location.json` / `setWorkspaceDirectory` 机制，可显式选择持久 workspace。

本次验收过程中手工写入 `workspace-location.json` 时一度带 UTF-8 BOM，RC 的 `readFileSync(..., "utf8") + JSON.parse` 会把 BOM 视为内容并解析失败，于是静默回退默认 workspace。该手工写入问题已通过改为 UTF-8 无 BOM 纠正；产品自身 `writeFileSync(..., "utf8")` 路径不会写 BOM，因此这不是产品代码缺陷。

纠正后重新登录，当前 RC sidecar 已真实打开旧 workspace：旧 `app.db` 与 `settings.db` 均出现当前进程的 WAL/SHM 活动，新空 workspace 停止活动。用户随后确认项目列表已经恢复显示。

负责人补充结论：

- **数据恢复 / 旧 workspace 读取：PASS**。
- **合法登录 → Workspace → sidecar 自动启动：PASS**。
- **便携版升级 UX：需要优化但当前不构成数据安全阻塞。** 当新版便携目录与旧版不同、且此前没有持久 workspace preference 时，用户会先看到空 workspace，容易误认为数据丢失。正式面向小白用户前，建议在“默认 workspace 为空”时提供显式的“选择旧项目数据目录 / 检测旧工作区”引导，而不是只显示“暂无项目”。

本补充不改变此前 RC-ready 结论，但将“便携版跨目录升级的数据目录发现体验”加入后续产品优化清单。

## 15. 真实旧项目 UI 深度验收与阻塞项修复（2026-09-11 深夜）

基于用户恢复后的旧 workspace，对 `project-679e2eab7f3a12cc` 继续执行真实 GUI 使用测试，而不是仅查询数据库。

已真实通过：

- 项目列表显示 3 个旧项目，并从 UI 打开 `测试3`（`project-679...`）。
- 文档模式打开《第02集 铁瓮入灶》，正文、标题、场景、台词和编辑工具栏正常渲染。
- 项目素材库切换到 `测试3` 后，旧分镜 PNG 缩略图与右侧预览正常。
- 生成历史 UI 可读取旧 Codex / ChatGPT 订阅 `image_generation` 结果、Prompt、预览、生成时间和耗时。
- API 密钥页能读取旧 OpenAI-compatible 配置且密钥保持掩码；旧 `deepseek-v4-pro-0813` 模型分配仍在。
- Codex 接入页显示 ChatGPT 官方订阅 / Codex 已登录 / 默认供应商。
- 在真实旧项目中将 Agent 从第三方 DeepSeek 切换到 Codex-GPT / ChatGPT OAuth / GPT-5.6-Sol，通过 UI 发送“只回复 UI_OK，不调用任何工具”，最终 UI 返回 `UI_OK`。
- 新建独立 `RC-SMOKE-IMAGE` 会话并通过 UI 调用 Codex `image_generation` 生成“纯白背景，中间一个黑色圆点，无文字。”，41 秒完成，UI 正常显示结果，输出文件真实落入旧 workspace `library/2026-09-11/`。

### 验收中发现的真实 release blocker

首次文档 UI 测试中，仅打开《第02集 铁瓮入灶》、未主动编辑，应用即自动产生 `workspace_save` 历史提交 `6999408`：

- frontmatter `version: 4 → 5`
- 文件尾部额外增加两个空行
- 正文语义未变，但文件从 6491 bytes 变成 6493 bytes，SHA-256 改变

根因定位为 `MarkdownHybridEditor` / TipTap 在解析旧 Markdown 后，其 `getMarkdown()` canonical 序列化与原始字符串存在尾部空行差异。该初始化 `onUpdate` 被误当成用户编辑，进入 `updateDocumentContent` / save queue，导致版本递增和历史污染。

修复方式：

- 在编辑器创建时捕获首次 canonical Markdown。
- 第一次 update 如果仍只是该初始化 canonical 结果，直接忽略，不触发 `onChange`。
- 一旦用户产生真实输入，仍按原逻辑进入保存队列，第一笔真实编辑不会被吞掉。

新增回归测试：

1. 旧 Markdown 只 render / 等待 / blur，不应触发 `onChange`。
2. 吞掉初始化 canonical 差异后，第一笔真实用户编辑必须正常触发一次 `onChange`。

定向测试 6/6 PASS；随后刷新 `win-unpacked` 做真实现场复测：打开《第02集》→ 不输入 → 等待 → 切到第03集。

最终完整性结果：

- 18 个 Markdown：**0 变化**
- 《第02集 铁瓮入灶》：6491 bytes / SHA-256 `41fe0172e6079cdd18625d25d4cf1cff07d6aec72f6839b6202875cc352c9d38`
- `document-history.git` HEAD：`04af0b8c6662a4be709666a1945433f154ab6009`，保持不变
- project-679：55 assets / 55 selected bindings / 110 generation tasks，保持不变

测试专用 `RC-SMOKE-IMAGE` 会话、task、asset、toolbox session dir、临时 `studio` generation preference 以及 `UI_OK` agent session 已按精确 ID 清理。清理后旧 workspace 恢复为 3 projects / 110 assets / 110 generation tasks / 2 generation conversations，settings generation_preferences 恢复为原有 `agent` 与 `agent:video` 两条。

修复后最终回归：

- Frontend Vitest：**226/226 Test Files PASS**
- Frontend Tests：**1467/1467 PASS**
- `pnpm build`：PASS
- `git diff --check`：PASS

最终 Windows RC 已在修复后重新执行 `pnpm electron:release:windows-x64 --publish never` 等价发布链并成功退出；最终 `app.asar` 包含修复后的 `renderer/assets/WritingWorkspace-DMZff6oO.js`，包内 Codex CLI 仍为 `0.151.0`。Section 5 的 EXE / ZIP / Blockmap SHA-256 已更新为这批最终包。

负责人结论：该真实 UI 阻塞项已关闭，旧项目真实 UI 验收达到 PASS。
