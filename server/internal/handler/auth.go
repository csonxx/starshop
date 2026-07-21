// Package handler auth 相关接口.
package handler

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"star/server/internal/config"
	"star/server/internal/middleware"
	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// AuthHandler 提供登录/验证码/me.
type AuthHandler struct {
	cfg   *config.Config
	users *repo.UserRepo
}

// NewAuthHandler 构造.
func NewAuthHandler(cfg *config.Config, users *repo.UserRepo) *AuthHandler {
	return &AuthHandler{cfg: cfg, users: users}
}

// Register 在路由组中绑定端点.
func (h *AuthHandler) Register(g *gin.RouterGroup) {
	g.POST("/auth/send-code", h.SendCode)
	g.POST("/auth/login", h.Login)
	g.GET("/me", middleware.Auth(h.cfg.JWTSecret, h.users), h.Me)
	g.POST("/auth/logout", middleware.Auth(h.cfg.JWTSecret, h.users), h.Logout)
}

// SendCode 发码. 仅当 StaticCode 配置时, 用静态码 (开发环境).
// 其它情况, 调用外部短信通道 (此处仅演示, 不会回传明文验证码).
func (h *AuthHandler) SendCode(c *gin.Context) {
	var in struct {
		Phone string `json:"phone"`
	}
	_ = c.ShouldBindJSON(&in)

	phone := strings.TrimSpace(in.Phone)
	if !validPhone(phone) {
		response.BadRequest(c, "invalid phone")
		return
	}

	resp := gin.H{
		"phone":     phone,
		"expiresIn": 300,
	}
	if h.cfg.StaticCode != "" {
		// 生产环境配置 STAR_STATIC_CODE 即视为关闭真实短信; 返回脱敏提示.
		if env("STAR_ENV") == "prod" {
			response.BadRequest(c, "sms service unavailable")
			return
		}
		// 开发期: 把静态码写入响应方便手测, 但提示仅本地开发可见.
		resp["devCode"] = maskCode(h.cfg.StaticCode)
		resp["notice"] = "开发模式: 验证码已脱敏, 后端不返回明文"
	}
	response.OK(c, resp)
}

// Login 登录.
func (h *AuthHandler) Login(c *gin.Context) {
	var in struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	phone := strings.TrimSpace(in.Phone)
	code := strings.TrimSpace(in.Code)
	if !validPhone(phone) || code == "" {
		response.BadRequest(c, "phone and code required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// 根据白名单决定角色.
	role := h.cfg.RoleForPhone(phone)

	if h.cfg.StaticCode == "" {
		response.BadRequest(c, "sms verify not configured")
		return
	}
	if code != h.cfg.StaticCode {
		response.Unauthorized(c, "code mismatch")
		return
	}

	// upsert 用户 (不要覆盖已存在用户的角色).
	user, err := h.users.UpsertOnLogin(ctx, phone, "", "")
	if err != nil {
		response.ServerError(c, err)
		return
	}
	if role == model.RoleAdmin && user.Role != model.RoleAdmin {
		if err := h.users.EnsureAdmin(ctx, phone); err != nil {
			response.ServerError(c, err)
			return
		}
		user.Role = model.RoleAdmin
	} else if user.Role != role {
		if err := h.users.SetRole(ctx, user.ID, role); err != nil {
			response.ServerError(c, err)
			return
		}
		user.Role = role
	}

	token, err := middleware.IssueToken(h.cfg.JWTSecret, user.ID.Hex(), user.Phone, user.Role)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID.Hex(),
			"phone":    maskPhone(phone),
			"role":     user.Role,
			"nickname": user.Nickname,
		},
	})
}

// Me 返回当前用户.
func (h *AuthHandler) Me(c *gin.Context) {
	value, ok := c.Get("claims")
	claims, ok := value.(*middleware.Claims)
	if !ok || claims == nil {
		response.Unauthorized(c, "no uid")
		return
	}
	oid, err := primitive.ObjectIDFromHex(claims.UID)
	if err != nil {
		response.Unauthorized(c, "invalid uid")
		return
	}
	user, err := h.users.Get(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{
		"id":       user.ID.Hex(),
		"phone":    maskPhone(user.Phone),
		"role":     user.Role,
		"nickname": user.Nickname,
	})
}

// Logout 注销. 真实系统应做 token 黑名单或版本号, 此处简化.
func (h *AuthHandler) Logout(c *gin.Context) {
	response.OK(c, gin.H{"ok": true})
}

// validPhone 简单校验: 11 位中国大陆手机号.
func validPhone(phone string) bool {
	if len(phone) != 11 || phone[0] != '1' {
		return false
	}
	for _, ch := range phone {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// maskPhone 中间四位打码.
func maskPhone(p string) string {
	if len(p) < 7 {
		return p
	}
	return p[:3] + "****" + p[7:]
}

// maskCode 验证码脱敏: 仅显示前 1 后 1 位.
func maskCode(c string) string {
	if len(c) <= 2 {
		return "**"
	}
	return string(c[0]) + "**" + string(c[len(c)-1])
}

// env 简化的本地 env 读取.
func env(k string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	if k == "STAR_ENV" {
		return "dev"
	}
	return ""
}
