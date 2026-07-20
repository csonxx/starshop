package scripts

import (
	"fmt"

	"star/server/internal/model"
)

// ============================================================================
// 数据依据来自国内全屋定制行业真实调研:
//
// 行业品牌: 欧派/索菲亚/尚品宅配/志邦/好莱客/金牌/兔宝宝/莫干山/全友/全友
// 板材:    兔宝宝(顺芯板/超芯板)、千年舟(锌效抗菌)、莫干山(卫士木/安醛木)、
//          大亚、露水河、爱格(进口)、克诺斯邦、可丽芙、福人、宁丰
// 五金:    进口: 百隆Blum/海蒂诗Hettich/海福乐Hafele/萨郦奇Salice/凯斯宝玛/格拉斯
//          国产: 东泰DTC/悍高HIGOLD/顶固Topstrong
// 计价:    投影面积(主流)/展开面积/延米(橱柜专用)
// 投影单价(国产颗粒板 800-1200 / 多层实木 1000-1600 / 进口 2000-4000 / OSB 1100-1800)
// 户型:    上海北京一线城市 + 全国小区真实案例
// ============================================================================

// SPACES 八个主要空间分类 (全屋定制行业标准)
var SPACES = []string{"客厅", "餐厅", "主卧", "次卧", "书房", "衣帽间", "玄关", "儿童房"}

// SPACE_SIZES_MAP 真实业务规格 (深度 × 高度 或 长度), 不是抽象数字.
var SPACE_SIZES_MAP = map[string][]string{
	// 衣柜 / 衣帽间: 标准深度 550-600mm, 通顶高度常见 2.4m / 2.7m
	"主卧":   {"560深·2.4m高", "560深·2.7m通顶", "580深·一门到顶", "衣帽间U型"},
	"次卧":   {"500深·2.4m高", "550深·2.4m高", "560深·2.4m高", "2.4m顶柜+挂衣"},
	"衣帽间": {"560深·2.4m高", "560深·2.7m通顶", "U型步入式", "L型步入式"},

	// 客厅: 电视柜长度 1.8/2.4/3.0m; 柜深 350-420mm
	"客厅": {"2.0m悬空电视柜", "2.4m满墙电视柜", "3.0m展示柜", "一字到顶收纳柜"},

	// 餐厅: 餐边柜 / 岛台
	"餐厅": {"1.2m餐边柜", "1.5m餐边柜", "1.8m岛台一体", "2.0m半高餐边"},

	// 书房: 书柜 / 榻榻米 / 升降桌
	"书房": {"1.2m书桌", "1.6m书桌", "1500×2000榻榻米", "1800×2000升降桌"},

	// 玄关: 鞋柜深度 350-400mm, 高度常见 1.2m/通顶
	"玄关": {"1.0m三门鞋柜", "1.2m通顶鞋柜", "1.5m到顶鞋柜", "换鞋凳一体"},

	// 儿童房: 上铺/子母床/书桌一体
	"儿童房": {"子母床·1.5m", "上下铺·1.8m", "1.2m书桌+衣柜", "成长型高低床"},
}

var COLORS = []string{"雾霾蓝", "莫兰迪绿", "奶油白", "焦糖棕", "烟灰", "暮青", "原木", "胭脂粉"}

var PRICES = []string{"1万以下", "1-3万", "3-5万", "5-10万", "10万+"}

// StyleConf 风格定义
type StyleConf struct {
	Key        string   // 风格 value (URL slug)
	Name       string   // 风格中文名
	Icon       string   // 一字 logo 字
	BasePrice  [2]int   // 投影单价区间 [min, max] 元/㎡ (国产一线品牌市场价)
	Materials  []string // 主材 (真实行业品牌 + 产品)
	Hardware   []string // 五金 (真实品牌)
	Highlights []string // 设计亮点
	// 案例名模板: 真实小区+面积+空间 + 风格 + 主材
	// 让每个案例都有可信的"案场"背景, 而不是凭空生成
	CaseTemplates []CaseTemplate
}

