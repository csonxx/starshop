package scripts

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"star/server/internal/db"
	"star/server/internal/model"
	"star/server/internal/repo"
)

// SeedData 把种子数据初始化封装成一个结构, 由 cmd/seed 调用
// 流程: 清空 collection -> 写用户 -> 写 Banner -> 写标签 -> 写案例
type SeedData struct {
	Users  *repo.UserRepo
	Banner *repo.BannerRepo
	Tags   *repo.TagRepo
	Cases  *repo.CaseRepo
}

// NewSeed 用 db.Store 构造全部 repo
func NewSeed(store *db.Store) *SeedData {
	return &SeedData{
		Users:  repo.NewUserRepo(store.Coll("users")),
		Banner: repo.NewBannerRepo(store.Coll("banners")),
		Tags:   repo.NewTagRepo(store.Coll("tags")),
		Cases:  repo.NewCaseRepo(store.Coll("cases")),
	}
}

// Run 全量覆盖: 先清空 (避免重复) 再写入种子
func (s *SeedData) Run(ctx context.Context, adminPhone string) error {
	log.Printf("[seed] start clearing all collections ...")
	if err := s.Banner.Clear(ctx); err != nil {
		return err
	}
	if err := s.Tags.Clear(ctx); err != nil {
		return err
	}
	if err := s.Cases.Clear(ctx); err != nil {
		return err
	}
	// 清空本地本地的旧上传图 (避免被引用到不存在的文件)
	publicUploads := "../web/public/uploads"
	_ = removeAll(publicUploads)
	_ = mkdirAll(publicUploads)

	log.Printf("[seed] start writing real seed data")
	if err := s.seedUsers(ctx, adminPhone); err != nil {
		return err
	}
	if err := s.seedBanners(ctx); err != nil {
		return err
	}
	if err := s.seedTags(ctx); err != nil {
		return err
	}
	if err := s.seedCases(ctx); err != nil {
		return err
	}
	log.Printf("[seed] DONE. 全部种子数据写入完成")
	return nil
}

// seedUsers upsert 管理员账号 (白名单手机号), 确保管理员存在
func (s *SeedData) seedUsers(ctx context.Context, adminPhone string) error {
	now := time.Now()
	_, err := s.Users.UpsertByPhone(ctx, adminPhone)
	if err != nil {
		return err
	}
	_, err = s.Users.Coll().UpdateOne(ctx, bson.M{"phone": adminPhone}, bson.M{
		"$set": bson.M{"role": model.RoleAdmin, "nickname": "星仔管理员", "updatedAt": now},
	})
	if err == nil {
		log.Printf("[seed] admin user upserted: %s", adminPhone)
	}
	return err
}

// seedBanners 4 张 Banner (工厂直营 + 三种主流风格)
func (s *SeedData) seedBanners(ctx context.Context) error {
	banners := []model.Banner{
		{
			Title:    "工厂直营 · 全屋定制",
			Subtitle: "自有 12,000㎡ 智造工厂 · 从板材到入住 28 天",
			Image:    imgW("Luxury whole-house custom furniture factory showroom, warm champagne gold interior, floor-to-ceiling wardrobe and display cabinet, soft ambient lighting, marble floor, editorial photography"),
			Link:     "/cases",
			Sort:     1,
			Enabled:  true,
		},
		{
			Title:    "新中式 · 静谧东方",
			Subtitle: "胡桃木 + 隐框门 · 一门到顶 2.7m 通顶设计",
			Image:    imgW("New Chinese style living room with walnut wood floor-to-ceiling cabinet, rice paper screen, jade green accent wall, warm lantern lighting, elegant zen mood"),
			Link:     "/style/new-chinese",
			Sort:     2,
			Enabled:  true,
		},
		{
			Title:    "奶油风 · 治愈系暖居",
			Subtitle: "奶咖色 PET 门板 · 圆弧工艺 · 治愈每一天",
			Image:    imgW("Cream style bedroom with cream colored PET wardrobe, rounded corners, soft linen textile, warm morning light, cozy healing interior photography"),
			Link:     "/style/cream",
			Sort:     3,
			Enabled:  true,
		},
		{
			Title:    "意式轻奢 · 高阶质感",
			Subtitle: "岩板 + 镀钛金属 · 入户到卧室一气呵成",
			Image:    imgW("Italian light luxury living room, sintered stone TV wall, titanium gold metal accents, dark walnut wood cabinetry, premium leather sofa, dramatic lighting, editorial magazine style"),
			Link:     "/style/italian-luxury",
			Sort:     4,
			Enabled:  true,
		},
	}
	for i := range banners {
		if err := s.Banner.Insert(ctx, &banners[i]); err != nil {
			return err
		}
	}
	log.Printf("[seed] 写入 %d 张 Banner", len(banners))
	return nil
}

