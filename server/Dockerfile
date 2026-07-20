# =============================================================================
# 星仔高端定制 - Go 后端 Docker 镜像
# 多阶段:
#   builder 阶段: golang:1.21-alpine 编译静态二进制
#   runner  阶段: distroless/alpine 把二进制 + imgage seed 兜底包
# =============================================================================

# ---------- 阶段 1: 编译 ----------
FROM golang:1.21-alpine AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# alpine 用镜像源加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /src

# 1) 利用 docker layer cache - go.mod / go.sum 先拷贝
COPY go.mod go.sum ./
RUN go mod download -x

# 2) 复制源代码
COPY . .

# 3) 编译 server (静态二进制, 剥离调试符号, 减小镜像)
RUN go build -trimpath -ldflags "-s -w" -o /out/server ./cmd/server

# 4) 编译 seed 工具 (一次性 init 容器用)
RUN go build -trimpath -ldflags "-s -w" -o /out/seed ./cmd/seed

# ---------- 阶段 2: 运行镜像 ----------
FROM alpine:3.19

ENV TZ=Asia/Shanghai

# alpine 时区 + ca-certificates + curl (健康检查)
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
  && apk add --no-cache tzdata ca-certificates curl \
  && cp /usr/share/zoneinfo/$TZ /etc/localtime \
  && echo "$TZ" > /etc/timezone \
  && update-ca-certificates

WORKDIR /app

# 拷贝两个二进制
COPY --from=builder /out/server /app/server
COPY --from=builder /out/seed   /app/seed

# 非 root 身份跑
RUN adduser -D -u 10001 appuser
USER appuser

EXPOSE 8080

# 环境变量默认值 - 可在 docker-compose 里覆盖
ENV STAR_HTTP_PORT=8080 \
    STAR_MONGO_URI=mongodb://mongo:27017 \
    STAR_DB_NAME=star \
    STAR_JWT_SECRET=star-dev-secret-please-change \
    STAR_ADMIN_PHONE=13800138000 \
    STAR_STATIC_CODE=1234 \
    STAR_CORS_ORIGINS=*

# 健康检查: GET /healthz 需返回 200
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -sI http://localhost:${STAR_HTTP_PORT}/healthz | grep "200" >/dev/null || exit 1

CMD ["/app/server"] 