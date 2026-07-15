# ACM Hot 100 · Post-MVP 上线与演进实施计划

> 面向执行能力弱于规划者的编码模型。每次只执行一个任务；每个任务必须独立验证、独立提交、完成后停止。
>
> 基线审计日期：2026-07-15。基线提交：`fa1be7d docs: 完成全栈部署和最终验收`。

## 0. 规划结论

MVP 已完成，下一阶段的主目标固定为：**把项目从“本地可演示”推进到“小规模公开 Beta 可安全运行、可恢复、可观察、可回滚”**。

本计划不直接进入管理后台、排行榜、收藏或完整 100 题。先完成 Phase 5 的上线闸门，再进入 Phase 6 产品扩展。原因是在线判题会执行不可信代码，且系统已经涉及账户、邮件、持久数据和异步队列；在缺少交付、隔离、监控和恢复能力时扩展功能，会放大上线风险。

### 0.1 默认部署假设

- 用户规模：小规模公开 Beta；设计验收基线为 20 个并发提交、100 个连续判题任务。
- 核心业务：一台 Linux x86_64 服务器，Docker Engine + Docker Compose。
- Judge0：生产环境使用独立服务器或至少独立安全边界；不与核心 MySQL、Redis、API 共用公网入口。
- 边缘入口：Caddy 2，负责域名、自动 HTTPS 和反向代理；现有 Web Nginx 继续只提供静态站点。
- 镜像仓库：默认 GHCR；如果目标服务器访问 GHCR 不稳定，在资源确认时改为用户指定的国内或云厂商镜像仓库。
- CI/CD：GitHub Actions。CI 自动验证；镜像在版本标签发布；生产部署默认需要人工批准。
- 数据：MySQL 是业务事实来源；Redis 可重建但仍需 AOF、内存上限和故障策略。
- 邮件：使用用户指定的生产 SMTP 服务，不绑定具体厂商。
- 编排：Phase 5 不引入 Kubernetes、Terraform、服务网格或多区域部署。

这些是假设，不是对外部资源的虚构。若用户给出不同目标，先单独更新本节和受影响任务，不允许编码模型自行改架构。

### 0.2 本阶段完成定义

Phase 5 全部完成后，必须同时满足：

- PR/提交有自动化 CI，任何失败都会阻止发布。
- API 与 Web 镜像由受控流水线构建，使用不可变版本或 digest，可追溯到 Git 提交。
- 生产环境没有开发默认 Secret，数据库、Redis、Judge0 不暴露公网端口。
- 域名、HTTPS、真实 SMTP、生产 Judge0 全部有真实验收证据。
- API、Worker、队列、Judge0 有结构化日志、指标和最小告警。
- MySQL 备份可自动执行，并完成过一次隔离环境恢复演练。
- 20 个并发提交、100 个任务下无任务丢失、无重复进度更新、全部进入终态。
- Worker、Redis 短暂中断和 Judge0 超时经过故障恢复验收。
- 有部署、回滚、备份、恢复、故障处理和值班检查文档。
- 上线操作只有获得用户明确授权后才能执行。

---

## 1. 强制执行规则

1. `docs/MVP_DESIGN_SPEC.md` 继续约束已有产品行为；本文件决定 Post-MVP 实施顺序。
2. 每次只执行本文件第一个未勾选任务。完成、验证、提交、汇报后立即停止。
3. 不允许跨任务“顺手实现”；发现额外问题，只记录到任务汇报的“剩余风险”。
4. 开工前完整阅读 `CLAUDE.md`、本文件和本任务涉及的代码。
5. 涉及库、GitHub Actions、Docker、Caddy、Judge0、OpenAPI 或监控时，先查最新官方文档。
6. Bug、安全策略和业务逻辑必须先写失败测试，再修实现。
7. 不得删除、skip、放宽断言、增加无条件 retry 或硬编码成功结果来制造绿色验证。
8. 不得提交真实密码、Token、Cookie、SMTP 凭据、私钥、生产数据库或用户代码。
9. 每个任务允许创建一个本地 Git commit；禁止 `git push`、创建远程分支、PR、Release、Package、Environment 或修改仓库设置，除非用户当轮明确授权。
10. 外部资源缺失时，只能完成明确写出的“仓库内准备”；需要真实验收的任务必须停止并列出所需资源，不能用 Mock 代替。
11. 生产变更、DNS、服务器、防火墙、数据库恢复和真实邮件发送都视为外部状态变更，必须有明确授权。
12. 禁止对 Judge0 执行节点与核心数据服务之间建立不必要的网络访问。
13. 所有新增脚本默认 `set -Eeuo pipefail`，不得把 Secret 放进命令行参数或日志。
14. 工作树有用户改动时不得覆盖；无法安全隔离就停止。

## 2. 已核实的真实基线

### 2.1 仓库与版本状态

- 当前分支为 `main`，相对 `origin/main` 领先 1 个提交：`fa1be7d`。
- 规划审计时工作树干净。
- 远端为 `origin https://github.com/wudiqiegaoleng63/ACM-Hot100.git`。
- 用户尚未授权 push，因此本计划不得自动同步远端。

### 2.2 2026-07-15 实测通过

- `npm run lint`
- `npm run type-check`
- `npm run test -- --run`：15 个测试文件、94 个测试通过。
- `npm run build`
- `go test ./...`
- `go vet ./...`
- `go build ./cmd/api ./cmd/judge-worker`
- `docker compose -f infra/docker-compose.yml config --quiet`

### 2.3 已有能力

- React + Vite 前端、Gin API、Go Worker、MySQL、Redis、Mailpit 全栈 Compose。
- 邮箱注册验证、登录、Refresh Rotation、退出、忘记/重置密码。
- 题库、Monaco、草稿、样例运行、正式提交、提交历史和训练进度。
- Redis Streams、Reconciliation、XAUTOCLAIM、幂等终态更新。
- Mock Judge 和可选 Judge0 Compose Profile。
- Go、Vitest、Playwright 自动化测试。

### 2.4 上线前真实缺口

