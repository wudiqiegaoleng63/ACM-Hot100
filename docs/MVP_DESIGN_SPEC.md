# ACM Hot 100 · 第一版完整设计与实现规格

> 文档用途：作为第一版产品、交互、视觉、数据、API、判题与验收的唯一事实来源。
> 目标读者：负责实现本项目的编码模型或工程师。
> 版本：MVP v1，2026-07-13。

## 0. 执行摘要

ACM Hot 100 是一个面向算法学习者的中文在线判题网站。它借鉴主流刷题网站的高效做题流程，但题目全部采用 ACM 标准输入输出形式：用户提交完整程序，程序从 `stdin` 读取输入并向 `stdout` 输出答案。

第一版只验证一条完整闭环：

```text
注册/登录 → 浏览题单 → 打开题目 → 编写完整程序 → 运行样例 → 正式提交
→ 查看判题状态 → 查看失败原因 → AC 后更新题单与个人进度
```

第一版必须能够真实运行，但不追求一次录入全部 100 道题。先放入 5 道自行编写 ACM 题面和测试数据的示例题，确认产品和判题闭环稳定后，再扩充内容。

## 1. 固定假设

- 主要用户：准备实习、校招、复试或算法竞赛入门的中文用户。
- 主要设备：宽度 1280px 以上的笔记本或桌面显示器。
- 移动端用途：查看题目和提交记录；移动端写代码只保证可用，不作为主要体验。
- 第一版语言：C++17、Python 3、Java 17。具体 Judge0 语言 ID 不允许硬编码，应根据语言名称初始化或写入配置。
- 第一版题目：5 道原创改写的 ACM 示例题；后续扩充至 100 道。
- 后端技术栈：Go、Gin、GORM、MySQL、Redis。
- 认证方式：邮箱验证注册 + 邮箱密码登录 + JWT Access/Refresh Token。
- 判题执行：使用 Go Judge Worker 从 Redis Streams 消费任务并调用自托管 Judge0；业务 API 进程不得直接执行用户代码。
- 主题：浅色题面、深色代码编辑器；第一版不实现完整暗色模式。
- 品牌：使用临时文字标识 `ACM HOT 100`，不得复制 LeetCode 的 Logo、商标素材或整套视觉资产。
- 内容：不得自动抓取或整段复制第三方题面、题解和测试数据；正式公开前需单独确认内容授权方案。

## 2. 产品目标与非目标

### 2.1 MVP 必须实现

1. 用户邮箱注册、邮箱验证、登录、JWT 刷新、退出和会话恢复。
2. Hot 100 题单浏览、搜索、难度/标签/状态筛选。
3. ACM 题目详情：描述、输入、输出、约束、样例、提示。
4. Monaco 代码编辑器、语言切换、自动保存草稿。
5. “运行样例”和“正式提交”两个不同动作。
6. 判题状态：Queued、Compiling、Running、AC、WA、TLE、MLE、RE、CE、System Error。
7. 提交列表和提交详情。
8. 用户做题进度：未开始、尝试中、已通过。
9. Docker Compose 一键启动本地依赖和服务。
10. 自动化测试与最小 E2E 测试。

### 2.2 第一版明确不做

- 排行榜、比赛、Rating、虚拟参赛。
- 讨论区、评论、私信和关注。
- AI 讲题、AI Debug、代码补全。
- 付费、订阅、会员体系。
- Special Judge、交互题、多文件题、输出任意合法答案的题。
- 全站暗色主题。
- 复杂后台管理界面。
- 一次性录入完整 100 道题。
- 微服务拆分、Kubernetes、事件总线等超前架构。

## 3. 产品原则

### 3.1 ACM 优先

每道题必须有明确的：

- 输入格式；
- 输出格式；
- 数据范围；
- 是否多组输入；
- 输出是否唯一；
- 时间和内存限制；
- 样例解释；
- 至少一个参考程序；
- 覆盖边界情况的隐藏测试。

不能把“函数签名题”机械地包一层输入输出后发布。

### 3.2 训练优先

最重要的信息依次是：当前题目、代码、判题反馈、学习进度。装饰性 Banner、虚构统计数字、无意义图标都不得挤占主视区。

### 3.3 反馈必须可行动

- CE：显示经过长度限制和敏感信息过滤的编译器输出。
- WA：显示通过测试点数量；隐藏测试默认不泄露完整输入。
- TLE/MLE：显示限制值和本次最大消耗。
- RE：显示运行错误摘要，不暴露宿主机路径。
- System Error：告诉用户是系统问题，可以重新提交，不计入错误尝试。

### 3.4 内容与品牌独立

产品可以借鉴成熟刷题产品的流程，但视觉、文案、题面和测试数据应形成自己的体系。页面底部应写明“独立学习项目，非 LeetCode 官方产品”。

## 4. 核心用户流程

### 4.1 首次用户

```text
进入首页 → 点击“开始训练” → 注册 → 进入题单
→ 默认打开推荐的第一道未完成题 → 选择语言
→ 阅读题面 → 编码 → 运行样例 → 提交 → 查看结果
```

### 4.2 回访用户

```text
登录 → 首页显示“继续上一题” → 恢复语言和草稿
→ 提交 → AC → 自动建议同一训练阶段的下一题
```

### 4.3 失败恢复

