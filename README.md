# ACM HOT 100

中文 ACM 在线判题网站 MVP。用户注册并验证邮箱后，可以浏览原创 Hot 100 训练题、使用完整程序进行样例自测与正式提交，并查看判题历史和训练进度。

> 独立学习项目，非 LeetCode 官方产品。

## 功能与技术栈

- React 19 + Vite + TypeScript + React Router + TanStack Query
- Go 1.26 + Gin + GORM
- MySQL 8、Redis 7 Streams、Mailpit
- Mock Judge（默认）或可选自托管 Judge0 CE
- Vitest、Go testing、Playwright

## 前置要求

### 全栈 Docker 启动

- Docker Engine 24+ 或 Docker Desktop
- Docker Compose v2

### 宿主机开发

- Go 1.26.1+（与 `apps/server/go.mod` 一致）
- Node.js 20+
- MySQL 8.0、Redis 7

## 最快启动：全栈 Compose

默认使用 Mock Judge；Web、API、Worker、MySQL、Redis 和 Mailpit 都在容器中运行。

```powershell
git clone https://github.com/wudiqiegaoleng63/ACM-Hot100.git
cd ACM-Hot100
Copy-Item .env.example .env

docker compose -f infra/docker-compose.yml up --build -d
docker compose -f infra/docker-compose.yml ps
```

访问：

- Web：<http://localhost:5173>
- API 健康检查：<http://localhost:8080/api/v1/health>
- Mailpit：<http://localhost:8025>

首次启动会自动执行版本化 SQL migration 和幂等 Seed。查看日志或关闭：

```powershell
docker compose -f infra/docker-compose.yml logs -f api worker web
docker compose -f infra/docker-compose.yml down
# 同时删除本地数据库与 Redis 数据：
docker compose -f infra/docker-compose.yml down -v
```

如宿主机端口已被占用，可在 `.env` 覆盖 `WEB_PORT`、`API_PORT`、`MYSQL_PORT`、`REDIS_PORT`、`MAILPIT_HTTP_PORT` 和 `MAILPIT_SMTP_PORT`。

## 宿主机开发

宿主机配置使用 `localhost`；容器配置由 Compose 覆盖为 `mysql`、`redis`、`mailpit` 等内部服务名，两者不要混用。

```powershell
Copy-Item .env.example .env
# 编辑 .env 中的本地 MySQL/Redis 密码和 JWT Secret

cd apps/server
go run cmd/api/main.go -migrate
go run cmd/api/main.go -seed
go run cmd/api/main.go
```

另开终端启动 Worker：

```powershell
cd apps/server
go run cmd/judge-worker/main.go
```

另开终端启动 Web：

```powershell
cd apps/web
npm ci
npm run dev
```

Vite 默认监听 `5173`，并将 `/api` 代理到 `http://localhost:8080`。如 API 使用其他地址，设置 `VITE_API_PROXY_TARGET`。

## Migration 与 Seed

```powershell
cd apps/server
go run cmd/api/main.go -migrate       # 执行所有未应用的 up migration
go run cmd/api/main.go -migrate-down  # 回滚最近一版 migration
go run cmd/api/main.go -seed          # 幂等导入语言配置与 5 道题
```

API 容器正常启动时也会执行未应用的 migration。生产环境仍以版本化 SQL 为准，不使用 GORM AutoMigrate。

## 判题模式

### Mock Judge（默认）

```text
JUDGE_MODE=mock
```

适用于开发、UI 联调和验收。样例与正式提交仍经过 Redis Streams 和独立 Worker；不会在 API 进程执行用户代码。

### Judge0 CE（可选 Profile）

Judge0 使用独立 PostgreSQL 和 Redis，不复用业务 MySQL/Redis。启动前在 `.env` 设置强密码：

```text
JUDGE_MODE=judge0
JUDGE0_POSTGRES_PASSWORD=<strong-password>
JUDGE0_REDIS_PASSWORD=<strong-password>
```