- 没有 `.github/workflows`，远端没有自动验证和镜像发布证据。
- 没有 OpenAPI 契约，前后端接口仍主要靠手写类型维持。
- `Makefile` 的 migrate/seed 路径和部分工具命令已过时。
- 存在未使用的占位 `AuthRequired` 中间件，容易被误接入。
- API `http.Server` 未设置读取、写入、空闲和 Header 超时。
- API 在启动时自动迁移；多副本或发布失败时不具备安全迁移纪律。
- `RequireAuth` 在 Redis deny-list 查询失败时放行，生产故障策略不安全。
- 健康接口混合了 liveness、readiness 和业务信息，不适合直接作为生产探针。
- 主要使用非结构化 `log`，缺少请求日志、队列指标和告警。
- 没有生产 Compose overlay、Secret 文件注入、自动备份和恢复演练。
- Judge0 Profile 使用 `privileged: true` 且适合本地集成，不应原样作为生产隔离方案。
- Judge0 Adapter 尚未支持私有认证 Header。
- 没有真实并发容量报告和故障注入报告。

---

## 3. 目标架构与信任边界

```text
Internet
   │ 80/443 only
   ▼
Caddy (Core VM)
   │
   ├── /api/* ──> Gin API ──> MySQL
   │                    └────> Redis Streams
   └── /* ──────> Web Nginx       │
                                  ▼
                            Go Judge Worker
                                  │ private network + auth token
                                  ▼
                         Judge0 API (Judge VM)
                                  │
                       Judge0 Workers / Postgres / Redis
```

信任边界：

- 互联网只能访问 Caddy 的 80/443。
- Core MySQL、Core Redis、Gin API 端口、监控端口默认不对公网开放。
- Judge0 API 只允许 Core Worker 的私网 IP，并要求认证 Header。
- Judge0 执行用户代码的节点不能访问 Core MySQL、Core Redis、SMTP Secret 或 Docker Socket。
- CI 构建环境不持有生产数据库和服务器 Root 凭据。
- 生产部署使用镜像 digest 或不可变版本，不在服务器现场构建源代码。

---

## 4. 依赖顺序和验收闸门

```text
Phase 5A 可信交付基线
  ↓ Gate A：CI、契约、扫描、镜像全部可重复
Phase 5B 生产运行基线
  ↓ Gate B：配置、迁移、HTTPS、SMTP、Staging 全通过
Phase 5C 可观测与数据恢复
  ↓ Gate C：告警可触发，备份可恢复，Runbook 可执行
Phase 5D 真实判题与上线
  ↓ Gate D：真实判题、容量、故障恢复和 Go/No-Go 通过
Phase 6 产品与内容扩展
```

阶段工作量与交付物：

| 阶段 | 任务数 | 主要交付物 | 外部依赖 |
|---|---:|---|---|
| Phase 5A | 6 | OpenAPI、CI、E2E、安全扫描、可追溯镜像 | GitHub push/Actions/Packages 权限 |
| Phase 5B | 8 | Secret 契约、运行加固、安全迁移、生产 Compose、HTTPS、SMTP、Staging | 服务器、域名、SMTP |
| Phase 5C | 6 | 结构化日志、指标、告警、备份、恢复、Runbook | 告警渠道、异机备份存储 |
| Phase 5D | 5 | 隔离 Judge0、真实判题矩阵、容量、故障恢复、Go/No-Go | Judge 节点、上线授权 |

最少需要 25 轮单任务执行。外部资源任务可能增加一次“仓库内准备”和一次“真实验收”，不建议为了减少轮次合并任务。

禁止事项：

- Gate A 前不部署 Staging。
- Gate B 前不接收真实公网用户。
- Gate C 前不进行生产开放。
- Gate D 前不宣称“可上线”。
- Phase 5 未完成前不启动管理后台和完整 100 题。

---

## 5. 外部资源与人工决策闸门

| 闸门 | 最迟需要时间 | 用户必须提供或确认 | 缺失时允许做什么 |
|---|---|---|---|
| G0 GitHub 同步 | P5-03 验证前 | 是否允许 push；仓库 Actions/Packages 权限 | 只可在本地创建 workflow 并做语法检查 |
| G1 部署区域 | P5-10 前 | 云厂商/区域、Linux 版本、CPU/内存/磁盘、是否兼容中国大陆网络 | 完成 provider-neutral 配置和文档 |
| G2 域名网络 | P5-11 前 | 域名、DNS 控制权、公网 IP、80/443、防火墙权限 | 只能用本地域名验证配置语法 |
| G3 生产邮件 | P5-12 前 | SMTP Host/Port/User/Password/From/TLS；DNS 发信配置 | 保持 Mailpit，不得声称真实邮件通过 |
| G4 备份存储 | P5-18 前 | 异机对象存储或备份服务器、保留周期、RPO/RTO | 只做本地加密备份和恢复演练 |
| G5 Judge 节点 | P5-21 前 | 独立服务器/私网/防火墙、CPU/内存/磁盘 | 只在隔离测试环境验证 Judge0 |
| G6 生产开放 | P5-25 前 | 明确“允许生产部署/开放访问” | 只完成 Go/No-Go 报告，不执行上线 |
| G7 内容授权 | P6-07 前 | 100 题来源、原创/授权策略 | 只做内容导入工具，不复制第三方题面 |

敏感值不得写入 Issue、计划文件、Git 提交或聊天汇报；只记录变量名和配置位置。

---

## 6. Phase 5A · 可信交付基线

### [x] P5-01 · 固化 Post-MVP 工程基线

目标：删除会误导后续模型的遗留入口，让一条命令能够验证当前仓库。

前置：规划文档已单独保存；不需要外部资源。

只做：

- 修正 `Makefile` 中不存在的 migrate/seed 入口，统一为 `cmd/api -migrate/-seed`。
- 不依赖未安装的全局 `golangci-lint`；现阶段使用 `go vet`，是否引入新 linter 留给独立任务。
- 新增聚合命令 `make verify`，覆盖 Web lint/type-check/test/build、Go test/vet/build、Compose config。
- 删除未使用且不校验 JWT 的占位 `AuthRequired`，确认所有受保护路由仍使用 `RequireAuth`。
- 在旧实施计划顶部标注“已完成历史计划”，并从 README 链接本文件。
- 添加当前基线说明，不重写 MVP 设计规格。

不得做：业务功能、认证策略、CI、OpenAPI、生产 Compose。

测试与验收：

```bash
make verify
rg -n 'AuthRequired\(' apps/server --glob '*.go'
git diff --check
```

人工确认：`rg` 不再找到占位中间件；README 能明确区分 MVP 历史计划和 Post-MVP 活跃计划。

建议提交信息：`chore: 固化 post-MVP 工程基线`

### [x] P5-02 · 建立 OpenAPI 单一接口契约

