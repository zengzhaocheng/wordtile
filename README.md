# WordTile 🧩

> 移动端优先的单词/词条匹配学习应用 —— 上下分屏点击匹配的"连连看"式玩法，配合自定义模式、闯关模式、题库市场，以及 **Git-like 题库版本管理与三路合并**。

基于 `WordTile_PRD.docx` 完整实现。前端 React 18 + TypeScript（Vite），后端 Go 1.22（Gin + GORM + PostgreSQL + Redis），`docker-compose` 一键启动。

---

## 目录

- [功能总览](#功能总览)
- [技术栈](#技术栈)
- [快速开始](#快速开始)
- [本地开发](#本地开发)
- [演示账号](#演示账号)
- [生产部署](#生产部署)
- [API 摘要](#api-摘要)
- [题库 JSON 导入 / 导出](#题库-json-导入--导出)
- [Git-like 版本合并模型](#git-like-版本合并模型)
- [架构与目录结构](#架构与目录结构)
- [测试](#测试)
- [假设与取舍](#假设与取舍)
- [PRD 里程碑映射](#prd-里程碑映射)

---

## 功能总览

| 模块 | 说明 |
| --- | --- |
| **核心玩法** | 上屏词条、下屏释义，点击匹配。正确→消除；错误→双方变红 1 秒。计时、错误计数、提示（高亮一对未匹配项） |
| **自定义模式** | 选择任意题库，10~50 块数量滑块，按标签筛选参与词条 |
| **闯关模式** | 13 关递进解锁（第 n 关 `min(6+2(n-1), 30)` 对词条），每关最多 3 次错误，记录最佳成绩，前 3 关试玩 |
| **题库管理** | 题库/词条/标签 CRUD，支持词条级 commit 快照、发布到市场 |
| **在线社区** | 题库市场（搜索/排序）、发布、Fork、点赞、评论 |
| **Git-like 版本合并** | 跨题库 Pull 合并、三路合并算法、冲突解决工作台（采用我方/采用对方/自定义） |
| **数据统计** | GitHub 风格年度热力图、每局用时折线、错误率演进折线、单题库进度 |
| **体验** | 移动端优先响应式布局、深色/浅色主题（跟随系统）、中/英双语、PWA 可安装、离线壳缓存 |

## 技术栈

**前端**：React 18 · TypeScript（strict）· Vite · Tailwind CSS · React Router · Zustand · TanStack Query · Axios · React Hook Form + Zod · Framer Motion · vite-plugin-pwa · Vitest

**后端**：Go 1.22 · Gin · GORM · PostgreSQL 16 · Redis 7 · golang-jwt/v5（Access + Refresh 双令牌）· bcrypt · golang-migrate（生产 SQL 迁移）· testify

**基础设施**：Docker + docker-compose（PostgreSQL / Redis / backend / frontend-nginx）

---

## 快速开始

### 前置要求

- Docker 与 Docker Compose

### 一键启动

```bash
docker compose up --build
```

启动完成后：

| 服务 | 地址 |
| --- | --- |
| 前端 | http://localhost:9001 |
| 后端 API | http://localhost:9000/api/v1 |
| 健康检查 | http://localhost:9000/healthz |
| OpenAPI 文档 | `docs/openapi.yaml`（可直接导入 Swagger Editor / Apifox） |

首次启动会自动：

1. 建表（GORM AutoMigrate）
2. 写入演示数据：账号 `demo / demo1234`、公开示例题库「基础英语词汇（示例）」（46 条词条，含数字/食物/旅游/动物/家庭标签）

> 数据持久化在 `pgdata` volume 中。想彻底重置：`docker compose down -v`。

### 本地开发（不走 Docker）

后端（需本地 PostgreSQL 与 Redis，或只起基础设施）：

```bash
docker compose up -d postgres redis   # 只起依赖
cd backend
export DATABASE_URL="postgres://wordtile:wordtile@localhost:5432/wordtile?sslmode=disable"
export REDIS_ADDR="localhost:6379"
go run ./cmd/server                    # 默认 :8080，AutoMigrate+Seed 开启
```

> Windows 提示：Hyper-V 会保留一段 TCP 端口（可用 `netsh interface ipv4 show excludedportrange protocol=tcp` 查看），本机 8080/8081 恰好在保留段内无法绑定，因此 compose 对外使用 **9000（后端）/ 9001（前端）**。本地直跑后端如遇同样问题，设 `PORT=9000` 并相应调整 Vite 代理目标。

前端（Vite dev server，`/api` 已代理到 `localhost:8080`）：

```bash
cd frontend
npm install
npm run dev                            # http://localhost:5173
```

---

## 生产部署

### 前置：买一台服务器

推荐 **香港/新加坡的轻量云服务器**（2 核 2G 起，约 100–300 元/年）。这两个地区**不需要 ICP 备案**，且国内访问延迟低（30–80ms，选 CN2/GIA 线路更好）。

> 备案规则：ICP 备案的触发条件是**服务器在中国大陆**。服务器在境外（含香港、新加坡、日本、美国，以及 Cloudflare 的免费 `*.pages.dev`）都不需要备案。个人无公司也可以用**个人主体**备案，但个人备案站点不得含经营性/企业性质内容。

想零成本可申请 [Oracle Cloud 永久免费 ARM 实例](https://www.oracle.com/cloud/free/)（4 核 24G，跑这套容器绰绰有余）。

### 部署步骤

```bash
# 1. 装 Docker（Ubuntu 22.04）
curl -fsSL https://get.docker.com | sh

# 2. 拉代码
git clone https://github.com/zengzhaocheng/wordtile.git && cd wordtile

# 3. 配置环境变量（脚本会自动生成随机密钥，只需改域名）
cp .env.example .env
vim .env          # 改 DOMAIN 和 CORS_ORIGINS 为你的真实域名

# 4. 一键构建启动
bash deploy.sh
```

### 需要改的配置

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `DOMAIN` | 你的域名，Caddy 用它自动签发 HTTPS 证书 | `wordtile.example.com` |
| `CORS_ORIGINS` | 允许跨域的来源，多个用逗号分隔 | 仅 localhost |
| `JWT_SECRET` | **必须改**！生成：`openssl rand -base64 32` | `change-me-in-production` |
| `POSTGRES_PASSWORD` | 数据库密码，脚本会随机生成 | `wordtile` |
| `SEED_DEMO` | 是否灌入演示账号与示例题库 | `true` |

生产 compose 与开发版的区别：Postgres/Redis **不映射宿主机端口**（只在容器内网可达），由 Caddy 统一对外，自动签发并续期 Let's Encrypt 证书。

### 关于 Cloudflare

- **前端**可以免费放 Cloudflare Pages（静态资源，全球 CDN）
- **后端不行**：Workers 是 V8 隔离环境而非容器，跑不了 Go 二进制，也没有原生 TCP 连 Postgres/Redis。要上就得把后端整体重写成 TS（Gin→Hono、GORM→Drizzle、Postgres→D1、Redis→KV）
- 想在家里/公司机器上零成本暴露服务，可以用 **Cloudflare Tunnel**，但国内访问 Cloudflare 域名不稳定，自用尚可，不适合给国内用户

---

## 演示账号

| 用户名 | 密码 | 说明 |
| --- | --- | --- |
| `demo` | `demo1234` | 种子账号，拥有示例公开题库 |

也可以直接注册新账号（登录页有注册入口）。

---

## API 摘要

完整定义见 [`docs/openapi.yaml`](docs/openapi.yaml)。JSON 字段一律为 **snake_case**；除标注外均需 `Authorization: Bearer <access_token>`；单个对象响应均有包装（如 `{"deck": ...}`、`{"decks": [...]}`）。

### 认证 `/api/v1/auth`（全局限流 20 次/分钟/IP）

| 方法 | 路径 | 请求 | 响应 |
| --- | --- | --- | --- |
| POST | `/auth/register` | `{username, email, password, display_name?}` | `201 {user, access_token, refresh_token}` |
| POST | `/auth/login` | `{identifier, password}`（identifier 为用户名或邮箱） | `200 {user, access_token, refresh_token}`（access 15min / refresh 7d） |
| POST | `/auth/refresh` | `{refresh_token}` | `200 {access_token, refresh_token}` |

### 用户

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/users/me` | `{user: {id, username, email, display_name, avatar_url, theme_preference, ...}}` |
| PATCH | `/users/me` | `{display_name?, avatar_url?, theme_preference?}` |

### 题库与词条

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/decks` | `{decks: [...]}` 我的题库 / `{name, description?}` 新建（→ `{deck}`） |
| GET/PATCH/DELETE | `/decks/:id` | 详情（含词条、origin_head、open_conflicts）/ `{name?, description?, visibility?}` / 删除：`?remove_from_market=true\|false`（见下） |
| GET/POST | `/decks/:id/entries` | `{entries: [...]}` / `{term, definition, tags?}` 新增（→ `{entry}`） |
| PATCH/DELETE | `/entries/:entryId` | `{term?, definition?, tags?}` 修改 / 删除（PRD 规定的顶层路径） |
| GET | `/decks/:id/export` | 导出题库 JSON 文档（公开题库任何人可导出） |
| POST | `/decks/import` | 从 JSON 文档**新建**题库 → `201 {deck, imported}` |
| POST | `/decks/:id/import` | 导入词条到已有题库，`mode: append`（默认）/ `replace` → `{deck_id, imported, mode}` |

题库字段：`id, owner_id, name, description, visibility(public/private), head_commit_id, forked_from, likes_count, forks_count, archived, entry_count, ...`

> 导入导出的完整格式见下方 [题库 JSON 导入 / 导出](#题库-json-导入--导出)。

### 题库删除与 Fork 语义

市场**不是独立副本**：`visibility=public` 的同一条题库记录就是市场里的题库。因此删除时提供两种模式（前端在「我的题库」列表和题库详情页给出勾选，公开题库默认不勾选即保留市场版）：

| `remove_from_market` | 公开题库 | 私有题库 |
| --- | --- | --- |
| `true`（默认，兼容旧客户端） | 彻底删除题库、词条、commit、点赞、评论，市场同步下架 | 彻底删除 |
| `false` | **归档**：只从「我的题库」隐藏（`archived=true`），市场公开版、词条、commit 全部保留 | 彻底删除（本就没有市场版本） |

响应：`{deleted, mode: "hard"|"archived", market_removed, market_retained, detached_forks}`。

**已 Fork 的用户永远不受影响**：Fork 生成的是独立题库行（词条与 commit 都是副本）。彻底删除源题库时，只会把这些副本的 `forked_from` 置空（返回 `detached_forks` 数量），避免悬空引用；副本本身、词条、成绩都继续可用，只是不能再从已删除的源 Pull 更新。选择「保留市场版本」时源记录仍在，既有 fork 连 Pull 都能继续用。归档版仍归原作者所有，作者可在市场详情页彻底删除它。

**Fork 幂等**：同一账号对同一来源只保留一个活跃 fork。重复点击返回 `200 {deck, created:false, existing:true}`，不新建副本、不重复增加 `forks_count`；首次 Fork 返回 `201 {deck, created:true, existing:false}`。已彻底删除或已归档的副本不算「活跃」，可重新 Fork。历史数据中已存在的重复副本不会自动合并。

### 社区

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/marketplace/decks` | 公开市场（无需登录），`?search=&sort=new|likes|forks` → `{decks: [...]}`（Redis 缓存） |
| POST | `/decks/:id/publish` | 发布/取消发布（visibility 切换） |
| POST | `/decks/:id/fork` | 幂等 Fork 到自己名下（保留 LogicalID 与 commit 历史）→ 首次 `201 {deck, created:true}`，重复 `200 {deck, existing:true}` |
| POST | `/decks/:id/like` | 点赞/取消 → `{liked, likes_count}` |
| GET/POST | `/decks/:id/comments` | `{comments: [...]}` / `{content}` 发表评论 |

### 版本与合并

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET/POST | `/decks/:id/commits` | `{commits: [...]}`（按 seq）/ `{message}` 创建 commit（快照全部词条变更） |
| POST | `/decks/:id/merge` | `{source_deck_id, source_commit_id?}` 从源题库 Pull：自动合并立即生效，返回冲突列表 |
| GET | `/decks/:id/conflicts` | `?status=open`（默认）未解决冲突列表 |
| POST | `/decks/:id/conflicts/resolve` | `{conflict_id, resolution: ours\|theirs\|custom, custom: {term, definition, tags}}`，全部解决后自动生成双父 merge commit |

### 游戏（服务端判定，客户端只持有 token，不知道 term↔definition 映射）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | `/games/start` | `{deck_id, mode: custom\|campaign, block_count?, level?, tags?}` → 会话 `{session_id, terms: [{token, text}], definitions: [{token, text}], ...}` |
| GET | `/games/:id` | 会话状态（刷新恢复） |
| POST | `/games/:id/match` | `{term_token, definition_token}` → `{correct, correct_count, wrong_count, finished}` |
| POST | `/games/:id/hint` | → `{term_token, definition_token}`（一对未匹配词条） |
| POST | `/games/:id/finish` | `{duration_ms}` 结算（幂等）→ `{correct_count, wrong_count, accuracy, ...}` |

### 统计

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/stats/contribution` | `{days: [{date, sessions, duration_ms, accuracy}]}` 近 26 周（热力图数据） |
| GET | `/stats/decks/:id/progress` | `{points: [{completed_at, duration_ms, correct_count, wrong_count, error_rate}]}` |
| GET | `/stats/campaign` | `{levels: [...]}` 闯关最佳成绩 |

---

## 题库 JSON 导入 / 导出

不用在界面上一条条点。可以用脚本、Excel、Anki、LLM 等外部工具生成 JSON，一次导入整份题库；也可以把已有题库导出成 JSON 备份或分享。

### 入口

| 想做的事 | 在哪操作 |
| --- | --- |
| 从 JSON 新建题库 | 「我的题库」页 → **导入 JSON**（可先点「下载模板」） |
| 追加 / 替换某个题库的词条 | 进入该题库 → **导入词条** + 选择「追加」或「替换」 |
| 备份 / 分享题库 | 进入该题库 → **导出 JSON**（浏览器下载文件） |
| 逛别人公开的题库 | 底部导航 **🌐 市场**，或「我的题库」页顶部的题库市场卡片 |

### 文件格式（v1）

最小可用文件——只有 `entries` 是必填的：

```json
{
  "entries": [
    { "term": "apple", "definition": "苹果" },
    { "term": "travel", "definition": "旅行", "tags": ["旅游"] }
  ]
}
```

也接受**纯数组**（适合脚本直接生成）：

```json
[
  { "term": "apple", "definition": "苹果" },
  { "term": "banana", "definition": "香蕉" }
]
```

导出的文件带完整元信息：

```json
{
  "format": "wordtile-deck",
  "version": 1,
  "name": "基础英语词汇",
  "description": "示例题库",
  "exported_at": "2026-11-15T00:00:00Z",
  "entries": [
    { "term": "apple", "definition": "苹果", "tags": ["食物", "水果"] }
  ]
}
```

| 字段 | 必填 | 说明 |
| --- | --- | --- |
| `entries` | 是 | 词条数组，至少 1 条，最多 5000 条 |
| `entries[].term` | 是 | 词条文本，去首尾空白后非空，≤200 字 |
| `entries[].definition` | 是 | 释义文本，同上，≤1000 字 |
| `entries[].tags` | 否 | 字符串数组，最多 20 个、每个 ≤50 字；空标签自动丢弃 |
| `name` / `description` | 否 | 仅「从 JSON 新建题库」时生效；导入已有题库时忽略 |
| `format` / `version` | 否 | 省略即可；存在时会校验，`format` 必须为 `wordtile-deck`，`version` 不得大于 1 |
| `mode` | 否 | 仅 `POST /decks/:id/import` 使用：`append`（默认）或 `replace` |

`replace` 会清空该题库原有词条后再写入，前端会二次确认。

### 与 Git-like 版本模型的关系

导入的每条词条都会生成新的 `LogicalID`，所以对版本系统而言这是一次「全量新增」：

- **追加**导入 → 下次 commit 记录为 N 条 added
- **替换**导入 → 下次 commit 记录为旧词条 removed + 新词条 added（不是就地修改）
- 因此**不要指望用 JSON 导入去「更新」某个已有词条**。要改内容请用界面编辑，或走 commit + Pull 合并，那样才能保留合并语义

导入与新建题库都在**单个事务**内完成，失败不会留下半份题库；导入请求体上限 8 MiB。校验规则由 `backend/internal/deckio`（服务端）与 `frontend/src/utils/deckJson.ts`（浏览器）双重实现，并有单元测试覆盖。

---

## Git-like 版本合并模型

这是本项目最核心的设计，用一套 **LogicalID + Commit 图** 实现跨题库的三路合并：

### 关键概念

- **LogicalID**：每个词条除了行级主键 `ID` 外，还有一个跨题库稳定的 `LogicalID`。Fork 时行 ID 重新生成，但 LogicalID 原样保留 —— 因此 fork 出的题库与源题库中"同一个词"永远可以对应起来。
- **Commit**：每次 commit 对全量词条做 diff（相对上一次 commit 的状态），把 `added / updated / removed`（均以 LogicalID 为键）写入 commit 记录。commit 拥有内容哈希 `Hash`、序号 `Seq` 和父提交 `ParentHashes`（merge commit 有两个父）。
- **Fork 复制 commit 历史**：fork 时把源题库的全部 commit（hash、seq、parents、CreatedAt）原样复制，使两个题库可以计算**最近公共祖先（LCA）**。

### 合并流程（Pull）

```
源题库 fork ──●──●──● (source 新提交)
              \
               ●──● (fork 后你的新提交)

LCA = fork 点 → 以 LCA 的词条状态为 base，
      以源题库最新状态为 theirs，你的最新状态为 ours，
      做三路合并。
```

三路合并规则（`internal/merge/merge.go`，纯函数、全覆盖单测）：

| base | ours | theirs | 结果 |
| --- | --- | --- | --- |
| 不变 | 不变 | 不变 | 保留 |
| X | X' | X | 采用 ours |
| X | X | X'' | 采用 theirs |
| X | X' | X''（≠X'） | **冲突**，挂起待解决 |
| X | X' | 删除 | **冲突**（删除 vs 修改） |
| X | 删除 | 删除 | 删除，不冲突 |
| — | +A | — | 新增 A |
| — | +A | +A | **冲突**（同 ID 各自新增） |

- 自动可合并部分**立即写入**工作区（entries 表）并生成变更 commit；
- 冲突项写入 `conflicts` 表，题库挂起 `PendingMergeSource/PendingMergeBase`；
- 冲突解决工作台（前端 MergePage）三栏对比 base/ours/theirs，逐条选择 **采用我方 / 采用对方 / 自定义**；
- 最后一个冲突解决后自动生成**双父 merge commit**，此后两个题库继续演化仍可再次合并。

---

## 架构与目录结构

```
worktile/
├── docker-compose.yml          # 一键启动
├── README.md
├── docs/openapi.yaml           # OpenAPI 3.0 接口文档
├── backend/
│   ├── cmd/server/main.go      # 装配入口
│   ├── internal/
│   │   ├── config/             # 环境变量配置
│   │   ├── model/              # GORM 模型（User/Deck/Entry/Commit/Conflict/...）
│   │   ├── apperr/             # 带状态码的业务错误
│   │   ├── auth/               # JWT 签发/校验、限流中间件
│   │   ├── database/           # 连接（重试）、AutoMigrate、种子数据
│   │   ├── repository/         # 数据访问层
│   │   ├── service/            # 业务层：auth/deck/community/merge/game/stats/cache
│   │   ├── handler/            # HTTP 层（参数绑定、响应）
│   │   ├── router/             # 路由注册
│   │   └── merge/              # ★ 三路合并纯函数库 + 单元测试
│   ├── migrations/             # golang-migrate SQL（生产用）
│   ├── Dockerfile
│   └── Makefile
└── frontend/
    ├── src/
    │   ├── api/                # axios 实例（401 自动刷新）+ 类型化端点
    │   ├── stores/             # zustand：auth / theme / game（游戏状态机）
    │   ├── i18n/               # 中/英双语字典
    │   ├── components/         # TileBlock（五态方块）/ Heatmap / LineChart / Layout
    │   ├── pages/              # 13 个页面
    │   ├── utils/              # 热力图日历/分级
    │   └── types.ts
    ├── Dockerfile              # 构建 → nginx（SPA 回退 + /api 反代）
    └── vite.config.ts          # PWA + 代理 + vitest
```

后端严格分层：`handler（HTTP）→ service（业务）→ repository（数据）`，合并算法独立为无依赖的纯函数包便于测试。

---

## 测试

### 后端

```bash
cd backend
go test ./...        # 合并算法 8 组用例：Diff/Apply 往返、无冲突合并、
                     # 双方同改、双方异改冲突、删除 vs 修改冲突、
                     # 双删不冲突、同 ID 双方各自新增冲突、LCA
```

Dockerfile 构建阶段也会先跑 `go test ./internal/merge` 再产出镜像。

### 前端

```bash
cd frontend
npm test             # vitest + jsdom
                     # - gameStore：开局重置 / 正确匹配消除 / 错误配对标记 / 取消选中
                     # - heatmap：levelOf 分级 / 26 周日历网格 / 日期对齐
```

---

## 假设与取舍

1. **JWT 无撤销机制（stateless）**：Access Token 15 分钟、Refresh 7 天，不做黑名单。代价是登出后 access 在最长 15 分钟内仍有效；换来零额外存储。多副本部署时若需强制下线，应引入 Redis 黑名单。
2. **限流为进程内固定窗口**：单副本够用；多副本部署应换成 Redis 令牌桶（接口签名已隔离在中间件，替换成本小）。
3. **AutoMigrate 与 golang-migrate 双轨**：开发/演示默认 AutoMigrate（快速迭代），生产提供 `migrations/*.sql`（与模型对齐）。README 假设演示环境直接用默认值。
4. **游戏判定全部在服务端**：客户端只拿 token 与文本，不知道 term↔definition 映射，`/games/:id/match` 由服务端判定，防止前端作弊；代价是每次点击一次网络请求（局域网/本地无感，弱网下可加乐观 UI）。
5. **合并冲突时"暂取本地侧"**：冲突未解决前工作区保留 ours 版本，保证题库随时可玩、可继续编辑；冲突列表独立展示。
6. **市场缓存 Redis 优先、进程内兜底**：Redis 不可用时自动降级为内存缓存，功能不中断。
7. **单实例 Seed**：种子数据幂等（存在即跳过），并发首启不会产生重复数据。
8. **闯关第 1~3 关试玩**：与 PRD 一致，未登录只能玩前 3 关；本实现中全部对局接口要求登录（简化权限模型），PRD 的"试玩"体现为前端引导注册。
9. **PWA 图标**为脚本生成的简洁图形占位（`gen_icons.py` 已删除，产物在 `frontend/public/`），正式上线请替换为设计稿。
10. **未做实时通信（WebSocket）**：PRD 未要求对战/协作实时性，评论与合并均为请求-响应模型。
11. **删除公开题库默认连市场一起删**：`remove_from_market` 缺省 `true`，保证旧客户端行为不变；「只删我的、保留市场」用 `archived` 标记实现，而不是复制一份市场快照 —— 避免同一内容出现两个 ID、让已有 fork 的来源指针失效。归档版仍归原作者所有，作者可从市场详情页彻底删除。
12. **Fork 幂等（GitHub 语义）**：同一用户对同一来源只保留一个活跃副本，避免重复副本携带相同 commit hash 集合、虚增 `forks_count`；代价是用户若确实想要第二份独立副本，需先删除或归档已有副本再 Fork。历史上已产生的重复副本不自动合并。

---

## PRD 里程碑映射

| 里程碑 | PRD 范围 | 实现 |
| --- | --- | --- |
| M1 核心玩法 | 匹配消除、错误反馈、计时、提示 | `GamePage` + `games` API（服务端判定） |
| M2 模式扩展 | 自定义模式（10~50 块+标签筛选）、闯关（13 关递进+3 错上限） | `CustomSetupPage`、`CampaignPage`、`games/start` |
| M3 题库 CRUD | 题库/词条/标签管理 | `DecksPage`、`DeckEditPage` + decks/entries API |
| M4 社区 | 发布、Fork、点赞、评论、市场搜索 | `MarketplacePage`、`MarketDeckPage` + community API |
| M5 Git-like 合并 | commit、三路合并、冲突解决工作台 | `MergePage` + version API + `internal/merge` |
| M6 数据统计 | 热力图、用时/错误率演进、题库进度 | `StatsPage` + stats API |

---

## 许可

演示项目，仅用于评估与学习。
