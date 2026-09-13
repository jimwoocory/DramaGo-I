# DramaGo-I 基线迁移验证

日期：2026-09-14

## 来源

- 源工作树：`D:\openai\MediaGo-Drama`
- 源分支：`mediago-clean-baseline-20260903`
- 源 HEAD：`78739afa269e4f634f4431b98630add38bcdb184`
- 新仓库：`https://github.com/jimwoocory/DramaGo-I`
- 新本地目录：`D:\openai\DramaGo-I`

旧工作树未删除、未重写历史，继续作为迁移回退点。

## 迁移方式

迁移采用受控快照，而不是整目录复制：

1. 复制源仓库全部 Git tracked 文件。
2. 补入当前真正需要的未跟踪 Canon / ShotManifest / Continuity / Production QA / docs 文件。
3. 明确排除源目录内嵌套 `MediaGO-Drama/` Git 仓库。
4. 不迁移 `node_modules`、前端 `dist`、release、cache、运行态 workspace/database 和本机私密配置。

迁移时统计：

- tracked：2045 个文件
- 选中 untracked：32 个文件

## 入库前扫描

- 未发现超过 25 MiB 的候选入库文件。
- 未发现 `node_modules`、release、cache、嵌套仓库进入迁移快照。
- 常见凭据模式扫描仅命中测试文件中的测试 fixture；非测试文件无命中。
- `.env.example` 保留为配置模板；真实 `.env` 不入库。

## 验证环境

- Node.js：v25.2.1
- pnpm：11.9.0
- Go：1.27.0 windows/amd64
- Git：2.55.0.windows.5

仓库建议最低版本仍以 README 为准。

## 验证结果

### 前端

`pnpm -C apps/workspace test`

- Test Files：226 / 226 PASS
- Tests：1470 / 1470 PASS

`pnpm -C apps/workspace build`

- PASS
- 仅存在既有 chunk-size warning，无编译错误。

### 服务端

首次直接执行服务端测试时，开发文档测试因新仓库没有被 Git 跟踪的 Swagger cache 而失败。该 cache 本来就是本机生成文件。

按 Taskfile 中同等命令生成 `.cache/server-swagger/swagger.json` 后：

`go test ./services/server/... -count=1`

- PASS

### Core / 其他 Go modules

`go test ./packages/core/... -count=1`

- PASS

`go test ./packages/instructions/... ./packages/jianyingdraft/... ./packages/mcp/... ./packages/tools/... ./packages/vendor/... -count=1`

- PASS

## 基线用途

该快照用于建立 DramaGo-I 的 B+C 开发起点：

- `main`：稳定迁移基线
- `develop`：后续 B+C 开发主线
- `v0.1.0-bc-baseline`：迁移基线标签

后续 P0/P1 开发应从 `develop` 开始，不再以旧 `MediaGO-Drama` 仓库作为跨电脑同步主线。