目标：让已有 API 的路径、Cookie、请求、响应、枚举和错误结构可被机器校验。

前置：P5-01。

只做：

- 新增 OpenAPI 3.1 文档，覆盖当前全部 `/api/v1` 路由，不新增业务端点。
- 定义统一 Error、分页、认证 Cookie、Problem、Language、Draft、Run、Submission、Profile Schema。
- 明确 Access/Refresh Cookie、Origin 要求、401/403/409/422/429/503。
- 将 OpenAPI lint 工具锁定在项目依赖中，不使用未锁版本的临时 `npx`。
- 添加契约校验脚本和最小测试，确保真实 Handler 的关键响应字段与契约一致。
- README 增加本地查看/校验契约命令。

不得做：自动生成整套 Handler、改变字段命名、绕过现有安全 Cookie、增加 Swagger 公网服务。

验证：

```bash
cd apps/web
npm run api:lint
npm run type-check
npm run test -- --run
cd ../server
go test ./internal/http/... -count=1
go test ./...
git diff --check
```

人工确认：契约不包含隐藏测试输入、参考答案、Refresh Token 内容或服务器内部 DSN。

建议提交信息：`docs: 建立 OpenAPI 接口契约`

### [ ] P5-03 · 添加 GitHub Actions 基础 CI

目标：每次 PR 和 main 更新自动执行与本地一致的基础验证。

前置：P5-02；G0 未授权时只能本地完成文件和语法检查。

只做：

- 新增 CI workflow，拆分 Web、Go、OpenAPI、Compose Config jobs。
- 使用与仓库一致的 Node、Go 版本和 lockfile 缓存。
- 设置最小 `permissions: contents: read`、并发取消和合理 timeout。
- 第三方 Action 必须查官方文档并固定到完整 commit SHA，旁注对应版本。
- CI 命令必须复用仓库脚本，不在 YAML 复制另一套逻辑。
- README 添加 CI 状态和失败排查说明；未 push 时不得伪造 badge 成功。

不得做：使用生产 Secret、SSH 部署、发布镜像、修改远端分支保护、push。

验证：

```bash
make verify
docker compose -f infra/docker-compose.yml config --quiet
git diff --check
```

人工确认：获得 push 授权后，在 GitHub 上用一次真实 workflow run 验收；未运行前任务结论只能写“仓库内实现完成，远端验证待授权”。

建议提交信息：`ci: 添加前后端基础验证`

### [ ] P5-04 · 将全栈 Mock E2E 接入 CI

目标：CI 在隔离环境真实启动全栈并走通注册到 AC 的核心流程。

前置：P5-03，且远端基础 CI 已真实运行成功。

只做：

- 使用现有 Mock Judge Compose 启动 MySQL、Redis、Mailpit、API、Worker、Web。
- 等待健康状态后执行 Playwright `full-flow.spec.ts`。
- 使用唯一 E2E 标识，保证重复运行不冲突。
- 失败时上传必要的 Playwright report 和脱敏容器日志；成功时也必须清理 Compose 资源。
- workflow 设定 timeout，服务日志不得包含密码、Cookie、Token 和源代码。

不得做：在 CI 启动 privileged Judge0、使用生产邮件、连接生产数据库、通过 retry 隐藏 flaky test。

验证：

```bash
docker compose -f infra/docker-compose.yml up -d --build
cd apps/web && npm run test:e2e
cd ../..
docker compose -f infra/docker-compose.yml down -v
git diff --check
```

人工确认：远端至少一次从干净 runner 全绿，失败场景能看到脱敏诊断材料。

建议提交信息：`ci: 添加全栈核心流程验收`

### [ ] P5-05 · 添加依赖与代码安全检查

目标：新依赖和常见安全问题在合并前被发现。

前置：P5-03；确认仓库公开性和可用 GitHub 安全功能。

只做：

- Go 使用官方 `govulncheck` 并锁定安装版本；Node 对生产依赖执行审计并记录例外策略。
- 对可用的仓库启用 Dependency Review workflow；不可用时记录限制，不伪造结果。
- 添加 CodeQL 的 Go/JavaScript-TypeScript 分析，使用最小权限和固定 Action SHA。
- 扫描失败必须显示依赖和修复方向；例外必须有 GHSA/CVE、理由、到期日。
- 添加 Dependabot 或等价的低频分组更新配置，避免每日噪音。

不得做：无评估地批量升级主版本、用全局 ignore 跳过漏洞、扫描 Secret 内容。

验证：

```bash
cd apps/server && govulncheck ./...
cd ../web && npm audit --omit=dev
cd ../..
make verify
git diff --check
```

人工确认：远端安全 workflow 实际运行；如果存在漏洞，任务必须列出处理决定，不能仅报告命令退出码。

建议提交信息：`ci: 添加依赖和代码安全检查`

### [ ] P5-06 · 发布可追溯容器镜像

目标：版本标签生成 Web/API-Worker 镜像，可追溯、不可变、可验证。

前置：P5-03 至 P5-05 全绿；G0 允许使用 GitHub Packages。

只做：

- 发布 `web` 和 `server` 两个 GHCR 镜像，Server 镜像继续同时包含 API 和 Worker 入口。
- PR 只 build 不 push；仅语义化版本 tag 或手动受控事件 push。
- 镜像标签至少包含版本和 Git SHA；生产文档优先使用 digest。
- 写入 OCI source/revision/version labels。
- 生成 SBOM 和 provenance/attestation；所有 Action 固定 SHA。
- 使用 `GITHUB_TOKEN` 最小 `packages: write` 权限，不新增长期 registry 密码。

不得做：自动部署、覆盖不可变版本、发布 `.env`、Seed 隐藏测试到 Web 镜像。

验证：

```bash
docker build -f apps/server/Dockerfile -t acmhot100-server:test .
docker build -f apps/web/Dockerfile -t acmhot100-web:test .
docker image inspect acmhot100-server:test
docker image inspect acmhot100-web:test
make verify
```

人工确认：在真实测试 tag 上检查两个 Package、digest、OCI labels 和 attestation；未经 push 授权不得执行该检查。

建议提交信息：`ci: 发布可追溯容器镜像`

### Gate A · 可信交付验收

- P5-01 至 P5-06 全部完成。
- GitHub 基础 CI、全栈 E2E 和安全扫描至少各有一次真实成功运行。
- 测试版本镜像能够按 digest 拉取并启动。
- OpenAPI lint 与 Handler 关键响应校验进入 CI。
- Gate A 未通过，不进入生产环境文件和服务器操作。

