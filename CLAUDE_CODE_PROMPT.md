# 交给 Claude Code 的第一版主提示词

把下面“主提示词”整体复制给 Claude Code。让它在本仓库根目录执行，不要只让它回答方案。

---

## 主提示词

你现在是这个仓库的实现负责人。请在当前仓库内实现一个可以真实运行的中文 ACM 在线判题网站 MVP。

### 你必须先读的文件

1. `docs/MVP_DESIGN_SPEC.md`：这是产品、视觉、数据、API、判题、安全和验收的唯一事实来源。
2. 当前仓库内所有已有文件，以及任何 `AGENTS.md`、`CLAUDE.md`、README 和配置文件。

如果代码和规格冲突，以 `docs/MVP_DESIGN_SPEC.md` 为准；如果规格内部矛盾，先指出具体章节和最小可行解释，再采用风险最低的实现继续工作。不要因为小问题停下来反复询问。

### 最终目标

实现以下完整闭环：

```text
注册/登录 → 浏览 Hot 100 题单 → 打开 ACM 题目 → 选择语言
→ 编辑完整程序 → 运行样例 → 正式提交 → 查看判题状态
→ 查看提交详情 → AC 后更新题单和个人进度
```

第一版只需要 5 道原创改写的 ACM 示例题，不要联网抓取、复制或批量导入 LeetCode 题面、题解、测试数据和 Logo。

### 固定技术约束

- Monorepo：`apps/web`、`apps/server`、`infra`、`seed`；Go Server 内含 API 和 Judge Worker 两个 `cmd`。
- Web：React + Vite + TypeScript SPA。
- 前端路由：React Router Data Mode；使用 `createBrowserRouter`、路由懒加载和 route error boundary。
- 前端服务端状态：TanStack Query；表单使用 React Hook Form + Zod；样式使用 Tailwind CSS；编辑器使用 Monaco Editor。
- MVP 不使用 Next.js、Redux、服务端渲染或 React Server Components。
- API：Go + Gin。
- ORM/数据库：GORM + MySQL Driver + MySQL；使用版本化 SQL migration，生产环境禁止依赖 GORM `AutoMigrate`。
- Redis：`go-redis/v9`；用于邮箱验证 Token、JWT Refresh Session/撤销/限流和判题 Redis Streams。
- Auth：邮箱验证注册、邮箱密码登录、`golang-jwt/jwt/v5` Access/Refresh JWT、Refresh Rotation。
- 邮件：SMTP；开发环境使用 Mailpit。
- Worker：Go，与 API 共享 `internal` domain/repository/queue 代码，通过 Redis Streams Consumer Group 消费。
- 判题适配：Judge0 HTTP API；同时提供 `JUDGE_MODE=mock` 用于开发和测试。
- 测试：Go `testing`/Testify、前端单元测试、Playwright E2E。
- 所有依赖必须锁定并提交 `go.mod`、`go.sum` 和前端 lockfile。
- 不允许业务 API 直接执行用户代码，不允许 Go `os/exec`、`exec`、`eval` 或本地 shell 编译用户程序。
- 不要引入 Celery、Kafka、RabbitMQ、Kubernetes、微服务框架或本规格之外的复杂基础设施。

### 工作方式

1. 先检查真实仓库状态和本机可用工具，不要假设目录为空。
2. 写一个可勾选的实现计划，严格按 `Phase 0 → Phase 4` 推进。
3. 每次只完成一个可验证阶段；阶段结束后运行对应测试和健康检查。
4. 发现失败时先定位根因并修复，不要用删除测试、放宽类型、写死返回值的方式绕过。
5. 保留用户已有改动，不得使用 `git reset --hard`、`git clean -fd` 等破坏性命令。
6. 不要在完成前只给我教程或代码片段；直接在当前仓库创建和修改文件。
7. 不要擅自添加排行榜、讨论区、AI 功能、付费、完整 100 题或全站暗色模式。
8. 遇到普通实现选择时按规格和最佳判断继续；只有缺少密钥、外部服务或会导致数据丢失时才询问。

### Phase 0：基础设施与骨架

必须完成：

