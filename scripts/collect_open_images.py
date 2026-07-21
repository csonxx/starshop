import hashlib
import html
import json
import os
import random
import re
import time
import urllib.parse
import urllib.request
from collections import defaultdict
from datetime import date
from io import BytesIO

from PIL import Image, ImageOps, ImageStat


ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
POOL = os.path.join(ROOT, "web", "public", "img-pool")
USER_AGENT = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 Chrome/124 Safari/537.36"
TARGET_PER_COMBINATION = int(os.environ.get("STAR_IMG_PER_COMBO", "5"))
MIN_WIDTH = 900
MIN_HEIGHT = 650
MAX_EDGE = 1800


STYLES = [
    ("new-chinese", "新中式", ["新中式", "中式", "东方", "胡桃", "深木", "茶色"]),
    ("cream", "奶油风", ["奶油", "奶白", "奶咖", "米白", "浅木", "暖白"]),
    ("italian-luxury", "意式轻奢", ["意式", "轻奢", "岩板", "玻璃", "黑金", "灰色"]),
    ("modern", "现代简约", ["现代", "简约", "高级灰", "无拉手", "悬浮", "极简"]),
    ("nordic", "北欧", ["北欧", "原木", "白橡", "浅木", "清新", "简约"]),
    ("japanese", "日式无印", ["日式", "原木", "无印", "榻榻米", "木色", "极简"]),
    ("american", "美式", ["美式", "复古", "胡桃", "樱桃木", "实木", "护墙"]),
    ("wabi-sabi", "侘寂", ["侘寂", "微水泥", "原木", "素色", "米灰", "自然"]),
    ("minimalist", "极简", ["极简", "纯白", "一门到顶", "无拉手", "悬空", "简约"]),
    ("french", "法式", ["法式", "奶油", "拱形", "复古", "白色", "线条"]),
    ("industrial", "工业风", ["工业", "黑色", "金属", "深木", "铁艺", "灰色"]),
]

SPACES = [
    ("living-room", "客厅", ["电视柜", "电视墙", "整墙柜", "收纳柜", "展示柜"]),
    ("dining-room", "餐厅", ["餐边柜", "酒柜", "岛台", "餐厅柜", "储物柜"]),
    ("master-bedroom", "主卧", ["衣柜", "床尾柜", "床头柜", "梳妆柜", "通顶衣柜"]),
    ("guest-bedroom", "次卧", ["衣柜", "榻榻米", "床柜一体", "组合柜", "次卧柜"]),
    ("study", "书房", ["书柜", "书桌柜", "书架", "整墙书柜", "书房柜"]),
    ("walk-in-closet", "衣帽间", ["衣帽间", "步入式衣柜", "开放衣柜", "衣柜岛台", "玻璃衣柜"]),
    ("entryway", "玄关", ["鞋柜", "玄关柜", "入户柜", "鞋帽柜", "换鞋凳"]),
    ("kids-room", "儿童房", ["儿童衣柜", "书桌柜", "高低床", "榻榻米", "儿童房收纳"]),
]

CABINET_WORDS = {
    "柜", "衣柜", "电视柜", "餐边柜", "鞋柜", "玄关柜", "书柜", "酒柜", "橱柜",
    "衣帽间", "收纳", "储物", "书架", "岛台", "榻榻米", "定制", "壁柜", "组合柜",
}

TO8TO_TAGS = {
    "living-room": ["dianshigui"],
    "dining-room": ["canbiangui"],
    "master-bedroom": ["yigui"],
    "guest-bedroom": ["yigui"],
    "study": ["shugui"],
    "walk-in-closet": ["yigui"],
    "entryway": ["xiegui"],
    "kids-room": ["yigui", "shugui"],
}

TO8TO_SPACE_IDS = {
    "living-room": 1,
    "dining-room": 3,
    "master-bedroom": 2,
    "guest-bedroom": 2,
    "study": 7,
    "walk-in-closet": 7,
    "entryway": 8,
    "kids-room": 2,
}