// seedTags 写入 5 类标签: 风格 / 空间 / 颜色 / 尺寸 / 价格
func (s *SeedData) seedTags(ctx context.Context) error {
	// 一级: 11 种主流风格
	styles := []model.Tag{
		{Name: "新中式", Value: "new-chinese", Icon: "中", Sort: 1},
		{Name: "奶油风", Value: "cream", Icon: "奶", Sort: 2},
		{Name: "意式轻奢", Value: "italian-luxury", Icon: "意", Sort: 3},
		{Name: "现代简约", Value: "modern", Icon: "现", Sort: 4},
		{Name: "北欧", Value: "nordic", Icon: "北", Sort: 5},
		{Name: "日式无印", Value: "japanese", Icon: "日", Sort: 6},
		{Name: "美式", Value: "american", Icon: "美", Sort: 7},
		{Name: "侘寂", Value: "wabi-sabi", Icon: "寂", Sort: 8},
		{Name: "极简", Value: "minimalist", Icon: "极", Sort: 9},
		{Name: "法式", Value: "french", Icon: "法", Sort: 10},
		{Name: "工业风", Value: "industrial", Icon: "工", Sort: 11},
	}
	for i := range styles {
		styles[i].Type = model.TagStyle
		styles[i].Enabled = true
		if err := s.Tags.Insert(ctx, &styles[i]); err != nil {
			return err
		}
	}

	// 二级: 空间 (9)
	spaces := []model.Tag{
		{Name: "客厅", Value: "客厅", Sort: 1},
		{Name: "餐厅", Value: "餐厅", Sort: 2},
		{Name: "主卧", Value: "主卧", Sort: 3},
		{Name: "次卧", Value: "次卧", Sort: 4},
		{Name: "书房", Value: "书房", Sort: 5},
		{Name: "衣帽间", Value: "衣帽间", Sort: 6},
		{Name: "玄关", Value: "玄关", Sort: 7},
		{Name: "儿童房", Value: "儿童房", Sort: 8},
		{Name: "多功能房", Value: "多功能房", Sort: 9},
	}
	for i := range spaces {
		spaces[i].Type = model.TagSpace
		spaces[i].Enabled = true
		if err := s.Tags.Insert(ctx, &spaces[i]); err != nil {
			return err
		}
	}

	// 二级: 颜色 (8 色卡 + 真实色值)
	colors := []model.Tag{
		{Name: "雾霾蓝", Value: "雾霾蓝", Color: "#7A8FA6", Sort: 1},
		{Name: "莫兰迪绿", Value: "莫兰迪绿", Color: "#8DA38F", Sort: 2},
		{Name: "奶油白", Value: "奶油白", Color: "#F5EFE3", Sort: 3},
		{Name: "焦糖棕", Value: "焦糖棕", Color: "#A56B3F", Sort: 4},
		{Name: "烟灰", Value: "烟灰", Color: "#9AA0A6", Sort: 5},
		{Name: "暮青", Value: "暮青", Color: "#1F3A3D", Sort: 6},
		{Name: "原木", Value: "原木", Color: "#C8A478", Sort: 7},
		{Name: "胭脂粉", Value: "胭脂粉", Color: "#D89A9E", Sort: 8},
	}
	for i := range colors {
		colors[i].Type = model.TagColor
		colors[i].Enabled = true
		if err := s.Tags.Insert(ctx, &colors[i]); err != nil {
			return err
		}
	}

	// 二级: 尺寸 (6 种)
	sizes := []model.Tag{
		{Name: "1.2m", Value: "1.2m", Sort: 1},
		{Name: "1.5m", Value: "1.5m", Sort: 2},
		{Name: "1.8m", Value: "1.8m", Sort: 3},
		{Name: "2.0m", Value: "2.0m", Sort: 4},
		{Name: "2.4m", Value: "2.4m", Sort: 5},
		{Name: "通顶", Value: "通顶", Sort: 6},
	}
	for i := range sizes {
		sizes[i].Type = model.TagSize
		sizes[i].Enabled = true
		if err := s.Tags.Insert(ctx, &sizes[i]); err != nil {
			return err
		}
	}

	// 二级: 价格区间 (5 档)
	prices := []model.Tag{
		{Name: "1万以下", Value: "1万以下", Sort: 1},
		{Name: "1-3万", Value: "1-3万", Sort: 2},
		{Name: "3-5万", Value: "3-5万", Sort: 3},
		{Name: "5-10万", Value: "5-10万", Sort: 4},
		{Name: "10万+", Value: "10万+", Sort: 5},
	}
	for i := range prices {
		prices[i].Type = model.TagPrice
		prices[i].Enabled = true
		if err := s.Tags.Insert(ctx, &prices[i]); err != nil {
			return err
		}
	}
	log.Printf("[seed] 写入风格 %d / 空间 %d / 颜色 %d / 尺寸 %d / 价格 %d",
		len(styles), len(spaces), len(colors), len(sizes), len(prices))
	return nil
}

// seedCases 程序化生成大量真实案例 (每个 风格×空间 组合多个变体)
func (s *SeedData) seedCases(ctx context.Context) error {
	cases := BuildCases()
	for i := range cases {
		if err := s.Cases.Insert(ctx, &cases[i]); err != nil {
			return err
		}
	}
	log.Printf("[seed] 写入案例 %d 条", len(cases))
	return nil
}