```text
提交 → WA/TLE/CE → 打开判题详情抽屉
→ 查看可行动反馈 → 返回编辑器，代码和滚动位置不丢失
→ 修改 → 再次提交
```

## 5. 信息架构与路由

| 路由 | 页面 | 权限 | 作用 |
|---|---|---|---|
| `/` | 首页 | 公开 | 产品介绍；登录后显示训练摘要 |
| `/problems` | Hot 100 题单 | 公开 | 搜索、筛选、状态与进度 |
| `/problems/[slug]` | 做题页 | 公开浏览；提交需登录 | 阅读题面、编码、自测、提交 |
| `/submissions` | 我的提交 | 登录 | 查询提交历史 |
| `/submissions/[id]` | 提交详情 | 本人 | 查看代码和判题结果 |
| `/profile` | 个人进度 | 登录 | 训练阶段和题型掌握情况 |
| `/login` | 登录 | 未登录 | 登录 |
| `/register` | 注册 | 未登录 | 注册 |
| `/verify-email` | 验证邮箱 | 持有验证 Token | 完成邮箱验证或重新发送 |
| `/forgot-password` | 忘记密码 | 未登录 | 申请密码重置邮件 |
| `/reset-password` | 重置密码 | 持有重置 Token | 设置新密码并撤销旧会话 |

导航只保留：品牌、题库、提交记录、进度、用户菜单。首页不使用传统“巨型 Hero + 三列功能卡 + 用户评价”模板。

## 6. 页面级设计

### 6.1 全局框架

- 顶部导航高度：56px；桌面端固定在顶部。
- 内容最大宽度：题单/首页 1200px；做题页占满视口。
- 背景：暖灰白；主要内容面为纯白。
- 边界：使用 1px 中性分隔线，减少悬浮卡片。
- 图标：只使用 Lucide 等统一线性图标；没有语义价值时不放图标。
- 页面切换不使用大幅动画；只保留 120–180ms 的颜色、透明度和位移反馈。

### 6.2 首页 `/`

叙事角色：入口和继续训练，不是营销落地页。

未登录状态：

- 左侧主标题：“用 ACM 模式刷完 Hot 100”。
- 一段不超过 60 字的说明。
- 主按钮“开始训练”，次按钮“浏览题单”。
- 右侧展示一段真实的 ACM 输入、输出和代码片段，而不是插画。
- 页面下方只解释三个差异：完整程序、真实判题、路线化进度。

登录状态：

- 主区显示“继续训练：题目标题”。
- 显示 `已通过 x / 100`、当前训练阶段、最近一次提交结果。
- 提供“继续上一题”和“打开题单”。
- 数据为空时展示诚实空状态，不虚构连续天数或排名。

### 6.3 题单 `/problems`

桌面端使用表格，不使用卡片瀑布流。

顶部：

- 标题 `ACM Hot 100`；
- 真实进度条；
- 搜索框；
- 难度、标签、状态筛选；
- “重置筛选”。

表格列：

| 列 | 说明 |
|---|---|
| 状态 | 空心圆=未开始；琥珀点=尝试中；绿色对勾=已通过 |
| 序号 | 训练路线中的顺序 |
| 题目 | 标题 + 最多两个关键标签 |
| 难度 | 简单/中等/困难，颜色只作为辅助 |
| 通过率 | 第一版可不展示；没有真实数据时必须隐藏 |
| 最近提交 | 相对时间；未登录时隐藏 |

默认排序为训练顺序。筛选结果必须同步到 URL 查询参数，刷新后保留。

移动端改为紧凑列表，每行显示状态、题目、难度和首个标签。

### 6.4 做题页 `/problems/[slug]`

这是第一版最重要的页面。

桌面布局：

```text
┌──────────────────── 顶部导航 56 ────────────────────┐
├──────────── 题面 46% ───────────┬──── 编辑器 54% ────┤
│ 题目标题 / 难度 / 标签           │ 语言 / 重置 / 设置   │
│ 描述                             │                     │
│ 输入格式                         │ Monaco Editor       │
│ 输出格式                         │                     │
│ 数据范围                         ├─────────────────────┤
│ 样例与解释                       │ 自测输入 / 输出 / 结果│
├─────────────────────────────────┴─────────────────────┤
│ 上一题 / 下一题        运行样例     正式提交             │
└───────────────────────────────────────────────────────┘
```

具体规则：

- 主分栏默认 46/54，可拖动，限制在 32%–68%，比例写入 `localStorage`。
- 题面和编辑器分别滚动，不能互相带动。
- 题面使用 Markdown 渲染，但禁用原始 HTML。
- 题面正文宽度控制在 72 个中文字符以内。
- 代码编辑器默认字体 14px、行高 1.6，可在 13–20px 调整。
- 语言切换时：若当前代码仍等于旧语言模板，直接替换；若用户已修改，弹出确认框。
- 草稿按 `userId + problemId + language` 保存；输入防抖 500ms 后同时写本地和服务器。
- “运行样例”只运行公开样例或用户自定义输入，不改变正式进度。
- “正式提交”创建不可变提交快照。
- 提交后右侧底部面板自动切到“判题结果”，但不得让编辑器内容丢失。

移动端：题面、代码、结果使用三个 Tab；底部保留固定的运行和提交按钮。

### 6.5 判题结果面板

状态展示：

