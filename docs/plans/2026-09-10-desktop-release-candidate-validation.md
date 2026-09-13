# MediaGo Desktop 发布候选验收矩阵

日期：2026-09-10
负责人任务：`tsk_2b7bce639015ac58`
仓库：`D:\openai\MediaGo-Drama`
基线分支：`mediago-clean-baseline-20260903`
基线 HEAD：`78739afa269e4f634f4431b98630add38bcdb184`
状态：未 commit / 未 push。

## 放行原则

本轮目标不是继续扩架构，而是验证“真实用户按原有操作是否能稳定使用新一致性链”。自动测试通过只是必要条件，不等于可发布。任何会导致身份漂移、造型无依据回退、旧项目数据破坏、安装后 sidecar/API 不可用的缺陷均视为阻塞。

## A. 新项目黄金路径

建立一个最小项目，固定同一角色 Character Core，并至少包含：

1. 默认造型镜头。
2. 显式切换婚礼 Look Variant。
3. 连续三个镜头不重复写“婚礼”，验证 Variant 自动继承。
4. 同一场景连续镜头不重复写 Scene Variant，验证场景继承。
5. 显式切换工作 Look Variant，验证只在此处换装。
6. 每镜检查 Canon Reference、Shot Manifest、Resolved State、Compiled Prompt 与最终生成结果。

通过标准：同一角色身份稳定；婚礼阶段不回退默认装；工作阶段只在显式切换后变化；场景固定结构和关键物件不漂移。

## B. 旧项目兼容

样本：`project-679e2eab7f3a12cc` 及至少一个普通旧项目。

验证：

- 打开前后原 Markdown hash 不变。
- Canon/Shot 首次惰性同步只新增结构化数据。
- 二次同步 Reused，不覆盖用户已调整 Shot。
- 有 Shot Manifest 的镜头走新 Compiler。
- 无 Shot Manifest 的旧入口仍可走 legacy fallback。
- 现有图片、视频、音频历史仍可查看。

## C. 网络与安全环境矩阵

至少覆盖：

- 普通 Windows，无 VPN/TUN。
- Clash/OpenClash 等 TUN/透明代理开启。
- 360 安全卫士/360 浏览器存在或残留策略环境。

验证 sidecar bind 仍为 `127.0.0.1`；Renderer/Agent/Codex Relay 使用 `localhost`；本地 API、第三方模型 Relay、登录/授权入口可用。

## D. Desktop 打包

候选命令：`pnpm electron:release:windows-x64`。

该命令必须保持 `--publish never`，先生成本地候选，不发布远端。

验证：安装/解压、首次启动、sidecar readiness、项目打开、Agent、Generation、退出/重启。

## E. 稳定性与恢复

- 长 Generation / Agent 任务至少一轮。
- 应用退出重启后项目和任务状态恢复。
- Generation 失败后 Retry。
- sidecar 异常退出后的错误提示/恢复。
- 不产生重复 Shot/Canon 或重复批次提交。

## F. 回归门禁

最终候选必须再次满足：

- `go test ./...` PASS。
- `go build ./...` PASS。
- Frontend 226/226 test files、1465/1465 tests 或更新后的完整测试集全部 PASS。
- `pnpm build` PASS。
- Instruction Pack tests PASS。
- `project-679...` Production QA regression PASS。
- `git diff --check` PASS。

## G. 发布控制

负责人可安排代码修复、测试、打包候选和本地验收；未经用户最终确认，不执行正式远端 push / publish。最终输出 Release Candidate 报告，列出：版本基线、改动范围、实机矩阵、已知限制、阻塞项、回滚方法和负责人结论。