---

## 7. Phase 5B · 生产运行基线

### [ ] P5-07 · 建立生产配置与 Secret 文件契约

目标：生产 Secret 不通过镜像、Git 或普通环境变量散布，配置错误能在启动时失败。

前置：Gate A。

只做：

- 为密码、JWT Secret、SMTP 密码、Judge0 Token 支持 `*_FILE` 读取，明确定义直接值与文件值冲突规则。
- 新增不含真实值的生产配置清单和 Secret 生成说明。
- 规定 Core VM 上 Secret 源文件由专用部署用户持有、权限最小化，并记录安全轮换步骤。
- 扩展 `ValidateProduction`：域名、SMTP TLS、Judge 模式、Judge0 URL、Trusted Proxy、TTL 和 Secret 强度。
- 日志只打印配置项名称和非敏感模式，不打印值。
- 为缺失、空文件、权限错误、冲突和不安全默认值写表驱动测试。
- `.gitignore` 覆盖本地 Secret 目录，但保留 `.example` 模板。

不得做：创建真实生产 Secret、提交 Secret 样例值、选择具体 SMTP 厂商。

验证：

```bash
cd apps/server
go test ./internal/config/... -count=1
go test ./...
go vet ./...
cd ../..
make verify
git diff --check
```

人工确认：使用临时文件能启动配置校验；日志和 `docker compose config` 输出不出现 Secret 内容。

建议提交信息：`feat: 建立生产配置和 secret 文件契约`

### [ ] P5-08 · 加固 API 生命周期与健康探针

目标：慢连接、依赖故障和发布退出都有可预测行为。

前置：P5-07。

先写失败测试，再实现：

- 为 `http.Server` 配置 `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout`、`IdleTimeout`、`MaxHeaderBytes`。
- 对 JSON 写接口设置统一请求体上限；超限返回稳定错误码。
- 拆分 `/api/v1/health/live` 与 `/api/v1/health/ready`：live 不查外部依赖；ready 检查 MySQL、Redis 和必要后台能力。
- 公开探针不返回 DSN、Secret、内部地址或详细错误；Judge 模式若仍供 UI 使用，迁移到受控公开配置端点。
- 请求 ID 可接收来自唯一受信边缘代理的合法值，否则重新生成。
- 优雅退出先进入 not-ready，再等待在途请求，最后关闭连接。

不得做：监控指标、结构化日志、业务 API 改造。

验证：

```bash
cd apps/server
go test ./internal/http/... ./internal/config/... -run 'Health|Timeout|Body|RequestID|Shutdown' -count=1
go test ./...
go vet ./...
go build ./cmd/api
cd ../..
make verify
```

人工确认：停止 MySQL/Redis 时 ready 失败但 live 保持；SIGTERM 在宽限期内退出。

建议提交信息：`fix: 加固 API 生命周期和健康探针`

### [ ] P5-09 · 固定 Redis 故障下的认证与限流策略

目标：Redis 故障不能绕过撤销、会话安全或敏感接口限流。

前置：P5-08。

先写失败测试，再实现：

- `RequireAuth` 查询 deny-list 失败时返回 503，不再放行。
- `OptionalAuth` 的 Redis 故障行为明确：公开只读请求降级为匿名，不设置用户上下文。
- 登录、刷新、注册、重置密码和判题入口的限流器访问 Redis 失败时 fail closed，返回稳定 503。
- 退出接口在依赖失败时不能虚假报告会话已撤销。
- 前端对 503 显示可恢复提示，不触发 Refresh 死循环。
- 添加 Redis 故障测试，确保错误响应不泄露内部地址。

不得做：更换认证方案、把 JWT 放入 LocalStorage、增加绕过限流的内存 fallback。

验证：

```bash
cd apps/server
go test ./internal/http/... ./internal/auth/... -run 'Redis|Deny|Rate|Unavailable' -count=1
go test ./...
cd ../web
npm run test -- --run
npm run build
cd ../..
make verify
```

人工确认：断开 Redis 后受保护接口不能访问，公开题单仍可匿名读取，敏感 POST 不放行。

建议提交信息：`fix: 固定 Redis 故障安全策略`

### [ ] P5-10 · 建立安全迁移与发布纪律

目标：生产 API 多次启动不会并发迁移，失败发布可以停止和回滚。

前置：P5-08；在 P5-14 前完成。

只做：

- 生产 API 启动不再隐式执行 migration；开发模式是否保留必须显式配置并测试。
- migration 使用独立一次性任务，并通过 MySQL 锁防止并发执行。
- 添加 migration status/check 命令，能区分已应用、待应用和未知版本。
- 每个迁移在空库和已有 MVP 数据库上测试；新增迁移必须有 forward/rollback 说明。
- 发布顺序固定为：备份 → 拉镜像 → migrate → 启 API/Worker → smoke → 标记成功。
- 回滚文档区分应用回滚与数据库回滚，禁止默认执行破坏性 down migration。

不得做：自动恢复真实数据库、在 API 多副本内抢跑 migration、删除历史 migration。

验证：

```bash
cd apps/server
go test ./internal/config/... -run 'Migration' -count=1
go test ./...
cd ../..
./scripts/test-migrations.sh
git diff --check
```

`test-migrations.sh` 必须使用唯一的测试 Compose project/volume、设置退出清理，并拒绝生产 DSN；不得删除默认开发卷或用户数据。

人工确认：并发启动两个 migrate 任务只有一个执行；第二次迁移幂等；生产 API 不自动修改 Schema。

建议提交信息：`fix: 建立安全数据库迁移流程`

### [ ] P5-11 · 新增生产 Compose 与 Caddy 边缘入口

目标：用预构建镜像在单台 Core VM 安全运行，不暴露内部数据服务。

前置：P5-07 至 P5-10；G1、G2 可暂用安全占位值做语法验证。

只做：

- 保留现有开发 Compose，新增 production override 或独立 production Compose。
- 生产服务使用 registry image + digest/version，不执行 `build:`。
- 增加 Caddy 2，只有 Caddy 映射宿主机 80/443；Web、API、MySQL、Redis、监控均只在内部网络。
- 使用 Compose Secrets 将 Secret 仅授予所需服务。
- 添加 restart policy、healthcheck、日志轮转、资源上限、`no-new-privileges`、只读文件系统/tmpfs（能安全应用的服务）。
- MySQL、Redis、Caddy 证书数据使用明确持久卷；Redis 设置内存和淘汰策略，不得淘汰队列/会话数据。
- 提供 `compose.production.example.yaml`/配置模板和操作文档，不包含真实域名与 Secret。

