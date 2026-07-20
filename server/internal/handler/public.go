package handler

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"

	"star/server/internal/middleware"
	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// PublicHandler 前台公开数据接口 (Banner / 标签 / 案例)
// 案例接口根据当前用户角色脱敏价格:
//   - 普通用户 / 匿名: 仅看价格区间 (priceLabel), price 置 0
//   - 销售 / 供应商 / 管理员: 看精准数字 (price)
type PublicHandler struct {
	banners *repo.BannerRepo
	tags    *repo.TagRepo
	cases   *repo.CaseRepo
}

func NewPublicHandler(b *repo.BannerRepo, t *repo.TagRepo, c *repo.CaseRepo) *PublicHandler {
	return &PublicHandler{banners: b, tags: t, cases: c}
}

// currentRole 从 gin context 读取当前用户角色 (匿名时默认 model.RoleUser)
func currentRole(c *gin.Context) string {
	if v, ok := c.Get("claims"); ok {
		if claims, ok2 := v.(*middleware.Claims); ok2 && claims != nil {
			return claims.Role
		}
	}
	return model.RoleUser
}

// sanitizeCase 单条案例脱敏: 普通用户隐藏精准价格
func sanitizeCase(c model.Case, role string) model.Case {
	if !model.CanSeePrice(role) {
		c.Price = 0
	}
	return c
}

// sanitizeCases 批量脱敏 (in-place)
func sanitizeCases(list []model.Case, role string) []model.Case {
	if model.CanSeePrice(role) {
		return list
	}
	for i := range list {
		list[i].Price = 0
	}
	return list
}

// Banners 首页轮播 Banner
func (h *PublicHandler) Banners(c *gin.Context) {
	out, err := h.banners.ListEnabled(c.Request.Context())
	if err != nil {
		log.Printf("[public] banners ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, out)
}

// Tags 标签列表 (type=style|space|color|size|price; 空则全量)
func (h *PublicHandler) Tags(c *gin.Context) {
	typ := c.Query("type")
	out, err := h.tags.ListEnabled(c.Request.Context(), typ)
	if err != nil {
		log.Printf("[public] tags ERROR type=%s: %v", typ, err)
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, out)
}

// Cases 案例列表 (支持 style/space/color/size/price 多维筛选 + 分页)
// 二级筛选采用 OR 语义: 任一命中即返回, 保证 UI 上"选择都有数据"
// 一级风格是 AND 精确匹配
func (h *PublicHandler) Cases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "24"))
	f := repo.CaseFilter{
		Style:  c.Query("style"),
		Space:  c.Query("space"),
		Color:  c.Query("color"),
		Size:   c.Query("size"),
		Price:  c.Query("price"),
		Search: c.Query("q"),
		Page:   page,
		Size2:  size,
	}
	list, total, err := h.cases.List(c.Request.Context(), f)
	if err != nil {
		log.Printf("[public] cases ERROR f=%+v: %v", f, err)
		response.ServerError(c, err.Error())
		return
	}
	role := currentRole(c)
	list = sanitizeCases(list, role)
	log.Printf("[public] cases role=%s style=%s space=%s color=%s size=%s price=%s -> total=%d page=%d",
		role, f.Style, f.Space, f.Color, f.Size, f.Price, total, page)
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// CaseDetail 案例详情, 含大图 / 价格 (价格按角色脱敏)
func (h *PublicHandler) CaseDetail(c *gin.Context) {
	id := c.Param("id")
	caseObj, err := h.cases.Get(c.Request.Context(), mustObjectID(id))
	if err != nil {
		response.BadRequest(c, "case not found")
		return
	}
	role := currentRole(c)
	out := sanitizeCase(*caseObj, role)
	log.Printf("[public] case-detail id=%s role=%s title=%s", id, role, out.Title)
	response.OK(c, out)
}

// Pinned 首页置顶案例 (后台可置顶, 最多 8 个)
func (h *PublicHandler) Pinned(c *gin.Context) {
	list, err := h.cases.ListPinned(c.Request.Context())
	if err != nil {
		log.Printf("[public] pinned ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	role := currentRole(c)
	list = sanitizeCases(list, role)
	response.OK(c, list)
}