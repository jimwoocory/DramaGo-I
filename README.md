# DramaGo-I

DramaGo-I 是基于开源项目 [mediago-dev/mediago-drama](https://github.com/mediago-dev/mediago-drama) 演进的独立开发主线，用于继续推进 B+C 架构：

- **B：专业生产控制骨架** —— Documents / Assets / Canon / Variant / ShotManifest / Continuity / Prompt Compiler / Generation / Timeline。
- **C：AI Command Plane** —— AI、页面、HTTP、MCP 与外部 Agent 通过同一业务动作层协作，不建立第二套业务真相。

本仓库不是 MediaGo 官方仓库，也不代表 MediaGo 官方发行版。

## 当前基线

2026-09-14 从本地 MediaGo 已验证工作树迁移建立。迁移来源见 [`docs/migration/source-baseline.txt`](docs/migration/source-baseline.txt)。

当前已验证主链包括：

- Documents / Assets / Agent / Generation / Episode / Timeline / ProductionShot
- Canon Core / Variant / CanonReference
- ShotManifest / Continuity Resolver
- Deterministic Prompt Compiler / Reference Policy
- Codex Image 与多 Provider 生成抽象
- Pippit / 小云雀 CLI generation provider
- Windows Electron / sidecar / ACP Agent / 重启与取消恢复
- 旧项目惰性迁移与兼容

B+C 后续产品与技术规划放在 `docs/plans/`。

## 仓库结构

```text
apps/workspace     React + Electron 工作台
services/server    Go 服务端 / Agent / MCP / 文档 / 生产控制面
packages/core      多模型与生成 Provider
packages/mcp       MCP 工具与协议
packages/*         共享能力与构建工具
docs/              方案、验收报告与迁移说明
scripts/           构建与辅助脚本
```

## 本地开发

建议环境：

- Node.js 24+
- pnpm 11.9+
- Go 1.25+
- go-task 3.x

安装依赖：

```bash
pnpm install --frozen-lockfile
```

开发运行：

```bash
pnpm dev:server
pnpm dev:desktop
```

常用验证：

```bash
# 前端
pnpm -C apps/workspace test
pnpm -C apps/workspace build

# Go
# 先生成开发 Swagger cache
# task -d services/server swagger

go test ./services/server/... -count=1
go test ./packages/core/... -count=1
go test ./packages/instructions/... ./packages/jianyingdraft/... ./packages/mcp/... ./packages/tools/... ./packages/vendor/... -count=1
```

## 跨电脑开发

公司电脑继续开发的标准流程见：

[`docs/migration/company-workstation.md`](docs/migration/company-workstation.md)

源码通过 GitHub 同步；项目 Workspace、SQLite、图片、视频等运行态数据不放 Git，使用 NAS 或单独的受控文件同步。账号登录、Provider 凭据和本机私密配置在各电脑独立配置。

## 分支策略

- `main`：稳定、经过验证的基线。
- `develop`：B+C 日常开发主线。
- `feature/*`：短期功能分支，完成后合入 `develop`。

迁移完成后的起始标签：`v0.1.0-bc-baseline`。

## 许可与上游

本仓库保留原项目 Apache License 2.0 授权与上游归属。第三方依赖继续遵循其各自许可证；原项目官方发行版中的专有组件与在线服务不属于本仓库授权范围，边界参考 [`COMMERCIAL_FEATURES.md`](COMMERCIAL_FEATURES.md)。

- 上游开源项目：https://github.com/mediago-dev/mediago-drama
- DramaGo-I：https://github.com/jimwoocory/DramaGo-I
