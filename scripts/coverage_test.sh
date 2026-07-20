#!/usr/bin/env bash
# Star 组合覆盖率测试 - 确保所有 5 维筛选组合都至少 1 条命中
set -u
BASE="${BASE:-http://localhost:8181/api/v1}"

STYLES=("new-chinese" "cream" "italian-luxury" "modern" "nordic" "japanese" "american" "wabi-sabi" "minimalist" "french" "industrial")
SPACES=("客厅" "餐厅" "主卧" "次卧" "书房" "衣帽间" "玄关" "儿童房")
COLORS=("雾霾蓝" "莫兰迪绿" "奶油白" "焦糖棕" "烟灰" "暮青" "原木" "胭脂粉")
SIZES=("1.2m" "1.5m" "1.8m" "2.0m" "2.4m" "通顶")
PRICES=("1万以下" "1-3万" "3-5万" "5-10万" "10万+")

# URL-encode via python helper
urlenc() {
  python3 -c "import sys, urllib.parse; print(urllib.parse.quote(sys.argv[1]))" "$1"
}

PASS=0
FAIL=0
echo "============================================"
echo " Star · 组合覆盖率测试 (OR 语义)"
echo "============================================"

# 1. 单维度 - 风格 (确保每个风格 ≥ 1 条)
echo ""
echo "[1] 单维度 - 风格 (11 种)"
for s in "${STYLES[@]}"; do
  n=$(curl -s "$BASE/cases?style=$s" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
  if [ "$n" -ge 1 ]; then
    PASS=$((PASS+1))
    printf "  \033[32m✓\033[0m %-20s %s 条\n" "$s" "$n"
  else
    FAIL=$((FAIL+1))
    printf "  \033[31m✗\033[0m %-20s %s 条\n" "$s" "$n"
  fi
done

# 2. 单维度 - 空间 (9 种)
echo ""
echo "[2] 单维度 - 空间 (9 种)"
for sp in "${SPACES[@]}"; do
  n=$(curl -s --get --data-urlencode "space=$sp" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
  if [ "$n" -ge 1 ]; then
    PASS=$((PASS+1))
    printf "  \033[32m✓\033[0m %-10s %s 条\n" "$sp" "$n"
  else
    FAIL=$((FAIL+1))
    printf "  \033[31m✗\033[0m %-10s %s 条\n" "$sp" "$n"
  fi
done

# 3. 单维度 - 颜色 (8 种)
echo ""
echo "[3] 单维度 - 颜色 (8 种)"
for cl in "${COLORS[@]}"; do
  n=$(curl -s --get --data-urlencode "color=$cl" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
  if [ "$n" -ge 1 ]; then
    PASS=$((PASS+1))
    printf "  \033[32m✓\033[0m %-10s %s 条\n" "$cl" "$n"
  else
    FAIL=$((FAIL+1))
    printf "  \033[31m✗\033[0m %-10s %s 条\n" "$cl" "$n"
  fi
done

# 4. 单维度 - 尺寸 (6 种)
echo ""
echo "[4] 单维度 - 尺寸 (6 种)"
for sz in "${SIZES[@]}"; do
  n=$(curl -s --get --data-urlencode "size=$sz" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
  if [ "$n" -ge 1 ]; then
    PASS=$((PASS+1))
    printf "  \033[32m✓\033[0m %-10s %s 条\n" "$sz" "$n"
  else
    FAIL=$((FAIL+1))
    printf "  \033[31m✗\033[0m %-10s %s 条\n" "$sz" "$n"
  fi
done

# 5. 单维度 - 价格 (5 档)
echo ""
echo "[5] 单维度 - 价格 (5 档)"
for pr in "${PRICES[@]}"; do
  n=$(curl -s --get --data-urlencode "price=$pr" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
  if [ "$n" -ge 1 ]; then
    PASS=$((PASS+1))
    printf "  \033[32m✓\033[0m %-10s %s 条\n" "$pr" "$n"
  else
    FAIL=$((FAIL+1))
    printf "  \033[31m✗\033[0m %-10s %s 条\n" "$pr" "$n"
  fi
done

# 6. 二维度 - 风格 × 空间 (11 × 9 = 88)
echo ""
echo "[6] 二维度 - 风格 × 空间 (88 组合)"
cnt=0
fail2=0
for s in "${STYLES[@]}"; do
  for sp in "${SPACES[@]}"; do
    n=$(curl -s --get --data-urlencode "style=$s" --data-urlencode "space=$sp" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
    cnt=$((cnt+1))
    if [ "$n" -ge 1 ]; then
      PASS=$((PASS+1))
    else
      FAIL=$((FAIL+1)); fail2=$((fail2+1))
      printf "  \033[31m✗\033[0m %s × %s = 0\n" "$s" "$sp"
    fi
  done
done
echo "  88 组合中 $((88-fail2)) 通过, $fail2 失败"

# 7. 三维度 - 风格 × 空间 × 颜色 (11 × 9 × 8 = 792) - 仅抽样
echo ""
echo "[7] 三维度 - 风格 × 空间 × 颜色 (抽样 30 个组合)"
sample_count=0
sample_pass=0
for s in "${STYLES[@]:0:4}"; do
  for sp in "${SPACES[@]:0:4}"; do
    for cl in "${COLORS[@]:0:3}"; do
      n=$(curl -s --get --data-urlencode "style=$s" --data-urlencode "space=$sp" --data-urlencode "color=$cl" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
      sample_count=$((sample_count+1))
      if [ "$n" -ge 1 ]; then
        PASS=$((PASS+1))
        sample_pass=$((sample_pass+1))
      else
        FAIL=$((FAIL+1))
        printf "  \033[31m✗\033[0m %s × %s × %s = 0\n" "$s" "$sp" "$cl"
      fi
    done
  done
done
echo "  抽样 $sample_count 中 $sample_pass 通过"

# 8. 五维度 - 风格 × 空间 × 颜色 × 尺寸 × 价格
echo ""
echo "[8] 五维度 - 风格 × 空间 × 颜色 × 尺寸 × 价格 (抽样)"
for s in "cream" "italian-luxury" "modern"; do
  for sp in "主卧" "客厅"; do
    for cl in "奶油白" "暮青"; do
      for sz in "1.8m" "通顶"; do
        for pr in "1-3万" "5-10万"; do
          n=$(curl -s --get \
            --data-urlencode "style=$s" --data-urlencode "space=$sp" \
            --data-urlencode "color=$cl" --data-urlencode "size=$sz" \
            --data-urlencode "price=$pr" "$BASE/cases" | python3 -c "import sys,json; print(json.load(sys.stdin)['data']['total'])")
          if [ "$n" -ge 1 ]; then
            PASS=$((PASS+1))
          else
            FAIL=$((FAIL+1))
            printf "  \033[31m✗\033[0m %s × %s × %s × %s × %s = 0\n" "$s" "$sp" "$cl" "$sz" "$pr"
          fi
        done
      done
    done
  done
done

echo ""
echo "============================================"
echo " 结果:  PASS=$PASS  FAIL=$FAIL"
echo "============================================"
exit $FAIL