// cmd/server: 星仔高端定制 API 服务入口.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"star/server/internal/config"
	"star/server/internal/db"
	"star/server/internal/handler"
	"star/server/internal/repo"
)

func main() {
	cfg, err := config.LoadValidated()
	if err != nil {
		log.Fatalf("[star-server] 配置错误: %v", err)
	}
	if err := cfg.RequireProdCheck(); err != nil {
		log.Fatalf("[star-server] 配置错误: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	store, err := db.Connect(ctx, cfg.MongoURI, cfg.DBName)
	cancel()
	if err != nil {
		log.Fatalf("[star-server] MongoDB 连接失败: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = store.Close(ctx)
	}()

	// 建索引.
	ensureIndexes(store.DB)

	// 确保管理员存在.
	adminPhone := cfg.AdminPhone
	if adminPhone != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		users := repo.NewUserRepo(store.DB)
		if err := users.EnsureAdmin(ctx, adminPhone); err != nil {
			log.Printf("[star-server] 警告: EnsureAdmin 失败: %v", err)
		}
		cancel()
	}

	routes := &handler.Routes{
		Cfg:     cfg,
		Users:   repo.NewUserRepo(store.DB),
		Banners: repo.NewBannerRepo(store.DB),
		Tags:    repo.NewTagRepo(store.DB),
		Cases:   repo.NewCaseRepo(store.DB),
		DB:      store.DB,
	}
	engine := routes.Build()

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("[star-server] 监听: %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[star-server] HTTP error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[star-server] 收到终止信号, 开始优雅退出 ...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("[star-server] 关闭服务出错: %v", err)
	} else {
		log.Println("[star-server] HTTP 已停止")
	}
}

func ensureIndexes(d *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cases := repo.NewCaseRepo(d)
	banners := repo.NewBannerRepo(d)
	tags := repo.NewTagRepo(d)
	users := repo.NewUserRepo(d)
	for _, fn := range []func(context.Context) error{
		cases.EnsureIndexes,
		banners.EnsureIndexes,
		tags.EnsureIndexes,
		users.EnsureIndexes,
	} {
		if err := fn(ctx); err != nil {
			log.Printf("[star-server] EnsureIndexes 警告: %v", err)
		}
	}
}
