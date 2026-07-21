// star-seed 入口 - 一次性种子数据写入
//
// 用法:
//
//	go run ./cmd/seed                # 本地开发
//	docker compose --profile seed run --rm seed   # 容器一次性
//
// 流程: 清空现有 collection → 写 Banner/标签/案例/管理员
// 注意: 这是破坏性的, 会清空当前 DB 中 banners/tags/cases
package main

import (
	"context"
	"log"
	"time"

	"star/server/internal/config"
	"star/server/internal/db"
	"star/server/scripts"
)

func main() {
	cfg, err := config.LoadValidated()
	if err != nil {
		log.Fatalf("[star-seed] 配置错误: %v", err)
	}
	if !cfg.CanSeed() {
		log.Fatalf("[star-seed] 拒绝清空数据库 %q；请使用允许的开发/测试库，或显式设置 STAR_ALLOW_DESTRUCTIVE_SEED=true", cfg.DBName)
	}
	log.Printf("[star-seed] mongo=%s db=%s", cfg.MongoURI, cfg.DBName)

	// 给种子 60s 超时 (374 条案例不会到这个量级)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	store, err := db.Connect(ctx, cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("[star-seed] MongoDB 连接失败: %v", err)
	}
	defer store.Close(ctx)

	seed := scripts.NewSeed(store.DB)
	if err := seed.Run(ctx, cfg.AdminPhone); err != nil {
		log.Fatalf("[star-seed] 初始化失败: %v", err)
	}
	log.Printf("[star-seed] 全部种子数据写入完成, db=%s", cfg.DBName)
}
