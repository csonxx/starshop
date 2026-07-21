// Package handler 公开接口.
package handler

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"star/server/internal/middleware"
	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// PublicHandler 暴露 Banner/Tag/Case 公开接口.
type PublicHandler struct {
	banners *repo.BannerRepo
	tags    *repo.TagRepo
	cases   *repo.CaseRepo
}

// NewPublicHandler 构造.
func NewPublicHandler(b *repo.BannerRepo, t *repo.TagRepo, c *repo.CaseRepo) *PublicHandler {
	return &PublicHandler{banners: b, tags: t, cases: c}
}

// Register 在路由组中绑定全部公开端点.
func (h *PublicHandler) Register(g *gin.RouterGroup) {
	g.GET("/banners", h.Banners)
	g.GET("/tags", h.Tags)
	g.GET("/cases", h.Cases)
	g.GET("/cases/pinned", h.Pinned)
	g.GET("/cases/:id", h.CaseDetail)
}

// currentRole 从 context 读取当前角色; 没有 claims 时默认匿名 (按普通用户处理).
func currentRole(c *gin.Context) string {
	if v, ok := c.Get("claims"); ok {
		if claims, ok2 := v.(*middleware.Claims); ok2 && claims != nil {
			return claims.Role
		}
	}
	return model.RoleUser
}

func sanitizeCase(c model.Case, role string) model.Case {
	if !model.CanSeePrice(role) {
		c.Price = 0
	}
	return c
}

// sanitizeCases 批量脱敏价格, 返回新切片.
func sanitizeCases(list []model.Case, role string) []model.Case {
	if model.CanSeePrice(role) {
		return list
	}
	out := make([]model.Case, len(list))
	for i := range list {
		out[i] = list[i]
		out[i].Price = 0
	}
	return out
}

// Banners 公开返回启用的 Banner, 数量截断到最多 5.
func (h *PublicHandler) Banners(c *gin.Context) {
	cur, err := h.banners.FindEnabled(c.Request.Context(), 5)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	if cur == nil {
		cur = []model.Banner{}
	}
	response.OK(c, cur)
}

// Tags 返回指定 type 的标签.
func (h *PublicHandler) Tags(c *gin.Context) {
	typ := strings.ToLower(strings.TrimSpace(c.Query("type")))
	if !validTagType(typ) {
		response.BadRequest(c, "type invalid")
		return
	}
	list, err := h.tags.ListEnabled(c.Request.Context(), typ)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, list)
}

// Cases 公开案例列表. 多选参数通过逗号分隔.
func (h *PublicHandler) Cases(c *gin.Context) {
	role := currentRole(c)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		response.BadRequest(c, "page invalid")
		return
	}
	size, err := strconv.Atoi(c.DefaultQuery("pageSize", "24"))
	if err != nil || size < 1 || size > 60 {
		response.BadRequest(c, "pageSize invalid")
		return
	}
	style := strings.TrimSpace(c.Query("style"))
	q := strings.TrimSpace(c.Query("q"))
	if len([]rune(style)) > tagNameMax || len([]rune(q)) > caseTitleMax {
		response.BadRequest(c, "filter too long")
		return
	}

	filter := repo.CaseFilter{
		Style:      style,
		Space:      repo.SplitLists(c.QueryArray("space")...),
		Color:      repo.SplitLists(c.QueryArray("color")...),
		Size:       repo.SplitLists(c.QueryArray("size")...),
		Price:      repo.SplitLists(c.QueryArray("price")...),
		Q:          q,
		Page:       page,
		Size2:      size,
		OnlyActive: true,
	}

	items, total, err := h.cases.List(c.Request.Context(), filter)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	log.Printf("[public] cases role=%s style=%s space=%v color=%v size=%v price=%v -> total=%d page=%d",
		role, filter.Style, filter.Space, filter.Color, filter.Size, filter.Price, total, filter.Page)

	out := sanitizeCases(items, role)

	normPage, normSize := normalizePage(filter.Page, filter.Size2)
	response.OK(c, gin.H{
		"list":     out,
		"total":    total,
		"page":     normPage,
		"pageSize": normSize,
	})
}

// Pinned 首页置顶.
func (h *PublicHandler) Pinned(c *gin.Context) {
	role := currentRole(c)
	items, err := h.cases.Pinned(c.Request.Context())
	if err != nil {
		response.ServerError(c, err)
		return
	}
	if items == nil {
		items = []model.Case{}
	}
	response.OK(c, sanitizeCases(items, role))
}

// CaseDetail 公开详情, 仅返回已启用案例.
func (h *PublicHandler) CaseDetail(c *gin.Context) {
	role := currentRole(c)
	oid, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	cc, err := h.cases.Get(c.Request.Context(), oid)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	response.OK(c, sanitizeCase(cc, role))
}

// normalizePage 返回客户端使用的规范 page/size.
func normalizePage(page, size int) (int, int) {
	if page < 1 {
		page = 1
	}
	if size <= 0 || size > 60 {
		size = 24
	}
	return page, size
}