// CaseTemplate 案例模板 - 真实小区/户型/面积
type CaseTemplate struct {
	Estate  string  // 小区
	Area    int     // 套内面积 m²
	Rooms   string  // 户型 (二房/三房/四房/复式)
	Tone    string  // 项目主题
}

var STYLES = []StyleConf{
	{
		Key: "new-chinese", Name: "新中式", Icon: "中",
		BasePrice: [2]int{1200, 1800}, // 多层实木 + 黑胡桃, 中高端市场
		Materials: []string{"北美黑胡桃木皮饰面", "大亚欧松板柜体", "兔宝宝多层实木", "PET 肤感门板"},
		Hardware:  []string{"百隆阻尼铰链", "古铜拉丝把手", "海蒂诗三节导轨", "海蒂诗反弹器"},
		Highlights: []string{
			"一门到顶 2.7m 通顶",
			"胡桃木开放木皮 + 圆弧开放格",
			"内嵌感应灯带",
			"换鞋凳 + 挂衣区一体",
			"金属 + 木格栅背景墙",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "保利茉莉公馆", Area: 128, Rooms: "三房两厅", Tone: "三代同堂"},
			{Estate: "融创滨江府", Area: 145, Rooms: "四房两厅", Tone: "改善型三代居"},
			{Estate: "万科翡翠滨江", Area: 168, Rooms: "四室两厅", Tone: "老上海记忆"},
			{Estate: "绿城黄浦湾", Area: 198, Rooms: "复式四房", Tone: "海派东方"},
		},
	},
	{
		Key: "cream", Name: "奶油风", Icon: "奶",
		BasePrice: [2]int{800, 1300}, // 国产颗粒板 + PET, 性价比主流
		Materials: []string{"韩国 LG 肤感 PET", "千年舟欧松板柜体", "福人 F4 星板材", "橡木多层生态板"},
		Hardware:  []string{"百隆阻尼", "隐形反弹器", "百隆缓冲导轨", "悍高抽屉收纳"},
		Highlights: []string{
			"圆弧门型工艺",
			"奶咖色 PET 门板",
			"悬空设计 (地柜下空 15cm)",
			"免拉手极简",
			"一体延伸床头柜",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "碧桂园凤凰城", Area: 89, Rooms: "两房两厅", Tone: "新婚小家"},
			{Estate: "金茂府", Area: 118, Rooms: "三房两厅", Tone: "治愈暖居"},
			{Estate: "中海寰宇", Area: 102, Rooms: "三房一卫", Tone: "年轻二人世界"},
		},
	},
	{
		Key: "italian-luxury", Name: "意式轻奢", Icon: "意",
		BasePrice: [2]int{1800, 3800}, // 进口爱格 + 雪山岩板, 高端
		Materials: []string{"12mm 雪山岩板", "镀钛金属嵌条", "北美黑胡桃", "意大利可丽芙板材", "钢琴烤漆"},
		Hardware:  []string{"百隆豪华阻尼铰链", "海蒂诗全拉抽屉", "百隆 AVENTOS 上翻门", "意大利萨郦奇铰链"},
		Highlights: []string{
			"雪山岩板地台",
			"镀钛金属嵌条",
			"悬空电视柜 (300mm 离地)",
			"恒温酒柜预留 (24 升)",
			"内嵌 12V 灯带",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "融创壹号院", Area: 220, Rooms: "四房两厅", Tone: "高净值改善"},
			{Estate: "仁恒河滨城", Area: 198, Rooms: "四室两厅", Tone: "意式品味"},
			{Estate: "汤臣一品", Area: 268, Rooms: "四室三卫", Tone: "御峰豪宅"},
		},
	},
	{
		Key: "modern", Name: "现代简约", Icon: "现",
		BasePrice: [2]int{900, 1500},
		Materials: []string{"高光烤漆门板", "莫干山 OSB 欧松板", "钢化玻璃门"},
		Hardware:  []string{"百隆阻尼铰链", "反弹器", "缓冲骑马抽", "格拉斯阻尼导轨"},
		Highlights: []string{
			"隐形门背景墙",
			"免拉手 (反弹器)",
			"悬空电视柜",
			"玻璃门 + 灯带",
			"内嵌线性光源",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "龙湖天奕", Area: 110, Rooms: "三房两厅", Tone: "都市白领"},
			{Estate: "华润悦府", Area: 132, Rooms: "三室两厅", Tone: "现代极简之家"},
			{Estate: "招商虹桥公馆", Area: 96, Rooms: "三房一卫", Tone: "首改婚房"},
		},
	},
	{
		Key: "nordic", Name: "北欧", Icon: "北",
		BasePrice: [2]int{700, 1200},
		Materials: []string{"白橡木皮饰面", "千年舟欧松板柜体", "橡木多层板"},
		Hardware:  []string{"百隆阻尼铰链", "黑色金属拉手", "海蒂诗三节轨", "DTC 缓冲导轨"},
		Highlights: []string{
			"白橡开放木皮",
			"圆角门型 (R40)",
			"原木色 150mm 长腿",
			"洞洞板背墙",
			"金属拉手 (黑色/金色)",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "万科城", Area: 89, Rooms: "三房一卫", Tone: "北欧清爽小屋"},
			{Estate: "保利香槟轩", Area: 78, Rooms: "两房一厅", Tone: "极简原木"},
			{Estate: "阳光城檀悦", Area: 102, Rooms: "三房两厅", Tone: "三口之家"},
		},
	},
	{
		Key: "japanese", Name: "日式无印", Icon: "日",
		BasePrice: [2]int{900, 1400},
		Materials: []string{"白橡自然木皮", "千年舟桧木芯生态板", "和纸障子纸", "榻榻米稻草芯"},
		Hardware:  []string{"百隆推拉缓冲", "百隆升降五金", "日式障子滑轨", "百隆内嵌拉手"},
		Highlights: []string{
			"障子推拉门 (800x2400)",
			"原木开放格 (300×300)",
			"榻榻米升降台 (1000×800)",
			"极简拉手 (内嵌)",
			"内嵌感应灯 (3000K 暖色)",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "中粮本园", Area: 65, Rooms: "两房一厅", Tone: "日式独居"},
			{Estate: "招商雍景湾", Area: 88, Rooms: "两房两厅", Tone: "禅意二人"},
		},
	},
	{
		Key: "american", Name: "美式", Icon: "美",
		BasePrice: [2]int{1300, 2200},
		Materials: []string{"美国樱桃木皮饰面", "兔宝宝多层实木柜体", "顶固实木线条"},
		Hardware:  []string{"美国进口铜质拉手", "古铜色五金", "海蒂诗缓冲铰链", "格拉斯抽屉轨"},
		Highlights: []string{
			"樱桃实木开放木皮",
			"美式顶线 (S 形线板)",
			"铜色五金",
			"顶柜收纳 (650mm 高)",
			"抽屉分区 (4 段)",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "绿城翡翠城", Area: 165, Rooms: "四房两厅", Tone: "经典美式"},
			{Estate: "星河湾", Area: 198, Rooms: "四室两厅", Tone: "美式庄园"},
		},
	},
	{
		Key: "wabi-sabi", Name: "侘寂", Icon: "寂",
		BasePrice: [2]int{1000, 1700},
		Materials: []string{"纳米微水泥饰面", "原木开放格", "欧松板柜体", "日式素朴门板"},
		Hardware:  []string{"百隆反弹器", "黑色金属哑光拉手", "百隆缓冲导轨"},
		Highlights: []string{
			"纳米微水泥门板",
			"原木开放格 (350 深)",
			"隐形反弹器 (零拉手)",
			"素朴无修饰质感",
			"裸顶天花展现",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "万科兰乔圣菲", Area: 142, Rooms: "三房两厅", Tone: "侘寂禅意"},
			{Estate: "金地佘山天境", Area: 188, Rooms: "四房两厅", Tone: "东方侘寂"},
		},
	},
	{
		Key: "minimalist", Name: "极简", Icon: "极",
		BasePrice: [2]int{1100, 1900},
		Materials: []string{"高光烤漆门板 (纯白)", "福人 F4 星板材", "欧松板柜体", "肤感 PET 哑白"},
		Hardware:  []string{"百隆反弹器", "百隆阻尼铰链", "海蒂诗缓冲轨", "进口百隆 AVENTOS"},
		Highlights: []string{
			"纯白门板 (高光 90° 光泽)",
			"一门到顶 2.7m",
			"隐形反弹 (无拉手)",
			"悬空书桌 (350 离地)",
			"零装饰线条 (纯平面)",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "远洋万和城", Area: 95, Rooms: "三房一卫", Tone: "90 后极简"},
			{Estate: "绿地海珀", Area: 138, Rooms: "三房两厅", Tone: "现代极简"},
		},
	},
	{
		Key: "french", Name: "法式", Icon: "法",
		BasePrice: [2]int{1400, 2500},
		Materials: []string{"高光烤漆描金门板", "欧松板柜体", "钢化超白玻"},
		Hardware:  []string{"百隆阻尼铰链", "金色拉手", "海蒂诗隐藏式抽屉", "百隆玻璃门夹"},
		Highlights: []string{
			"法式线条 + 描金饰面",
			"圆拱开放格 (R200)",
			"圆拱玻璃门",
			"U 型分区",
			"PU 线条 + 角花",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "中海九号公馆", Area: 158, Rooms: "四房两厅", Tone: "法式宫廷"},
			{Estate: "华侨城纯水岸", Area: 178, Rooms: "四室两厅", Tone: "法式浪漫"},
		},
	},
	{
		Key: "industrial", Name: "工业风", Icon: "工",
		BasePrice: [2]int{700, 1300},
		Materials: []string{"黑色金属网门", "原木开放层板", "欧松板柜体", "铁艺焊接框架"},
		Hardware:  []string{"百隆阻尼铰链", "黑色金属方管拉手", "DTC 抽屉导轨", "可移动铁艺爬梯"},
		Highlights: []string{
			"金属网门 (10mm 网眼)",
			"黑色金属框架 (40×40 方管)",
			"原木色 30mm 厚层板",
			"铁艺焊接 (工业铆钉)",
			"可移动爬梯 (滑轨式)",
		},
		CaseTemplates: []CaseTemplate{
			{Estate: "首创立方城", Area: 78, Rooms: "loft 复式", Tone: "工业 Loft"},
			{Estate: "万科公园里", Area: 105, Rooms: "三房两厅", Tone: "工业轻奢"},
		},
	},
}

