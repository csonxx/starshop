// Package handler 路由总入口.
package handler

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"star/server/internal/config"
	"star/server/internal/middleware"
	"star/server/internal/repo"
)

// Routes 聚合全部 handler 后暴露给 cmd/server.
type Routes struct {
	Cfg     *config.Config
	Users   *repo.UserRepo
	Banners *repo.BannerRepo
	Tags    *repo.TagRepo
	Cases   *repo.CaseRepo
	DB      *mongo.Database
}

// Build 配置全部路由并返回 gin.Engine.
func (r *Routes) Build() *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.CORS(r.Cfg.CORSOrigins))

	engine.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })
	engine.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := r.DB.Client().Ping(ctx, nil); err != nil {
			c.JSON(503, gin.H{"ok": false, "msg": "mongo down"})
			return
		}
		c.JSON(200, gin.H{"ok": true})
	})

	api := engine.Group("/api/v1")
	api.Use(middleware.OptionalAuth(r.Cfg.JWTSecret, r.Users))

	pub := NewPublicHandler(r.Banners, r.Tags, r.Cases)
	pub.Register(api)

	auth := NewAuthHandler(r.Cfg, r.Users)
	auth.Register(api)

	admin := NewAdminHandler(r.Banners, r.Tags, r.Cases, r.Users, r.DB)
	adminGroup := api.Group("/admin",
		middleware.Auth(r.Cfg.JWTSecret, r.Users),
		middleware.Admin(),
	)
	admin.Register(adminGroup)
	return engine
}
