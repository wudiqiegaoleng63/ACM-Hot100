# ACM Hot 100 · 后续实施计划

> 面向执行能力弱于主设计者的编码模型。每次只执行一个任务；每个任务都必须有独立验证和独立 Git 提交。
> 基线审计日期：2026-07-13。
> 基线提交：`5f5045e feat: Phase 1 - Auth, data models, seed data, and read-only pages`。

## 0. 使用规则

1. `docs/MVP_DESIGN_SPEC.md` 仍是产品与架构的唯一事实来源。
2. 本文件决定从当前提交继续实现的顺序，不允许跳过验收闸门。
3. 每次只执行第一个未完成任务；验证通过并提交后立即停止，不自动开始下一个任务。
4. 开工前必须读 `CLAUDE.md`，按其中要求查询所用库的最新官方文档。
5. 禁止顺手重构无关模块、改技术栈、加入新业务功能或补完整 100 题。
6. 不得把“能编译”写成“测试通过”；没有测试文件必须明确报告为没有覆盖。
7. 计划中的 `[ ]` 只有在验证命令真实通过后才能改为 `[x]`；勾选与代码放在同一个任务提交中，提交哈希只在执行汇报中记录。

## 1. 已核实的真实基线

### 1.1 已通过

- Git 工作树干净，当前分支 `main`。
- `go test ./...`、`go vet ./...`、API/Worker build 当前退出码为 0。
- 但 Go 输出全部为 `[no test files]`，这只证明编译成功，不代表业务逻辑被验证。
- `npm run build` 通过。
- API `GET /api/v1/health` 返回 MySQL/Redis `ok`。
- Web `http://127.0.0.1:5173` 返回 200。
- API 能返回 5 道题和 10 个标签；题目详情只返回公开样例。

### 1.2 当前失败

- `npm run test -- --run` 失败：没有任何测试文件。
- `npm run lint` 失败：ESLint 9 缺少 `eslint.config.js/mjs/cjs`。
- `docker compose -f infra/docker-compose.yml ps` 当前没有运行中的 Compose 服务；现有 API/Web 使用宿主机进程与本地 MySQL/Redis。

### 1.3 不能再视为“已完成”的部分

- `ProblemListPage`、`ProblemDetailPage`、`ProfilePage`、`SubmissionsPage`、`SubmissionDetailPage` 仍是 `coming soon` 占位页。
- `cmd/judge-worker/main.go` 只有等待退出信号的占位循环。
- `internal/queue/queue.go` 没有 Redis Streams 实现。
- 前端尚未安装 `@monaco-editor/react`。
- Router 已定义 Profile/提交页面，但没有实际套用 `ProtectedRoute`。
- README 仍有 Vue/React 混写、错误 Worker 路径和迁移命令格式。

### 1.4 开始 Phase 2 前必须修复的高风险点

1. **Refresh Rotation Family 不一致**：轮换时生成的新 Refresh JWT 自带新 Family ID，但 Redis 保存的是旧 Family ID；下一次刷新会发生不一致。
2. **邮箱验证/重置 Token 非原子消费**：当前为 `GET` 后 `DEL`，并发请求可能重复使用同一个 Token；应使用 `GETDEL` 或 Lua。
3. **开发环境不会发到 Mailpit**：`APP_ENV=development` 或 `JUDGE_MODE=mock` 时邮件只写日志。
4. **前后端错误码不一致**：前端判断 `EMAIL_ALREADY_EXISTS/USERNAME_ALREADY_EXISTS`，后端返回 `EMAIL_TAKEN/USERNAME_TAKEN`。
5. **前后端题目 DTO 不一致**：字段名、Difficulty 大小写、Tag 结构、状态名称、内存单位均不一致。
6. **API Client 会对登录失败的 401 尝试刷新**：登录/刷新等认证端点需要明确跳过自动 Refresh。
7. **题面含数学标记但前端没有数学渲染**：Seed 使用 `$...$`，仅 `react-markdown` 会显示原始标记。

## 2. 依赖顺序

```text
Phase 1.5：建立可信基线
  ↓
Phase 2A：语言配置 + 编辑器 + 做题布局
  ↓
Phase 2B：草稿 + Mock 样例运行
  ↓
Phase 3A：提交领域 + Redis Streams + Worker
  ↓
Phase 3B：Judge0 + 前端结果/历史/进度
  ↓
Phase 4：安全、E2E、全栈 Compose、视觉与文档
```

