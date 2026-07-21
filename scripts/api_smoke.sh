#!/usr/bin/env bash
# Star API 端到端冒烟测试 - 全自动
set -u
BASE="${BASE:-http://127.0.0.1:8181/api/v1}"
PASS=0
FAIL=0

ok()   { PASS=$((PASS+1)); echo "  \033[32m✓\033[0m $1"; }
bad()  { FAIL=$((FAIL+1)); echo "  \033[31m✗\033[0m $1  =>  $2"; }

# 调用：GET /path
G() {
  curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE$1"
}

# 调用带 token
GA() {
  curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $1" "$BASE$2"
}

# 调用 POST/PUT/DELETE
PJSON() {
  curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X "$1" -H 'Content-Type: application/json' -d "$3" "$BASE$2"
}

PJSONA() {
  curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X "$1" -H 'Content-Type: application/json' -H "Authorization: Bearer $2" -d "$4" "$BASE$3"
}

assert_code_data() {
  local name="$1" code="$2" expect="$3"
  if [ "$code" = "$expect" ]; then
    if [ -s /tmp/star_resp.json ]; then
      ok "$name (HTTP $code)"
    else
      bad "$name" "empty body"
    fi
  else
    bad "$name" "HTTP $code"
    cat /tmp/star_resp.json
  fi
}

echo ""
echo "============================================"
echo " STAR · API 自动化冒烟测试"
echo "============================================"

echo ""
echo "[1] 公开接口 - Banner / 标签 / 案例"
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/banners")
[ "$code" = "200" ] && ok "GET /banners (HTTP 200)" || bad "GET /banners" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/tags?type=style")
[ "$code" = "200" ] && ok "GET /tags?type=style (HTTP 200)" || bad "GET /tags?type=style" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/tags?type=space")
[ "$code" = "200" ] && ok "GET /tags?type=space (HTTP 200)" || bad "GET /tags?type=space" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/tags?type=color")
[ "$code" = "200" ] && ok "GET /tags?type=color (HTTP 200)" || bad "GET /tags?type=color" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/tags?type=size")
[ "$code" = "200" ] && ok "GET /tags?type=size (HTTP 200)" || bad "GET /tags?type=size" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/tags?type=price")
[ "$code" = "200" ] && ok "GET /tags?type=price (HTTP 200)" || bad "GET /tags?type=price" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/cases")
[ "$code" = "200" ] && ok "GET /cases (HTTP 200)" || bad "GET /cases" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/cases?style=cream")
[ "$code" = "200" ] && ok "GET /cases?style=cream (HTTP 200)" || bad "GET /cases?style=cream" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/cases/pinned")
[ "$code" = "200" ] && ok "GET /cases/pinned (HTTP 200)" || bad "GET /cases/pinned" "HTTP $code"

echo ""
echo "[2] 鉴权 - 登录 + 验证码"
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X POST -H 'Content-Type: application/json' -d '{"phone":"13800138000"}' "$BASE/auth/send-code")
[ "$code" = "200" ] && ok "POST /auth/send-code (HTTP 200)" || bad "POST /auth/send-code" "HTTP $code"
CODE_EXPOSED=$(cat /tmp/star_resp.json | python3 -c "import sys,json; d=json.load(sys.stdin); print('1' if d.get('data',{}).get('code') else '0')" 2>/dev/null)
[ "$CODE_EXPOSED" = "0" ] && ok "验证码响应不回传明文" || bad "验证码响应" "泄露验证码"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X POST -H 'Content-Type: application/json' -d '{"phone":"13800138000","code":"1234"}' "$BASE/auth/login")
[ "$code" = "200" ] && ok "POST /auth/login (admin)" || bad "POST /auth/login" "HTTP $code"

TOKEN=$(cat /tmp/star_resp.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('token',''))" 2>/dev/null)
ROLE=$(cat /tmp/star_resp.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('user',{}).get('role',''))" 2>/dev/null)

if [ -n "$TOKEN" ]; then
  ok "JWT token 已签发 (role=$ROLE)"
else
  bad "JWT token" "未拿到 token"
fi

# 普通用户登录
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X POST -H 'Content-Type: application/json' -d '{"phone":"13900000001","code":"1234"}' "$BASE/auth/login")
[ "$code" = "200" ] && ok "POST /auth/login (普通用户)" || bad "POST /auth/login user" "HTTP $code"
USER_TOKEN=$(cat /tmp/star_resp.json | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('token',''))" 2>/dev/null)

# 错误验证码
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X POST -H 'Content-Type: application/json' -d '{"phone":"13800138000","code":"0000"}' "$BASE/auth/login")
[ "$code" = "401" ] && ok "POST /auth/login (错误验证码应被拒)" || bad "wrong code" "HTTP $code"

echo ""
echo "[3] /me 鉴权"
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/me")
[ "$code" = "401" ] && ok "GET /me (无 token 应 401)" || bad "/me no token" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/me")
[ "$code" = "200" ] && ok "GET /me (admin token)" || bad "/me admin" "HTTP $code"

