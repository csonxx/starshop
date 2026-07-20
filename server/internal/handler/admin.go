package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// AdminHandler 后台管理 (admin only).
// 职责:
//  1. 数据总览 (Banner / Tag / Case 数量)
//  2. Banner / Tag / Case 的 CRUD
//  3. Case 置顶开关
type AdminHandler struct {
	banners *repo.BannerRepo
	tags    *repo.TagRepo
	cases   *repo.CaseRepo
}

func NewAdminHandler(b *repo.BannerRepo, t *repo.TagRepo, c *repo.CaseRepo) *AdminHandler {
	return &AdminHandler{banners: b, tags: t, cases: c}
}

// Overview 后台主页统计概览
func (h *AdminHandler) Overview(c *gin.Context) {
	ctx := c.Request.Context()
	banners, _ := h.banners.ListAll(ctx)
	styleTags, _ := h.tags.ListAll(ctx, model.TagStyle)
	allTags, _ := h.tags.ListAll(ctx, "")
	allCases, _ := h.cases.ListAll(ctx)
	pinned := 0
	for _, cc := range allCases {
		if cc.Pinned {
			pinned++
		}
	}
	log.Printf("[admin] overview banners=%d styles=%d tags=%d cases=%d pinned=%d",
		len(banners), len(styleTags), len(allTags), len(allCases), pinned)
	response.OK(c, gin.H{
		"bannerCount":  len(banners),
		"styleTagCount": len(styleTags),
		"tagCount":      len(allTags),
		"caseCount":     len(allCases),
		"pinnedCount":   pinned,
	})
}

// ===== Banner CRUD =====

