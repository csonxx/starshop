# Star 全屋定制官网 — 技术架构文档

## 1. 总体架构

```
+------------------+        HTTP/JSON        +-------------------+
|  React (Vite)    |  <--------------------> |   Go (Gin)        |
|  H5 / PC Web     |                         |   RESTful API     |
+------------------+                         +---------+---------+
                                                        |
                                                MongoDB Driver
                                                        |
                                                +---------+---------+
                                                |     MongoDB       |
                                                |   (本地默认)       |
                                                +-------------------+
```

- 前端：React 18 + Vite + React Router + Axios + CSS Modules + Framer Motion。
- 后端：Go 1.21+ + Gin + go.mongodb.org/mongo-driver + JWT (golang-jwt/jwt)。
- 数据库：MongoDB 6.x (本地 `mongodb://localhost:27017`，库名 `star`)。
- 通讯：HTTP/JSON，所有响应遵循统一信封：
  ```json
  { "code": 0, "msg": "ok", "data": { ... } }
  ```

## 2. 目录结构

```
star/
├── server/                # Go 后端
│   ├── cmd/server/        # 入口
│   ├── cmd/seed/          # 种子数据命令
│   ├── internal/
│   │   ├── config/        # 配置
│   │   ├── db/            # MongoDB 初始化
│   │   ├── handler/       # HTTP 路由
│   │   ├── middleware/    # JWT / CORS
│   │   ├── model/         # 数据模型
│   │   ├── repo/          # 数据访问
│   │   └── service/       # 业务逻辑
│   ├── pkg/response/      # 统一响应
│   ├── scripts/seed.go    # 种子数据源 (真实数据)
│   └── go.mod
├── web/                   # React 前端
│   ├── public/
│   ├── src/
│   │   ├── api/           # axios 封装
│   │   ├── components/    # 通用组件
│   │   ├── pages/         # 页面
│   │   ├── store/         # 用户状态
│   │   ├── styles/        # 全局样式
│   │   ├── App.jsx
│   │   └── main.jsx
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
└── docs/
    ├── PRD.md
    └── TECH.md
```

## 3. 后端设计

### 3.1 配置

`.env` (通过 `os.Getenv` 读取，提供默认值)：

| Key | 默认值 | 说明 |
| --- | --- | --- |
| `STAR_HTTP_PORT` | 8080 | HTTP 监听端口 |
| `STAR_MONGO_URI` | `mongodb://localhost:27017` | MongoDB 连接串 |
| `STAR_DB_NAME` | `star` | 数据库名 |
| `STAR_JWT_SECRET` | `star-dev-secret` | JWT 密钥 |
| `STAR_CORS_ORIGINS` | `*` | CORS 白名单 |

### 3.2 模块

| 模块 | 文件 | 说明 |
| --- | --- | --- |
| 入口 | `cmd/server/main.go` | 装载配置、连接 Mongo、注册路由 |
| 种子 | `cmd/seed/main.go` | 调用 `scripts/seed.go` 写入真实数据 |
| 配置 | `internal/config/config.go` | 环境变量解析 |
| 数据库 | `internal/db/mongo.go` | Mongo 客户端单例 |
| 模型 | `internal/model/*.go` | User / Banner / Tag / Case |
| 仓库 | `internal/repo/*.go` | 数据访问封装 |
| 中间件 | `internal/middleware/{auth,cors,log}.go` | 鉴权、CORS、日志 |
| Handler | `internal/handler/*.go` | HTTP 路由 |

### 3.3 数据模型

```go
// User
{ _id, phone, role: "user|admin", createdAt }

// Banner
{ _id, title, image, link, sort, enabled, createdAt }

// Tag (一二级标签统一)
{ _id, type: "style|space|color|size|price", name, value, color?, icon?, sort, enabled }

// Case
{
  _id, title, styleTag, spaceTag, colors[], sizeLabel, area, price,
  images[], highlights[], materials[], hardware[], pinned, enabled, createdAt
}
```

### 3.4 API 列表