不得做：把 Docker Socket 挂入容器、把数据库端口绑定 `0.0.0.0`、在 Core Compose 原样加入 privileged Judge0 Worker。

验证：

```bash
docker compose -f infra/docker-compose.yml config --quiet
docker compose -f infra/docker-compose.production.yml config --quiet
! docker compose -f infra/docker-compose.production.yml config | rg '0\.0\.0\.0:|privileged: true'
make verify
git diff --check
```

人工确认：最后一条风险搜索无内部服务公网暴露和 privileged Core 服务；Secret 不出现在渲染配置与 Git diff。

建议提交信息：`ops: 添加生产 compose 和边缘入口`

### [ ] P5-12 · 配置真实域名与 HTTPS

目标：公网只通过受信 Caddy 访问，HTTP 自动跳转 HTTPS，Cookie 和代理链正确。

前置：P5-11；必须具备 G1、G2；属于外部状态变更，执行前需用户授权。

只做：

- DNS A/AAAA 指向 Core VM；开放 80/443，关闭不必要公网端口。
- Caddy 持久化 ACME 数据并自动续期，反向代理 Web/API。
- API `APP_BASE_URL`、Cookie Secure、CORS、Origin、Trusted Proxies 与唯一域名一致。
- 验证 HSTS、CSP、Referrer-Policy、Permissions-Policy、X-Content-Type-Options 和页面静态资源。
- 记录 DNS、证书续期和代理排错 Runbook。

不得做：信任任意 `X-Forwarded-*`、关闭 TLS 验证、把管理/监控端点公开、记录证书私钥。

验证：

```bash
curl -I http://<DOMAIN>
curl -fsS https://<DOMAIN>/api/v1/health/live
curl -fsS https://<DOMAIN>/api/v1/health/ready
curl -I https://<DOMAIN>
openssl s_client -connect <DOMAIN>:443 -servername <DOMAIN> </dev/null
```

人工确认：HTTP 301/308 到 HTTPS；浏览器无证书告警；注册 Cookie 带 Secure/HttpOnly/SameSite；外网无法访问 3306/6379/8080/9090/Judge0 端口。

建议提交信息：`ops: 配置生产域名和 HTTPS`

### [ ] P5-13 · 接入生产 SMTP 与发信域名

目标：真实用户能可靠收到验证与重置邮件，日志不泄露 Token。

前置：P5-07、P5-12；必须具备 G3；执行前需用户授权。

只做：

- 通过 Secret 文件配置 SMTP 凭据，强制与供应商匹配的 TLS 模式。
- 使用正式 From 地址和 HTTPS 链接。
- 完成发信域名要求的 DNS 配置；具体记录由供应商给出并由用户确认。
- 对注册验证、重发、忘记密码、密码重置进行真实邮箱验收。
- 对 SMTP 超时、拒信和临时错误设置明确返回/重试策略；请求仍不能枚举邮箱。
- 记录凭据轮换、额度和故障切回维护提示流程。

不得做：提交凭据、在日志打印收件 Token、无限重试、用个人邮箱密码作为长期生产 Secret。

验证：

```bash
curl -fsS https://<DOMAIN>/api/v1/health/ready
./scripts/verify-auth-flow.sh --base-url https://<DOMAIN> --mail-mode external
```

人工确认：至少两个不同邮箱服务真实收件；链接域名和 HTTPS 正确；Token 单次使用；日志脱敏。

建议提交信息：`ops: 接入生产邮件服务`

### [ ] P5-14 · 完成 Staging 部署与发布回滚演练

目标：在不影响生产用户的环境验证预构建镜像、迁移、Smoke 和回滚。

前置：P5-10 至 P5-13；需要 Staging 主机或隔离命名空间。

只做：

- 编写版本化部署脚本：预检查、拉取 digest、备份、迁移、启动、ready、Smoke、记录版本。
- 编写应用回滚脚本，仅切回上一个兼容镜像；数据库回滚需要单独人工确认。
- 在 Staging 从空环境部署一次，再从 MVP 版本升级一次。
- 运行完整 E2E；验证日志不含 Secret；保存版本和验收结果到发布报告模板。
- 故意部署一个不能 ready 的测试版本，确认自动停止并保持旧版本可用。

不得做：直接部署生产、服务器现场 build、用 `latest`、自动执行 destructive down migration。

验证：

```bash
./scripts/deploy.sh --env staging --version <VERSION> --dry-run
./scripts/deploy.sh --env staging --version <VERSION>
./scripts/smoke.sh https://<STAGING_DOMAIN>
cd apps/web && E2E_PUBLIC_URL=https://<STAGING_DOMAIN> npm run test:e2e
```

人工确认：部署和回滚各成功一次；当前/上一版本可识别；失败版本不接管流量。

建议提交信息：`ops: 建立 staging 发布和回滚流程`

### Gate B · 生产运行验收

- P5-07 至 P5-14 全部完成。
- 生产配置校验拒绝所有开发默认值。
- 内部服务无公网端口，域名和 HTTPS 有真实证据。
- 真实 SMTP 验收通过。
- Staging 完成一次部署、升级、失败发布和回滚演练。

---

## 8. Phase 5C · 可观测、备份与恢复

### [ ] P5-15 · 接入结构化日志与请求关联

目标：一次用户请求、入队和 Worker 判题可以通过 ID 关联排查。

前置：Gate B。

只做：

- 使用 Go 标准库 `log/slog` 输出 JSON；开发可配置文本格式。
- API 请求日志包含 request_id、method、route pattern、status、duration、client class；不记录完整 URL query 和 Body。
- 提交日志可记录 submission_id、run_id、queue_message_id、worker_id、verdict，但不记录源码、测试数据、Cookie、Token、邮箱 Token。
- 错误使用稳定 code 和 cause 分类；内部错误不直接返回用户。
- 给 API 与 Worker 注入统一 logger，不保留分散的无上下文 `log.Printf`。
- 添加日志脱敏和关联字段测试。

不得做：引入重量日志平台、记录高基数字段为标签、打印完整邮箱或用户代码。

验证：

```bash
cd apps/server
go test ./... -run 'Log|Redact|RequestID' -count=1
go test ./...
go vet ./...
cd ../..
make verify
```

