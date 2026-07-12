# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 编码前必做：查阅最新文档

每次编写代码之前，**必须**先通过以下方式查阅相关库/框架的最新文档和用法，不要仅凭训练数据中的记忆编写代码：

1. **Context7 MCP** — 使用 `mcp__context7__resolve-library-id` 解析库 ID，再用 `mcp__context7__query-docs` 查询最新文档。适用于所有库、框架、SDK、CLI 工具的 API 语法、配置、版本迁移等问题。
2. **网络搜索** — 使用 `WebSearch` 或 `mcp__aliyun__bailian_web_search` 搜索最新实践、变更日志、社区方案等 Context7 未覆盖的内容。

**规则：** 即使你认为已经知道答案，也必须先查文档再写代码。训练数据可能已过时。

## Git 版本管理

使用 Git 进行代码版本管理，遵循以下规则：

- **每个有意义的改动都要提交** — 完成一个功能、修复一个 bug、重构一段代码后，立即 commit，不要积攒大量改动后一次性提交。
- **提交信息规范** — 使用简洁明了的中文或英文描述改动内容，格式：`<类型>: <描述>`，类型包括：
  - `feat:` 新功能
  - `fix:` 修复 bug
  - `refactor:` 重构
  - `docs:` 文档变更
  - `chore:` 构建/工具变更
- **提交前检查** — 每次 commit 前，用 `git diff` 确认变更内容，确保没有遗漏或误改。
- **不要推送未经确认的代码** — 只有在用户明确要求时才执行 `git push`。
- **分支策略** — 在 main 分支上直接开发小改动；较大功能应创建 `feat/<功能名>` 分支，完成后再合并回 main。

常用命令：
```bash
git status                    # 查看当前状态
git diff                      # 查看未暂存的变更
git add -A                    # 暂存所有变更
git commit -m "feat: xxx"     # 提交
git log --oneline -10         # 查看最近提交记录
git branch feat/xxx           # 创建分支
git checkout feat/xxx         # 切换分支
git merge feat/xxx            # 合并分支到当前分支
```

## 项目概述

ACM Hot 100 — 中文 ACM 在线判题网站 MVP。用户注册/登录后浏览 Hot 100 题单，使用 ACM 标准输入输出形式编写完整程序，运行样例、正式提交、查看判题状态和进度。

**唯一事实来源：** `docs/MVP_DESIGN_SPEC.md`。代码与规格冲突时以规格为准。

## 技术栈

- **前端：** React + Vite + TypeScript SPA，React Router Data Mode，TanStack Query，React Hook Form + Zod，Tailwind CSS，Monaco Editor，Lucide Icons
- **后端：** Go + Gin，GORM + MySQL Driver，go-redis/v9，golang-jwt/jwt/v5
- **判题：** Judge0 HTTP API（开发用 `JUDGE_MODE=mock`）
- **邮件：** SMTP（开发用 Mailpit）
- **测试：** Go testing/Testify，Vitest/React Testing Library，Playwright E2E

## 仓库结构

```
ACMHOT100/
├─ apps/
│  ├─ web/                    # React + Vite + TypeScript SPA
│  │  ├─ src/
│  │  │  ├─ app/              # Router、Provider、全局错误边界
│  │  │  ├─ components/       # 跨功能通用组件
│  │  │  ├─ features/         # auth/problems/editor/submissions/profile
│  │  │  ├─ lib/              # API Client、Query Client、工具
│  │  │  └─ styles/           # Token 与全局样式
│  │  ├─ nginx.conf           # SPA fallback 与 /api 反向代理
│  │  └─ vite.config.ts
│  └─ server/                 # 单个 Go Module
│     ├─ cmd/
│     │  ├─ api/              # Gin API 入口
│     │  └─ judge-worker/     # Redis Streams 判题消费者
│     ├─ internal/
│     │  ├─ auth/             # JWT、邮箱验证、密码与会话
│     │  ├─ config/
│     │  ├─ http/             # Gin handlers/middleware/routes
│     │  ├─ judge/            # Judge0 adapter 与结果比较
│     │  ├─ model/            # GORM models
│     │  ├─ queue/            # Redis Streams
│     │  ├─ repository/
│     │  └─ service/
│     ├─ migrations/          # 版本化 MySQL SQL 迁移
│     ├─ go.mod / go.sum
├─ packages/
│  └─ contracts/              # OpenAPI 生成的前端类型或共享 schema
├─ infra/
│  ├─ docker-compose.yml
│  └─ judge0/                 # Judge0 相关说明/覆盖配置
├─ seed/
│  └─ problems/               # 版本化题目 YAML/JSON 与测试数据
├─ docs/
├─ .env.example
└─ README.md
```

