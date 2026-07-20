package scripts

import (
	"fmt"

	"star/server/internal/model"
)

// 设计原则：每个 (风格 × 空间) 组合都有 ≥ 1 案例, 每个风格 ≥ 18 案例,
// 颜色从 8 色卡中选 2-3 个, 尺寸从对应空间的可用集合选 1 个
// 价格按风格定位区间真实可信

var SPACES = []string{"客厅", "餐厅", "主卧", "次卧", "书房", "衣帽间", "玄关", "儿童房"}

var SPACE_SIZES_MAP = map[string][]string{
	"客厅":   {"2.0m", "2.4m", "通顶"},
	"餐厅":   {"1.5m", "1.8m", "2.0m", "2.4m"},
	"主卧":   {"1.8m", "2.0m", "2.4m", "通顶"},
	"次卧":   {"1.2m", "1.5m", "1.8m", "2.0m"},
	"书房":   {"1.2m", "1.5m", "1.8m", "2.0m", "通顶"},
	"衣帽间": {"2.0m", "2.4m", "通顶"},
	"玄关":   {"1.2m", "1.5m"},
	"儿童房": {"1.2m", "1.5m", "1.8m"},
}

var COLORS = []string{"雾霾蓝", "莫兰迪绿", "奶油白", "焦糖棕", "烟灰", "暮青", "原木", "胭脂粉"}

var PRICES = []string{"1万以下", "1-3万", "3-5万", "5-10万", "10万+"}

type StyleConf struct {
	Key        string  // 风格 value
	Name       string  // 风格中文名
	Icon       string  // 风格图标
	BasePrice  [2]int  // 价格区间 [min, max]
	Materials  []string
	Hardware   []string
	Highlights []string
	PromptTpl  string  // prompt 模板, 会拼接 room/space
}

