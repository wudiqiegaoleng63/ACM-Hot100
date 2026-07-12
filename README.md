# ACM HOT 100

中文 ACM 在线判题网站

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker Desktop

## Quick Start

```powershell
# 1. Clone the repository
git clone https://github.com/<your-org>/ACMHOT100.git
cd ACMHOT100

# 2. Copy environment file
Copy-Item .env.example .env

# 3. Start infrastructure (MySQL, Redis, Mailpit)
docker compose -f infra/docker-compose.yml up -d

# 4. Start backend
cd apps/server
go mod tidy
go run cmd/api/main.go

# 5. Start frontend (in a new terminal)
cd apps/web
npm install
npm run dev
```

## Architecture

This project follows a monorepo structure:

- **apps/server** — Go backend (API, judge integration, auth)
- **apps/web** — Frontend (Vue / React SPA)
- **infra/** — Docker Compose configurations for local infrastructure
- **packages/** — Shared libraries and utilities
- **seed/** — Database seed data

## Available Commands

### Web (Frontend)

| Command | Description |
|---------|-------------|
| `cd apps/web && npm run dev` | Start dev server |
| `cd apps/web && npm run build` | Production build |
| `cd apps/web && npm run lint` | Lint code |

### Server (Backend)

| Command | Description |
|---------|-------------|
| `cd apps/server && go run cmd/api/main.go` | Start API server |
| `cd apps/server && go run cmd/worker/main.go` | Start background worker |
| `cd apps/server && go test ./...` | Run tests |

### Infrastructure

| Command | Description |
|---------|-------------|
| `docker compose -f infra/docker-compose.yml up -d` | Start all services |
| `docker compose -f infra/docker-compose.yml down` | Stop all services |
| `docker compose -f infra/docker-compose.yml logs -f` | Stream logs |

## Judge Mode

Set the `JUDGE_MODE` environment variable to switch between judge backends:

- **`mock`** — Returns canned responses, no real code execution (default for local development)
- **`judge0`** — Forwards submissions to a Judge0 server (set `JUDGE0_BASE_URL` accordingly)

## Note

独立学习项目，非 LeetCode 官方产品