| Method | Path | 说明 | 鉴权 |
| --- | --- | --- | --- |
| POST | `/api/v1/auth/send-code` | 发送验证码 (固定 1234) | 否 |
| POST | `/api/v1/auth/login` | 手机号 + 验证码登录 | 否 |
| GET  | `/api/v1/me` | 当前用户 | 是 |
| GET  | `/api/v1/banners` | 首页 Banner | 否 |
| GET  | `/api/v1/tags?type=style` | 一级 / 二级标签 | 否 |
| GET  | `/api/v1/cases` | 案例列表 (支持 style/space/color/size/price 过滤) | 否 |
| GET  | `/api/v1/cases/:id` | 案例详情 | 否 |
| GET  | `/api/v1/cases/pinned` | 置顶案例 | 否 |
| --- | --- | --- | --- |
| POST | `/api/v1/admin/banners` | 新建 Banner | admin |
| PUT  | `/api/v1/admin/banners/:id` | 更新 Banner | admin |
| DELETE | `/api/v1/admin/banners/:id` | 删除 Banner | admin |
| POST | `/api/v1/admin/tags` | 新建标签 | admin |
| PUT  | `/api/v1/admin/tags/:id` | 更新标签 | admin |
| DELETE | `/api/v1/admin/tags/:id` | 删除标签 | admin |
| POST | `/api/v1/admin/cases` | 新建案例 | admin |
| PUT  | `/api/v1/admin/cases/:id` | 更新案例 | admin |
| DELETE | `/api/v1/admin/cases/:id` | 删除案例 | admin |
| POST | `/api/v1/admin/cases/:id/pin` | 切换置顶 | admin |
| GET  | `/api/v1/admin/overview` | 后台首页统计 | admin |

### 3.5 鉴权

- 登录成功后下发 JWT (`Authorization: Bearer <token>`)。
- 中间件 `Auth()` 解析 token 写入 `c.Set("user", user)`。
- 中间件 `Admin()` 在 `Auth()` 之上校验 `role == "admin"`。

## 4. 前端设计

### 4.1 技术栈

- React 18 + Vite
- React Router 6
- Axios (拦截器统一注入 token)
- Framer Motion (列表 / 详情切换动效)
- 自研设计系统 (CSS Variables + CSS Modules)

### 4.2 设计语言 — "星仔高端定制"

- 调色板：墨黑 `#0E0E10` / 香槟金 `#C8A45C` / 月白 `#F5F2EB` / 砂金 `#D6B988` / 暮青 `#1F3A3D`。
- 字体：标题 `Cormorant Garamond` (Display) + 正文 `Noto Serif SC`。
- 留白：克制、纵深感、大间距、栅格 12 列。
- 动效：缓慢 ease-out 渐显 + 轻微 y 位移；详情切换使用 mask + clip-path。

### 4.3 路由

| Path | Page |
| --- | --- |
| `/` | 首页 |
| `/cases/:id` | 详情页 |
| `/login` | 登录 |
| `/me` | 我的 |
| `/admin` | 后台首页 |
| `/admin/banners` | Banner 管理 |
| `/admin/tags` | 标签管理 |
| `/admin/cases` | 案例管理 |
| `/admin/cases/new` `/admin/cases/:id` | 案例编辑 |

### 4.4 适配策略

- `<=768` 启用 H5 布局 (单列 / 顶部抽屉 / 底部 TabBar)。
- `768-1280` 启用平板布局 (双列)。
- `>=1280` PC 布局 (4 列 / 侧边栏 / 大 Banner)。

## 5. 数据初始化

`cmd/seed/main.go` 启动时执行：

1. 清空目标集合 (避免重复)。
2. 写入 1 个管理员 `13800138000`。
3. 写入 4 张 Banner (主题与设计文案)。
4. 写入一级风格标签 11 个 (覆盖市面主流风格)。
5. 写入二级标签：空间 7 个 / 颜色 8 个色卡 / 尺寸 6 个 / 价格 5 档。
6. 写入真实风格的真实案例，每种风格 ≥ 5 条，价格、面积、亮点真实可信。

## 6. 部署与运行

### 后端
```bash
cd server
go mod tidy
go run ./cmd/server      # 启动 API
go run ./cmd/seed        # 写入种子数据
```

### 前端
```bash
cd web
npm install
npm run dev              # 开发
npm run build            # 生产构建
```

### 前端环境变量
- `VITE_API_BASE`：默认 `/api/v1`，开发期通过 Vite 代理转发到 `http://localhost:8080`。

## 7. 安全

- 所有后台接口校验 JWT + role。
- 入参校验：手机号格式、验证码长度、必填字段。
- 前端不存敏感数据；后台图片上传 (本版本使用 URL 占位)。

## 8. 风险与限制

- 验证码为开发期固定值，生产环境需替换为短信通道。
- 图片使用本地占位 (Unsplash + 风格色块生成)，生产环境需接入 OSS。
- 暂无 SSR / SEO 优化 (二期)。