echo ""
echo "[4] 案例详情 + 过滤"
CASES=$(curl -s -m 6 "$BASE/cases?pageSize=1" | python3 -c "import sys,json; d=json.load(sys.stdin); lst=d.get('data',{}).get('list') or []; print(lst[0].get('id','') if lst else '')" 2>/dev/null)
if [ -n "$CASES" ]; then
  code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/cases/$CASES")
  [ "$code" = "200" ] && ok "GET /cases/:id ($CASES)" || bad "/cases/:id" "HTTP $code"
else
  bad "拿案例 ID" "cases 为空"
fi

# 多条件筛选
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/cases?style=cream&space=%E4%B8%BB%E5%8D%A7&size=2.0m")
[ "$code" = "200" ] && ok "GET /cases 多条件过滤" || bad "/cases filter" "HTTP $code"

echo ""
echo "[5] 管理员接口"
code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" "$BASE/admin/overview")
[ "$code" = "401" ] && ok "GET /admin/overview (无 token 应 401)" || bad "/admin no token" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $USER_TOKEN" "$BASE/admin/overview")
[ "$code" = "403" ] && ok "GET /admin/overview (普通用户应 403)" || bad "/admin 普通用户" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/admin/overview")
[ "$code" = "200" ] && ok "GET /admin/overview (admin)" || bad "/admin admin" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/admin/stats/styles")
[ "$code" = "200" ] && ok "GET /admin/stats/styles" || bad "/admin stats" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/admin/banners")
[ "$code" = "200" ] && ok "GET /admin/banners" || bad "/admin/banners" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/admin/tags")
[ "$code" = "200" ] && ok "GET /admin/tags" || bad "/admin/tags" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/admin/cases")
[ "$code" = "200" ] && ok "GET /admin/cases" || bad "/admin/cases" "HTTP $code"

code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -H "Authorization: Bearer $TOKEN" "$BASE/admin/cases/not-an-id")
[ "$code" = "400" ] && ok "GET /admin/cases/:id (非法 ID 应 400)" || bad "/admin/cases invalid id" "HTTP $code"

echo ""
echo "[6] 后台 CRUD - 标签 / Banner / 案例"
NEW_TAG=$(curl -s -m 6 -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
  -d '{"type":"color","name":"雾霾蓝","value":"smog-blue-test","color":"#7A8FA6","sort":99,"enabled":true}' \
  "$BASE/admin/tags" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
if [ -n "$NEW_TAG" ]; then
  ok "POST /admin/tags (新建测试标签 $NEW_TAG)"
  code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X DELETE -H "Authorization: Bearer $TOKEN" "$BASE/admin/tags/$NEW_TAG")
  [ "$code" = "200" ] && ok "DELETE /admin/tags/:id" || bad "del tag" "HTTP $code"
else
  bad "新建测试标签" "未拿到 id"
fi

NEW_BN=$(curl -s -m 6 -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"测试 Banner","image":"https://example.com/x.jpg","sort":99,"enabled":true}' \
  "$BASE/admin/banners" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
if [ -n "$NEW_BN" ]; then
  ok "POST /admin/banners (新建测试 Banner $NEW_BN)"
  code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X DELETE -H "Authorization: Bearer $TOKEN" "$BASE/admin/banners/$NEW_BN")
  [ "$code" = "200" ] && ok "DELETE /admin/banners/:id" || bad "del banner" "HTTP $code"
else
  bad "新建 Banner" "未拿到 id"
fi

NEW_CASE=$(curl -s -m 6 -X POST -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"测试案例","style":"new-chinese","space":"客厅","colors":["原木"],"size":"1.8m","area":"10㎡","price":12345,"priceLabel":"1-3万","cover":"https://example.com/c.jpg","images":[],"highlights":["亮点1"],"materials":["板材1"],"hardware":["五金1"],"pinned":false,"enabled":true}' \
  "$BASE/admin/cases" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('data',{}).get('id',''))" 2>/dev/null)
if [ -n "$NEW_CASE" ]; then
  ok "POST /admin/cases (新建测试案例 $NEW_CASE)"
  # 切换置顶
  code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X POST -H "Authorization: Bearer $TOKEN" "$BASE/admin/cases/$NEW_CASE/pin")
  [ "$code" = "200" ] && ok "POST /admin/cases/:id/pin" || bad "toggle pin" "HTTP $code"
  # 更新
  code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X PUT -H 'Content-Type: application/json' -H "Authorization: Bearer $TOKEN" \
    -d '{"title":"测试案例-更新","style":"new-chinese","space":"客厅","colors":[],"size":"1.8m","area":"10㎡","price":22222,"priceLabel":"1-3万","cover":"https://example.com/c.jpg","images":[],"highlights":[],"materials":[],"hardware":[],"pinned":true,"enabled":true}' \
    "$BASE/admin/cases/$NEW_CASE")
  [ "$code" = "200" ] && ok "PUT /admin/cases/:id" || bad "update case" "HTTP $code"
  # 删除
  code=$(curl -s -m 6 -o /tmp/star_resp.json -w "%{http_code}" -X DELETE -H "Authorization: Bearer $TOKEN" "$BASE/admin/cases/$NEW_CASE")
  [ "$code" = "200" ] && ok "DELETE /admin/cases/:id" || bad "del case" "HTTP $code"
else
  bad "新建案例" "未拿到 id"
fi

echo ""
echo "============================================"
echo " 结果:  PASS=$PASS  FAIL=$FAIL"
echo "============================================"
exit $FAIL