任何阶段失败时，回到该阶段第一个未通过任务。禁止在基础测试失败时继续堆新功能。

---

## 3. Phase 1.5 · 基线修复与验收闸门

### [x] P15-01 · 修复工程验证命令与文档漂移

目标：让仓库声明的所有基础命令真实可运行，文档与代码一致。

只做：

- 查询 ESLint 9 Flat Config 官方文档，新增正确的 `eslint.config.*`。
- 补齐 `npm run type-check` 脚本；不得用关闭规则的方式制造通过。
- 修复 README：前端只写 React；Worker 路径改为 `cmd/judge-worker`；迁移/Seed 命令与 `flag` 实际语法一致。
- 核对 README 的 Go 最低版本与 `go.mod`，选择一个真实一致的版本并说明原因。
- README 明确区分“宿主机开发启动”和“Docker 基础设施启动”。

不得做：认证、页面、Monaco、Judge 业务修改。

验证：

```powershell
cd apps/web
npm run lint
npm run type-check
npm run build

cd ../server
go test ./...
go vet ./...
go build ./cmd/api ./cmd/judge-worker
```

建议提交信息：`chore: 修复工程验证命令和开发文档`

### [x] P15-02 · 建立最小但真实的测试基线

目标：测试命令不能再因为“没有测试文件”失败或产生虚假安全感。

只做：

- 前端添加 API Client 或 Router 的第一个真实 Vitest 测试。
- 后端为 JWT claim 校验和 Token Hash/生成添加表驱动测试。
- 测试必须验证行为，不允许只写 `expect(true).toBe(true)`。
- 如果加入测试依赖，先查询最新官方文档并锁定版本。

验证：

```powershell
cd apps/web
npm run test -- --run
npm run lint
npm run build

cd ../server
go test ./...
go vet ./...
```

建议提交信息：`test: 建立前后端测试基线`

### [x] P15-03 · 修复 JWT Refresh Rotation 与 API Client 刷新边界

目标：连续刷新、旧 Token 复用检测与登录失败行为正确。

先写失败测试，再实现：

- 新 Refresh JWT 必须保留原 Token Family ID；不得生成后丢弃新旧 Family 差异。
- 连续刷新至少两次都成功。
- 第一次轮换后的旧 Token 再次使用会撤销整个 Family。
- Access/Refresh 的 issuer、audience、算法和 JTI 必须严格校验。
- Refresh API 只从 HttpOnly Cookie 读取 Refresh Token，不从 JSON Body 接收。
- 前端 API Client 对 `/auth/login`、`/auth/register`、`/auth/refresh` 等认证失败不得触发自动 Refresh。
- 并发普通请求 401 时仍只能产生一个 Refresh 请求，原请求最多重试一次。

验证：

```powershell
cd apps/server
go test ./internal/auth/... ./internal/service/... -run "Refresh|JWT|Reuse" -count=1
go test ./...

cd ../web
npm run test -- --run
npm run build
```

建议提交信息：`fix: 修复 JWT 刷新轮换和客户端刷新边界`

### [x] P15-04 · 原子化邮箱 Token 并打通 Mailpit 认证闭环

目标：验证邮件、重置邮件和单次 Token 行为得到端到端证据。

先写失败测试，再实现：

- 邮箱验证与密码重置 Token 使用 Redis `GETDEL` 或 Lua 原子消费。
- 并发使用同一 Token 时只有一个请求成功。
- Development 环境在配置 SMTP Host 后真实发送到 Mailpit；只有显式 `MAIL_MODE=log` 才写日志替代。
- 宿主机开发 `.env.example` 使用 `localhost:1025`；容器环境通过 Compose 覆盖为 `mailpit:1025`。
- 统一注册错误码，前后端使用相同常量。
- 正确区分“密码正确但邮箱未验证”与普通错误，同时避免通过错误密码枚举账号。
- 新增可重复的认证集成验证：注册 → Mailpit 取验证链接 → 验证 → 登录 → 连续刷新 → 退出 → 忘记/重置密码 → 旧会话失效。

验证必须包含 Mailpit 真实收件证据，不能只看日志。

建议提交信息：`fix: 打通邮箱认证与密码重置闭环`

### [x] P15-05 · 固定题目与语言 API 契约

目标：在编写页面和编辑器前消除前后端 DTO 漂移。

先做一份单一 API DTO 定义，再实现：