- 创建文档规定的目录结构。
- 初始化 React + Vite + TypeScript SPA。
- 初始化 React Router、TanStack Query、React Hook Form、Zod、Tailwind、Vitest 和 Playwright。
- 配置 Vite 开发代理 `/api -> Gin`；生产使用 Nginx 提供静态文件、SPA fallback 和 `/api` 反向代理。
- 初始化 Go Module、Gin、GORM、MySQL Driver、go-redis 和共享 `internal` 包。
- 建立 `cmd/api` 和 `cmd/judge-worker` 两个入口。
- 添加 MySQL、Redis、Mailpit Docker Compose 服务、健康检查、`.env.example`。
- 添加版本化 MySQL SQL migration runner；禁止把 `AutoMigrate` 当正式迁移方案。
- 添加 Web/API/Worker 的格式化、Lint、类型检查和测试命令。
- 添加根 README，写 PowerShell 可复制的启动命令。

阶段验收：

- Web 首页可打开。
- `GET /api/v1/health` 返回健康状态。
- MySQL、Redis 连接和初始 SQL migration 成功。
- Mailpit SMTP 和 Web UI 健康。
- 后端与前端测试骨架可运行。

### Phase 1：邮箱认证、数据模型、Seed 与只读页面

必须完成：

- 按规格实现 users、problems、tags、test_cases、language_configs、drafts、submissions、submission_case_results、user_problem_progress。
- 使用 Argon2id；使用 HttpOnly SameSite Cookie，并实现 Origin/CSRF 防护。
- 实现邮箱注册、验证邮件、重发验证、邮箱密码登录、忘记密码、重置密码、当前用户接口。
- Access JWT 15 分钟、Refresh JWT 7 天；二者使用不同 Cookie 和 `aud`。
- 实现 Refresh Token Rotation、旧 Token 复用检测、单设备退出和全部设备退出。
- 邮箱验证 Raw Token 只发邮件，Redis 只存 SHA-256 哈希，TTL 30 分钟且只能使用一次。
- 密码重置使用独立 Raw Token/Redis Key，Hash-only 存储、TTL 20 分钟；成功后撤销用户全部 Refresh Token Family。
- 注册/登录/重发/刷新接口使用 Redis 限流。
- 忘记密码对存在和不存在的邮箱返回完全相同的状态与文案，禁止账号枚举。
- 实现题单、题目详情、标签、导航接口。
- 创建 5 道原创 ACM 示例题；每题至少 2 个公开样例和 8 个隐藏测试。
- 普通题目接口绝不能返回隐藏测试、参考程序或其他用户代码。
- 实现 React 全局导航、首页、题单和只读题目页；路由组件按页面懒加载。
- Auth Provider 只通过 `/auth/me` 获取用户，不得读取或保存 HttpOnly JWT。

视觉必须遵守规格：暖灰背景、白色内容面、砖橙强调色、状态色语义化、表格式题单；禁止紫色渐变、玻璃拟态、Emoji 图标和无意义卡片网格。

阶段验收：Mailpit 收到验证和重置邮件；未验证账号不能登录；验证后可登录和刷新；旧 Refresh Token 复用会撤销 Token Family；重置密码和退出后旧会话失效；题单筛选可用；隐藏测试不会出现在浏览器网络响应。

### Phase 2：编辑器、草稿和样例运行

必须完成：

- 集成 Monaco Editor。
- 支持 C++17、Python 3、Java 17，模板来自数据库配置。
- 实现语言切换保护逻辑。
- 实现本地备份 + 500ms 防抖服务端草稿保存。
- 实现题面/编辑器可拖动分栏和 LocalStorage 记忆。
- 实现 `JUDGE_MODE=mock` 的样例运行状态流。
- 实现桌面分栏和移动端题面/代码/结果 Tab。

阶段验收：编辑后刷新草稿不丢；样例运行有 Queued/Running/终态；移动端基本可用。

### Phase 3：正式提交、Worker 与 Judge0 Adapter

必须完成：