- Queued：灰色，文案“等待判题”。
- Compiling：蓝色，文案“正在编译”。
- Running：蓝色，显示 `当前测试点 / 总测试点`。
- AC：绿色，显示总耗时、最大内存、通过测试点。
- WA：红色，显示首个失败测试序号和规范化后的实际/期望输出；隐藏测试输入默认不展示。
- TLE/MLE/RE/CE：红色，显示对应限制与摘要。
- System Error：中性红，显示“系统判题失败，可重新提交”。

编译输出、实际输出和期望输出使用等宽字体，并设最大高度和复制按钮。输出超过 8KB 时截断并明确标注。

### 6.6 提交列表与详情

提交列表列出：提交时间、题目、语言、状态、耗时、内存。支持按题目、状态、语言筛选。

提交详情包括：

- 状态摘要；
- 判题指标；
- 提交源代码；
- 编译器输出或首个失败信息；
- “复制代码”和“回到题目继续修改”。

用户只能读取自己的提交。不得在公开接口返回其他用户源代码。

### 6.7 个人进度 `/profile`

- 总进度：已通过数量 / 已发布题目数量。
- 三态分布：未开始、尝试中、已通过。
- 按训练阶段展示：数组与哈希、链表与栈、树与回溯、二分与贪心、图与动态规划。
- 标签掌握度只根据真实 AC 数量计算，不给虚构评分。
- “需要复习”“连续刷题”“二刷”都放到后续版本。

### 6.8 注册、验证与登录页面

- 注册字段：邮箱、用户名、密码、确认密码；密码规则在输入前可见。
- 注册成功后进入“检查邮箱”状态，显示脱敏邮箱和 60 秒重发倒计时。
- 验证页处理：验证中、成功、Token 过期、Token 已使用四种状态；成功后跳转登录。
- 登录字段：邮箱、密码；未验证账号显示“重新发送验证邮件”，但不得泄露不存在的邮箱。
- 忘记密码始终返回统一成功文案：“如果该邮箱已注册，我们已发送重置邮件”。
- 重置成功后撤销该用户全部 Refresh Token Family，要求重新登录。
- 所有认证页面不得把 Raw Token、密码或 JWT 写入日志、埋点和 LocalStorage。

## 7. 视觉设计系统

### 7.1 设计方向

采用“信息建筑式训练控制台”：理性、克制、准确、轻微纸张温度。它不是赛博控制台，也不是游戏大厅。

关键词：`清晰`、`专注`、`可信`、`高密度`、`不压迫`。

### 7.2 色彩 Token

```css
:root {
  --bg: #f6f5f1;
  --surface: #ffffff;
  --surface-subtle: #f0efeb;
  --text: #191918;
  --text-muted: #686863;
  --border: #deddd7;

  --accent: #c45724;
  --accent-hover: #a9471b;
  --accent-soft: #f7e5dc;

  --success: #17753a;
  --success-soft: #e5f3e9;
  --warning: #956400;
  --warning-soft: #f8edcc;
  --danger: #bd352f;
  --danger-soft: #f8e4e2;
  --info: #2d62b7;
  --info-soft: #e4ecf8;

  --editor-bg: #181a1b;
  --editor-panel: #202224;
  --editor-border: #343638;
}
```

状态不能只靠颜色区分，必须同时有图形和文字。

### 7.3 字体

- 品牌英文与关键数字：`Bricolage Grotesque` 或同类开源 Display Sans。
- 中文正文：`Noto Sans SC`。
- 代码与输出：`JetBrains Mono`。
- 字体加载失败时使用明确 fallback，不阻塞首屏。

字号：正文 15–16px；辅助信息 13px；页面标题 28–36px；题目标题 22–26px。中文正文行高 1.75。

### 7.4 间距与形状

- 间距基线：4、8、12、16、24、32、48、64。
- 小控件圆角 6px；面板 8px；禁止所有元素都使用胶囊圆角。
- 主要按钮高度 40px；移动端点击区域至少 44×44px。
- 阴影只用于浮层和模态框，普通面板使用边界分隔。

### 7.5 明确禁区

- 不使用紫粉蓝全屏渐变、玻璃拟态和霓虹光。
- 不使用 Emoji 充当图标。
- 不用“大量同形圆角卡片”组织题单。
- 不使用虚构用户评价、虚构通过率、虚构在线人数。
- 不用装饰插画替代真实的输入、输出和代码示例。
- 不复制第三方品牌 Logo 和整套配色。

## 8. 响应式与无障碍

- `>= 1100px`：做题页左右分栏。
- `768–1099px`：分栏比例调整为 42/58，导航收缩。
- `< 768px`：题面/代码/结果 Tab 模式。
- 所有表单有可见 Label；错误信息与字段关联。
- 键盘可完成登录、筛选、切换 Tab、运行和提交。
- 焦点环不可移除。
- 正文对比度至少达到 WCAG AA。
- Monaco 外的按钮和筛选控件需要正确 aria-label。
- 支持 `prefers-reduced-motion`。

## 9. 推荐技术架构

### 9.1 仓库结构

