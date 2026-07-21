package scripts

import (
	"fmt"
	"hash/fnv"
)

// imageRoot 部署时为 /img-pool/... 静态 URL, 由 nginx/web serve.
const imageRoot = "/img-pool"

var imageSpaceSlugs = map[string]string{
	"客厅":  "living-room",
	"餐厅":  "dining-room",
	"主卧":  "master-bedroom",
	"次卧":  "guest-bedroom",
	"书房":  "study",
	"衣帽间": "walk-in-closet",
	"玄关":  "entryway",
	"儿童房": "kids-room",
}

var poolEntries = buildPoolEntries()

type poolEntry struct {
	File  string
	Style string
	Space string
}

func buildPoolEntries() []poolEntry {
	out := make([]poolEntry, 0, len(STYLES)*len(SPACES))
	for _, style := range STYLES {
		for _, space := range SPACES {
			out = append(out, poolEntry{
				File:  fmt.Sprintf("case_%s_%s_01.jpg", style.Key, imageSpaceSlugs[space]),
				Style: style.Key,
				Space: space,
			})
		}
	}
	return out
}

// LocalImageMap 由 SortLocalMap() 自动构建, 这里是结构占位.
var LocalImageMap = map[string]map[string][]string{}

func init() {
	rebuildLocalImageMap()
}

// rebuildLocalImageMap 把手工分配的图池投影为 LocalImageMap.
func rebuildLocalImageMap() {
	m := map[string]map[string][]string{}
	for _, e := range poolEntries {
		if m[e.Style] == nil {
			m[e.Style] = map[string][]string{}
		}
		m[e.Style][e.Space] = append(m[e.Style][e.Space], fmt.Sprintf("%s/%s", imageRoot, e.File))
	}
	LocalImageMap = m
}

// PickCaseImage 按 (style, space) 选一张图片, 缺图时跨风格/空间降级, 不会再随机.
func PickCaseImage(style, space string, v int) string {
	if n := len(LocalImageMap[style][space]); n > 0 {
		idx := v % n
		if idx < 0 {
			idx = -idx
		}
		return LocalImageMap[style][space][idx]
	}
	// 1) 同空间跨风格兜底.
	for _, alt := range fallbackStyles {
		if n := len(LocalImageMap[alt][space]); n > 0 {
			return LocalImageMap[alt][space][0]
		}
	}
	// 2) 同风格其它空间兜底.
	for _, sp := range fallbackSpaces {
		if n := len(LocalImageMap[style][sp]); n > 0 {
			return LocalImageMap[style][sp][0]
		}
	}
	// 3) 最终兜底.
	if n := len(poolEntries); n > 0 {
		idx := v % n
		if idx < 0 {
			idx = -idx
		}
		return fmt.Sprintf("%s/%s", imageRoot, poolEntries[idx].File)
	}
	return fmt.Sprintf("%s/case_modern_living-room_01.jpg", imageRoot)
}

// PickBannerImage 给 banner 选图: 按 idx 决定 4 张轮播.
func PickBannerImage(idx int) string {
	// 4 张 banner 使用 4 个不同风格的代表图:
	// 0 工厂直营 -> 整体感强的现代客厅
	// 1 新中式 -> 新中式客厅
	// 2 奶油风 -> 奶油客厅
	// 3 意式轻奢 -> 意式轻奢客厅
	switch idx {
	case 0:
		return PickCaseImage("industrial", "客厅", 0) // 工业感宽阔, 适合"工厂直营"主题
	case 1:
		return PickCaseImage("new-chinese", "客厅", 0)
	case 2:
		return PickCaseImage("cream", "客厅", 0)
	default:
		return PickCaseImage("italian-luxury", "客厅", 0)
	}
}

var fallbackStyles = []string{
	"modern", "minimalist", "cream", "new-chinese",
	"nordic", "japanese", "italian-luxury", "american",
	"wabi-sabi", "french", "industrial",
}

var fallbackSpaces = []string{
	"客厅", "主卧", "次卧", "餐厅", "书房",
	"衣帽间", "玄关", "儿童房",
}

func hashV(s string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32())
}

// LoadBannerImages 返回 4 张 banner 轮播图 URL.
func LoadBannerImages() []string {
	out := make([]string, 4)
	for i := 0; i < 4; i++ {
		out[i] = PickBannerImage(i)
	}
	return out
}