ZHUXIAOBANG_CASES = [
    "7169513127468827172",
    "7172434160458564110",
    "7205760016463168059",
    "7239614854519554615",
    "7356445839771828755",
    "7359866536527397427",
    "7070713148419867149",
]

BRAND_PAGES = [
    ("兔宝宝", "https://www.tubaobao.com/products/137", "奶油风"),
    ("兔宝宝", "https://www.tubaobao.com/products/135", "法式"),
    ("兔宝宝", "https://www.tubaobao.com/products/136", "现代轻奢"),
    ("莫干山", "https://www.mgsyg.com/proInfo/722.html", "现代轻奢"),
    ("莫干山", "https://www.mgsyg.com/proInfo/190.html", "极简轻奢"),
    ("莫干山", "https://www.mgsyg.com/proInfo/723.html", "现代"),
]


def request_bytes(url, referer="", attempts=4):
    headers = {"User-Agent": USER_AGENT, "Accept": "*/*"}
    if referer:
        headers["Referer"] = referer
    last_error = None
    for attempt in range(attempts):
        try:
            request = urllib.request.Request(url, headers=headers)
            with urllib.request.urlopen(request, timeout=60) as response:
                return response.read(35 * 1024 * 1024)
        except Exception as error:
            last_error = error
            if attempt + 1 < attempts:
                time.sleep(1.5 * (attempt + 1))
    raise last_error


def request_text(url, referer=""):
    return request_bytes(url, referer).decode("utf-8", "ignore")


def clean(value):
    return re.sub(r"\s+", " ", html.unescape(re.sub(r"<[^>]+>", " ", value or ""))).strip()


def collect_to8to():
    candidates = []
    detail_ids = defaultdict(list)
    seen_ids = defaultdict(set)
    for space_slug, tags in TO8TO_TAGS.items():
        for tag in tags:
            list_url = f"https://m.to8to.com/xiaoguotu/hot/{tag}"
            try:
                text = request_text(list_url)
            except Exception as error:
                print(f"to8to list failed: {list_url}: {error}")
                continue
            for detail_id in re.findall(r'/xiaoguotu/p(\d+)\.html', text):
                if detail_id not in seen_ids[space_slug]:
                    seen_ids[space_slug].add(detail_id)
                    detail_ids[space_slug].append(detail_id)
    required_space_count = TARGET_PER_COMBINATION * 4 + 12
    for space_slug, space_id in TO8TO_SPACE_IDS.items():
        for page in range(1, 4):
            if len(detail_ids[space_slug]) >= required_space_count:
                break
            list_url = f"https://xiaoguotu.to8to.com/pic_space{space_id}?page={page}"
            try:
                text = request_text(list_url)
            except Exception as error:
                print(f"to8to space failed: {list_url}: {error}")
                continue
            for detail_id in re.findall(r'/p(\d+)\.html', text):
                if detail_id != "1" and detail_id not in seen_ids[space_slug]:
                    seen_ids[space_slug].add(detail_id)
                    detail_ids[space_slug].append(detail_id)
    for space_slug, ids in detail_ids.items():
        saved = 0
        for detail_id in ids:
            if saved >= required_space_count:
                break
            detail_url = f"https://m.to8to.com/xiaoguotu/p{detail_id}.html"
            try:
                text = request_text(detail_url, "https://m.to8to.com/")
            except Exception:
                continue
            match = re.search(r"['\"]detail_data['\"]\s*:\s*(\[\{[\s\S]*?\}\])\s*,\s*['\"]dataurl['\"]", text)
            if not match:
                continue
            try:
                items = json.loads(match.group(1))
            except json.JSONDecodeError:
                continue
            if not items:
                continue
            item = items[0]
            image_url = item.get("newFileName") or item.get("originFileName") or ""
            title = clean(item.get("title") or item.get("element") or "")
            element = clean(item.get("element") or "")
            width = int(item.get("width") or 0)
            height = int(item.get("height") or 0)
            if image_url and width >= MIN_WIDTH and height >= MIN_HEIGHT and any(word in f"{title} {element}" for word in CABINET_WORDS):
                candidates.append({
                    "provider": "土巴兔",
                    "source_page": detail_url,
                    "original_url": image_url,
                    "title": title,
                    "creator": clean(item.get("authorName") or "土巴兔设计师"),
                    "width": width,
                    "height": height,
                    "space_hint": space_slug,
                    "text": f"{title} {element}",
                })
                saved += 1
        print(f"to8to {space_slug}: {saved}")
    return candidates


