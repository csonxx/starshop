# =============================================================================
# 星仔高端定制 - 前端 Docker 镜像 (多阶段构建)
# 阶段 1: node:18-alpine 安装依赖并 build 出 dist/
# 阶段 2: nginx:alpine 把 dist/ 服务出来, 同时代理 /api 到 backend
# =============================================================================

# ---------- 阶段 1: 构建静态资源 ----------
FROM node:18-alpine AS builder

# 设置时区 (PRC)
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata \
  && cp /usr/share/zoneinfo/$TZ /etc/localtime \
  && echo "$TZ" > /etc/timezone

WORKDIR /app

# 1) 先 copy lockfile 利用 Docker layer cache
COPY package*.json ./

# 2) 装依赖 (国内镜像加速)
RUN npm config set registry https://registry.npmmirror.com \
  && npm install --no-audit --no-fund

# 3) 源代码 (ci 时被 .dockerignore 过滤掉)
COPY . .

# 4) 编译生产包
RUN npm run build

# ---------- 阶段 2: nginx 服务 ----------
FROM nginx:1.25-alpine

# 时区 + alpine 包
ENV TZ=Asia/Shanghai
RUN apk add --no-cache tzdata curl \
  && cp /usr/share/zoneinfo/$TZ /etc/localtime \
  && echo "$TZ" > /etc/timezone

# nginx 主配置 + 服务前端 + 反向代理后端
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 把阶段 1 build 产物拷过来
COPY --from=builder /app/dist /usr/share/nginx/html

# 健康检查
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl --fail --silent --show-error http://localhost/healthz >/dev/null || exit 1

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
