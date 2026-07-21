#!/usr/bin/env bash
# Star 图片访问 200 验证 - 防止裂图
set -uo pipefail
HERE="$(cd "$(dirname "$0")/.." && pwd)"
WEB_ROOT="$HERE/web"
SERVER="${SERVER:-http://localhost:8181}"
FRONT="${FRONT:-http://localhost:5173}"

PASS=0
FAIL=0

ok()  { PASS=$((PASS+1)); echo "  \033[32m✓\033[0m $1"; }
bad() { FAIL=$((FAIL+1)); echo "  \033[31m✗\033[0m $1"; }

echo "============================================"
echo " Star · 图片加载验证"
echo "============================================"

echo ""
echo "[1] 本地图池 440 张风格空间图全部有效"
i=0
zero=0
invalid=0
for f in $WEB_ROOT/public/img-pool/case_*_*.jpg; do
  i=$((i+1))
  name=$(basename "$f")
  size=$(stat -f %z "$f" 2>/dev/null || stat -c %s "$f")
  if [ "$size" -lt 1000 ]; then
    zero=$((zero+1))
  fi
  if ! python3 - "$f" <<'PY'
import sys
from PIL import Image
with Image.open(sys.argv[1]) as image:
    image.verify()
PY
  then
    invalid=$((invalid+1))
  fi
done
echo "  共 $i 张, 小于 1KB $zero 张, 无法解码 $invalid 张"
if [ "$zero" = "0" ] && [ "$invalid" = "0" ] && [ "$i" = "440" ]; then
  ok "440 张图全部 ≥ 1KB 且可解码"
else
  bad "图池异常"
fi

echo ""
echo "[2] 后端返回的 cover URL 在前端可达 (前端 dev proxy)"
# 拉案例拿 cover, 前端代理访问
ANY_ID=$(curl -s "$SERVER/api/v1/cases?pageSize=1" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['list'][0]['id'])")
COVER_PATH=$(curl -s "$SERVER/api/v1/cases/$ANY_ID" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['cover'])")
echo "  测试 URL: $COVER_PATH"
code=$(curl -s -o /dev/null -w "%{http_code}" "$FRONT$COVER_PATH")
if [ "$code" = "200" ]; then
  ok "前端代理访问 cover $COVER_PATH -> 200"
else
  bad "前端访问 $COVER_PATH -> $code"
fi

# 直接本地文件 (绕过前端, 验证文件确实存在)
FULL="$WEB_ROOT/public/${COVER_PATH#/}"
if [ -f "$FULL" ]; then
  bytes=$(stat -f %z "$FULL" 2>/dev/null || stat -c %s "$FULL")
  ok "文件存在 $COVER_PATH ($bytes bytes)"
else
  bad "文件不存在 $FULL"
fi

echo ""
echo "[3] 前端 dev server 代理 案例数据"
code=$(curl -s -o /dev/null -w "%{http_code}" "$FRONT/api/v1/cases?pageSize=2")
[ "$code" = "200" ] && ok "前端代理 /api/v1/cases -> 200" || bad "前端代理 /api/v1/cases -> $code"

# 批量抽样 30 张 cover 检查全部 200
echo ""
echo "[4] 案例 cover 全量 200 检查 (抽样)"
COVERS=$(curl -s "$SERVER/api/v1/cases?pageSize=60" | python3 -c "import sys,json; print('\n'.join(c['cover'] for c in json.load(sys.stdin)['data']['list']))")
cnt=0
fails=0
while IFS= read -r url; do
  cnt=$((cnt+1))
  code=$(curl -s -o /dev/null -w "%{http_code}" "$FRONT$url")
  if [ "$code" != "200" ]; then
    fails=$((fails+1))
    echo "  ✗ $url -> $code"
  fi
done <<< "$COVERS"
echo "  抽样 $cnt 个 cover, 失败 $fails 个"
if [ "$fails" = "0" ]; then
  ok "所有抽样 cover 都 200"
else
  bad "$fails 个 cover 不是 200"
fi

# 批量抽样 images 检查
echo ""
echo "[5] 案例 images[] 全量 200 检查 (抽样)"
IMG_URLS=$(curl -s "$SERVER/api/v1/cases?pageSize=60" | python3 -c "
import sys, json
for c in json.load(sys.stdin)['data']['list']:
    for u in c.get('images', []):
        if u: print(u)
")
cnt=0
fails=0
while IFS= read -r url; do
  cnt=$((cnt+1))
  code=$(curl -s -o /dev/null -w "%{http_code}" "$FRONT$url")
  if [ "$code" != "200" ]; then
    fails=$((fails+1))
    echo "  ✗ $url -> $code"
  fi
done <<< "$IMG_URLS"
echo "  抽样 $cnt 个 images, 失败 $fails 个"
if [ "$fails" = "0" ]; then
  ok "所有抽样 images 都 200"
else
  bad "$fails 个 images 不是 200"
fi

# Banner 图
echo ""
echo "[6] Banner 图 200 检查"
BN=$(curl -s "$SERVER/api/v1/banners" | python3 -c "import sys,json; print('\n'.join(b['image'] for b in json.load(sys.stdin)['data']))")
while IFS= read -r url; do
  [ -z "$url" ] && continue
  code=$(curl -s -o /dev/null -w "%{http_code}" "$FRONT$url")
  if [ "$code" = "200" ]; then
    ok "$url -> 200"
  else
    bad "$url -> $code"
  fi
done <<< "$BN"

echo ""
echo "============================================"
echo " 结果:  PASS=$PASS  FAIL=$FAIL"
echo "============================================"
exit "$FAIL"
