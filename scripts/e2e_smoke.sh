#!/usr/bin/env bash
# Star E2E - 模拟用户访问完整链路
set -uo pipefail
FRONT="${FRONT:-http://localhost:5173}"
PASS=0
FAIL=0
ok()  { PASS=$((PASS+1)); echo "  \033[32m✓\033[0m $1"; }
bad() { FAIL=$((FAIL+1)); echo "  \033[31m✗\033[0m $1"; }

echo "============================================"
echo " STAR · E2E 集成测试 (前端 → 后端 → DB)"
echo "============================================"

echo ""
echo "[1] 用户打开首页 (/)"
HTML=$(curl -s "$FRONT/")
echo "$HTML" | grep -Eiq "star|星仔" && ok "首页 HTML 含品牌标题" || bad "首页 title"
echo "$HTML" | grep -q "root" && ok "首页挂载点正确" || bad "root 容器"

echo ""
echo "[2] 首页加载数据 (Banner / 风格 / 案例)"
BN=$(curl -s "$FRONT/api/v1/banners" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['data']))")
[ "$BN" -ge 1 ] && ok "首页 Banner 加载 ($BN 张)" || bad "Banner 数量"

ST=$(curl -s "$FRONT/api/v1/tags?type=style" | python3 -c "import sys,json; d=json.load(sys.stdin); print(len(d['data']))")
[ "$ST" -ge 11 ] && ok "首页风格标签加载 ($ST 个, 目标 ≥11)" || bad "风格标签"

CASE_TOTAL=$(curl -s "$FRONT/api/v1/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
[ "$CASE_TOTAL" -ge 55 ] && ok "案例库 ($CASE_TOTAL 条, 目标 ≥55)" || bad "案例数量"

PIN=$(curl -s "$FRONT/api/v1/cases/pinned" | python3 -c "import sys,json; print(len(json.load(sys.stdin)['data'] or []))")
[ "$PIN" -le 8 ] && ok "置顶案例 ($PIN 条, 上限 8)" || bad "置顶数量超限"

echo ""
echo "[3] 点击风格标签 '奶油风' 触发筛选"
N=$(curl -s "$FRONT/api/v1/cases?style=cream" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
[ "$N" -ge 5 ] && ok "奶油风 筛选 ($N 条)" || bad "奶油风 筛选"

# 多风格测试
for style in new-chinese cream italian-luxury modern nordic japanese american wabi-sabi minimalist french industrial; do
  n=$(curl -s "$FRONT/api/v1/cases?style=$style" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
  [ "$n" -ge 5 ] && ok "  - $style: $n 条" || bad "  - $style 仅 $n 条"
done

echo ""
echo "[4] 二级筛选 - 空间 / 颜色 / 尺寸 / 价格"
N=$(curl -s --get --data-urlencode "space=主卧" "$FRONT/api/v1/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
[ "$N" -ge 1 ] && ok "空间=主卧 ($N 条)" || bad "空间筛选"

N=$(curl -s --get --data-urlencode "color=奶油白" "$FRONT/api/v1/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
[ "$N" -ge 1 ] && ok "颜色=奶油白 ($N 条)" || bad "颜色筛选"

SIZE=$(curl -s "$FRONT/api/v1/tags?type=size" | python3 -c "import sys,json; data=json.load(sys.stdin)['data']; print(data[0]['value'] if data else '')")
N=$(curl -s --get --data-urlencode "size=$SIZE" "$FRONT/api/v1/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
[ -n "$SIZE" ] && [ "$N" -ge 1 ] && ok "尺寸=$SIZE ($N 条)" || bad "尺寸筛选"

N=$(curl -s --get --data-urlencode "price=3-5万" "$FRONT/api/v1/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
[ "$N" -ge 1 ] && ok "价格=3-5万 ($N 条)" || bad "价格筛选"

echo ""
echo "[5] 用户点击列表项进详情页"
CASE_ID=$(curl -s "$FRONT/api/v1/cases?pageSize=1" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['list'][0]['id'])")
DETAIL=$(curl -s "$FRONT/api/v1/cases/$CASE_ID")
DETAIL_FAIL=$(echo "$DETAIL" | python3 -c "
import sys, json
d = json.load(sys.stdin)
c = d['data']
checks = [
    ('标题', bool(c.get('title'))),
    ('价格区间', bool(c.get('priceLabel')) and c.get('price') == 0),
    ('封面', bool(c.get('cover'))),
    ('风格', bool(c.get('style'))),
    ('空间', bool(c.get('space'))),
    ('面积', bool(c.get('area'))),
    ('亮点', bool(c.get('highlights'))),
    ('主材', bool(c.get('materials'))),
    ('五金', bool(c.get('hardware'))),
    ('多图', bool(c.get('images'))),
]
for name, ok in checks:
    print('  ✓ ' if ok else '  ✗ ', name)
print(sum(not ok for _, ok in checks))
"
)
DETAIL_RESULT=$(echo "$DETAIL_FAIL" | tail -1)
echo "$DETAIL_FAIL" | sed '$d'
PASS=$((PASS+10-DETAIL_RESULT))
FAIL=$((FAIL+DETAIL_RESULT))

echo ""
echo "[6] 用户点击登录"
RESP=$(curl -s -X POST "$FRONT/api/v1/auth/send-code" -H 'Content-Type: application/json' -d '{"phone":"13800138000"}')
CODE_EXPOSED=$(echo "$RESP" | python3 -c "import sys,json; d=json.load(sys.stdin); print('1' if d['data'].get('code') else '0')")
[ "$CODE_EXPOSED" = "0" ] && ok "验证码已发送且未回传明文" || bad "验证码响应异常"

TOKEN=$(curl -s -X POST "$FRONT/api/v1/auth/login" -H 'Content-Type: application/json' -d '{"phone":"13800138000","code":"1234"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['token'])")
[ -n "$TOKEN" ] && ok "JWT 签发成功" || bad "JWT 签发"

ROLE=$(curl -s -H "Authorization: Bearer $TOKEN" "$FRONT/api/v1/me" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['role'])")
[ "$ROLE" = "admin" ] && ok "管理员权限 ($ROLE)" || bad "角色: $ROLE"

echo ""
echo "[7] 进入运营后台 - 数据 / 增删改"
OV=$(curl -s -H "Authorization: Bearer $TOKEN" "$FRONT/api/v1/admin/overview")
if echo "$OV" | python3 -c "
import sys, json
o = json.load(sys.stdin)['data']
print(f'  ✓ Banner: {o[\"bannerCount\"]}')
print(f'  ✓ 一级风格: {o[\"styleTagCount\"]}')
print(f'  ✓ 全部标签: {o[\"tagCount\"]}')
print(f'  ✓ 案例总数: {o[\"caseCount\"]}')
print(f'  ✓ 置顶案例: {o[\"pinnedCount\"]}')
"; then
  PASS=$((PASS+5))
else
  bad "后台概览"
fi

if STATS=$(curl -s -H "Authorization: Bearer $TOKEN" "$FRONT/api/v1/admin/stats/by-style" | python3 -c "
import sys, json
data = json.load(sys.stdin)['data']
assert data
print('\n'.join(f'  ✓ {s[\"name\"]}: {s[\"count\"]} 案例' for s in data))
"); then
  echo "$STATS"
  PASS=$((PASS+1))
else
  bad "风格统计"
fi

echo ""
echo "============================================"
echo " 结果:  PASS=$PASS  FAIL=$FAIL"
echo "============================================"
exit "$FAIL"