人工确认：完成一次提交后，可用 request_id/submission_id 串起 API、队列和 Worker；日志中搜索 Secret/源码样例无命中。

建议提交信息：`feat: 添加结构化日志和请求关联`

### [ ] P5-16 · 添加应用与判题指标

目标：可以量化 API、依赖、队列和 Worker 的健康状态。

前置：P5-15。

只做：

- 使用 Prometheus Go client，暴露仅内部网络可访问的 `/metrics`。
- API：请求数、状态码类、route pattern 延迟；不得使用原始 URL/user_id 作为 label。
- Runtime：Go 指标、MySQL 连接池、Redis 操作错误。
- Queue/Worker：队列长度、pending、最老消息年龄、reclaim、处理耗时、终态、system error。
- Judge0：请求耗时、超时、状态分布、可用 worker/queue 信息（若官方接口可用）。
- 指标名称、单位和 label 有文档；测试注册冲突和关键计数。

不得做：公网暴露 `/metrics`、把邮箱/提交 ID/题目 slug 全量作为高基数 label、用 metrics 替代业务事实表。

验证：

```bash
cd apps/server
go test ./... -run 'Metric|Prometheus' -count=1
go test ./...
go vet ./...
curl -fsS http://127.0.0.1:<INTERNAL_METRICS_PORT>/metrics | rg 'http_|queue_|judge_'
```

人工确认：正常提交、WA、Judge 超时分别改变预期指标；公网请求 `/metrics` 不可达。

建议提交信息：`feat: 添加 API 和判题指标`

### [ ] P5-17 · 建立监控面板与最小告警

目标：关键故障能在用户大量报告前被发现。

前置：P5-16；用户确认告警接收渠道。

只做：

- Staging/Production 运维 Profile 添加 Prometheus；Grafana 可选但若加入必须有预置 dashboard。
- 告警至少覆盖：API not ready、5xx 比例、Redis/MySQL 不可用、队列最老消息、pending 持续增长、Worker 消失、Judge0 无 worker、SYSTEM_ERROR、磁盘空间、备份过期、证书到期风险。
- 为告警设置持续时间，避免瞬时抖动；区分 warning/critical。
- Alertmanager 发送到用户指定渠道；凭据通过 Secret，不提交。
- 编写每条告警的含义、影响和第一响应动作。

不得做：公网无认证暴露 Prometheus/Grafana、没有 Runbook 的告警、把所有异常设为 critical。

验证：

```bash
promtool check config infra/monitoring/prometheus.yml
promtool check rules infra/monitoring/rules/*.yml
docker compose -f infra/docker-compose.production.yml --profile monitoring config --quiet
```

人工确认：人为停止 Worker 和 API，各收到一条预期测试告警；恢复后告警自动解除。

建议提交信息：`ops: 添加监控和最小告警`

### [ ] P5-18 · 实现 MySQL 自动备份与保留策略

目标：按明确 RPO 生成可校验、加密、异机保存的业务备份。

前置：P5-10、P5-17；G4 必须在真实异机备份前提供。

只做：

- 使用 MySQL 官方支持的一致性逻辑备份方式，凭据从 Secret 文件读取，不放命令行。
- 备份包含时间、应用版本、Schema 版本、校验和；压缩后加密。
- 本机临时文件权限最小化，成功上传异机存储后按策略清理。
- 默认建议 RPO 24 小时、每日备份保留 7 天、每周保留 4 周；用户可在 G4 调整。
- 备份 job 成功/失败和最新成功时间进入监控。
- 说明 Redis 丢失的影响与 MySQL Reconciliation 恢复边界，不把 Redis 当唯一业务备份。

不得做：将备份提交 Git、在日志打印密码、仅保存在同一磁盘、未经确认删除历史备份。

验证：

```bash
./scripts/backup-mysql.sh --env staging --dry-run
./scripts/backup-mysql.sh --env staging
./scripts/verify-backup.sh <BACKUP_FILE>
```

人工确认：异机目标存在加密文件、校验和正确、监控显示最新成功时间；尚无 G4 时只能标记“本地实现完成，异机验收未完成”。

建议提交信息：`ops: 添加数据库自动备份`

### [ ] P5-19 · 完成隔离恢复与灾难演练

目标：证明备份真的可用，并测出恢复时间。

前置：P5-18 有至少一个真实备份。

只做：

- 恢复只能进入全新隔离 MySQL 实例，脚本默认拒绝生产 DSN/主机名。
- 校验解密、校验和、Schema 版本、核心表数量、用户/题目/提交关联。
- 恢复后启动隔离 API/Worker，执行登录、题目读取和一次 Mock 提交 Smoke。
- 模拟 Core Redis 全丢失，验证旧会话失效、QUEUED 提交可由 MySQL Reconciliation 恢复且不重复进度。
- 记录实测 RPO、RTO、数据差异和改进项。

不得做：覆盖生产数据库、把生产用户邮件发出、对生产 Judge0 提交测试代码。

验证：

```bash
./scripts/restore-mysql.sh --target isolated --backup <BACKUP_FILE>
./scripts/verify-restored-data.sh --target isolated
./scripts/smoke.sh http://127.0.0.1:<ISOLATED_PORT>
```

人工确认：恢复数据可查询，核心行数符合备份清单，记录 RTO；演练报告不含个人信息和 Secret。

建议提交信息：`test: 完成数据库恢复和灾难演练`

### [ ] P5-20 · 完成运维 Runbook 与 SLO

目标：非原作者也能判断故障、回滚和恢复。

前置：P5-14 至 P5-19。

只做：

- 定义 Beta SLI/SLO：可用性、API 延迟、提交入队、判题最终完成、邮件成功、备份新鲜度。
- Runbook 覆盖 API/MySQL/Redis/Worker/Judge0/SMTP/磁盘/证书/队列积压。
- 每个 Runbook 包含症状、确认命令、止损、恢复、验证、升级条件。
- 写每日/每周检查清单、Secret 轮换、依赖更新、发布/回滚清单。
- 建立发布报告、事故报告、恢复演练模板。

不得做：承诺未测量的 SLO、在文档放生产地址和 Secret、用“重启全部”作为唯一处理方式。

验证：

```bash
! rg -n 'TODO|TBD|<DOMAIN>|<SECRET>' docs/runbooks docs/operations
make verify
git diff --check
```

人工确认：由另一位执行者按 Runbook 完成一次 Staging Worker 故障恢复，不依赖口头提示。

