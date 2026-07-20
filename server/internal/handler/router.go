// Package handler 把 HTTP 路由表集中在一个文件里.
// 通过 cmd/server Build() 注入 gin engine.
package handler

import (
	"github.com/gin-gonic/gin"

	"star/server/internal/config"
	"star/server/internal/db"
	"star/server/internal/middleware"
	"star/server/internal/repo"
)

// Router 持有配置 + DB, 在 Build() 里集中装配所有 Handler 和中间件
type Router struct {
	cfg   *config.Config
	store *db.Store
}

func NewRouter(cfg *config.Config, store *db.Store) *Router {
	return &Router{cfg: cfg, store: store}
}

// Build 把全部路由注册到 gin.Engine. 三个分组:
//
//   /api/v1                - 公开 + 可选鉴权
//     ├── /auth/*          - 鉴权 (发送码 + 登录)
//     ├── /banners /tags /cases - 公开接口, 可选鉴权用于价格脱敏
//     └── /me              - 强鉴权
//
//   /api/v1/admin/*        - 强鉴权 + admin 角色检查, 完整 CRUD
//
// 中间件:
//   middleware.OptionalAuth - 尝试解析 token, 失败也继续(匿名访问)
//   middleware.Auth         - 必须 token, 不通过 abort 401
//   middleware.Admin        - 必须在 authed 且 claims.Role == admin
func (r *Router) Build(engine *gin.Engine) {
	// ---------- 装配依赖 ----------
	users := repo.NewUserRepo(r.store.Coll("users"))
	bannerRepo := repo.NewBannerRepo(r.store.Coll("banners"))
	tagRepo := repo.NewTagRepo(r.store.Coll("tags"))
	caseRepo := repo.NewCaseRepo(r.store.Coll("cases"))

	auth := NewAuthHandler(r.cfg, users)
	pub := NewPublicHandler(bannerRepo, tagRepo, caseRepo)
	adm := NewAdminHandler(bannerRepo, tagRepo, caseRepo)

	// ---------- v1 API ----------
	api := engine.Group("/api/v1")

	// 鉴权
	api.POST("/auth/send-code", auth.SendCode)
	api.POST("/auth/login", auth.Login)

	// 公开接口 - 也支持可选 auth, 识别当前用户角色以脱敏价格
	optionalAuth := middleware.OptionalAuth(r.cfg.JWTSecret, users)
	api.GET("/banners", optionalAuth, pub.Banners)
	api.GET("/tags", optionalAuth, pub.Tags)
	api.GET("/cases", optionalAuth, pub.Cases)
	api.GET("/cases/pinned", optionalAuth, pub.Pinned)
	api.GET("/cases/:id", optionalAuth, pub.CaseDetail)

	// 已登录用户
	authed := api.Group("")
	authed.Use(middleware.Auth(r.cfg.JWTSecret, users))
	authed.GET("/me", auth.Me)

	// ---------- admin 后台 ----------
	admin := api.Group("/admin")
	admin.Use(middleware.Auth(r.cfg.JWTSecret, users), middleware.Admin())
	admin.GET("/overview", adm.Overview)
	admin.GET("/stats/by-style", adm.StatsByStyle)

	// Banner CRUD
	admin.GET("/banners", adm.ListBanners)
	admin.POST("/banners", adm.CreateBanner)
	admin.PUT("/banners/:id", adm.UpdateBanner)
	admin.DELETE("/banners/:id", adm.DeleteBanner)

	// Tag CRUD
	admin.GET("/tags", adm.ListTags)
	admin.POST("/tags", adm.CreateTag)
	admin.PUT("/tags/:id", adm.UpdateTag)
	admin.DELETE("/tags/:id", adm.DeleteTag)

	// Case CRUD
	admin.GET("/cases", adm.ListCases)
	admin.POST("/cases", adm.CreateCase)
	admin.PUT("/cases/:id", adm.UpdateCase)
	admin.DELETE("/cases/:id", adm.DeleteCase)
	admin.POST("/cases/:id/pin", adm.TogglePin)
}