```text
ACMHOT100/
├─ apps/
│  ├─ web/                       # React + Vite + TypeScript SPA
│  │  ├─ src/
│  │  │  ├─ app/                # Router、Provider、全局错误边界
│  │  │  ├─ components/         # 跨功能通用组件
│  │  │  ├─ features/           # auth/problems/editor/submissions/profile
│  │  │  ├─ lib/                # API Client、Query Client、工具
│  │  │  └─ styles/             # Token 与全局样式
│  │  ├─ nginx.conf             # SPA fallback 与 /api 反向代理
│  │  ├─ vite.config.ts
│  │  └─ package-lock.json
│  └─ server/                    # 单个 Go Module
│     ├─ cmd/
│     │  ├─ api/                 # Gin API 入口
│     │  └─ judge-worker/        # Redis Streams 判题消费者
│     ├─ internal/
│     │  ├─ auth/                # JWT、邮箱验证、密码与会话
│     │  ├─ config/
│     │  ├─ http/                # Gin handlers/middleware/routes
│     │  ├─ judge/               # Judge0 adapter 与结果比较
│     │  ├─ model/               # GORM models
│     │  ├─ queue/               # Redis Streams
│     │  ├─ repository/
│     │  └─ service/
│     ├─ migrations/             # 版本化 MySQL SQL 迁移
│     ├─ go.mod
│     └─ go.sum
├─ packages/
│  └─ contracts/            # OpenAPI 生成的前端类型或共享 schema
├─ infra/
│  ├─ docker-compose.yml
│  └─ judge0/               # Judge0 相关说明/覆盖配置
├─ seed/
│  └─ problems/             # 版本化题目 YAML/JSON 与测试数据
├─ docs/
├─ .env.example
├─ Makefile                 # Windows 不强依赖；只作为快捷入口
└─ README.md
```

### 9.2 技术选择

- Web：React、Vite、TypeScript、Tailwind CSS、Monaco Editor、Lucide Icons。
- 路由：React Router Data Mode，使用 `createBrowserRouter`，按路由懒加载并设置 route error boundary。
- 服务端状态：TanStack Query；所有 API Query Key 集中定义，禁止页面各自发明缓存键。
- 表单：React Hook Form + Zod；前端校验只改善体验，Gin 后端仍必须独立校验。
- 客户端状态：优先组件状态和 Context；MVP 不引入 Redux。编辑器草稿与服务端数据不放全局 Store。
- API：Go + Gin。
- ORM：GORM + MySQL Driver；数据库迁移使用版本化 SQL，不在生产环境依赖 `AutoMigrate`。
- 数据库：MySQL，统一使用 `utf8mb4`、UTC 和严格 SQL 模式。
- Redis：`go-redis/v9`，负责邮箱验证 Token、Refresh Session、JWT 撤销/限流、判题 Streams 和短期缓存。
- JWT：`golang-jwt/jwt/v5`，Access/Refresh 分离并执行 Refresh Token Rotation。
- 邮件：标准 SMTP；开发环境使用 Mailpit 捕获验证邮件。
- 判题：Judge0 CE，HTTP JSON API。
- Worker：Go，与 API 共享 `internal` domain/repository/queue 代码，使用 Redis Streams Consumer Group。
- 测试：Go `testing`、Testify、Vitest/React Testing Library、Playwright。

React 前端使用 Vite 构建 SPA，React Router 负责 URL/页面映射，TanStack Query 负责服务端状态；Gin 提供中间件式 HTTP 路由；GORM 官方支持 MySQL Driver；Redis 官方 Go 客户端支持 Streams；Judge0 提供沙箱化代码执行和 HTTP API。实现时应以官方文档和实际安装版本为准，提交 `go.mod`、`go.sum` 和前端 `package-lock.json`。

### 9.3 MySQL 与 Redis 的职责边界

- MySQL 是用户、题目、草稿、提交、测试数据和进度的最终事实来源。
- Redis 只保存有 TTL 或可重建的数据，以及异步任务流。
- 判题提交必须先写入 MySQL，再写入 Redis Stream。若写 Stream 失败，MySQL 中的 `QUEUED` 记录由 reconciliation job 补发。
- 不得仅在 Redis 中保存提交或 AC 结果。
- 第一版不额外引入 Celery、Kafka 或 RabbitMQ。

### 9.4 Redis Key 与 Stream 约定

| Key/Stream | Value | TTL/说明 |
|---|---|---|
| `auth:verify:{sha256}` | `user_id` | 30 分钟，一次性消费 |
| `auth:verify:user:{user_id}` | 当前 Token Hash | 30 分钟，重发时删除旧 Token |
| `auth:reset:{sha256}` | `user_id` | 20 分钟，一次性消费 |
| `auth:refresh:{jti}` | user/family/session metadata | 与 Refresh JWT 同寿命 |
| `auth:family:{family_id}` | 有效 Refresh JTI 集合 | 全设备/Family 撤销 |
| `auth:deny:{jti}` | `1` | Access JWT 剩余寿命 |
| `rate:{scope}:{subject}` | 计数器 | 不同接口独立窗口 |
| `judge:submissions` | Redis Stream | 字段只有 `submission_id` |
| `judge:lock:{submission_id}` | worker ID | 300 秒，可续期 |

Redis Key 必须集中在一个包中生成，禁止 handlers 内散落字符串。测试环境使用独立 DB 或统一 Key Prefix。

### 9.5 必需环境变量