建议提交信息：`docs: 完成运维 runbook 和 SLO`

### Gate C · 可恢复运行验收

- P5-15 至 P5-20 全部完成。
- 日志可关联，指标无敏感高基数标签，关键告警真实触发过。
- 真实备份已异机保存，并在隔离环境完整恢复过。
- Runbook 经非原作者演练。

---

## 9. Phase 5D · 真实判题、容量与上线

### [ ] P5-21 · 建立生产 Judge0 隔离与私有认证

目标：用户代码只在隔离 Judge 节点执行，Judge0 API 不对公网开放且需要认证。

前置：Gate C；必须具备 G5。

只做：

- 将 Judge0 运行在独立 Judge VM/安全边界，防火墙只允许 Core Worker 私网访问 Judge0 API。
- 设置 Judge0 AUTHN Header/Token，并让 Go Adapter 从 Secret 文件发送 Header。
- 关闭不需要的 wait result、compiler options、command arguments、callbacks、additional files、batch/delete 等能力。
- 禁止用户程序网络访问，并禁止调用方重新开启网络。
- 设置 CPU、wall time、memory、process/thread、file size、queue size、worker count 上限。
- Judge 节点不能访问 Core MySQL、Core Redis、SMTP、监控写接口或 Docker Socket。
- 记录 Judge0 版本 digest、配置和升级/回滚方法。

不得做：公网暴露 2358、复用 Core Redis/Postgres、在 Core VM 原样运行 privileged Judge worker、把认证 Token 放 URL query。

验证：

```bash
curl -fsS http://<JUDGE_PRIVATE_IP>:2358/workers -H '<AUTH_HEADER>: <TOKEN>'
test "$(curl -sS -o /dev/null -w '%{http_code}' http://<JUDGE_PRIVATE_IP>:2358/workers)" = "401"
! timeout 3 bash -c '</dev/tcp/<CORE_MYSQL_PRIVATE_IP>/3306'
cd apps/server && go test ./internal/judge/... -run Judge0 -count=1
```

人工确认：无 Token 请求失败；公网不可达；Judge 执行容器无法连接 Core 数据网段；网络型用户代码不能出网。

建议提交信息：`ops: 隔离并保护生产 Judge0`

### [ ] P5-22 · 完成真实语言与判题矩阵验收

目标：C++17、Java 17、Python 3 在真实 Judge0 上产生可信终态。

前置：P5-21。

只做：

- 从目标 Judge0 `/languages` 核对并固定三种语言映射，不凭记忆硬编码 ID。
- 每种语言验证模板、Unicode、标准输入输出、时间和内存限制单位。
- 对每种语言至少验证 AC、WA、CE、TLE、RE；MLE 若环境稳定支持则真实验证。
- 验证输出比较、8KB 截断、编译信息清洗、隐藏测试不泄露。
- Judge0 超时/5xx/未知状态映射到可恢复的 SYSTEM_ERROR，不能错误标 AC/WA。
- 生成不含隐藏输入的验收报告。

不得做：放宽资源限制让测试通过、把参考解答或隐藏输入返回浏览器、使用 Mock 作为证据。

验证：

```bash
cd apps/server
JUDGE_INTEGRATION=1 go test ./internal/judge/... -run 'RealJudge0|VerdictMatrix' -count=1
go test ./...
cd ../web
npm run test -- --run
```

人工确认：3 种语言 × 5 种终态有真实 submission ID/时间证据，报告不包含用户源码和隐藏测试。

建议提交信息：`test: 验证真实 Judge0 判题矩阵`

### [ ] P5-23 · 建立可重复容量测试工具

目标：用受控负载验证 API、队列、Worker 和 Judge0，而不是手工连点。

前置：P5-16 指标、P5-22 真实判题。

只做：

- 创建独立负载工具/脚本，生成测试用户和提交，所有数据带可清理前缀。
- 支持并发数、任务数、语言、题目、超时和目标 URL 参数。
- 收集入队响应延迟、最终完成时间、状态分布、丢失、重复终态和卡住数量。
- 默认基线：20 并发、100 个任务；API 入队错误率 <1%，任务 100% 最终进入终态，无重复 Progress 副作用。
- 输出机器可读 JSON 和简短 Markdown 报告；不输出 Cookie/Token/源码。
- 提供只清理测试数据的命令，禁止模糊 DELETE。

不得做：对未授权生产环境压测、无限并发、把压测用户混入排行榜/真实统计。

验证：

```bash
./scripts/load-judge.sh --target staging --concurrency 20 --total 100
./scripts/verify-load-report.sh artifacts/load/latest.json
```

人工确认：报告总数守恒，MySQL 提交数与工具结果一致，所有任务终态且用户进度无重复更新。

建议提交信息：`test: 添加判题容量测试工具`

### [ ] P5-24 · 执行判题故障恢复与持续运行测试

目标：真实依赖中断不会造成丢任务、重复终态或永久卡住。

前置：P5-23。

只做：

- 在 Staging 压测中依次停止/恢复一个 Worker、全部 Worker、Core Redis、Judge0 API。
- 验证 XAUTOCLAIM、Reconciliation、幂等更新、队列满、Judge0 超时和 retry 上限。
- Queue 饱和时 API 返回明确 429/503，不无限接受任务。
- 失败恢复后所有可恢复任务进入终态；不可恢复任务进入 SYSTEM_ERROR 并可追踪。
- 执行至少 2 小时受控 soak test，记录内存、goroutine、连接池、队列和错误率趋势；测试必须作为可恢复的 CI/远程任务运行并记录任务 ID，不得用前台长时间 `sleep` 占住执行会话。
- 形成故障矩阵报告和未解决风险。

不得做：在生产故障注入、无限 retry、通过手工改数据库伪造终态、隐藏失败样本。

验证：

```bash
./scripts/fault-test.sh --target staging --scenario worker-restart
./scripts/fault-test.sh --target staging --scenario redis-restart
./scripts/fault-test.sh --target staging --scenario judge-timeout
./scripts/soak-test.sh --target staging --duration 2h --concurrency 5
```

人工确认：每个场景任务总数守恒；没有重复 Progress；恢复时间和告警时间写入报告。

建议提交信息：`test: 验证判题故障恢复和持续运行`

### [ ] P5-25 · 上线前安全审计与 Go/No-Go

目标：用证据决定是否开放生产，而不是以“代码完成”替代上线验收。

前置：P5-01 至 P5-24；需要 G6 才能执行真实上线。

只做：