def find_dicts(value, output):
    if isinstance(value, dict):
        if any(key in value for key in ("dynamic_url", "dynamic_back_up_url", "watermark_url")):
            output.append(value)
        for child in value.values():
            find_dicts(child, output)
    elif isinstance(value, list):
        for child in value:
            find_dicts(child, output)


def collect_zhuxiaobang():
    candidates = []
    for case_id in ZHUXIAOBANG_CASES:
        detail_url = f"https://m.zhuxiaobang.com/room/detail/{case_id}"
        try:
            text = request_text(detail_url)
        except Exception as error:
            print(f"zhuxiaobang failed: {detail_url}: {error}")
            continue
        match = re.search(r"window\._SSR_DATA\s*=\s*(\{[\s\S]*?\});?\s*</script>", text)
        if not match:
            continue
        try:
            data = json.loads(match.group(1))
        except json.JSONDecodeError:
            continue
        images = []
        find_dicts(data, images)
        page_title = clean(re.search(r"<title>(.*?)</title>", text, re.S | re.I).group(1)) if re.search(r"<title>(.*?)</title>", text, re.S | re.I) else "住小帮全屋定制案例"
        seen = set()
        for item in images:
            image_url = item.get("dynamic_url") or item.get("dynamic_back_up_url") or item.get("url") or ""
            if not image_url or image_url in seen:
                continue
            seen.add(image_url)
            width = int(item.get("width") or 0)
            height = int(item.get("height") or 0)
            title = clean(item.get("paragraph_title") or item.get("title") or page_title)
            context = f"{page_title} {title}"
            if not any(word in context for word in CABINET_WORDS):
                continue
            candidates.append({
                "provider": "住小帮",
                "source_page": detail_url,
                "original_url": image_url,
                "title": title,
                "creator": "住小帮公开案例",
                "width": width,
                "height": height,
                "space_hint": "",
                "text": context,
            })
    return candidates


def extract_page_images(page_url):
    text = request_text(page_url)
    urls = []
    for pattern in (
        r'(?:src|data-src|data-original)=["\']([^"\']+)',
        r'url\(["\']?([^\)"\']+)',
        r'(/upfiles/[^"\' <>()]+\.(?:jpg|jpeg|png|webp))',
        r'(/uploads/image/[^"\' <>()]+\.(?:jpg|jpeg|png|webp))',
    ):
        for value in re.findall(pattern, text, re.I):
            url = urllib.parse.urljoin(page_url, html.unescape(value))
            if re.search(r'\.(?:jpg|jpeg|png|webp)(?:\?|$)', url, re.I) and url not in urls:
                urls.append(url)
    title_match = re.search(r"<title>(.*?)</title>", text, re.S | re.I)
    return clean(title_match.group(1)) if title_match else page_url, urls


def collect_brands():
    candidates = []
    for brand, page_url, style in BRAND_PAGES:
        try:
            title, urls = extract_page_images(page_url)
        except Exception as error:
            print(f"brand failed: {page_url}: {error}")
            continue
        for image_url in urls:
            if any(word in image_url.lower() for word in ("logo", "weixin", "douyin", "xiaohongshu", "taobao", "bottom", "banner", "code", "ewm")):
                continue
            candidates.append({
                "provider": brand,
                "source_page": page_url,
                "original_url": image_url,
                "title": title,
                "creator": brand,
                "width": 0,
                "height": 0,
                "space_hint": "",
                "text": f"{title} {style} 全屋定制 柜体",
            })
    return candidates