然后启动：

```powershell
docker compose -f infra/docker-compose.yml --profile judge0 up --build -d
docker compose -f infra/docker-compose.yml --profile judge0 ps
```

Judge0 API 仅绑定 `127.0.0.1:2358`，业务 Worker 通过 Compose 内网访问 `http://judge0-server:2358`。Judge0 Server/Workers 需要 `privileged` 沙箱权限；不应部署在不可信共享宿主机上。

## 自动化验证

### 前端

```powershell
cd apps/web
npm ci
npm run lint
npm run type-check
npm run test -- --run
npm run build
```

### 后端

```powershell
cd apps/server
go test ./... -count=1
go vet ./...
go build ./cmd/api ./cmd/judge-worker
```

### Playwright 核心 E2E

E2E 覆盖注册、Mailpit 邮箱验证、登录、题单、草稿、样例运行、正式提交 AC、Profile 和提交详情。需要先启动隔离测试服务或使用项目验证流程：

```powershell
cd apps/web
npx playwright install chromium
npm run test:e2e
```

Playwright 不启用 retry 或 skip。CI/本机可通过 `E2E_BASE_URL`、`E2E_MAILPIT_URL`、`E2E_EXTERNAL_SERVERS` 和 `E2E_CHROME_CHANNEL` 指定已有服务与浏览器。

## 常用命令

| 范围 | 命令 | 作用 |
|---|---|---|
| Web | `cd apps/web && npm run dev` | 启动 Vite 开发服务器 |
| Web | `cd apps/web && npm run build` | 生产构建 |
| Web | `cd apps/web && npm run test:e2e` | Playwright E2E |
| API | `cd apps/server && go run cmd/api/main.go` | 启动 API |
| Worker | `cd apps/server && go run cmd/judge-worker/main.go` | 启动判题 Worker |
| Compose | `docker compose -f infra/docker-compose.yml up --build -d` | 构建并启动完整 Mock 栈 |
| Compose | `docker compose -f infra/docker-compose.yml ps` | 查看容器和健康状态 |
| Compose | `docker compose -f infra/docker-compose.yml logs -f api worker` | 查看 API/Worker 日志 |

## 排错

### 端口已占用

在 `.env` 修改对应宿主机端口。例如：

```text
WEB_PORT=15173
API_PORT=18080
MYSQL_PORT=13306
REDIS_PORT=16379
MAILPIT_HTTP_PORT=18025
MAILPIT_SMTP_PORT=11025
```

### API 或 Worker 不健康

```powershell
docker compose -f infra/docker-compose.yml ps
docker compose -f infra/docker-compose.yml logs mysql redis migrate seed api worker
```

确认 `migrate`、`seed` 的退出码为 0，且 MySQL/Redis 为 `healthy`。

### 收不到验证邮件

- Compose：确认 `MAIL_MODE=smtp`、API 使用 `SMTP_HOST=mailpit`，并打开 Mailpit Web UI。
- 宿主机：确认 `.env` 使用 `SMTP_HOST=localhost`、`SMTP_PORT=1025`。
- 只有显式设置 `MAIL_MODE=log` 时才不会发送 SMTP 邮件。

### 页面路由刷新 404

生产 Web 镜像已经配置 Nginx SPA fallback。若自行部署，必须保留 `apps/web/nginx.conf` 中的 `try_files $uri $uri/ /index.html`。

### 清理并重新验收

```powershell
docker compose -f infra/docker-compose.yml down -v
docker compose -f infra/docker-compose.yml up --build -d
```

不要在含有需要保留数据的环境执行 `down -v`。

## 仓库结构

```text
apps/web        React SPA、Nginx 配置与 Web Dockerfile
apps/server     Go API、Judge Worker、migration 与 Server Dockerfile
infra           全栈 Compose 与可选 Judge0 Profile
seed/problems   5 道版本化原创题、公开样例与隐藏测试
packages        共享契约目录
docs            产品规格与实施计划
```