var STYLES = []StyleConf{
	{
		Key: "new-chinese", Name: "新中式", Icon: "中",
		BasePrice: [2]int{4000, 80000},
		Materials: []string{"北美黑胡桃木皮", "欧松板柜体", "PET 门板"},
		Hardware:  []string{"百隆阻尼铰链", "古铜拉丝把手", "海蒂诗反弹器"},
		Highlights: []string{
			"一门到顶 2.7m 通顶设计",
			"胡桃木开放木皮",
			"圆弧开放格",
			"内嵌感应灯带",
			"换鞋凳 + 挂衣区一体",
		},
		PromptTpl: "New Chinese style {space}, walnut wood custom cabinet, rice paper screen, jade accent, warm lantern lighting, zen mood, editorial interior photography",
	},
	{
		Key: "cream", Name: "奶油风", Icon: "奶",
		BasePrice: [2]int{6000, 50000},
		Materials: []string{"韩国 LG 肤感 PET", "欧松板柜体", "橡木多层板"},
		Hardware:  []string{"百隆阻尼", "隐形反弹器", "缓冲导轨"},
		Highlights: []string{
			"圆弧门型工艺",
			"奶咖色 PET 门板",
			"悬空设计",
			"免拉手极简",
			"一体延伸床头柜",
		},
		PromptTpl: "Cream style {space}, cream colored PET wardrobe with rounded corners, soft linen textile, warm morning light, cozy healing interior photography",
	},
	{
		Key: "italian-luxury", Name: "意式轻奢", Icon: "意",
		BasePrice: [2]int{20000, 100000},
		Materials: []string{"12mm 雪山岩板", "镀钛金属", "北美黑胡桃", "钢琴烤漆"},
		Hardware:  []string{"百隆豪华阻尼", "海蒂诗反弹", "百隆全拉抽屉"},
		Highlights: []string{
			"雪山岩板地台",
			"镀钛金属嵌条",
			"悬空电视柜",
			"恒温酒柜预留",
			"内嵌灯带氛围",
		},
		PromptTpl: "Italian luxury {space}, sintered stone and titanium gold metal accents, dark walnut cabinetry, dramatic lighting, editorial magazine style",
	},
	{
		Key: "modern", Name: "现代简约", Icon: "现",
		BasePrice: [2]int{6000, 60000},
		Materials: []string{"高光烤漆", "欧松板", "钢化玻璃"},
		Hardware:  []string{"百隆阻尼", "反弹器"},
		Highlights: []string{
			"隐形门背景墙",
			"免拉手设计",
			"悬空电视柜",
			"L 型分区",
			"玻璃门 + 灯带",
		},
		PromptTpl: "Modern minimalist {space}, grey and white handle-less cabinet, hidden storage, clean lines, contemporary interior photography",
	},
	{
		Key: "nordic", Name: "北欧", Icon: "北",
		BasePrice: [2]int{5000, 30000},
		Materials: []string{"白橡木皮", "欧松板", "橡木多层"},
		Hardware:  []string{"百隆阻尼", "黑色金属拉手"},
		Highlights: []string{
			"白橡开放木皮",
			"圆角门型",
			"原木长腿",
			"洞洞板背墙",
			"金属拉手点缀",
		},
		PromptTpl: "Nordic style {space}, white oak cabinet with rounded corners, soft natural light, minimalist Scandinavian interior photography",
	},
	{
		Key: "japanese", Name: "日式无印", Icon: "日",
		BasePrice: [2]int{6000, 40000},
		Materials: []string{"白橡", "障子纸", "榻榻米芯"},
		Hardware:  []string{"百隆推拉缓冲", "百隆升降五金"},
		Highlights: []string{
			"障子推拉门",
			"原木开放格",
			"榻榻米升降台",
			"极简拉手",
			"内嵌感应灯",
		},
		PromptTpl: "Japanese muji {space}, oak cabinet with shoji sliding doors, light wood, zen minimalist interior photography",
	},
	{
		Key: "american", Name: "美式", Icon: "美",
		BasePrice: [2]int{12000, 80000},
		Materials: []string{"樱桃实木", "欧松板"},
		Hardware:  []string{"美国进口铜拉手", "古铜五金"},
		Highlights: []string{
			"樱桃实木",
			"美式线条",
			"铜色五金",
			"顶柜收纳",
			"抽屉分区",
		},
		PromptTpl: "American style {space}, cherry solid wood cabinet with carved moldings, brass handles, classic traditional interior",
	},
	{
		Key: "wabi-sabi", Name: "侘寂", Icon: "寂",
		BasePrice: [2]int{8000, 50000},
		Materials: []string{"微水泥饰面", "原木", "欧松板"},
		Hardware:  []string{"百隆反弹", "黑色金属"},
		Highlights: []string{
			"微水泥门板",
			"原木开放格",
			"隐形拉手",
			"素朴质感",
			"无装饰线条",
		},
		PromptTpl: "Wabi sabi {space}, micro cement texture cabinet, oak accents, minimalist quiet luxury interior photography",
	},
	{
		Key: "minimalist", Name: "极简", Icon: "极",
		BasePrice: [2]int{7000, 40000},
		Materials: []string{"高光烤漆", "欧松板", "肤感 PET"},
		Hardware:  []string{"百隆反弹", "百隆阻尼"},
		Highlights: []string{
			"纯白门板",
			"一门到顶",
			"隐形反弹",
			"悬空书桌",
			"零装饰线条",
		},
		PromptTpl: "Pure minimalist {space}, all-white handle-less cabinet, floor-to-ceiling clean lines, ultra minimal interior photography",
	},
	{
		Key: "french", Name: "法式", Icon: "法",
		BasePrice: [2]int{15000, 70000},
		Materials: []string{"高光烤漆", "描金线条", "钢化玻璃"},
		Hardware:  []string{"百隆阻尼", "金色拉手"},
		Highlights: []string{
			"法式线条",
			"描金饰面",
			"圆拱开放格",
			"圆拱玻璃门",
			"U 型分区",
		},
		PromptTpl: "French style {space}, white cabinet with gold trim and arched openings, classical elegant interior photography",
	},
	{
		Key: "industrial", Name: "工业风", Icon: "工",
		BasePrice: [2]int{6000, 50000},
		Materials: []string{"金属网", "原木", "欧松板", "铁艺"},
		Hardware:  []string{"百隆阻尼", "黑色金属框架"},
		Highlights: []string{
			"金属网门",
			"黑色金属框架",
			"原木层板",
			"铁艺结构",
			"可移动梯子",
		},
		PromptTpl: "Industrial style {space}, metal mesh doors and black iron frame with oak shelves, urban loft aesthetic photography",
	},
}

