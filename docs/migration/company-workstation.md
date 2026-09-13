# 公司电脑继续开发 DramaGo-I

日期：2026-09-14

## 1. 克隆与分支

```bash
git clone https://github.com/jimwoocory/DramaGo-I.git
cd DramaGo-I
git checkout develop
```

建议额外保留官方上游：

```bash
git remote add upstream https://github.com/mediago-dev/mediago-drama.git
git remote -v
```

`origin` 只指向 `jimwoocory/DramaGo-I`；日常开发从 `develop` 开始，不直接在 `main` 上写功能。

## 2. 开发环境

建议安装并验证：

```text
Node.js 24+
pnpm 11.9+
Go 1.25+
go-task 3.x
Git
Codex / 当前团队开发工具
```

验证：

```bash
node --version
pnpm --version
go version
task --version
git --version
```

## 3. 安装依赖

```bash
pnpm install --frozen-lockfile
```

不要从另一台电脑复制 `node_modules`、前端 `dist`、Go build cache 或 Electron release 目录。

## 4. 首次验证

开发 Swagger cache 是本机生成内容，不入 Git。安装 go-task 后运行：

```bash
task -d services/server swagger
```

再运行：

```bash
pnpm -C apps/workspace test
pnpm -C apps/workspace build

go test ./services/server/... -count=1
go test ./packages/core/... -count=1
go test ./packages/instructions/... ./packages/jianyingdraft/... ./packages/mcp/... ./packages/tools/... ./packages/vendor/... -count=1
```

## 5. 启动

两个终端：

```bash
pnpm dev:server
```

```bash
pnpm dev:desktop
```

改动服务端、Agent MCP 或生成 MCP 后，重新构建/重启对应进程。

## 6. Git 与项目数据分离

GitHub 只同步：

- 源码
- 测试
- migration/schema
- docs
- 配置模板

不要通过 GitHub 同步：

- `.env` 或本机私密配置
- API Key / OAuth / 登录态
- Workspace SQLite / 运行态数据库
- 项目图片、视频、音频与大体积生成结果
- `node_modules`
- `dist` / release 安装包 / build cache

项目 Workspace 与素材使用 NAS 或其他受控文件同步；每台电脑单独配置 Codex、小云雀和其他 Provider 登录/凭据。

## 7. 日常同步

家里电脑结束工作：

```bash
git status
git pull --rebase origin develop
# 提交经过验证的工作
git push origin develop
```

公司电脑开始工作：

```bash
git checkout develop
git pull --rebase origin develop
```

不要在两台电脑分别保留长期未提交的同一批改动。切换电脑前至少做一个可说明的 commit，或者明确保留在单机上不要同时修改同一范围。

## 8. 基线恢复

B+C 开发前的迁移基线标签：

```bash
git checkout v0.1.0-bc-baseline
```

该标签只用于回看/恢复；继续开发时回到 `develop`。