## 常用命令

### 前端 (apps/web)
```bash
cd apps/web
npm install          # 安装依赖
npm run dev          # Vite 开发服务器 (端口 5173，代理 /api → :8080)
npm run build        # 生产构建
npm run lint         # ESLint
npm run type-check   # TypeScript 类型检查
npm run test         # Vitest 单元测试
npm run test:e2e     # Playwright E2E 测试
```

### 后端 (apps/server)
```bash
cd apps/server
go mod tidy          # 整理依赖
go run cmd/api/main.go              # 启动 API 服务 (端口 8080)
go run cmd/judge-worker/main.go     # 启动判题 Worker
go test ./...                        # 运行所有测试
go test ./internal/auth/...          # 运行 auth 包测试
go vet ./...                         # 静态检查
```

### 基础设施
```bash
cd infra
docker compose up -d                # 启动 MySQL、Redis、Mailpit
docker compose down                 # 停止
docker compose logs -f mailpit      # 查看 Mailpit 日志
```

### 数据库迁移与 Seed
```bash
cd apps/server
go run cmd/api/main.go migrate      # 执行 SQL 迁移
go run cmd/api/main.go seed         # 执行 Seed 数据
```

## 关键架构决策

- **MySQL 是最终事实来源**，Redis 只存有 TTL 或可重建的数据。提交必须先写 MySQL 再写 Redis Stream。
- **JWT 双 Token**：Access 15min + Refresh 7day，HttpOnly SameSite Cookie，Refresh Rotation + 复用检测。
- **邮箱验证/密码重置 Token**：Redis 只存 SHA-256 哈希，不存 Raw Token。
- **判题 Worker**：通过 Redis Streams Consumer Group 消费，分布式锁 + MySQL 行锁保证幂等。
- **前端 API Client**：`credentials: 'include'`，并发 401 只触发一次 Refresh，禁止死循环。
- **版本化 SQL migration**：生产环境禁止依赖 GORM AutoMigrate。

## 实现红线

- API 错误统一 `error + request_id` 结构
- 公开题目接口只返回公开样例，绝不返回隐藏测试
- 认证 Token 不得写入 LocalStorage
- React Router 受保护路由必须等 `/auth/me` 完成后再判断跳转
- JWT 必须固定签名算法，严格验证 `iss/aud/exp/nbf/jti`
- 忘记密码接口不得泄露邮箱是否存在
- 日志不得包含密码、Cookie、完整隐藏测试或完整用户代码
- 任何提交终态不能回退到非终态
- 没有真实数据时隐藏通过率、连续天数、排名
- 不要用 `any`、`@ts-ignore` 或吞异常作为常规方案
- 不要把整个页面写成超过 1000 行的组件
- Markdown 禁用原始 HTML
- 用户源代码最大 64KB，自测输入最大 16KB

## 视觉设计要点

- 暖灰背景 `#f6f5f1`，白色内容面，砖橙强调色 `#c45724`
- 状态色语义化：绿=AC/成功，红=WA/错误，琥珀=尝试中，蓝=编译/运行中
- 题单用表格不用卡片，做题页左右分栏（默认 46/54）
- 图标只用 Lucide 线性图标，不用 Emoji
- 禁止紫色渐变、玻璃拟态、虚构统计数字