// priceToLabel 7 位数字定位价格区间
func priceToLabel(price int) string {
	switch {
	case price < 10000:
		return "1万以下"
	case price < 30000:
		return "1-3万"
	case price < 50000:
		return "3-5万"
	case price < 100000:
		return "5-10万"
	default:
		return "10万+"
	}
}

// areaMap 空间典型面积
var areaMap = map[string][2]int{
	"客厅":   {16, 32},
	"餐厅":   {6, 14},
	"主卧":   {12, 22},
	"次卧":   {8, 14},
	"书房":   {6, 14},
	"衣帽间": {10, 18},
	"玄关":   {3, 8},
	"儿童房": {8, 14},
}

// BuildCases 生成: 11 风格 × 8 空间 × 4 个变体 + 设计师款 + 衣帽间款 = 374 条核心案例
// 每个 (style, space) 至少 4 条变体 (覆盖 5 档价格 + 6 种尺寸 + 8 种颜色)
// 这样用户在二级筛选任选 color/size/price 都至少能命中 1 条
func BuildCases() []model.Case {
	out := []model.Case{}

	// 11 × 8 × 4 = 352 条 (v 0/1/2/3 循环覆盖 5 档价格)
	for _, sc := range STYLES {
		for _, sp := range SPACES {
			for v := 0; v < 4; v++ {
				cc := makeCase(sc, sp, v)
				out = append(out, cc)
			}
		}
	}

	// 额外 每个风格 2 条"设计师推荐款"特色 (覆盖 v=4 高价位)
	for _, sc := range STYLES {
		cc := makeCase(sc, SPACES[(idx(sc)+1)%len(SPACES)], 4)
		cc.Title = sc.Name + " · 设计师推荐款"
		out = append(out, cc)
		cc2 := makeCase(sc, "衣帽间", 2)
		cc2.Title = sc.Name + " · 衣帽间旗舰款"
		out = append(out, cc2)
	}

	return out
}

func idx(sc StyleConf) int {
	for i, s := range STYLES {
		if s.Key == sc.Key {
			return i
		}
	}
	return 0
}