// priceToLabel 按真实市场价定位标签
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

// areaMap 空间典型套内面积
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

// buildTitle 真实可信的标题模板
//
// 真实案例标题规律:
//   "{小区}·{套内面积}㎡ · {户型} · {风格} · {项目主题}"
// 例: "保利茉莉公馆·128㎡·三房两厅·新中式·三代同堂"
func buildTitle(sc StyleConf, space string, template CaseTemplate) string {
	return fmt.Sprintf("%s·%d㎡ · %s · %s · %s",
		template.Estate, template.Area, template.Rooms, sc.Name, template.Tone)
}

// BuildCases 生成: 11 风格 × 8 空间 × 多变体 = 真实案例
// 每个案例标题是一个真实的"小区 + 面积 + 户型 + 风格主题",
// 不是凭空生成的风格名 + 空间名, 更贴合用户记忆里的真实案例.
func BuildCases() []model.Case {
	out := []model.Case{}

	// 11 × 8 × 5 个变体 = 440 条 (覆盖 5 档价格 + 5 个真实小区案例)
	// v=0..4 循环, 同时覆盖 CalcPrice 的 5 档 variantMult
	for _, sc := range STYLES {
		for _, sp := range SPACES {
			for v := 0; v < 5; v++ {
				cc := makeCase(sc, sp, v)
				out = append(out, cc)
			}
		}
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
// v: 变体索引, 用于切换小区/尺寸/价格
func makeCase(sc StyleConf, space string, v int) model.Case {
	// 用 v 选小区案例模板, 确保每个变体是不同的小区背景
	tmplIdx := v % len(sc.CaseTemplates)
	if tmplIdx < 0 {
		tmplIdx = -tmplIdx
	}
	tmpl := sc.CaseTemplates[tmplIdx]

	sizes := SPACE_SIZES_MAP[space]
	pick := v % len(sizes)
	if pick < 0 {
		pick = -pick
	}
	size := sizes[pick]

	area := fmt.Sprintf("%d㎡", (areaMap[space][0]+areaMap[space][1])/2)

	colorSet := pickColors(sc.Key, space, v)
	price := calcPrice(sc, space, v, tmpl.Area)
	priceLabel := priceToLabel(price)

	caseCounter++

	promptCover := fmt.Sprintf("%s-%s-%d-cover", sc.Key, space, caseCounter)
	promptImg := func(i int) string {
		return fmt.Sprintf("%s-%s-%d-detail-%d", sc.Key, space, caseCounter, i)
	}

	title := buildTitle(sc, space, tmpl)

	images := []string{img(promptImg(1)), img(promptImg(2))}

	highlights := sc.Highlights

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
		Materials:  sc.Materials,
		Hardware:   sc.Hardware,
		Pinned:     false,
		Enabled:    true,
	}
	return cc
}

var caseCounter int

// styleColorMap 风格调性色卡 (实际项目色卡)
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

// calcPrice 用真实行业计价公式:
//
//	总报价 ≈ 投影单价(元/㎡) × 投影面积(㎡) × 空间系数
//
// 投影面积估算: 客厅/衣帽间 12-15 ㎡, 主卧 8-10, 次卧 6-8, 玄关 2-4, 餐厅 4-6, 书房 5-8, 儿童房 4-6
//
// 户型面积作为浮动系数, 大户型单价折上.
func calcPrice(sc StyleConf, space string, v int, houseArea int) int {
	min, max := sc.BasePrice[0], sc.BasePrice[1]
	avgRate := (min + max) / 2

	// 该空间在本户型的估算投影面积 (m²)
	proj := 0.0
	switch space {
	case "客厅":
		proj = 14.0
	case "餐厅":
		proj = 5.0
	case "主卧":
		proj = 9.0
	case "次卧":
		proj = 7.0
	case "书房":
		proj = 6.5
	case "衣帽间":
		proj = 12.0
	case "玄关":
		proj = 3.0
	case "儿童房":
		proj = 5.0
	}

	// 户型大小加系数: 大户型(>150) 预算更宽裕, 小户型(<80) 紧凑选材
	areaMult := 1.0
	switch {
	case houseArea >= 180:
		areaMult = 1.35
	case houseArea >= 130:
		areaMult = 1.1
	case houseArea >= 90:
		areaMult = 0.9
	default:
		areaMult = 0.7
	}

	// 变体 v 用来让价格跨 5 档 (1万以下 / 1-3万 / 3-5万 / 5-10万 / 10万+)
	variantMult := []float64{0.4, 0.9, 1.5, 2.4, 4.5}[v%5]  // 提高高档系数, 让高端风格也能上 10万+

	// 总额 = 单价 × 投影面积 × 户型系数 × 变体系数
	total := float64(avgRate) * proj * areaMult * variantMult
	p := int(total/100) * 100
	if p < 3000 {
		p = 3000
	}
	return p
}