```text
APP_ENV=development
APP_BASE_URL=http://localhost:5173
API_ADDR=:8080

MYSQL_DSN=app:app@tcp(mysql:3306)/acmhot100?charset=utf8mb4&parseTime=True&loc=UTC
REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_KEY_PREFIX=acmhot100:

JWT_ISSUER=acmhot100
JWT_ACCESS_AUDIENCE=acmhot100-web
JWT_REFRESH_AUDIENCE=acmhot100-refresh
JWT_ACCESS_SECRET=<至少32字节随机值>
JWT_REFRESH_SECRET=<与Access不同的至少32字节随机值>
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h

SMTP_HOST=mailpit
SMTP_PORT=1025
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=ACM Hot 100 <no-reply@example.local>
SMTP_TLS_MODE=none

JUDGE_MODE=mock
JUDGE0_BASE_URL=http://judge0-server:2358
```

`.env.example` 只放安全占位值。生产 Secret 不能提交 Git，也不能复用开发默认值。

## 10. 数据模型

所有时间使用 UTC 存储，MySQL 使用 `datetime(6)`，API 输出 ISO 8601。UUID 第一版使用 `char(36)`，优先保证实现清晰；后续再评估 `binary(16)`。

### 10.1 `users`

| 字段 | 类型 | 约束 |
|---|---|---|
| id | UUID | PK |
| email | varchar(320) | unique, lower-case |
| username | varchar(32) | unique, 3–32 字符 |
| password_hash | text | Argon2id |
| email_verified_at | datetime(6) nullable | 验证成功时间 |
| status | varchar(20) | PENDING/ACTIVE/DISABLED |
| created_at/updated_at | datetime(6) | not null |

### 10.2 `problems`

| 字段 | 类型 |
|---|---|
| id | UUID |
| slug | varchar(80), unique |
| order_index | int, unique |
| title | varchar(120) |
| difficulty | enum EASY/MEDIUM/HARD |
| stage | varchar(40) |
| statement_md | text |
| input_format_md | text |
| output_format_md | text |
| constraints_md | text |
| hints_md | text nullable |
| time_limit_ms | int |
| memory_limit_kb | int |
| is_published | bool |
| created_at/updated_at | datetime(6) |

### 10.3 `tags` 与 `problem_tags`

`tags(id, slug, name)`；`problem_tags(problem_id, tag_id)` 使用联合主键。

### 10.4 `test_cases`

| 字段 | 类型 | 说明 |
|---|---|---|
| id | UUID | PK |
| problem_id | UUID | FK |
| order_index | int | 测试顺序 |
| input_data | text | 标准输入 |
| expected_output | text | 标准输出 |
| is_sample | bool | 是否公开 |
| explanation_md | text nullable | 样例解释 |
| created_at | datetime(6) |  |

约束：同一题的 `(problem_id, order_index)` 唯一。公开题目接口只能返回 `is_sample = true` 的数据。

### 10.5 `language_configs`

| 字段 | 类型 |
|---|---|
| key | varchar(20), PK，例如 `cpp17` |
| display_name | varchar(40) |
| judge0_language_name | varchar(120) |
| judge0_language_id | int nullable |
| editor_language | varchar(30) |
| source_template | text |
| enabled | bool |

### 10.6 `drafts`

`drafts(user_id, problem_id, language_key, source_code, updated_at)`；前三列联合唯一，时间为 `datetime(6)`。

### 10.7 `submissions`

| 字段 | 类型 |
|---|---|
| id | UUID |
| user_id | UUID |
| problem_id | UUID |
| language_key | varchar(20) |
| source_code | text |
| status | submission_status enum |
| passed_cases | int default 0 |
| total_cases | int |
| time_ms | int nullable |
| memory_kb | int nullable |
| compiler_output | text nullable |
| error_message | text nullable |
| stream_message_id | varchar(64) nullable |
| enqueued_at | datetime(6) nullable |
| claimed_at | datetime(6) nullable |
| retry_count | int default 0 |
| created_at | datetime(6) |
| judged_at | datetime(6) nullable |

终态：`AC, WA, TLE, MLE, RE, CE, SYSTEM_ERROR`。非终态：`QUEUED, COMPILING, RUNNING`。

### 10.8 `submission_case_results`

记录每个测试点的状态、时间、内存以及截断后的输出。隐藏测试点的输入不得通过用户 API 返回。

### 10.9 `user_problem_progress`

| 字段 | 类型 |
|---|---|
| user_id/problem_id | 联合主键 |
| state | enum NOT_STARTED/ATTEMPTED/SOLVED |
| attempt_count | int |
| first_ac_at | datetime(6) nullable |
| last_submitted_at | datetime(6) nullable |

规则：创建正式提交后进入 `ATTEMPTED`；首次 AC 后进入 `SOLVED`，以后失败不得降级。

## 11. API 契约