- 按资产、入口、身份、Secret、数据、Judge 隔离、供应链、备份、监控逐项审计。
- 验证仓库 Secret 扫描、生产端口、TLS、Cookie、Origin、限流、IDOR、日志脱敏、镜像 digest。
- 运行完整 CI、Staging E2E、真实判题矩阵、容量和恢复报告检查。
- 记录已知风险、Owner、严重度和截止时间；Critical/High 未关闭则 No-Go。
- 制定开放步骤、观察指标、回滚触发条件和首日检查。
- 未获 G6 时输出 Go/No-Go 报告并停止；获得明确授权后才部署指定版本。

不得做：忽略 High 风险、在审计中修新功能、未经授权 push/deploy/改 DNS、使用 `latest`。

验证：

```bash
make verify
./scripts/security-audit.sh --env staging
./scripts/smoke.sh https://<STAGING_DOMAIN>
./scripts/release-check.sh --version <VERSION>
```

人工确认：所有 Gate 有证据；备份时间满足 RPO；告警接收人在线；回滚版本可拉取；用户明确给出 Go 或 No-Go。

建议提交信息：`docs: 完成公开 beta 上线审计`

### Gate D · Phase 5 最终完成条件

- P5-01 至 P5-25 全部完成且每项有独立提交。
- 所有自动化命令通过，远端 CI 和镜像发布有真实证据。
- Staging 和生产配置不包含开发默认值；外网端口符合清单。
- 真实 SMTP、真实 Judge0、容量和故障恢复全部通过。
- 备份和恢复有真实时间戳与报告。
- 生产开放必须有用户 G6 明确授权；没有授权时，Phase 5 可标记“技术就绪、未上线”，不能写“已正式上线”。

---

## 10. Phase 6 · 产品与内容扩展候选计划

Phase 6 只有 Gate D 通过后才能启动。开始前必须由用户重新选择优先级，不允许一次实现全部。

### Phase 6A · 管理与内容能力

1. `P6-01` 角色与权限模型：先做 Admin RBAC 和权限矩阵，默认无管理员。
2. `P6-02` 管理操作审计：所有题目、测试、用户状态变更写审计日志。
3. `P6-03` 题目管理 API：草稿/发布状态、版本、乐观锁、隐藏测试隔离。
4. `P6-04` 测试数据校验管线：格式、大小、参考解一致性、三语言回归。
5. `P6-05` 管理前端：仅消费已验收 Admin API，不把隐藏测试放普通页面缓存。
6. `P6-06` 用户管理：搜索、禁用/恢复、会话撤销；禁止查看密码和 Token。
7. `P6-07` 100 题内容扩展：先确认原创/授权，再分批导入、校验和发布。

### Phase 6B · 学习体验

1. `P6-08` 收藏：最小数据模型、列表和取消收藏。
2. `P6-09` 错题本：由非 AC 提交派生，不复制完整提交源代码到新表。
3. `P6-10` 训练计划：基于阶段和进度生成，不做虚假 AI 推荐。
4. `P6-11` 代码历史：版本上限、删除策略、隐私和存储配额先行。
5. `P6-12` 排行榜：先明确排名规则、防刷、隐私和匿名选择。
6. `P6-13` 深度移动端与无障碍：键盘、屏幕阅读器、性能预算和真实设备矩阵。

Phase 6 每项仍需在启动前扩写为与 Phase 5 同等级的单任务规格。

---

## 11. 固定验证矩阵

除任务特定命令外，每个代码任务结束前至少运行：

```bash
cd apps/web
npm run lint
npm run type-check
npm run test -- --run
npm run build

cd ../server
go test ./...
go vet ./...
go build ./cmd/api ./cmd/judge-worker

cd ../..
docker compose -f infra/docker-compose.yml config --quiet
git diff --check
```

规则：

- 只改文档时可不运行全量业务测试，但必须运行文档/配置对应校验和 `git diff --check`，并解释原因。
- 涉及 Compose/Dockerfile 时必须真实 build 或明确说明缺少的环境条件。
- 涉及外部服务时，Mock 测试只能作为前置，不能替代真实验收。
- 任一命令失败，任务不得勾选、不得提交“完成”。

---

## 12. 固定汇报格式

```text
任务：P5-xx <名称>
结论：完成 / 仓库内完成待外部验收 / 未完成
基线：<开工 commit + 工作树状态>
官方文档：<本任务实际查阅的官方链接，最多 5 个>
改动：<最多 6 条>
测试：<新增或修改的测试及其证明的行为>
验证：<逐条命令 + exit code；不得省略失败>
人工验收：<已完成证据 / 待用户提供资源>
安全检查：<Secret、权限、日志、外部暴露检查>
提交：<commit hash；未提交说明原因>
推送：未推送 / 用户明确授权后已推送
剩余风险：<真实风险；没有则写“无新增已知风险”>
下一任务：<只写编号和名称，不开始执行>
```

---

## 13. 参考的官方资料

执行模型仍需在每个任务开始时核对最新版本，不得只依赖本列表：

- GitHub Actions 发布容器镜像：<https://docs.github.com/en/actions/tutorials/publish-packages/publish-docker-images>
- GitHub Artifact Attestations：<https://docs.github.com/en/actions/how-tos/secure-your-work/use-artifact-attestations/use-artifact-attestations>
- GitHub Dependency Review：<https://docs.github.com/en/code-security/how-tos/secure-your-supply-chain/manage-your-dependency-security/configure-dependency-review-action>
- Docker Compose 生产使用：<https://docs.docker.com/compose/how-tos/production/>
- Docker Compose Secrets：<https://docs.docker.com/compose/how-tos/use-secrets/>
- Caddy Automatic HTTPS：<https://caddyserver.com/docs/automatic-https>
- Caddy Reverse Proxy：<https://caddyserver.com/docs/caddyfile/directives/reverse_proxy>
- OpenAPI Specification：<https://spec.openapis.org/oas/>
- Go `log/slog`：<https://go.dev/blog/slog>
- Prometheus Go Client：<https://github.com/prometheus/client_golang>
- Prometheus Instrumentation Practices：<https://prometheus.io/docs/practices/instrumentation/>
- MySQL 8.0 Backup and Recovery：<https://dev.mysql.com/doc/refman/8.0/en/backup-and-recovery.html>
- Judge0 官方仓库与配置：<https://github.com/judge0/judge0>、<https://github.com/judge0/judge0/blob/master/judge0.conf>