- 明确 JSON 使用的字段：`order_index`、`difficulty`、`tags.slug`、`progress_state`、Markdown 字段、`memory_limit_kb`。
- 前后端统一 Difficulty/Progress 的大小写和枚举值。
- 添加公开 `GET /api/v1/languages`，从 `language_configs` 返回启用语言、Editor Language 和 Source Template；不返回无关内部字段。
- 题目详情继续只返回 `is_sample=true` 的测试数据。
- 前端类型必须与真实 JSON 一致，不允许用错误类型后在组件里猜字段。
- 添加后端响应测试与前端 API 解析测试。

验证：

```powershell
cd apps/server
go test ./internal/http/... ./internal/repository/... -count=1
go test ./...

cd ../web
npm run test -- --run
npm run type-check
npm run build
```

建议提交信息：`fix: 统一题目和语言 API 契约`

### [x] P15-06 · 完成 Phase 1 只读页面与路由保护

目标：把已声明完成但仍为占位的首页、题单和题目详情真正做完。

只做：

- 首页实现登录/未登录双状态，不增加虚构统计。
- 题单使用真实 API 表格、搜索、难度/标签/状态筛选，并同步 URL。
- 题目详情渲染真实 Markdown、公开样例、限制、标签与上一题/下一题。
- Seed 含数学标记，按官方文档接入 `remark-math + rehype-katex` 或等价可靠方案。
- Profile、Submissions 和 SubmissionDetail 在未实现前保留路由，但必须套 `ProtectedRoute` 并展示明确“后续阶段”空状态。
- 删除所有 `coming soon` 英文占位文案。
- 添加页面 loading/error/empty 状态测试。
- 用 Playwright 或浏览器检查 1280×720、1440×900、390×844；保存截图到临时验证目录，不提交无必要的图片。

验收闸门：

```powershell
cd apps/web
npm run lint
npm run type-check
npm run test -- --run
npm run build

cd ../server
go test ./...
go vet ./...
```

还必须人工确认：题单不再是占位页、数学公式可读、隐藏测试不出现在 Network Response。

建议提交信息：`feat: 完成 Phase 1 只读页面和路由保护`

> 只有 P15-01 至 P15-06 全部完成，才允许进入 Phase 2。

---

## 4. Phase 2A · 编辑器与做题页面

### [x] P2-01 · 接入 Monaco 基础组件

目标：只完成可复用 CodeEditor，不同时重写整页。

- 查询 `@monaco-editor/react` 当前官方用法并安装。
- 新建 `features/editor/components/CodeEditor.tsx`。
- Props 固定为：`value`、`language`、`onChange`、`fontSize`、`readOnly`。
- 深色主题使用现有 Editor Token；字号限制 13–20px，默认 14px，行高 1.6。
- 处理加载态和 Monaco 加载失败态。
- 添加组件测试；不得在测试中加载真实 Web Worker 才能完成的复杂能力。

建议提交信息：`feat: 添加 Monaco 代码编辑器组件`

### [x] P2-02 · 语言选择与切换保护

- 使用 `/languages` 的真实配置初始化模板和 Monaco Language。
- 当前代码等于旧模板时可直接切换。
- 当前代码已修改时必须确认；取消后语言和代码都不变。
- 不得把三种模板重新硬编码在前端。
- 添加“未修改直接切换、已修改确认、已修改取消”测试。

建议提交信息：`feat: 实现语言切换保护`

### [x] P2-03 · 做题页桌面分栏与移动 Tab

- 题面/编辑器默认 46/54；拖动范围 32%–68%。
- 两侧独立滚动，分栏比例写入 LocalStorage。
- 底部区域先实现 `自测输入 / 输出 / 结果` 空壳，不接判题。
- 桌面显示上一题/下一题、运行样例、正式提交；未登录点击运行/提交跳登录并保留返回地址。
- `<768px` 使用题面/代码/结果三个 Tab。
- 添加 Resizer 边界和移动 Tab 测试。

建议提交信息：`feat: 完成做题页响应式工作区`

## 5. Phase 2B · 草稿与 Mock 样例运行

### [x] P2-04 · 实现可恢复草稿

规则必须固定：