统一前缀 `/api/v1`。错误格式固定为：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "给用户看的简洁说明",
    "details": {}
  },
  "request_id": "uuid"
}
```

### 11.1 Auth

- `POST /auth/register`
- `POST /auth/verify-email`
- `POST /auth/resend-verification`
- `POST /auth/login`
- `POST /auth/refresh`
- `POST /auth/logout`
- `POST /auth/logout-all`
- `POST /auth/forgot-password`
- `POST /auth/reset-password`
- `GET /auth/me`

注册流程：

1. 用户提交邮箱、用户名和密码。
2. MySQL 创建 `PENDING` 用户；同一邮箱只能存在一个账号。
3. 生成 32 字节密码学安全随机 Token，只把 Token 的 SHA-256 哈希写入 Redis：`auth:verify:{hash} -> user_id`，TTL 30 分钟。
4. 通过 SMTP 发送 `APP_BASE_URL/verify-email?token=<raw-token>`。
5. 验证接口原子消费 Redis Token，更新 `email_verified_at` 和 `status=ACTIVE`。
6. Token 单次使用；重发会让旧 Token 失效，并设置 60 秒冷却。

登录规则：

- 仅支持已验证邮箱 + 密码登录；未验证账号返回 `EMAIL_NOT_VERIFIED`，但普通密码错误统一返回相同文案，避免枚举。
- Access JWT 有效期 15 分钟；Refresh JWT 有效期 7 天。
- Access/Refresh 使用不同 Cookie、不同 `aud`，都包含 `sub`、`jti`、`iat`、`exp`、`iss`、`aud`。
- 使用 `golang-jwt/jwt/v5`，固定允许的签名算法；不能信任 Token Header 任意指定的 `alg`。
- Refresh Token 每次使用都轮换。Redis 保存 `auth:refresh:{jti}` 会话和 Token Family；重复使用旧 Refresh Token 时撤销整个 Family。
- Logout 删除 Refresh Session，并将尚未过期的 Access `jti` 写入 Redis denylist；`logout-all` 撤销该用户全部 Token Family。
- JWT 只放 HttpOnly、SameSite=Lax Cookie。开发环境允许非 Secure，生产环境必须 Secure。
- Cookie 认证的写接口验证 `Origin`，并采用双提交 CSRF Token 或等价防护。

密码重置规则：

- `forgot-password` 对存在与不存在的邮箱返回完全相同的状态和文案。
- 生成独立的 32 字节 Raw Token，Redis 只存 SHA-256 Hash：`auth:reset:{hash}`，TTL 20 分钟。
- `reset-password` 原子消费 Token，更新 Argon2id Hash，并撤销该用户所有 Refresh Token Family。
- 邮箱验证 Token 与密码重置 Token 不得通用，Key Namespace 和邮件链接必须分离。

开发邮件：Docker Compose 启动 Mailpit，SMTP 端口只供内部服务使用，Web UI 只绑定本机。生产环境通过环境变量连接真实 SMTP 服务。

### 11.2 Problems

- `GET /problems?q=&difficulty=&tag=&state=&page=&page_size=`
- `GET /problems/{slug}`
- `GET /problems/{slug}/navigation`
- `GET /tags`

题目详情必须包含公开样例和当前用户草稿/进度，但不能包含隐藏测试。

### 11.3 Drafts

- `PUT /problems/{slug}/drafts/{language_key}`
- `GET /problems/{slug}/drafts/{language_key}`

限制源代码最大 64KB。

### 11.4 Run sample

- `POST /problems/{slug}/run`

请求：语言、源代码、`sample_case_id` 或 `custom_input`。自定义输入最大 16KB。返回运行任务 ID，前端轮询同一提交状态接口；样例运行不写入正式提交表，可写入短期 `runs` 表或统一任务表。

### 11.5 Submissions

- `POST /problems/{slug}/submissions`
- `GET /submissions?problem=&status=&language=&page=`
- `GET /submissions/{id}`

创建成功返回 `202 Accepted`：

```json
{
  "id": "uuid",
  "status": "QUEUED",
  "created_at": "2026-07-13T00:00:00Z"
}
```

前端首 10 秒每 800ms 轮询，之后每 2 秒轮询，60 秒超时后停止自动轮询并提供刷新按钮。第一版不强制 WebSocket。

### 11.6 Profile

- `GET /profile/summary`
- `GET /profile/progress-by-stage`

## 12. 判题设计

### 12.1 状态机

```text
QUEUED → COMPILING → RUNNING → AC
                           ├→ WA
                           ├→ TLE
                           ├→ MLE
                           ├→ RE
              COMPILING ──┴→ CE
