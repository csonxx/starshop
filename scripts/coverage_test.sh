#!/usr/bin/env bash
# Star 组合覆盖率测试 - 验证跨维度 AND、同维度多选 OR，并只查询真实存在组合
set -uo pipefail
BASE="${BASE:-http://localhost:8181/api/v1}"
PAGE_SIZE=60
PASS=0
FAIL=0
TMP_DIR=$(mktemp -d)
trap 'rm -rf "$TMP_DIR"' EXIT

ok() {
  PASS=$((PASS+1))
  printf "  \033[32m✓\033[0m %s\n" "$1"
}

bad() {
  FAIL=$((FAIL+1))
  printf "  \033[31m✗\033[0m %s\n" "$1"
}

fetch_cases() {
  local output="$1"
  shift
  curl --fail --silent --show-error --max-time 10 --get \
    --data-urlencode "pageSize=$PAGE_SIZE" "$@" "$BASE/cases" >"$output"
}

validate_response() {
  local response="$1"
  local label="$2"
  local style="${3:-}"
  local spaces="${4:-}"
  local colors="${5:-}"
  local sizes="${6:-}"
  local prices="${7:-}"
  if result=$(python3 - "$response" "$style" "$spaces" "$colors" "$sizes" "$prices" <<'PY'
import json
import sys

path, style, spaces_raw, colors_raw, sizes_raw, prices_raw = sys.argv[1:]
with open(path, encoding="utf-8") as handle:
    body = json.load(handle)

if body.get("code") != 0:
    raise SystemExit(f"业务 code={body.get('code')} message={body.get('message')}")

data = body.get("data") or {}
items = data.get("list")
total = data.get("total")
if not isinstance(items, list) or not isinstance(total, int):
    raise SystemExit("响应缺少 data.list 或 data.total")
if total < 1 or not items:
    raise SystemExit("真实组合查询结果为空")

spaces = set(filter(None, spaces_raw.split("\x1f")))
colors = set(filter(None, colors_raw.split("\x1f")))
sizes = set(filter(None, sizes_raw.split("\x1f")))
prices = set(filter(None, prices_raw.split("\x1f")))
errors = []
for item in items:
    item_id = item.get("id", "<missing-id>")
    if style and item.get("style") != style:
        errors.append(f"{item_id}: style={item.get('style')!r}")
    if spaces and item.get("space") not in spaces:
        errors.append(f"{item_id}: space={item.get('space')!r}")
    item_colors = set(item.get("colors") or [])
    if colors and item_colors.isdisjoint(colors):
        errors.append(f"{item_id}: colors={sorted(item_colors)!r}")
    if sizes and item.get("size") not in sizes:
        errors.append(f"{item_id}: size={item.get('size')!r}")
    if prices and item.get("priceLabel") not in prices:
        errors.append(f"{item_id}: priceLabel={item.get('priceLabel')!r}")

if errors:
    raise SystemExit("返回项不满足筛选条件: " + "; ".join(errors[:5]))
print(f"total={total}, checked={len(items)}")
PY
  ); then
    ok "$label ($result)"
  else
    bad "$label ($result)"
  fi
}

load_real_cases() {
  local first="$TMP_DIR/page-1.json"
  if ! fetch_cases "$first"; then
    bad "读取真实案例失败"
    return 1
  fi
  if ! python3 - "$first" "$BASE" "$PAGE_SIZE" "$TMP_DIR/all-cases.json" <<'PY'
import json
import math
import sys
import urllib.parse
import urllib.request

first_path, base, page_size_raw, output_path = sys.argv[1:]
page_size = int(page_size_raw)
with open(first_path, encoding="utf-8") as handle:
    first = json.load(handle)
if first.get("code") != 0:
    raise SystemExit("首个案例请求业务失败")
data = first.get("data") or {}
items = list(data.get("list") or [])
total = data.get("total")
if not isinstance(total, int) or total < 1 or not items:
    raise SystemExit("案例库为空或响应格式错误")
for page in range(2, math.ceil(total / page_size) + 1):
    query = urllib.parse.urlencode({"page": page, "pageSize": page_size})
    with urllib.request.urlopen(f"{base}/cases?{query}", timeout=10) as response:
        body = json.load(response)
    if body.get("code") != 0:
        raise SystemExit(f"第 {page} 页业务失败")
    items.extend((body.get("data") or {}).get("list") or [])
if len(items) != total:
    raise SystemExit(f"分页汇总数量 {len(items)} != total {total}")
with open(output_path, "w", encoding="utf-8") as handle:
    json.dump(items, handle, ensure_ascii=False)
print(total)
PY
  then
    bad "汇总真实案例失败"
    return 1
  fi
}

build_samples() {
  python3 - "$TMP_DIR/all-cases.json" "$TMP_DIR" <<'PY'
import json
import os
import sys

source, output = sys.argv[1:]
with open(source, encoding="utf-8") as handle:
    cases = json.load(handle)

def unique_rows(fields, limit=None):
    seen = set()
    rows = []
    for case in cases:
        color_values = case.get("colors") or [""] if "color" in fields else [""]
        for color in color_values:
            row = []
            for field in fields:
                if field == "color":
                    row.append(color)
                else:
                    row.append(case.get(field, ""))
            key = tuple(row)
            if all(row) and key not in seen:
                seen.add(key)
                rows.append(row)
                if limit and len(rows) >= limit:
                    return rows
    return rows

datasets = {
    "styles.tsv": unique_rows(["style"]),
    "spaces.tsv": unique_rows(["space"]),
    "colors.tsv": unique_rows(["color"]),
    "sizes.tsv": unique_rows(["size"]),
    "prices.tsv": unique_rows(["priceLabel"]),
    "two.tsv": unique_rows(["style", "space"]),
    "three.tsv": unique_rows(["style", "space", "color"], 48),
    "five.tsv": unique_rows(["style", "space", "color", "size", "priceLabel"], 48),
}
multi_rows = []
for case in cases:
    colors = list(dict.fromkeys(case.get("colors") or []))
    if len(colors) >= 2:
        multi_rows.append([case.get("style", ""), case.get("space", ""), colors[0], colors[1]])
        break
datasets["multi.tsv"] = multi_rows
for name, rows in datasets.items():
    with open(os.path.join(output, name), "w", encoding="utf-8") as handle:
        for row in rows:
            handle.write("\t".join(row) + "\n")
PY
}

