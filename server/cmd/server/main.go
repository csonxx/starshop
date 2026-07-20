// star-server 入口 - Go 后端 HTTP 服务
//
// 启动顺序:
//   1. Load config (env STAR_*)
//   2. Connect MongoDB
//   3. Ensure admin (按白名单手机号)
//   4. Build gin engine + router
//   5. 阻塞监听, 收到 SIGINT/SIGTERM 优雅退出
//
// 健康检查: GET /healthz -> {"status":"ok","ts":<unix>}
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"star/server/internal/config"
	"star/server/internal/db"
	"star/server/internal/handler"
	"star/server/internal/middleware"
	"star/server/internal/repo"
)

func main() {
	// 1) 配置
	cfg := config.Load()
	gin.SetMode(gin.ReleaseMode)
	log.Printf("[star-server] start  port=%s  mongo=%s  db=%s  cors=%s",
		cfg.HTTPPort, cfg.MongoURI, cfg.DBName, cfg.CORSOrigins)

	// 2) MongoDB 连接 (10s 超时)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	store, err := db.Connect(ctx, cfg.MongoURI, cfg.DBName)
	if err != nil {
		log.Fatalf("[star-server] MongoDB 连接失败: %v", err)
	}
	defer func() {
		closeCtx, c := context.WithTimeout(context.Background(), 3*time.Second)
		defer c()
		_ = store.Close(closeCtx)
	}()
	log.Printf("[star-server] MongoDB OK")

	// 3) 确保管理员账户存在
	users := repo.NewUserRepo(store.Coll("users"))
	if err := users.EnsureAdmin(ctx, cfg.AdminPhone); err != nil {
		log.Printf("[star-server] warn 初始化管理员失败 (可忽略): %v", err)
	} else {
		log.Printf("[star-server] admin phone=%s 已就绪", cfg.AdminPhone)
	}

	// 4) gin 引擎
	engine := gin.New()
	engine.Use(
		gin.Recovery(),
		middleware.CORS(cfg.CORSOrigins),
		middleware.RequestLog(),
	)

	// 5) 公开 healthz (Docker healthcheck 用)
	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "ts": time.Now().Unix()})
	})

	// 6) 路由
	router := handler.NewRouter(cfg, store)
	router.Build(engine)

	// 7) 启动 HTTP (异步 goroutine, 主线程负责信号监听)
	go func() {
		log.Printf("[star-server] HTTP listening on :%s", cfg.HTTPPort)
		if err := engine.Run(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("[star-server] HTTP error: %v", err)
		}
	}()

	// 8) 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[star-server] shutting down (收到终止信号)")
}