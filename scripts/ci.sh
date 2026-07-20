#!/usr/bin/env bash
# Star CI - 全自动启动 MongoDB + 后端 + 前端 + 测试 + 关闭
# 用法：bash scripts/ci.sh
set -u
HERE="$(cd "$(dirname "$0")/.." && pwd)"
cd "$HERE"

LOG_DIR="$HERE/.ci"
mkdir -p "$LOG_DIR"

cleanup() {
  echo ""
  echo "[CI] 关闭服务..."
  [ -n "${SERVER_PID:-}" ] && kill "$SERVER_PID" 2>/dev/null
  [ -n "${FRONT_PID:-}" ] && kill "$FRONT_PID" 2>/dev/null
  [ -n "${MONGO_PID:-}" ] && kill "$MONGO_PID" 2>/dev/null
}
trap cleanup EXIT

echo "============================================"
echo " STAR · 一键 CI 流水线"
echo "============================================"

# ---------- 1. MongoDB ----------
echo ""
echo "[1/4] 启动 MongoDB..."
MONGO_BIN="/tmp/star-mongo-bin/mongod"
if [ ! -x "$MONGO_BIN" ]; then
  echo "  ✗ MongoDB 未安装，请先解压到 /tmp/star-mongo-bin/mongod"
  exit 1
fi
mkdir -p /tmp/star-mongo-data
"$MONGO_BIN" --dbpath /tmp/star-mongo-data --port 27017 --bind_ip 127.0.0.1 --logpath /tmp/star-mongo.log --fork > /dev/null 2>&1
MONGO_PID=$(pgrep -f "mongod --dbpath /tmp/star-mongo-data" | head -1)
echo "  ✓ MongoDB 启动 (PID $MONGO_PID, :27017)"

# ---------- 2. 后端 ----------
echo ""
echo "[2/4] 启动 Go 后端..."
cd "$HERE/server"
go build -o /tmp/star-server ./cmd/server || { echo "  ✗ 后端编译失败"; exit 1; }
go run ./cmd/seed > "$LOG_DIR/seed.log" 2>&1 || { echo "  ✗ seed 失败"; cat "$LOG_DIR/seed.log"; exit 1; }
STAR_HTTP_PORT=8181 /tmp/star-server > "$LOG_DIR/server.log" 2>&1 &
SERVER_PID=$!
sleep 3
echo "  ✓ 后端 PID=$SERVER_PID  http://localhost:8181"

# ---------- 3. 前端 ----------
echo ""
echo "[3/4] 启动 React 前端..."
cd "$HERE/web"
nohup npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
FRONT_PID=$!
sleep 5
echo "  ✓ 前端 PID=$FRONT_PID  http://localhost:5173"

# ---------- 4. 测试 ----------
echo ""
echo "[4/4] 执行测试套件..."
echo ""
echo "--- API 冒烟 ---"
BASE=http://localhost:8181/api/v1 bash "$HERE/scripts/api_smoke.sh" | tail -50
echo ""
echo "--- E2E 集成 ---"
bash "$HERE/scripts/e2e_smoke.sh" | tail -40

echo ""
echo "============================================"
echo " 全部完成 · 前端: http://localhost:5173"
echo "             后端: http://localhost:8181/api/v1"
echo "============================================"
echo ""
echo "按 Ctrl+C 关闭服务..."
wait