# ACM HOT 100

中文 ACM 在线判题网站 — 用 ACM 标准输入输出模式刷完 Hot 100。

## Prerequisites

- Go 1.26+（与 `go.mod` 中 `go 1.26.1` 一致）
- Node.js 20+
- Docker Desktop（仅基础设施需要；开发也可用宿主机 MySQL/Redis）

## Quick Start

### 方式一：宿主机开发（推荐）

需要本机已运行 MySQL 8.0 和 Redis 7。

```powershell
# 1. Clone the repository
git clone https://github.com/<your-org>/ACMHOT100.git
cd ACMHOT100

# 2. Copy and edit environment file
Copy-Item .env.example .env
# 编辑 .env，设置 MYSQL_DSN、REDIS_ADDR 等指向本机服务

# 3. Start backend
cd apps/server
go mod tidy
go run cmd/api/main.go          # API server on :8080

# 4. Start frontend (in a new terminal)
cd apps/web
npm install
npm run dev                      # Vite dev server on :5173, proxy /api -> :8080
```

### 方式二：Docker 基础设施 + 宿主机服务

```powershell
# 1. Clone and configure
git clone https://github.com/<your-org>/ACMHOT100.git
cd ACMHOT100
Copy-Item .env.example .env

# 2. Start infrastructure only (MySQL, Redis, Mailpit)
docker compose -f infra/docker-compose.yml up -d

# 3. Start backend (uses Docker MySQL/Redis via localhost)
cd apps/server
go mod tidy
go run cmd/api/main.go

# 4. Start frontend (in a new terminal)
cd apps/web
npm install
npm run dev
```

## Architecture

Monorepo 结构：

- **apps/server** — Go 后端（Gin API + Judge Worker，共享 `internal` 包）
- **apps/web** — React + Vite + TypeScript SPA
- **infra/** — Docker Compose（MySQL 8.0、Redis 7、Mailpit）
- **seed/** — 数据库种子数据（5 道 ACM 示例题）
- **packages/** — 共享类型与 schema

## Available Commands

### Web (Frontend)

| Command | Description |
|---------|-------------|
| `cd apps/web && npm run dev` | Start Vite dev server (port 5173) |
| `cd apps/web && npm run build` | Production build |
| `cd apps/web && npm run type-check` | TypeScript type check |
| `cd apps/web && npm run lint` | ESLint 9 |
| `cd apps/web && npm run test` | Vitest unit tests |
| `cd apps/web && npm run test:e2e` | Playwright E2E tests |

### Server (Backend)

| Command | Description |
|---------|-------------|
| `cd apps/server && go run cmd/api/main.go` | Start API server (port 8080) |
| `cd apps/server && go run cmd/judge-worker/main.go` | Start Judge Worker |
| `cd apps/server && go run cmd/api/main.go -migrate` | Run database migrations |
| `cd apps/server && go run cmd/api/main.go -migrate-down` | Rollback last migration |
| `cd apps/server && go run cmd/api/main.go -seed` | Seed problems and language configs |
| `cd apps/server && go test ./...` | Run all Go tests |
| `cd apps/server && go vet ./...` | Static analysis |

### Infrastructure

| Command | Description |
|---------|-------------|
| `docker compose -f infra/docker-compose.yml up -d` | Start MySQL, Redis, Mailpit |
| `docker compose -f infra/docker-compose.yml down` | Stop all services |
| `docker compose -f infra/docker-compose.yml logs -f mailpit` | View Mailpit logs |

## Judge Mode

Set the `JUDGE_MODE` environment variable to switch between judge backends:

- **`mock`** — Returns canned responses, no real code execution (default for development)
- **`judge0`** — Forwards submissions to a Judge0 server (set `JUDGE0_BASE_URL` accordingly)

## Note

独立学习项目，非 LeetCode 官方产品
