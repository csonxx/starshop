#!/usr/bin/env bash
# Star 总体一键验收测试
# 覆盖: API 冒烟 / 组合覆盖率 / 角色权限 / 图片裂图
set -uo pipefail
cd "$(dirname "$0")/.."

BASE="${BASE:-http://localhost:8181/api/v1}"
SERVER="${SERVER:-http://localhost:8181}"
FRONT="${FRONT:-http://localhost:5173}"
FAIL=0

run_test() {
  local name="$1"
  shift
  local log
  log=$(mktemp)
  if "$@" >"$log" 2>&1; then
    if ! grep -Eq "结果:.*FAIL=0" "$log"; then
      cat "$log"
      echo "  ✗ $name 输出缺少成功汇总"
      FAIL=$((FAIL+1))
    else
      grep -E "(结果:|✓|✗|通过|FAIL)" "$log" | tail -10 || true
    fi
  else
    local code=$?
    cat "$log"
    echo "  ✗ $name 失败 (exit=$code)"
    FAIL=$((FAIL+1))
  fi
  rm -f "$log"
}

echo "============================================================"
echo " Star 高端定制官网 · 总体验收"
echo " BASE=$BASE  FRONT=$FRONT"
echo "============================================================"

# 1. API 冒烟
echo ""
echo "============================================================"
echo "  ① API 冒烟"
echo "============================================================"
run_test "API 冒烟" env BASE="$BASE" bash scripts/api_smoke.sh

# 2. 组合覆盖
echo ""
echo "============================================================"
echo "  ② 组合覆盖率 (单维+二维+三维+五维)"
echo "============================================================"
run_test "组合覆盖率" env BASE="$BASE" bash scripts/coverage_test.sh

# 3. 角色权限
echo ""
echo "============================================================"
echo "  ③ 角色权限 + 价格脱敏"
echo "============================================================"
run_test "角色权限" env BASE="$BASE" bash scripts/role_test.sh

# 4. 图片
echo ""
echo "============================================================"
echo "  ④ 图片裂图检查"
echo "============================================================"
run_test "图片裂图检查" env SERVER="$SERVER" FRONT="$FRONT" bash scripts/image_test.sh

echo ""
echo "============================================================"
if [ "$FAIL" -eq 0 ]; then
  echo "  全部通过 · 访问首页 $FRONT 验证 UI"
else
  echo "  验收失败 · $FAIL 个测试套件未通过"
fi
echo "============================================================"
exit "$FAIL"