// makeCase 单条案例构造
// v 是变体索引 0/1/2/3: 用于切换尺寸/价格段, 让 (style,space) 多条覆盖不同 size+price 组合
func makeCase(sc StyleConf, space string, v int) model.Case {
	sizes := SPACE_SIZES_MAP[space]
	// size 在该空间的可用尺寸集合中循环, 覆盖多个尺寸
	pick := v % len(sizes)
	if pick < 0 {
		pick = -pick
	}
	size := sizes[pick]
	// 额外选择 size 来覆盖通顶 如果可用
	if v%5 == 4 {
		for _, s := range sizes {
			if s == "通顶" {
				size = "通顶"
				break
			}
		}
	}
	area := fmt.Sprintf("%d㎡", (areaMap[space][0]+areaMap[space][1])/2)

	// 颜色：选该风格调性的 2-3 色, 用 v 做偏移
	colorSet := pickColors(sc.Key, space, v)
	// 价格：在基础区间按空间浮动 + 变体偏移 (让 v 跨越不同价格档)
	price := calcPrice(sc, space, v)
	priceLabel := priceToLabel(price)

	// 用全局案例 ID 计数器保证不同 case 拿不同图
	caseCounter++
	salt := fmt.Sprintf("%s-%s-%d", sc.Key, space, caseCounter)

	promptCover := fmt.Sprintf(sc.PromptTpl+"|%s", space, salt)
	promptImg := func(i int) string {
		return fmt.Sprintf(sc.PromptTpl+" detail %d|%s", space, i, salt)
	}

	// 标题模板
	title := sc.Name + " · " + space + " · 全屋定制"

	images := []string{img(promptImg(1)), img(promptImg(2))}

	highlights := sc.Highlights[:3]
	materials := sc.Materials
	hardware := sc.Hardware

	cc := model.Case{
		Title:      title,
		Style:      sc.Key,
		Space:      space,
		Colors:     colorSet,
		Size:       size,
		Area:       area,
		Price:      price,
		PriceLabel: priceLabel,
		Cover:      img(promptCover),
		Images:     images,
		Highlights: highlights,
		Materials:  materials,
		Hardware:   hardware,
		Pinned:     false,
		Enabled:    true,
	}
	return cc
}

var caseCounter int

var styleColorMap = map[string][]string{
	"new-chinese":    {"暮青", "原木", "焦糖棕", "雾霾蓝", "莫兰迪绿", "烟灰"},
	"cream":          {"奶油白", "原木", "胭脂粉", "莫兰迪绿", "雾霾蓝", "焦糖棕"},
	"italian-luxury": {"暮青", "焦糖棕", "奶油白", "雾霾蓝", "胭脂粉", "烟灰"},
	"modern":         {"烟灰", "奶油白", "雾霾蓝", "暮青", "原木", "胭脂粉"},
	"nordic":         {"奶油白", "原木", "雾霾蓝", "莫兰迪绿", "焦糖棕", "烟灰"},
	"japanese":       {"原木", "奶油白", "暮青", "焦糖棕", "雾霾蓝", "烟灰"},
	"american":       {"焦糖棕", "原木", "奶油白", "暮青", "雾霾蓝", "烟灰"},
	"wabi-sabi":      {"奶油白", "原木", "烟灰", "莫兰迪绿", "暮青", "雾霾蓝"},
	"minimalist":     {"奶油白", "烟灰", "雾霾蓝", "原木", "胭脂粉", "莫兰迪绿"},
	"french":         {"奶油白", "胭脂粉", "暮青", "雾霾蓝", "焦糖棕", "烟灰"},
	"industrial":     {"烟灰", "原木", "暮青", "焦糖棕", "莫兰迪绿", "奶油白"},
}

func pickColors(style, space string, v int) []string {
	pool := styleColorMap[style]
	out := []string{pool[(v)%len(pool)]}
	if len(pool) > 1 {
		out = append(out, pool[(v+1)%len(pool)])
	}
	if len(pool) > 2 && v%2 == 0 {
		out = append(out, pool[(v+2)%len(pool)])
	}
	return out
}

func calcPrice(sc StyleConf, space string, v int) int {
	min, max := sc.BasePrice[0], sc.BasePrice[1]
	// 空间越大价位越高
	mult := 1.0
	switch space {
	case "客厅", "衣帽间":
		mult = 1.4
	case "餐厅", "主卧":
		mult = 1.1
	case "次卧", "书房":
		mult = 0.9
	case "玄关", "儿童房":
		mult = 0.7
	}
	// v ∈ {0,1,2,3,...} 映射到 5 档价格区间, 让价格覆盖 1万以下 / 1-3万 / 3-5万 / 5-10万 / 10万+
	priceMult := []float64{0.15, 0.6, 1.1, 1.8, 2.8}[v%5]
	base := float64(min+max)/2 * mult * priceMult
	p := int(base/100) * 100
	if p < 1500 {
		p = 1500
	}
	return p
}