func (h *AdminHandler) ListBanners(c *gin.Context) {
	out, err := h.banners.ListAll(c.Request.Context())
	if err != nil {
		log.Printf("[admin] list-banners ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, out)
}

type bannerReq struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	Image    string `json:"image"`
	Link     string `json:"link"`
	Sort     int    `json:"sort"`
	Enabled  bool   `json:"enabled"`
}

func (h *AdminHandler) CreateBanner(c *gin.Context) {
	var req bannerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	b := &model.Banner{
		Title:    req.Title,
		Subtitle: req.Subtitle,
		Image:    req.Image,
		Link:     req.Link,
		Sort:     req.Sort,
		Enabled:  req.Enabled,
	}
	if err := h.banners.Insert(c.Request.Context(), b); err != nil {
		log.Printf("[admin] create-banner ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] create-banner id=%s title=%q", b.ID.Hex(), b.Title)
	response.OK(c, b)
}

func (h *AdminHandler) UpdateBanner(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	var req bannerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	update := map[string]interface{}{
		"title":    req.Title,
		"subtitle": req.Subtitle,
		"image":    req.Image,
		"link":     req.Link,
		"sort":     req.Sort,
		"enabled":  req.Enabled,
	}
	if err := h.banners.Update(c.Request.Context(), id, update); err != nil {
		log.Printf("[admin] update-banner ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] update-banner id=%s", id.Hex())
	response.OK(c, gin.H{"id": id.Hex()})
}

func (h *AdminHandler) DeleteBanner(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	if err := h.banners.Delete(c.Request.Context(), id); err != nil {
		log.Printf("[admin] delete-banner ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] delete-banner id=%s", id.Hex())
	response.OK(c, gin.H{"id": id.Hex()})
}

// ===== Tag CRUD =====

func (h *AdminHandler) ListTags(c *gin.Context) {
	out, err := h.tags.ListAll(c.Request.Context(), c.Query("type"))
	if err != nil {
		log.Printf("[admin] list-tags ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, out)
}

type tagReq struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Color   string `json:"color"`
	Icon    string `json:"icon"`
	Sort    int    `json:"sort"`
	Enabled bool   `json:"enabled"`
}

func (h *AdminHandler) CreateTag(c *gin.Context) {
	var req tagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	t := &model.Tag{
		Type:    req.Type,
		Name:    req.Name,
		Value:   req.Value,
		Color:   req.Color,
		Icon:    req.Icon,
		Sort:    req.Sort,
		Enabled: req.Enabled,
	}
	if err := h.tags.Insert(c.Request.Context(), t); err != nil {
		log.Printf("[admin] create-tag ERROR type=%s name=%s: %v", t.Type, t.Name, err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] create-tag id=%s type=%s name=%s", t.ID.Hex(), t.Type, t.Name)
	response.OK(c, t)
}

func (h *AdminHandler) UpdateTag(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	var req tagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	update := map[string]interface{}{
		"type":    req.Type,
		"name":    req.Name,
		"value":   req.Value,
		"color":   req.Color,
		"icon":    req.Icon,
		"sort":    req.Sort,
		"enabled": req.Enabled,
	}
	if err := h.tags.Update(c.Request.Context(), id, update); err != nil {
		log.Printf("[admin] update-tag ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] update-tag id=%s", id.Hex())
	response.OK(c, gin.H{"id": id.Hex()})
}

func (h *AdminHandler) DeleteTag(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	if err := h.tags.Delete(c.Request.Context(), id); err != nil {
		log.Printf("[admin] delete-tag ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] delete-tag id=%s", id.Hex())
	response.OK(c, gin.H{"id": id.Hex()})
}

// ===== Case CRUD =====

func (h *AdminHandler) ListCases(c *gin.Context) {
	out, err := h.cases.ListAll(c.Request.Context())
	if err != nil {
		log.Printf("[admin] list-cases ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, out)
}

type caseReq struct {
	Title      string   `json:"title"`
	Style      string   `json:"style"`
	Space      string   `json:"space"`
	Colors     []string `json:"colors"`
	Size       string   `json:"size"`
	Area       string   `json:"area"`
	Price      int      `json:"price"`
	PriceLabel string   `json:"priceLabel"`
	Cover      string   `json:"cover"`
	Images     []string `json:"images"`
	Highlights []string `json:"highlights"`
	Materials  []string `json:"materials"`
	Hardware   []string `json:"hardware"`
	Pinned     bool     `json:"pinned"`
	Enabled    bool     `json:"enabled"`
}

func (h *AdminHandler) CreateCase(c *gin.Context) {
	var req caseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	cc := &model.Case{
		Title:      req.Title,
		Style:      req.Style,
		Space:      req.Space,
		Colors:     req.Colors,
		Size:       req.Size,
		Area:       req.Area,
		Price:      req.Price,
		PriceLabel: req.PriceLabel,
		Cover:      req.Cover,
		Images:     req.Images,
		Highlights: req.Highlights,
		Materials:  req.Materials,
		Hardware:   req.Hardware,
		Pinned:     req.Pinned,
		Enabled:    req.Enabled,
	}
	if err := h.cases.Insert(c.Request.Context(), cc); err != nil {
		log.Printf("[admin] create-case ERROR: %v", err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] create-case id=%s style=%s space=%s title=%q",
		cc.ID.Hex(), cc.Style, cc.Space, cc.Title)
	response.OK(c, cc)
}

func (h *AdminHandler) UpdateCase(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	var req caseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	update := map[string]interface{}{
		"title":      req.Title,
		"style":      req.Style,
		"space":      req.Space,
		"colors":     req.Colors,
		"size":       req.Size,
		"area":       req.Area,
		"price":      req.Price,
		"priceLabel": req.PriceLabel,
		"cover":      req.Cover,
		"images":     req.Images,
		"highlights": req.Highlights,
		"materials":  req.Materials,
		"hardware":   req.Hardware,
		"pinned":     req.Pinned,
		"enabled":    req.Enabled,
	}
	if err := h.cases.Update(c.Request.Context(), id, update); err != nil {
		log.Printf("[admin] update-case ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] update-case id=%s", id.Hex())
	response.OK(c, gin.H{"id": id.Hex()})
}

func (h *AdminHandler) DeleteCase(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	if err := h.cases.Delete(c.Request.Context(), id); err != nil {
		log.Printf("[admin] delete-case ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] delete-case id=%s", id.Hex())
	response.OK(c, gin.H{"id": id.Hex()})
}

// TogglePin 切换案例置顶状态 (前端 admin 后台置顶按钮)
func (h *AdminHandler) TogglePin(c *gin.Context) {
	id := mustObjectID(c.Param("id"))
	existing, err := h.cases.Get(c.Request.Context(), id)
	if err != nil {
		response.BadRequest(c, "case not found")
		return
	}
	if err := h.cases.Update(c.Request.Context(), id, map[string]interface{}{
		"pinned": !existing.Pinned,
	}); err != nil {
		log.Printf("[admin] toggle-pin ERROR id=%s: %v", id.Hex(), err)
		response.ServerError(c, err.Error())
		return
	}
	log.Printf("[admin] toggle-pin id=%s -> pinned=%v", id.Hex(), !existing.Pinned)
	response.OK(c, gin.H{"id": id.Hex(), "pinned": !existing.Pinned})
}

// StatsByStyle 后台统计各风格案例数
func (h *AdminHandler) StatsByStyle(c *gin.Context) {
	cs, _ := h.cases.ListAll(c.Request.Context())
	stats := map[string]int{}
	for _, x := range cs {
		stats[x.Style]++
	}
	tags, _ := h.tags.ListAll(c.Request.Context(), model.TagStyle)
	out := make([]gin.H, 0, len(tags))
	for _, t := range tags {
		out = append(out, gin.H{"name": t.Name, "value": t.Value, "count": stats[t.Value]})
	}
	response.OK(c, out)
}

// 引用占位避免 lint
var _ = primitive.NilObjectID