def image_features(data):
    with Image.open(BytesIO(data)) as source:
        image = ImageOps.exif_transpose(source).convert("RGB")
        width, height = image.size
        if width < MIN_WIDTH or height < MIN_HEIGHT:
            raise ValueError(f"small image {width}x{height}")
        if width / height > 3.2 or height / width > 3.2:
            raise ValueError(f"extreme aspect {width}x{height}")
        sample = ImageOps.fit(image, (8, 8), Image.Resampling.LANCZOS).convert("L")
        mean = sum(sample.getdata()) / 64
        phash = 0
        for value in sample.getdata():
            phash = (phash << 1) | int(value >= mean)
        color_sample = image.resize((1, 1), Image.Resampling.LANCZOS)
        rgb = color_sample.getpixel((0, 0))
        variance = sum(ImageStat.Stat(image.resize((64, 64))).var) / 3
        if variance < 90:
            raise ValueError("flat image")
        return image, width, height, phash, rgb


def hamming(left, right):
    return (left ^ right).bit_count()


def color_family(rgb):
    red, green, blue = rgb
    brightness = (red + green + blue) / 3
    if brightness > 205 and max(rgb) - min(rgb) < 28:
        return "light"
    if brightness < 95:
        return "dark"
    if red > blue + 18 and red >= green:
        return "warm"
    if blue > red + 15:
        return "cool"
    return "wood"


def style_color_families(style_slug):
    return {
        "new-chinese": {"dark", "wood", "warm"},
        "cream": {"light", "warm", "wood"},
        "italian-luxury": {"dark", "cool", "wood"},
        "modern": {"light", "cool", "wood"},
        "nordic": {"light", "wood", "warm"},
        "japanese": {"wood", "light", "warm"},
        "american": {"warm", "wood", "dark"},
        "wabi-sabi": {"wood", "warm", "light"},
        "minimalist": {"light", "cool", "wood"},
        "french": {"light", "warm", "wood"},
        "industrial": {"dark", "wood", "cool"},
    }[style_slug]


def text_score(candidate, style_terms, space_terms, style_slug, space_slug):
    text = candidate["text"].lower()
    score = 0
    score += sum(9 for term in space_terms if term.lower() in text)
    score += sum(4 for term in style_terms if term.lower() in text)
    if candidate.get("space_hint") == space_slug:
        score += 18
    elif candidate.get("space_hint"):
        score -= 24
    if candidate["provider"] in {"兔宝宝", "莫干山"}:
        score += 8
    if candidate["provider"] == "住小帮":
        score += 6
    return score


def normalize_and_save(image, output_path):
    image.thumbnail((MAX_EDGE, MAX_EDGE), Image.Resampling.LANCZOS)
    image.save(output_path, "JPEG", quality=86, optimize=True, progressive=True)
    data = open(output_path, "rb").read()
    with Image.open(output_path) as saved:
        width, height = saved.size
    return width, height, len(data), hashlib.sha256(data).hexdigest()


