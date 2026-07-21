// Package scripts 提供种子数据, 仅用于开发/演示场景.
package scripts

import (
	"context"
	"fmt"
	"log"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"

	"star/server/internal/model"
	"star/server/internal/repo"
)

// SeedData 把种子数据初始化封装成一个结构, 由 cmd/seed 调用.
type SeedData struct {
	DB     *mongo.Database
	Users  *repo.UserRepo
	Banner *repo.BannerRepo
	Tags   *repo.TagRepo
	Cases  *repo.CaseRepo
}

// NewSeed 基于 Store 构造.
func NewSeed(store *mongo.Database) *SeedData {
	return &SeedData{
		DB:     store,
		Users:  repo.NewUserRepo(store),
		Banner: repo.NewBannerRepo(store),
		Tags:   repo.NewTagRepo(store),
		Cases:  repo.NewCaseRepo(store),
	}
}

// Run 全量覆盖: 先清空再写入种子. 已在 cmd/seed 层做环境保护.
func (s *SeedData) Run(ctx context.Context, adminPhone string) error {
	log.Printf("[seed] start clearing all collections ...")
	for _, fn := range []func(context.Context) error{
		s.Banner.Clear, s.Tags.Clear, s.Cases.Clear,
	} {
		if err := fn(ctx); err != nil {
			return fmt.Errorf("clear: %w", err)
		}
	}
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
	log.Printf("[seed] DONE")
	return nil
}

// seedUsers upsert 管理员账号, 其他角色不通过 seed 引入.
func (s *SeedData) seedUsers(ctx context.Context, adminPhone string) error {
	if strings.TrimSpace(adminPhone) == "" {
		return nil
	}
	if _, err := s.Users.UpsertByPhone(ctx, adminPhone, "星仔管理员", ""); err != nil {
		return fmt.Errorf("upsert user: %w", err)
	}
	if err := s.Users.EnsureAdmin(ctx, adminPhone); err != nil {
		return fmt.Errorf("ensure admin: %w", err)
	}
	log.Printf("[seed] admin user upserted: %s", adminPhone)
	return nil
}

// seedBanners 4 张 Banner, 启用 4 张均包含在 5 张上限内.
func (s *SeedData) seedBanners(ctx context.Context) error {
	imgs := LoadBannerImages()
	banners := []model.Banner{
		{
			Title: "工厂直营 · 全屋定制", Image: imgs[0],
			Link: "/cases", Enabled: true, Sort: 1,
		},
		{
			Title: "新中式 · 静谧东方", Image: imgs[1],
			Link: "/style/new-chinese", Enabled: true, Sort: 2,
		},
		{
			Title: "奶油风 · 治愈系暖居", Image: imgs[2],
			Link: "/style/cream", Enabled: true, Sort: 3,
		},
		{
			Title: "意式轻奢 · 高阶质感", Image: imgs[3],
			Link: "/style/italian-luxury", Enabled: true, Sort: 4,
		},
	}
	for _, b := range banners {
		if _, err := s.Banner.Insert(ctx, b); err != nil {
			return fmt.Errorf("insert banner: %w", err)
		}
	}
	log.Printf("[seed] %d banners inserted", len(banners))
	return nil
}

// seedTags 写风格/空间/颜色/尺寸/价格. 尺寸来自 cases.go 的 SPACE_SIZES_MAP 去重.
func (s *SeedData) seedTags(ctx context.Context) error {
	tags := BuildTagSeed()
	for _, t := range tags {
		if _, err := s.Tags.Insert(ctx, t); err != nil {
			return fmt.Errorf("insert tag: %w", err)
		}
	}
	log.Printf("[seed] %d tags inserted", len(tags))
	return nil
}

// seedCases 写案例.
func (s *SeedData) seedCases(ctx context.Context) error {
	cases := BuildCases()
	for _, cc := range cases {
		if _, err := s.Cases.Insert(ctx, cc); err != nil {
			return fmt.Errorf("insert case: %w", err)
		}
	}
	log.Printf("[seed] %d cases inserted", len(cases))
	return nil
}