任何非终态 ────────────────→ SYSTEM_ERROR
```

状态更新必须单向；终态不得被后续轮询覆盖。

### 12.2 Redis Streams 入队与 Worker 消费

Stream 名称：`judge:submissions`；Consumer Group：`judge-workers`。

1. API 在 MySQL 事务中创建 `QUEUED` 提交。
2. 事务提交后调用 Redis `XADD`，消息只包含 `submission_id`；成功后回写 `stream_message_id/enqueued_at`。
3. 若 `XADD` 失败，API 仍返回已创建的提交；后台 reconciliation job 扫描缺少 `enqueued_at` 的 `QUEUED` 记录并补发。
4. Worker 使用 `XREADGROUP` 消费，先以 `SET judge:lock:{submission_id} <worker-id> NX EX 300` 防止重复执行。
5. Worker 在 MySQL 事务中锁定提交，确认仍为 `QUEUED` 后设置 `claimed_at/COMPILING`；重复或终态任务直接 `XACK`。
6. 读取语言配置与测试数据，使用 Judge0 batch API 提交测试点并聚合结果。
7. 结果与 `user_problem_progress` 在同一 MySQL 事务中落库。
8. 成功落库后 `XACK` 并删除分布式锁。

Worker 崩溃恢复：使用 `XAUTOCLAIM` 领取超过 5 分钟未确认的 Pending 消息；同时 reconciliation job 扫描 MySQL 中超时的非终态提交。最多重试 2 次，之后写入 `SYSTEM_ERROR`。所有处理必须幂等。

### 12.3 输出比较

第一版只支持普通文本判题：

1. `\r\n` 和 `\r` 统一为 `\n`。
2. 每行移除行尾空白。
3. 移除文件末尾多余空行。
4. 其余字符严格比较，区分大小写。

比较函数必须有单元测试。题目若允许多种合法输出，不得在第一版发布。

### 12.4 资源与输出限制

- 时间、内存从题目配置传给 Judge0。
- 单次源代码最大 64KB。
- stdin 最大 64KB，用户自测最大 16KB。
- stdout/stderr 每项最多保存 8KB；超出截断。
- API 和 Worker 日志不得记录完整隐藏测试数据、密码、Cookie、邮箱验证 Token、Access JWT 或 Refresh JWT。

### 12.5 Judge0 降级模式

开发环境支持 `JUDGE_MODE=mock`：非空代码在 500–1000ms 后返回 AC，用于前端联调。页面必须显示“Mock Judge”开发标记。测试环境通过 Mock HTTP Server 覆盖 CE、WA、TLE、AC 状态映射。生产环境禁止 mock。

## 13. 认证、邮箱与安全

- 密码使用 Argon2id 哈希。
- 邮箱验证 Token 使用 `crypto/rand` 生成，Redis 只存 SHA-256 哈希，不存原 Token。
- 密码重置 Token 采用相同的安全生成和 Hash-only 存储策略，但使用独立命名空间与 20 分钟 TTL。
- 邮件模板不得包含密码、JWT 或内部服务地址。
- Access/Refresh JWT 分离，固定算法并严格校验 `iss/aud/exp/nbf/jti`。
- Refresh Token Rotation 与复用检测状态存入 Redis。
- 登录 Cookie 为 HttpOnly；生产环境 Secure；写请求执行 CSRF/Origin 检查。
- 注册、登录、重发验证邮件、刷新 Token 和提交接口使用 Redis 限流。
- 用户源代码视为私有数据。
- Markdown 禁用原始 HTML并进行链接协议过滤。
- 开发环境由 Vite Dev Server 把 `/api` 代理到 Gin；生产环境由 Nginx 同源反向代理 `/api`，因此默认不开放宽泛 CORS。
- Nginx 必须为 React Router 配置 SPA fallback：非静态文件、非 `/api` 路径回退到 `index.html`。
- 业务容器不挂载 Docker Socket。
- 用户代码只能进入 Judge0，不能使用 Go `os/exec`、Python `exec`、Node `eval` 或本地 shell 直接运行。
- Judge0 不暴露公网端口；只在 Compose 内网被 Worker 访问。
- 所有对象级接口检查当前用户 ID，防止 IDOR。
- MySQL 连接必须启用 `charset=utf8mb4&parseTime=True&loc=UTC`；生产凭据不得使用 root 用户。
- Redis 不暴露公网端口，生产环境启用认证和网络隔离。

## 14. 示例题内容规范

Seed 至少包含 5 道题，覆盖不同输入结构：

1. 两数目标和：一维数组输入。
2. 括号序列是否合法：字符串输入。
3. 合并两个有序序列：两个数组输入。
4. 最大连续子段和：大整数与负数边界。
5. 二叉树层序遍历：使用明确的层序序列和空节点标记。

每道题必须包含：原创中文题面、2 个公开样例、至少 8 个隐藏测试、三种语言参考程序、预期复杂度、数据生成/校验说明。参考程序不通过普通用户 API 暴露。

## 15. 前端状态与错误处理

- 服务端数据使用 TanStack Query + 统一 API Client；不要在每个组件内手写 fetch。
- API Client 默认 `credentials: 'include'`，不能读取 HttpOnly JWT。
- 收到 Access Token 过期的 401 时，API Client 只允许执行一次并发合并的 `/auth/refresh`；刷新成功后重试原请求一次，失败则清空用户缓存并跳转登录，禁止无限刷新循环。
- Auth 状态以 `/auth/me` 为准；React Context 只保存当前用户快照和认证动作，不保存 JWT。
- React Router 的受保护路由在用户状态加载完成后再决定跳转，避免刷新页面时闪回登录页。
- URL 保存题单筛选状态。
- LocalStorage 只保存分栏比例、编辑器偏好和草稿备份；认证信息不得放 LocalStorage。
- API 草稿保存失败时保留本地备份并显示非阻塞警告。
- 提交按钮防重复；请求成功后以提交 ID 去重。
- 网络断开时不伪造成功状态。
- 空状态、加载状态、错误状态必须分别设计。

## 16. 实现阶段

### Phase 0：基础设施与契约

- 建立 monorepo 目录。
- 创建 React/Vite Web 和 Go Server；Server 内提供 API 与 Judge Worker 两个 `cmd`。
- 初始化 React Router、TanStack Query、React Hook Form、Zod、Tailwind、Vitest 和 Playwright。
- 配置 Vite `/api` 开发代理、Nginx SPA fallback 和生产 `/api` 反向代理。
- Docker Compose 启动 MySQL、Redis、Mailpit；Judge0 可先以 mock 替代。
- 初始化 Gin、GORM、MySQL Driver、go-redis 和版本化 SQL migrations。
- 建立环境变量、迁移、健康检查和 CI。

验收：`docker compose up` 后 Web 和 API 健康检查可访问，测试命令通过。

### Phase 1：邮箱认证、题库和设计系统

- 实现注册、验证邮件、重发验证、登录、忘记/重置密码、JWT 刷新轮换、退出和会话恢复。
- 建表、迁移和 5 道 Seed 题。
- 实现全局 Shell、首页、题单、题目详情只读版。

验收：Mailpit 能收到验证和重置邮件；未验证账号不能登录；验证后可登录和刷新；重置密码后旧会话失效；退出后 Refresh Token 失效；隐藏测试不出现在网络响应中。

### Phase 2：编辑器、草稿和样例运行

- Monaco、三语言模板、语言切换。
- 本地+服务端草稿。
- Mock Judge 样例运行。

验收：刷新后代码恢复；移动端可切换题面、代码和结果。

### Phase 3：正式判题闭环

- Submission API、Redis Streams、Go Worker、reconciliation job、Judge0 Adapter。
- 状态轮询、结果面板、进度更新。

验收：至少分别验证 AC、WA、CE、TLE；AC 后题单和 Profile 同步更新。

### Phase 4：提交历史、质量与文档

- 提交列表和详情。
- 限流、安全检查、错误边界。
- E2E、README、运维说明。

验收：新环境按 README 能启动；核心 E2E 全部通过。

## 17. 测试计划

### 17.1 API/Domain 单元测试

- 邮箱验证 Token 单次使用、过期和重发失效。
- 密码重置不枚举邮箱、Token 单次使用、过期后拒绝、成功后撤销全部旧会话。
- Access/Refresh `aud` 隔离、固定签名算法和 Refresh Rotation。
- 旧 Refresh Token 复用时撤销整个 Token Family。
- 输出规范化和比较。
- Submission 状态机拒绝倒退。
- Judge0 状态映射。
- progress：失败后 ATTEMPTED，AC 后 SOLVED，之后不降级。
- 用户只能读取自己的提交。
- 公开题目响应不含隐藏测试。

### 17.2 Worker 集成测试

- Redis Streams Consumer Group 下并发 Worker 不重复产生终态副作用。
- `XADD` 失败后 reconciliation 能补发 MySQL 中的 QUEUED 提交。
- Pending 消息可通过 `XAUTOCLAIM` 恢复。
- AC、WA、CE、TLE、Judge0 超时。
- 崩溃后的超时重领。
- 输出截断和宿主路径清理。

### 17.3 前端测试

- React Router 公开路由、受保护路由与 SPA 刷新回退。
- API Client 并发 401 只触发一次 Refresh，请求最多重试一次。
- Auth Provider 不接触或持久化 JWT。
- 题单筛选与 URL 同步。
- 语言切换保护已编辑代码。
- 草稿防抖保存和失败回退。
- 判题轮询在终态停止。
- 不同状态的结果面板内容。

### 17.4 Playwright E2E

1. 注册并登录。
2. 打开题单并筛选。
3. 打开题目，修改代码，刷新后草稿恢复。
4. 运行样例。
5. 正式提交并等待 AC。
6. 题单状态变绿，Profile 数量增加。
7. 打开提交详情并复制代码。

## 18. MVP 完成定义

只有同时满足以下条件才算完成：

- `docker compose up --build` 能在干净环境启动。
- 数据库迁移和 Seed 可重复运行。
- Mailpit 能收到邮箱验证邮件，验证 Token 单次使用且会过期。
- 忘记密码不泄露邮箱是否注册，重置后旧 Refresh Token 全部失效。
- Access/Refresh JWT、轮换、退出和 Redis 撤销流程通过测试。
- 5 道题可浏览，隐藏测试不泄露。
- 三种语言可选择并保存草稿。
- Mock Judge 和真实 Judge0 Adapter 都有测试。
- 能得到 AC、WA、CE、TLE 四种真实或受控集成结果。
- 提交状态、提交历史和用户进度一致。
- 核心 E2E 通过。
- 无阻断级控制台错误。
- 1280×720、1440×900、390×844 三个视口完成视觉检查。
- README 写清 Windows/PowerShell 和 Docker 的启动方式。

## 19. 官方技术参考

- React 官方文档：https://react.dev/
- React + Vite 起步：https://react.dev/learn/build-a-react-app-from-scratch
- Vite 官方文档：https://vite.dev/guide/
- React Router Data Mode：https://reactrouter.com/start/data/routing
- TanStack Query React：https://tanstack.com/query/latest/docs/framework/react/overview
- Gin 官方文档：https://gin-gonic.com/en/docs/
- GORM MySQL 官方文档：https://gorm.io/docs/connecting_to_the_database.html
- Redis Go 客户端与 Streams：https://redis.io/docs/latest/develop/clients/go/
- golang-jwt 官方仓库：https://github.com/golang-jwt/jwt
- Mailpit SMTP 文档：https://mailpit.axllent.org/docs/configuration/smtp/
- Judge0 官方仓库：https://github.com/judge0/judge0
- Judge0 API 文档：https://ce.judge0.com/

实现模型必须在安装依赖时查看当前官方文档，提交 lockfile，并避免凭记忆硬编码已经变化的 CLI 参数或语言 ID。