- LocalStorage Key：`draft:{user-or-guest}:{problemSlug}:{languageKey}`。
- 本地草稿包含 `source_code` 和 `updated_at`。
- 登录用户同时加载服务端草稿，取时间更新的一份；Guest 只用本地草稿。
- 输入后立即保存本地，500ms 防抖保存服务端。
- 切题、切语言和卸载时处理未完成保存，不得把 A 题草稿写到 B 题。
- 服务端失败显示非阻塞警告，本地草稿继续保留。
- 草稿最大 64KB，前后端同时限制。

建议提交信息：`feat: 实现本地和服务端草稿恢复`

### [x] P2-05 · 建立 Sample Run 数据与 Mock Worker

目标：样例运行也使用异步任务，不在 Gin Handler 中 `sleep` 或执行代码。

- 添加版本化 migration：`sample_runs`，保存用户、题目、语言、源码快照、输入、状态、输出、错误和 TTL/时间。
- 新增 `POST /problems/:slug/run` 与 `GET /runs/:id`，均要求登录。
- 校验源码 64KB、自定义输入 16KB、语言启用状态和 Sample Case 所属题目。
- 使用独立 Redis Stream `judge:runs`；消息只放 `run_id`。
- `cmd/judge-worker` 在 `JUDGE_MODE=mock` 下消费 Sample Run，500–1000ms 后写入 AC；不得在 API 进程启动 goroutine 假装异步。
- 先实现最小 Consumer Group、幂等检查和 XACK，为 Phase 3 复用。

建议提交信息：`feat: 添加异步 Mock 样例运行`

### [ ] P2-06 · 样例运行前端闭环

- 接入公开样例选择和自定义输入。
- 状态为 Queued → Running → 终态；开发环境显示 `Mock Judge`。
- 轮询遇终态停止，组件卸载时停止，网络失败可重试。
- 自测不更新正式提交历史和 Progress。
- 添加前端轮询、终态停止、输入上限测试。

Phase 2 验收：刷新后草稿不丢；三种语言可切换；样例运行完整；移动端可用；所有全局命令通过。

建议提交信息：`feat: 完成样例运行交互闭环`

---

## 6. Phase 3A · 正式提交与判题内核

### [ ] P3-01 · Submission Repository 与 API

- 实现创建提交、本人列表、本人详情。
- 创建时在 MySQL 事务中写入 `QUEUED` 和 `ATTEMPTED` Progress。
- 只接受启用语言，源码最大 64KB。
- 提交源码为不可变快照。
- 列表支持 problem/status/language/page；详情严格校验 User ID。
- 添加 IDOR、输入限制和 Progress 测试。

建议提交信息：`feat: 添加正式提交 API`

### [ ] P3-02 · Redis 提交入队与 Reconciliation

- MySQL 提交成功后 `XADD judge:submissions submission_id=...`。
- XADD 成功后回写 `stream_message_id/enqueued_at`。
- XADD 失败不删除 MySQL 提交。
- Reconciliation 扫描缺少 `enqueued_at` 的 QUEUED 记录并幂等补发。
- 并发补发不能产生重复终态副作用。

建议提交信息：`feat: 实现提交队列与丢消息补偿`

### [ ] P3-03 · Judge Adapter 接口、Mock 与比较器

- 定义小接口，不让 Worker 直接依赖 HTTP 细节。
- 实现确定性的 Fake/Mock Adapter，测试可注入 AC/WA/CE/TLE/MLE/RE/SYSTEM_ERROR。
- 输出比较：换行统一、行尾空白移除、末尾空行移除，其余严格比较。
- stdout/stderr 8KB 截断，宿主路径清理。
- 状态机终态不可回退。

建议提交信息：`feat: 添加判题适配接口和结果比较器`

### [ ] P3-04 · Worker 消费、幂等与崩溃恢复

- XREADGROUP 消费 `judge:submissions`。
- Redis Lock + MySQL 行锁确认 QUEUED 后进入 COMPILING。
- 结果、Case Results、Progress 在一个 MySQL 事务落库。
- AC 设置 SOLVED；后续失败不降级。
- 成功落库后才 XACK。
- XAUTOCLAIM 恢复 5 分钟 Pending；最多重试 2 次后 SYSTEM_ERROR。
- Worker 收到 SIGTERM 停止领取新任务并等待当前任务在限制时间内结束。

建议提交信息：`feat: 实现幂等判题 Worker`

## 7. Phase 3B · Judge0 与用户可见闭环

### [ ] P3-05 · Judge0 HTTP Adapter 与基础设施

