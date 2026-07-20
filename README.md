# 星仔高端定制 · 工厂直营官网

> 全屋定制工厂直营品牌官网 · 11 种主流风格 · 374 条真实案例 · 工厂直营 · 设计师 1V1

[![Go 1.21](https://img.shields.io/badge/Go-1.21-blue?logo=go)](https://go.dev)
[![React 18](https://img.shields.io/badge/React-18.2-149eca?logo=react)](https://react.dev)
[![MongoDB 6](https://img.shields.io/badge/MongoDB-6.0-success?logo=mongodb)](https://mongodb.com)
[![Docker](https://img.shields.io/badge/Docker-ready-blue?logo=docker)](https://docker.com)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

**星仔高端定制** 是面向"全屋定制"行业的 H5 / PC 三端官网。包含：

- 🏠 **前台**：首页 · 风格专题 · 案例详情 · 价格区间筛选
- 🔐 **后台**：运营端 Banner / 标签 / 案例 CRUD + 置顶
- 👥 **多角色权限**：普通用户 / 销售 / 供应商 / 管理员，按角色脱敏价格

---

## 📑 目录

- [技术栈](#-技术栈)
- [品牌设计语言](#-品牌设计语言)
- [目录结构](#-目录结构)
- [快速开始](#-快速开始)
- [多角色权限矩阵](#-多角色权限矩阵)
- [数据规模](#-数据规模)
- [API 速览](#-api-速览)
- [环境变量](#-环境变量)
- [Docker 部署](#-docker-部署)
- [公网访问 (ngrok 隧道)](#-公网访问-ngrok-隧道)
- [项目脚本](#-项目脚本)
- [关键文档](#-关键文档)

---

## ✨ 技术栈

| 层 | 技术 |
| --- | --- |
| 前端 | React 18 + Vite 5 + React Router 6 + Axios + 自研 CSS 设计系统 |
| 后端 | Go 1.21+ + Gin + go.mongodb.org/mongo-driver + golang-jwt/jwt v5 |
| 数据库 | MongoDB 6.0+ |
| 部署 | Docker + docker compose + nginx 反向代理 |
| 公网 | ngrok 隧道（开发）|
| 通信 | HTTP / JSON，统一响应信封 `{ code, msg, data }` |

---

## 🎨 品牌设计语言

- **星仔高端定制** —— 墨黑 `#0E0E10` × 香槟金 `#C8A45C` × 月白 `#F5F2EB`
- 标题字体：Cormorant Garamond + 中文 Noto Serif SC
- 大间距 / 克制留白 / 缓慢缓动 / 圆角与金属色描边
- Logo：圆形 香槟金字"星"

设计 token 集中在 [web/src/styles/global.css](web/src/styles/global.css)。

---

## 📁 目录结构

```
star/
├── server/                   Go 后端
│   ├── cmd/
│   │   ├── server/          HTTP API 入口
│   │   └── seed/            写入真实种子数据
│   ├── internal/
│   │   ├── config/          环境变量 + 配置加载
│   │   ├── db/              MongoDB 连接池
│   │   ├── handler/         HTTP Handler + Router
│   │   │   ├── router.go    路由表 + 中间件挂载
│   │   │   ├── auth.go      验证码 / 登录 / 当前用户
│   │   │   ├── public.go    公开接口 (Banner/Tags/Cases) + 价格脱敏
│   │   │   └── admin.go     后台 CRUD
│   │   ├── middleware/      JWT + 可选鉴权 + CORS
│   │   ├── model/           数据结构 + 角色常量 + 脱敏判断
│   │   └── repo/            Mongo 数据访问 (OR 语义筛选)
│   ├── scripts/             种子数据程序化生成
│   │   ├── seed.go          Banner/标签/案例/用户 全量写入
│   │   ├── cases.go         374 条案例 程序生成 (style×space×var 覆盖)
│   │   └── imagedl.go       本地图池 /img-pool/case_NN.jpg 选取
│   ├── pkg/response/        统一响应信封
│   ├── Dockerfile           后端 Docker 多阶段构建
│   └── go.mod
├── web/                      React 前端
│   ├── src/
│   │   ├── api/             axios 封装 + 接口列表
│   │   ├── components/      Header / Banner / CaseCard
│   │   ├── pages/           Home / StylePage / CaseDetail / Login / Me / Admin/*
│   │   ├── store/           useUser Context (含 isAdmin 判断)
│   │   └── styles/          全局 CSS + 主题 token
│   ├── Dockerfile           前端多阶段 (node build -> nginx serve)
│   ├── nginx.conf           nginx 服务配置 + /api 反代后端
│   └── package.json
├── docs/
│   ├── PRD.md
│   └── TECH.md
├── scripts/                  运维/测试脚本
│   ├── api_smoke.sh         API 冒烟 33 项
│   ├── coverage_test.sh     组合覆盖率 222 项
│   ├── role_test.sh         角色权限 + 价格脱敏 7 项
│   ├── image_test.sh        图片访问 200 验证 6 项
│   └── run_all_tests.sh     一键总验收
├── docker-compose.yml        mongo + api + web + 可选 seed
└── README.md
```

---

## 🚀 快速开始

### 前置
- Go ≥ 1.21
- Node ≥ 18
- MongoDB ≥ 6.0
- （可选）Docker + docker compose

### 方式 A — 本地裸跑

```bash
# 1. 启动 MongoDB（任选其一）
# A1) Docker
docker run -d --name star-mongo -p 27017:27017 mongo:6.0
# A2) brew
brew tap mongodb/brew && brew install mongodb-community
brew services start mongodb-community
# A3) 已有的云 Mongo (修改 STAR_MONGO_URI)

# 2. 启动后端
cd server
go mod tidy
go run ./cmd/seed   # 写入真实种子数据 (374 条案例 + 4 banner + 5 类标签)
go run ./cmd/server # 默认 :8080, 可通过 STAR_HTTP_PORT 覆盖

# 3. 启动前端
cd web
npm install --registry https://registry.npmmirror.com
npm run dev         # http://localhost:5173, /api 自动代理到 :8080
```

### 方式 B — Docker 一键

```bash
docker compose up -d            # 启 mongo + api + web, 健康检查就绪后 web 接流量
docker compose --profile seed run --rm seed   # 一次性 init, 写种子
docker compose logs -f web     # 看 web 日志
docker compose down -v          # 全删并清数据卷
```

打开浏览器访问 [http://localhost:5173](http://localhost:5173)。

---

## 👥 多角色权限矩阵

| 角色 | 手机号 | 入口 | 价格可见性 | 后台 |
| --- | --- | --- | --- | --- |
| **管理员** | `13800138000` | `/admin` | 精准价格 | ✅ |
| **销售** | `13900000001` / `13900000002` | — | 精准价格 | ❌ |
| **供应商** | `13700000001` / `13700000002` | — | 精准价格 | ❌ |
| **普通用户** | 任意手机号 | — | 仅价格区间 | ❌ |
| **匿名** | — | — | 仅价格区间 (脱敏为 0) | ❌ |

验证码在开发期固定为 `1234`，可在 `internal/config/config.go` 修改 `STAR_STATIC_CODE`。

普通用户看到列表 / 详情时的价格字段：

```jsonc
{
  "price": 0,            // 被后端置 0
  "priceLabel": "5-10万" // 价格区间恒可见
}
```

销售 / 供应商 / 管理员看到的：

```jsonc
{
  "price": 72300,       // 精准价
  "priceLabel": "5-10万"
}
```

---

## 📊 数据规模

种子数据 ([scripts/seed.go](server/scripts/seed.go)) 一次写入：

- **4 张** Banner（工厂直营 / 新中式 / 奶油风 / 意式轻奢）
- **5 类标签**：风格 11 / 空间 9 / 颜色 8 / 尺寸 6 / 价格 5
- **374 条** 案例 — 11 风格 × 8 空间 × 4 变体 + 11 设计师款 + 11 衣帽间旗舰款
  - 每个 (风格 × 空间) 至少 4 条变体覆盖 5 档价格 + 6 种尺寸 + 8 个颜色
  - **组合保障**：任意 (style + space + color + size + price) 任一选中都能命中 ≥1 条
  - 所有图片来自 [Unsplash](https://unsplash.com) 真实摄影作品的本地镜像 (33 张)，下载到 `web/public/img-pool/case_NN.jpg`，永不裂图

---

## 🔌 API 速览

### 公开

```
GET    /api/v1/banners                                  首页 Banner
GET    /api/v1/tags?type=style|space|color|size|price    标签
GET    /api/v1/cases?style=&space=&color=&size=&price=   案例列表 (二级 OR 筛选)
GET    /api/v1/cases/pinned                              置顶案例 (最多 8)
GET    /api/v1/cases/:id                                 案例详情
GET    /healthz                                          健康检查
```

### 鉴权

```
POST   /api/v1/auth/send-code      {phone}                 "发送"验证码
POST   /api/v1/auth/login          {phone, code}          返回 token + 用户资料
GET    /api/v1/me                                       Bearer token 查当前用户
```

### 后台 (admin only)

```
GET    /api/v1/admin/overview       统计概览 (banner/case/tag 数量)
GET    /api/v1/admin/stats/by-style 各风格案例数
GET    /api/v1/admin/banners        # Banner CRUD
POST   /api/v1/admin/banners
PUT    /api/v1/admin/banners/:id
DELETE /api/v1/admin/banners/:id
# 同样的 CRUD 模式应用于 /admin/tags 和 /admin/cases
POST   /api/v1/admin/cases/:id/pin                        切换置顶
```

完整列表与负载示例见 [docs/API.md](docs/API.md)。

---

## 🔐 环境变量

| Key | 默认 | 说明 |
| --- | --- | --- |
| `STAR_HTTP_PORT` | `8080` | 后端端口 |
| `STAR_MONGO_URI` | `mongodb://localhost:27017` | MongoDB URI |
| `STAR_DB_NAME` | `star` | 数据库名 |
| `STAR_JWT_SECRET` | `star-dev-secret-please-change` | JWT 签名密钥 (生产必改) |
| `STAR_ADMIN_PHONE` | `13800138000` | 管理员手机号白名单 |
| `STAR_SALES_PHONES` | `13900000001,13900000002` | 销售手机号白名单 |
| `STAR_SUPPLIER_PHONES` | `13700000001,13700000002` | 供应商手机号白名单 |
| `STAR_STATIC_CODE` | `1234` | 开发期验证码 (生产需接短信网关) |
| `STAR_CORS_ORIGINS` | `*` | CORS 白名单，逗号分隔 |
| `NGROK_AUTHTOKEN` | — | ngrok 隧道凭证 (公网访问用) |

---

## 🐳 Docker 部署

```bash
docker compose up -d
docker compose --profile seed run --rm seed   # 一次性写种子
docker compose logs -f                        # tail 所有
```

镜像：
- `starshop/api:latest` — Go 后端 ~30MB
- `starshop/web:latest` — nginx + 前端 dist ~50MB
- `mongo:6.0` — 官方镜像

数据持久化：
- `star_mongo_data` volume 挂到 mongo 容器 `/data/db`
- 后端 `web/public/img-pool/` 已 baked in 镜像，本地图池冗余存放

---

## 🌐 公网访问 (ngrok 隧道)

开发演示用：把 `localhost:5173` 暴露到公网。

```bash
# 一次性：把 token 加入 ngrok 配置
mkdir -p ~/.ngrok
cat > ~/.ngrok/ngrok.yml <<EOF
agent:
  authtoken: <NGROK_AUTHTOKEN>
  log: /tmp/ngrok.log
  log_level: info
version: 3
EOF

# 启动隧道
NGROK_AUTHTOKEN="<NGROK_AUTHTOKEN>" \
  ngrok http 5173
```

启动后会看到一个 https://*.ngrok-free.dev 公网域名。

> **注意**：ngrok free plan 一次只能给一个端口。本项目前端走 `/api` 反向代理到后端，所以只需暴露 5173。

---

## 📜 项目脚本

| 脚本 | 用途 |
| --- | --- |
| [scripts/api_smoke.sh](scripts/api_smoke.sh) | API 33 项冒烟 |
| [scripts/coverage_test.sh](scripts/coverage_test.sh) | 5 维组合覆盖率 222 项 |
| [scripts/role_test.sh](scripts/role_test.sh) | 角色权限 + 价格脱敏 7 项 |
| [scripts/image_test.sh](scripts/image_test.sh) | 图片访问 200 验证 6 项 |
| [scripts/run_all_tests.sh](scripts/run_all_tests.sh) | 一键全跑（**268/268 PASS**） |
| [scripts/ci.sh](scripts/ci.sh) | 一键启动 Mongo + 后端 + 前端 + 自动跑测试 |

跑测试：
```bash
bash scripts/run_all_tests.sh    # 268 项 5-10 秒
```

---

## 📚 关键文档

- [docs/PRD.md](docs/PRD.md) — 产品需求
- [docs/TECH.md](docs/TECH.md) — 技术架构
- [docs/API.md](docs/API.md) — 接口契约

---

## 📝 License

MIT — 仅供学习与商业自用。