run_single_dimension() {
  local title="$1"
  local key="$2"
  local file="$3"
  local value response
  echo ""
  echo "$title"
  while IFS=$'\t' read -r value; do
    [ -z "$value" ] && continue
    response="$TMP_DIR/single-${key}-${PASS}-${FAIL}.json"
    if fetch_cases "$response" --data-urlencode "$key=$value"; then
      case "$key" in
        style) validate_response "$response" "$key=$value" "$value" ;;
        space) validate_response "$response" "$key=$value" "" "$value" ;;
        color) validate_response "$response" "$key=$value" "" "" "$value" ;;
        size) validate_response "$response" "$key=$value" "" "" "" "$value" ;;
        price) validate_response "$response" "$key=$value" "" "" "" "" "$value" ;;
      esac
    else
      bad "$key=$value 请求失败"
    fi
  done <"$file"
}

echo "============================================"
echo " Star · 组合覆盖率测试 (AND 语义)"
echo "============================================"

if total=$(load_real_cases); then
  ok "读取数据库真实案例 ($total 条)"
else
  echo ""
  echo "============================================"
  echo " 结果:  PASS=$PASS  FAIL=$FAIL"
  echo "============================================"
  exit "$FAIL"
fi

build_samples

run_single_dimension "[1] 单维度 - 风格" style "$TMP_DIR/styles.tsv"
run_single_dimension "[2] 单维度 - 空间" space "$TMP_DIR/spaces.tsv"
run_single_dimension "[3] 单维度 - 颜色" color "$TMP_DIR/colors.tsv"
run_single_dimension "[4] 单维度 - 尺寸" size "$TMP_DIR/sizes.tsv"
run_single_dimension "[5] 单维度 - 价格" price "$TMP_DIR/prices.tsv"

echo ""
echo "[6] 二维度 - 全量真实 风格 × 空间"
while IFS=$'\t' read -r style space; do
  response="$TMP_DIR/two-${PASS}-${FAIL}.json"
  if fetch_cases "$response" --data-urlencode "style=$style" --data-urlencode "space=$space"; then
    validate_response "$response" "$style × $space" "$style" "$space"
  else
    bad "$style × $space 请求失败"
  fi
done <"$TMP_DIR/two.tsv"

echo ""
echo "[7] 三维度 - 真实 风格 × 空间 × 颜色 (最多 48 组)"
while IFS=$'\t' read -r style space color; do
  response="$TMP_DIR/three-${PASS}-${FAIL}.json"
  if fetch_cases "$response" --data-urlencode "style=$style" --data-urlencode "space=$space" --data-urlencode "color=$color"; then
    validate_response "$response" "$style × $space × $color" "$style" "$space" "$color"
  else
    bad "$style × $space × $color 请求失败"
  fi
done <"$TMP_DIR/three.tsv"

echo ""
echo "[8] 五维度 - 真实 风格 × 空间 × 颜色 × 尺寸 × 价格 (最多 48 组)"
while IFS=$'\t' read -r style space color size price; do
  response="$TMP_DIR/five-${PASS}-${FAIL}.json"
  if fetch_cases "$response" \
    --data-urlencode "style=$style" \
    --data-urlencode "space=$space" \
    --data-urlencode "color=$color" \
    --data-urlencode "size=$size" \
    --data-urlencode "price=$price"; then
    validate_response "$response" "$style × $space × $color × $size × $price" "$style" "$space" "$color" "$size" "$price"
  else
    bad "$style × $space × $color × $size × $price 请求失败"
  fi
done <"$TMP_DIR/five.tsv"

echo ""
echo "[9] 同维度多选 - OR，跨维度仍为 AND"
IFS=$'\t' read -r multi_style multi_space multi_color_one multi_color_two <"$TMP_DIR/multi.tsv"
if [ -n "${multi_style:-}" ] && [ -n "${multi_space:-}" ] && [ -n "${multi_color_one:-}" ] && [ -n "${multi_color_two:-}" ]; then
  response="$TMP_DIR/multi.json"
  color_values="$multi_color_one"$'\x1f'"$multi_color_two"
  if fetch_cases "$response" \
    --data-urlencode "style=$multi_style" \
    --data-urlencode "space=$multi_space" \
    --data-urlencode "color=$multi_color_one,$multi_color_two"; then
    validate_response "$response" "$multi_style × $multi_space × ($multi_color_one OR $multi_color_two)" "$multi_style" "$multi_space" "$color_values"
  else
    bad "同维度多选请求失败"
  fi
else
  bad "真实数据颜色不足 2 种，无法验证同维度多选"
fi

echo ""
echo "============================================"
echo " 结果:  PASS=$PASS  FAIL=$FAIL"
echo "============================================"
exit "$FAIL"