def main():
    os.makedirs(POOL, exist_ok=True)
    candidates = collect_brands() + collect_zhuxiaobang() + collect_to8to()
    random.Random(20260721).shuffle(candidates)
    print(f"candidates collected: {len(candidates)}")

    prepared = []
    exact_hashes = set()
    perceptual_hashes = []
    for index, candidate in enumerate(candidates, 1):
        try:
            data = request_bytes(candidate["original_url"], candidate["source_page"])
            digest = hashlib.sha256(data).hexdigest()
            if digest in exact_hashes:
                continue
            image, width, height, phash, rgb = image_features(data)
            if any(hamming(phash, old) <= 5 for old in perceptual_hashes):
                continue
            exact_hashes.add(digest)
            perceptual_hashes.append(phash)
            prepared.append({**candidate, "image": image.copy(), "original_width": width, "original_height": height, "color_family": color_family(rgb), "phash": phash})
        except Exception:
            continue
        if index % 100 == 0:
            print(f"prepared {len(prepared)} / scanned {index}")
    print(f"quality images: {len(prepared)}")
    for name in os.listdir(POOL):
        if name.startswith("case_") and name.endswith(".jpg"):
            os.remove(os.path.join(POOL, name))

    images = []
    usage = defaultdict(int)
    combo_hashes = defaultdict(list)
    for style_slug, style_name, style_terms in STYLES:
        for space_slug, space_name, space_terms in SPACES:
            ranked = sorted(prepared, key=lambda item: (
                text_score(item, style_terms, space_terms, style_slug, space_slug),
                item["color_family"] in style_color_families(style_slug),
                -usage[item["original_url"]],
                item["original_width"] * item["original_height"],
            ), reverse=True)
            selected = []
            for item in ranked:
                if item.get("space_hint") and item["space_hint"] != space_slug:
                    continue
                if item["color_family"] not in style_color_families(style_slug):
                    continue
                if any(hamming(item["phash"], old) <= 8 for old in combo_hashes[(style_slug, space_slug)]):
                    continue
                selected.append(item)
                if len(selected) == TARGET_PER_COMBINATION:
                    break
            if len(selected) < TARGET_PER_COMBINATION:
                for item in ranked:
                    if item in selected:
                        continue
                    if item.get("space_hint") and item["space_hint"] != space_slug:
                        continue
                    if any(hamming(item["phash"], old) <= 8 for old in combo_hashes[(style_slug, space_slug)]):
                        continue
                    selected.append(item)
                    if len(selected) == TARGET_PER_COMBINATION:
                        break
            if len(selected) < TARGET_PER_COMBINATION:
                raise RuntimeError(f"insufficient images: {style_slug}/{space_slug} {len(selected)}")
            for image_index, item in enumerate(selected, 1):
                file_name = f"case_{style_slug}_{space_slug}_{image_index:02d}.jpg"
                output_path = os.path.join(POOL, file_name)
                width, height, byte_count, sha256 = normalize_and_save(item["image"].copy(), output_path)
                usage[item["original_url"]] += 1
                combo_hashes[(style_slug, space_slug)].append(item["phash"])
                images.append({
                    "file": file_name,
                    "style": style_slug,
                    "style_name": style_name,
                    "space": space_name,
                    "cabinet_type": space_terms[0],
                    "color_family": item["color_family"],
                    "title": item["title"],
                    "creator": item["creator"],
                    "provider": item["provider"],
                    "source_page": item["source_page"],
                    "original_url": item["original_url"],
                    "original_width": item["original_width"],
                    "original_height": item["original_height"],
                    "width": width,
                    "height": height,
                    "bytes": byte_count,
                    "sha256": sha256,
                    "fetched_at": str(date.today()),
                })
            print(f"saved {style_name}/{space_name}: {len(selected)}")

    payload = {
        "generated_at": str(date.today()),
        "catalog": "土巴兔、住小帮、兔宝宝、莫干山公开案例页",
        "usage_note": "测试项目案例素材；按柜体品类、空间、色系统一筛选，保留来源页、原图地址与校验值。",
        "images": images,
    }
    with open(os.path.join(POOL, "SOURCES.json"), "w", encoding="utf-8") as handle:
        json.dump(payload, handle, ensure_ascii=False, indent=2)
        handle.write("\n")
    with open(os.path.join(POOL, "SOURCES.md"), "w", encoding="utf-8") as handle:
        handle.write("# 图片来源清单\n\n")
        handle.write("图片来自土巴兔、住小帮、兔宝宝、莫干山公开案例页，按柜体品类、空间与色系筛选。完整字段见 `SOURCES.json`。\n\n")
        handle.write("| 文件 | 风格 | 空间 | 柜体 | 来源 | 原页面 |\n")
        handle.write("| --- | --- | --- | --- | --- | --- |\n")
        for item in images:
            title = item["title"].replace("|", "\\|")
            handle.write(f"| `{item['file']}` | {item['style_name']} | {item['space']} | {item['cabinet_type']} | {item['provider']} | [{title}]({item['source_page']}) |\n")
    print(f"completed: {len(images)} images")


if __name__ == "__main__":
    main()