- 查询 Judge0 当前官方 API。
- Language ID 从数据库配置读取，不硬编码在 Adapter。
- 设置时间、内存和输出限制；HTTP Client 有连接/请求/总超时。
- 映射 AC/WA/CE/TLE/MLE/RE/SYSTEM_ERROR。
- `infra` 增加明确的 Judge0 启动方案；Judge0 不暴露公网。
- 使用参考程序对至少 2 道题跑真实 AC，并构造 WA/CE/TLE 验证。

建议提交信息：`feat: 集成 Judge0 判题服务`

### [ ] P3-06 · 正式提交与结果面板

- 正式提交后切换到判题结果。
- 前 10 秒每 800ms，之后每 2 秒，60 秒停止并显示刷新按钮。
- 每种状态有文字、图形和颜色，不只靠颜色。
- WA 仅显示允许的失败信息；隐藏输入不泄露。
- CE/RE 输出截断、清理并支持复制。
- AC 后刷新题单和 Profile Query。

建议提交信息：`feat: 完成正式提交和判题结果面板`

### [ ] P3-07 · 提交历史、详情与 Profile

- 实现本人提交列表筛选和分页。
- 实现详情：状态、指标、代码、允许显示的错误、复制、回题修改。
- 实现 Profile 总进度、三态分布和阶段分组。
- 未登录路由保护；其他用户提交返回 404 或 403，策略统一。

Phase 3 验收：Mock/真实 Judge0 覆盖 AC、WA、CE、TLE；队列恢复测试通过；提交历史与 Progress 一致。

建议提交信息：`feat: 完成提交历史和训练进度`

---

## 8. Phase 4 · 安全、测试与完整交付

### [ ] P4-01 · 安全硬化

- Redis 限流：注册、登录、重发、刷新、忘记密码、样例运行、正式提交。
- 写接口 Origin/CSRF 防护；开发与生产白名单分离。
- Markdown 禁止原始 HTML，过滤危险链接协议。
- 日志脱敏；生产禁止默认 JWT Secret、Root MySQL 用户和开放 Redis。
- Cookie Secure/Proxy Trust 规则明确，不能无条件信任任意 `X-Forwarded-Proto`。

建议提交信息：`fix: 加固认证和输入安全边界`

### [ ] P4-02 · 自动化验收套件

- Go：认证、Token 原子性、JWT Rotation、输出比较、状态机、IDOR、Queue 幂等、Reconciliation、XAUTOCLAIM。
- React：路由守卫、401 单次刷新、语言切换、草稿防抖、轮询停止。
- Playwright：注册 → Mailpit 验证 → 登录 → 题单 → 草稿 → 样例 → 提交 AC → Profile → 提交详情。
- 测试失败不得通过 retry 或 skip 隐藏。

建议提交信息：`test: 添加核心业务自动化验收`

### [ ] P4-03 · 全栈 Compose、文档与最终视觉验收

- `docker compose up --build` 启动 Web、API、Worker、MySQL、Redis、Mailpit，并提供可选 Judge0 Profile。
- 容器使用内部主机名，宿主机开发使用 localhost；两套环境变量不混淆。
- 修正 README 全部命令，记录 migration/seed/mock/judge0/测试/排错。
- 检查 1280×720、1440×900、390×844。
- 浏览器 console 无阻断错误。
- 干净环境完整执行 MVP 完成定义。

建议提交信息：`docs: 完成全栈部署和最终验收`

## 9. 每个任务的固定结束格式

执行模型完成一个任务后只能按以下格式汇报，然后停止：

```text
任务：Pxx-xx <名称>
结论：完成 / 未完成
改动：<最多5条>
验证：<逐条命令 + exit code>
提交：<commit hash；未提交则说明原因>
剩余风险：<真实存在才写>
下一任务：<只写编号和名称，不开始执行>
```

## 10. 最终完成条件

除 `docs/MVP_DESIGN_SPEC.md` 第 18 节外，还必须满足：

- `npm run lint`、`npm run type-check`、`npm run test -- --run`、`npm run build` 全部通过。
- `go test ./...` 中存在真实测试且全部通过；`go vet ./...` 和两个 Go 入口 build 通过。
- 连续 Refresh、旧 Refresh 复用、邮箱 Token 并发消费都有自动化测试。
- Redis 入队失败可补发，Pending 可恢复，重复消息不重复更新 Progress。
- README 不再包含错误框架、路径或命令。
- 仓库工作树干净，每个任务有独立提交。
