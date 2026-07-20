package handler

import (
	"log"
	"regexp"

	"github.com/gin-gonic/gin"

	"star/server/internal/config"
	"star/server/internal/middleware"
	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// phoneRE 中国大陆 11 位手机号正则 (开发期固定 1234 验证码)
var phoneRE = regexp.MustCompile(`^1[3-9]\d{9}$`)

// AuthHandler 鉴权控制器, 负责验证码 / 登录 / 当前用户查询
type AuthHandler struct {
	cfg   *config.Config
	users *repo.UserRepo
}

func NewAuthHandler(cfg *config.Config, users *repo.UserRepo) *AuthHandler {
	return &AuthHandler{cfg: cfg, users: users}
}

// sendCodeReq 入参
type sendCodeReq struct {
	Phone string `json:"phone"`
}

// SendCode "发送" 验证码 (开发期固定为 cfg.StaticCode, 生产环境需接入短信网关)
func (h *AuthHandler) SendCode(c *gin.Context) {
	var req sendCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if !phoneRE.MatchString(req.Phone) {
		response.BadRequest(c, "invalid phone")
		return
	}
	log.Printf("[auth] send-code phone=%s (dev fixed 1234)", req.Phone)
	response.OK(c, gin.H{
		"phone":     req.Phone,
		"code":      h.cfg.StaticCode,
		"message":   "开发期验证码固定为 " + h.cfg.StaticCode,
		"expiresIn": 300,
	})
}

type loginReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// Login 手机号 + 验证码登录.
// 角色按 cfg 中的白名单分配 (admin / sales / supplier / user)
// 验证码固定为 cfg.StaticCode (开发期 1234)
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if !phoneRE.MatchString(req.Phone) {
		response.BadRequest(c, "invalid phone")
		return
	}
	if req.Code != h.cfg.StaticCode {
		log.Printf("[auth] login FAILED wrong code phone=%s", req.Phone)
		response.BadRequest(c, "验证码错误 (开发期固定 1234)")
		return
	}

	// upsert 用户
	user, err := h.users.UpsertByPhone(c.Request.Context(), req.Phone)
	if err != nil {
		log.Printf("[auth] login ERROR upsert phone=%s err=%v", req.Phone, err)
		response.ServerError(c, err.Error())
		return
	}

	// 根据手机号白名单确定角色
	role := model.RoleUser
	switch {
	case h.cfg.IsAdminPhone(req.Phone):
		role = model.RoleAdmin
		_ = h.users.EnsureAdmin(c.Request.Context(), req.Phone)
	case h.cfg.IsSupplierPhone(req.Phone):
		role = model.RoleSupplier
	case h.cfg.IsSalesPhone(req.Phone):
		role = model.RoleSales
	}

	// 签发 JWT (7 天有效期)
	token, err := middleware.IssueToken(h.cfg.JWTSecret, user.ID.Hex(), user.Phone, role)
	if err != nil {
		log.Printf("[auth] login ERROR issue-token phone=%s err=%v", req.Phone, err)
		response.ServerError(c, err.Error())
		return
	}

	nickname := map[string]string{
		model.RoleAdmin:    "星仔管理员",
		model.RoleSales:    "星仔销售",
		model.RoleSupplier: "星仔供应商",
		model.RoleUser:     "星仔用户",
	}[role]

	log.Printf("[auth] login OK phone=%s role=%s uid=%s", req.Phone, role, user.ID.Hex())
	response.OK(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID.Hex(),
			"phone":    user.Phone,
			"role":     role,
			"nickname": nickname,
		},
	})
}

// Me 返回当前登录用户的资料
func (h *AuthHandler) Me(c *gin.Context) {
	v, _ := c.Get("claims")
	claims := v.(*middleware.Claims)
	response.OK(c, gin.H{
		"id":    claims.UserID,
		"phone": claims.Phone,
		"role":  claims.Role,
	})
}