- 实现提交创建、列表、详情和状态轮询。
- API 先把 QUEUED 提交写入 MySQL，再 `XADD judge:submissions`。
- Worker 使用 Redis Streams `XREADGROUP` 消费，使用 Redis 锁和 MySQL 行锁保证幂等。
- 实现 `XADD` 失败后的 MySQL reconciliation 补发，以及 Pending 消息的 `XAUTOCLAIM` 恢复。
- 实现 Judge0 Adapter，不硬编码不可靠的语言 ID；把语言映射放入配置/Seed。
- 实现 AC、WA、CE、TLE、MLE、RE、SYSTEM_ERROR 映射。
- 实现输出规范化、8KB 截断、失败信息清理。
- 实现 Worker 超时重领和最多两次重试。
- AC 后原子更新 `user_problem_progress`；后续失败不能让 SOLVED 降级。
- 前端实现判题结果面板、轮询退避和终态停止。

阶段验收：用 Mock Judge HTTP 服务或真实 Judge0 分别覆盖 AC、WA、CE、TLE；提交历史和进度一致。

### Phase 4：质量、安全和完整交付

必须完成：

- 完成提交详情和 Profile 页面。
- 完成 API 对象级权限检查、输入长度限制、基础限流和日志脱敏。
- 完成输出比较、状态机、进度、Worker 并发领取、隐藏测试保护的单元/集成测试。
- 完成 Playwright 核心 E2E：注册 → 题单 → 草稿 → 样例 → 提交 AC → 进度更新 → 提交详情。
- 检查 1280×720、1440×900、390×844 三种视口。
- 更新 README：架构、环境变量、启动、迁移、Seed、测试、Judge0/Mock 切换和常见错误。

### 实现细节红线

- API 错误统一使用规格中的 `error + request_id` 结构。
- 公开题目接口只返回公开样例。
- 认证 Token 不得写入 LocalStorage。
- React API Client 必须使用 `credentials: 'include'`；并发 401 只能触发一次 Refresh，原请求最多重试一次，禁止死循环。
- React Router 受保护路由必须等待 `/auth/me` 完成后再判断跳转。
- Access/Refresh JWT 必须严格验证固定签名算法、`iss/aud/exp/nbf/jti`。
- Redis 不得保存邮箱验证 Raw Token，只保存 SHA-256 哈希。
- 忘记密码接口不得泄露邮箱是否存在；密码重置 Token 也只存 Hash。
- MySQL 是提交和进度的最终事实来源，Redis 队列消息不得作为唯一记录。
- Markdown 禁用原始 HTML。
- 用户源代码最大 64KB，自测输入最大 16KB。
- 日志不得包含密码、Cookie、完整隐藏测试或完整用户代码。
- 任何提交终态都不能回退到非终态。
- 没有真实数据时隐藏通过率、连续天数、排名，不允许编造。
- 不要为了“好看”加入第三方图片、插画、虚构评价或统计数字。
- 不要把整个页面写成一个超过 1000 行的组件；按路由、feature 和 domain 拆分。
- 不要用 `any`、`@ts-ignore` 或吞异常作为常规解决方案。

### 每个阶段结束时的汇报格式

只汇报可验证事实：

```text
已完成：<阶段和具体能力>
验证：<运行的命令及结果>
剩余：<下一阶段>
风险/占位：<只有真实存在时才写>
```

然后继续下一阶段。除非遇到必须由我提供外部信息的硬阻塞，不要停在计划或说明阶段。

### 最终完成标准

以 `docs/MVP_DESIGN_SPEC.md` 第 18 节为准。全部满足后，给出：

1. 是否完成的明确结论；
2. 启动命令；
3. 测试结果；
4. Mock Judge 和真实 Judge0 的切换方式；
5. 仍然存在的真实限制。

现在开始：先读取规格和仓库，建立计划，然后直接实施 Phase 0。

---

## 使用建议

如果 Claude Code 一次上下文不足，不要重新解释整个项目。下一轮只发：

```text
继续当前实现。重新读取 docs/MVP_DESIGN_SPEC.md 和 CLAUDE_CODE_PROMPT.md，检查 git diff、测试结果和未完成计划，从第一个未通过的验收项继续。不要重做已完成且测试通过的部分。
```
