#!/usr/bin/env bash
# Star 角色权限 + 价格脱敏测试
set -u
BASE="${BASE:-http://localhost:8181/api/v1}"

PASS=0
FAIL=0

echo "============================================"
echo " Star · 角色权限 + 价格脱敏测试"
echo "============================================"

# 1. 匿名访问 - price 应为 0
echo ""
echo "[1] 匿名访问 - price 应被脱敏 (置 0)"
RESP=$(curl -s "$BASE/cases?pageSize=1")
ANON_PRICE=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['price'] if d['data']['list'] else 0)")
ANON_LABEL=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['priceLabel'] if d['data']['list'] else '')")
if [ "$ANON_PRICE" = "0" ] && [ -n "$ANON_LABEL" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 匿名 price=0 (脱敏), priceLabel=\"$ANON_LABEL\" (可见)"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 匿名 price=$ANON_PRICE label=$ANON_LABEL (期望 0)"
fi

# 2. 普通用户登录 (用任意手机号 + 1234) - 应得到 user 角色
echo ""
echo "[2] 普通用户 13912345678 - role=user"
TOKEN_USER=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"phone":"13912345678","code":"1234"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
RESP=$(curl -s -H "Authorization: Bearer $TOKEN_USER" "$BASE/cases?pageSize=1")
USER_PRICE=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['price'] if d['data']['list'] else 0)")
if [ "$USER_PRICE" = "0" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 普通用户 price=0 (脱敏)"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 普通用户 price=$USER_PRICE (期望 0)"
fi

# 3. 销售登录 (白名单: 13900000001) - 应得到 sales 角色, 看精准价
echo ""
echo "[3] 销售 13900000001 - role=sales"
TOKEN_SALES=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"phone":"13900000001","code":"1234"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
RESP=$(curl -s -H "Authorization: Bearer $TOKEN_SALES" "$BASE/cases?pageSize=1")
SALES_PRICE=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['price'] if d['data']['list'] else 0)")
if [ "$SALES_PRICE" != "0" ] && [ "$SALES_PRICE" != "None" ] && [ -n "$SALES_PRICE" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 销售 price=$SALES_PRICE (精准, 非 0)"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 销售 price=$SALES_PRICE (期望非 0)"
fi

# 4. 供应商登录 (白名单: 13700000001) - supplier 角色, 看精准价
echo ""
echo "[4] 供应商 13700000001 - role=supplier"
TOKEN_SUP=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"phone":"13700000001","code":"1234"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
RESP=$(curl -s -H "Authorization: Bearer $TOKEN_SUP" "$BASE/cases?pageSize=1")
SUP_PRICE=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['price'] if d['data']['list'] else 0)")
if [ "$SUP_PRICE" != "0" ] && [ "$SUP_PRICE" != "None" ] && [ -n "$SUP_PRICE" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 供应商 price=$SUP_PRICE (精准)"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 供应商 price=$SUP_PRICE (期望非 0)"
fi

# 5. 管理员登录 - 全权限
echo ""
echo "[5] 管理员 13800138000 - role=admin"
TOKEN_ADM=$(curl -s -X POST "$BASE/auth/login" -H 'Content-Type: application/json' -d '{"phone":"13800138000","code":"1234"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
RESP=$(curl -s -H "Authorization: Bearer $TOKEN_ADM" "$BASE/cases?pageSize=1")
ADM_PRICE=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d['data']['list'][0]['price'] if d['data']['list'] else 0)")
# 管理员同时能访问 /admin/overview
OV=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN_ADM" "$BASE/admin/overview")
if [ "$ADM_PRICE" != "0" ] && [ "$OV" = "200" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 管理员 price=$ADM_PRICE + 后台 200"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 管理员 price=$ADM_PRICE or /admin/overview=$OV"
fi

# 6. 销售访问后台 - 应该 403
echo ""
echo "[6] 销售访问 /admin/overview - 应被拒 (403)"
CODE=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $TOKEN_SALES" "$BASE/admin/overview")
if [ "$CODE" = "403" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 销售 /admin/overview = 403"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 销售 /admin/overview = $CODE (期望 403)"
fi

# 7. 详情页同样按角色脱敏
echo ""
echo "[7] 详情页价格脱敏"
ANY_ID=$(curl -s "$BASE/cases?pageSize=1" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['list'][0]['id'])")

# 匿名查详情
ANON_D=$(curl -s "$BASE/cases/$ANY_ID" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['price'])")
ANON_DL=$(curl -s "$BASE/cases/$ANY_ID" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['priceLabel'])")

# 销售查详情
SALES_D=$(curl -s -H "Authorization: Bearer $TOKEN_SALES" "$BASE/cases/$ANY_ID" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['price'])")

if [ "$ANON_D" = "0" ] && [ -n "$ANON_DL" ] && [ "$SALES_D" != "0" ]; then
  PASS=$((PASS+1))
  echo "  ✓ 详情: 匿名 price=0+label=\"$ANON_DL\" / 销售 price=$SALES_D"
else
  FAIL=$((FAIL+1))
  echo "  ✗ 详情: 匿名 price=$ANON_D label=$ANON_DL / 销售 price=$SALES_D"
fi

echo ""
echo "============================================"
echo " 结果:  PASS=$PASS  FAIL=$FAIL"
echo "============================================"
exit $FAIL