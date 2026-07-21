// Package handler admin 后台接口.
package handler

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// 数值上限常量.
const (
	maxBannerActive   = 5
	maxPinnedCase     = 8
	caseTitleMax      = 80
	caseDescMax       = 500
	caseHighlightsMax = 12
	caseMaterialsMax  = 16
	caseHardwareMax   = 16
	caseColorsMax     = 6
	caseImagesMax     = 12
	bannerTitleMax    = 60
	tagNameMax        = 60
)

// AdminHandler 后台 CRUD 处理.
type AdminHandler struct {
	banners *repo.BannerRepo
	tags    *repo.TagRepo
	cases   *repo.CaseRepo
	users   *repo.UserRepo
	store   *mongo.Database
}

// NewAdminHandler 构造.
func NewAdminHandler(b *repo.BannerRepo, t *repo.TagRepo, c *repo.CaseRepo, u *repo.UserRepo, store *mongo.Database) *AdminHandler {
	return &AdminHandler{banners: b, tags: t, cases: c, users: u, store: store}
}

// Register 在路由组中绑定后台端点.
func (h *AdminHandler) Register(g *gin.RouterGroup) {
	g.GET("/overview", h.Overview)
	g.GET("/stats/styles", h.StatsByStyle)
	g.GET("/op-logs", h.ListOpLogs)

	g.GET("/banners", h.ListBanners)
	g.POST("/banners", h.CreateBanner)
	g.PUT("/banners/:id", h.UpdateBanner)
	g.DELETE("/banners/:id", h.DeleteBanner)

	g.GET("/tags", h.ListTags)
	g.POST("/tags", h.CreateTag)
	g.PUT("/tags/:id", h.UpdateTag)
	g.DELETE("/tags/:id", h.DeleteTag)

	g.GET("/cases", h.ListCases)
	g.POST("/cases", h.CreateCase)
	g.GET("/cases/:id", h.GetCase)
	g.PUT("/cases/:id", h.UpdateCase)
	g.DELETE("/cases/:id", h.DeleteCase)
	g.POST("/cases/:id/pin", h.TogglePin)

	g.GET("/users", h.ListUsers)
	g.PUT("/users/:id/role", h.UpdateUserRole)
}

// ---------- 概览 ----------

// Overview 后台首页统计. 仅统计启用数据.
func (h *AdminHandler) Overview(c *gin.Context) {
	ctx := c.Request.Context()
	bannerCount, err := h.banners.CountEnabled(ctx)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	styleCount, err := h.tags.CountAny(ctx, model.TagStyle)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	tagCount, err := h.tags.CountAny(ctx, "")
	if err != nil {
		response.ServerError(c, err)
		return
	}
	caseColl := h.store.Collection(model.CollCase)
	caseTotal, err := caseColl.CountDocuments(ctx, bson.M{"enabled": true})
	if err != nil {
		response.ServerError(c, err)
		return
	}
	pinned, err := h.cases.CountPinned(ctx)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, gin.H{
		"bannerCount":   bannerCount,
		"styleTagCount": styleCount,
		"tagCount":      tagCount,
		"caseCount":     caseTotal,
		"pinnedCount":   pinned,
	})
}

// StatsByStyle 风格占比.
func (h *AdminHandler) StatsByStyle(c *gin.Context) {
	ctx := c.Request.Context()
	styles, err := h.tags.ListAny(ctx, model.TagStyle)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	caseColl := h.store.Collection(model.CollCase)
	out := make([]gin.H, 0, len(styles))
	for _, s := range styles {
		n, err := caseColl.CountDocuments(ctx, bson.M{"enabled": true, "style": s.Value})
		if err != nil {
			response.ServerError(c, err)
			return
		}
		out = append(out, gin.H{
			"name":  s.Name,
			"value": s.Value,
			"count": n,
		})
	}
	response.OK(c, out)
}

// ---------- Banner ----------

func (h *AdminHandler) ListBanners(c *gin.Context) {
	list, err := h.banners.FindAny(c.Request.Context())
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *AdminHandler) CreateBanner(c *gin.Context) {
	var in model.Banner
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		response.BadRequest(c, "title is required")
		return
	}
	if len([]rune(in.Title)) > bannerTitleMax {
		response.BadRequest(c, "title too long")
		return
	}
	if !validImageURL(in.Image) {
		response.BadRequest(c, "image must be http(s) or /img-pool/...")
		return
	}
	if in.Link != "" {
		if _, err := url.ParseRequestURI(in.Link); err != nil {
			response.BadRequest(c, "link invalid")
			return
		}
	}
	enabledCount, err := h.banners.CountEnabled(c.Request.Context())
	if err != nil {
		response.ServerError(c, err)
		return
	}
	if in.Enabled && enabledCount >= maxBannerActive {
		response.Conflict(c, fmt.Sprintf("最多启用 %d 张 Banner", maxBannerActive))
		return
	}
	out, err := h.banners.Insert(c.Request.Context(), in)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	logOp(c, "banner.create", "banner", out.ID.Hex(), "ok")
	response.OK(c, out)
}

func (h *AdminHandler) UpdateBanner(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var in model.Banner
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if strings.TrimSpace(in.Title) == "" {
		response.BadRequest(c, "title is required")
		return
	}
	if len([]rune(in.Title)) > bannerTitleMax {
		response.BadRequest(c, "title too long")
		return
	}
	if !validImageURL(in.Image) {
		response.BadRequest(c, "image invalid")
		return
	}
	if in.Link != "" {
		if _, err := url.ParseRequestURI(in.Link); err != nil {
			response.BadRequest(c, "link invalid")
			return
		}
	}
	existing, err := h.banners.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "banner not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	if in.Enabled && !existing.Enabled {
		enabledCount, _ := h.banners.CountEnabled(c.Request.Context())
		if enabledCount >= maxBannerActive {
			response.Conflict(c, fmt.Sprintf("最多启用 %d 张 Banner", maxBannerActive))
			return
		}
	}
	fields := bson.M{
		"title":   strings.TrimSpace(in.Title),
		"image":   strings.TrimSpace(in.Image),
		"link":    strings.TrimSpace(in.Link),
		"enabled": in.Enabled,
		"sort":    in.Sort,
	}
	if err := h.banners.Update(c.Request.Context(), id, fields); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "banner not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "banner.update", "banner", id.Hex(), "ok")
	response.OK(c, gin.H{"id": id.Hex(), "ok": true})
}

func (h *AdminHandler) DeleteBanner(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.banners.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "banner not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "banner.delete", "banner", id.Hex(), "ok")
	response.OK(c, gin.H{"id": id.Hex(), "ok": true})
}

// ---------- Tag ----------

func (h *AdminHandler) ListTags(c *gin.Context) {
	typ := strings.TrimSpace(c.Query("type"))
	list, err := h.tags.ListAny(c.Request.Context(), typ)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, list)
}

func (h *AdminHandler) CreateTag(c *gin.Context) {
	var in model.Tag
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if !validTagType(in.Type) {
		response.BadRequest(c, "type invalid")
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Value) == "" {
		response.BadRequest(c, "name and value required")
		return
	}
	if len([]rune(in.Name)) > tagNameMax {
		response.BadRequest(c, "name too long")
		return
	}
	if len([]rune(in.Value)) > tagNameMax {
		response.BadRequest(c, "value too long")
		return
	}
	in.Type = strings.TrimSpace(in.Type)
	in.Name = strings.TrimSpace(in.Name)
	in.Value = strings.TrimSpace(in.Value)
	out, err := h.tags.Insert(c.Request.Context(), in)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Conflict(c, "tag already exists")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "tag.create", "tag", out.ID.Hex(), "ok")
	response.OK(c, out)
}

func (h *AdminHandler) UpdateTag(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var in model.Tag
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if !validTagType(in.Type) {
		response.BadRequest(c, "type invalid")
		return
	}
	if strings.TrimSpace(in.Name) == "" || strings.TrimSpace(in.Value) == "" {
		response.BadRequest(c, "name and value required")
		return
	}
	if len([]rune(in.Name)) > tagNameMax || len([]rune(in.Value)) > tagNameMax {
		response.BadRequest(c, "name or value too long")
		return
	}
	existing, err := h.tags.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "tag not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	identityChanged := strings.TrimSpace(in.Type) != existing.Type || strings.TrimSpace(in.Value) != existing.Value
	if (!in.Enabled && existing.Enabled) || identityChanged {
		n, err := h.tags.CountReferenced(c.Request.Context(), existing.Type, existing.Value)
		if err != nil {
			response.ServerError(c, err)
			return
		}
		if n > 0 {
			response.Conflict(c, fmt.Sprintf("当前 %d 个启用案例仍引用该标签, 不允许禁用", n))
			return
		}
	}
	fields := bson.M{
		"type":    in.Type,
		"name":    strings.TrimSpace(in.Name),
		"value":   strings.TrimSpace(in.Value),
		"enabled": in.Enabled,
		"sort":    in.Sort,
	}
	if err := h.tags.Update(c.Request.Context(), id, fields); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			response.Conflict(c, "tag already exists")
			return
		}
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "tag not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "tag.update", "tag", id.Hex(), "ok")
	response.OK(c, gin.H{"id": id.Hex(), "ok": true})
}

func (h *AdminHandler) DeleteTag(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	existing, err := h.tags.Get(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "tag not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	n, err := h.tags.CountReferenced(c.Request.Context(), existing.Type, existing.Value)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	if n > 0 {
		response.Conflict(c, fmt.Sprintf("当前 %d 个启用案例仍引用该标签, 请先改值或禁用", n))
		return
	}
	if err := h.tags.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "tag not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "tag.delete", "tag", id.Hex(), "ok")
	response.OK(c, gin.H{"id": id.Hex(), "ok": true})
}

// ---------- Case ----------

func (h *AdminHandler) ListCases(c *gin.Context) {
	list, err := h.cases.ListAll(c.Request.Context())
	if err != nil {
		response.ServerError(c, err)
		return
	}
	if list == nil {
		list = []model.Case{}
	}
	response.OK(c, list)
}

func (h *AdminHandler) GetCase(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	cc, err := h.cases.GetAny(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	response.OK(c, cc)
}

func (h *AdminHandler) CreateCase(c *gin.Context) {
	var in model.Case
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	normalizeCase(&in)
	if err := h.validateCase(in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.checkPinLimit(c, &in); err != nil {
		response.Conflict(c, err.Error())
		return
	}
	in.PriceLabel = labelForPrice(in.Price)
	out, err := h.cases.Insert(c.Request.Context(), in)
	if err != nil {
		response.ServerError(c, err)
		return
	}
	logOp(c, "case.create", "case", out.ID.Hex(), "ok")
	response.OK(c, out)
}

func (h *AdminHandler) UpdateCase(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var in model.Case
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	normalizeCase(&in)
	if err := h.validateCase(in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	existing, err := h.cases.GetAny(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	if in.Pinned && in.Enabled && (!existing.Pinned || !existing.Enabled) {
		n, err := h.cases.CountPinned(c.Request.Context())
		if err != nil {
			response.ServerError(c, err)
			return
		}
		if n >= maxPinnedCase {
			response.Conflict(c, fmt.Sprintf("置顶案例最多 %d 个", maxPinnedCase))
			return
		}
	}
	in.PriceLabel = labelForPrice(in.Price)
	fields := bson.M{
		"title":      strings.TrimSpace(in.Title),
		"style":      strings.TrimSpace(in.Style),
		"space":      strings.TrimSpace(in.Space),
		"colors":     cleanStrings(in.Colors),
		"size":       strings.TrimSpace(in.Size),
		"area":       strings.TrimSpace(in.Area),
		"estate":     strings.TrimSpace(in.Estate),
		"rooms":      strings.TrimSpace(in.Rooms),
		"tone":       strings.TrimSpace(in.Tone),
		"price":      in.Price,
		"priceLabel": in.PriceLabel,
		"cover":      strings.TrimSpace(in.Cover),
		"images":     cleanStrings(in.Images),
		"highlights": cleanStrings(in.Highlights),
		"materials":  cleanStrings(in.Materials),
		"hardware":   cleanStrings(in.Hardware),
		"pinned":     in.Pinned,
		"enabled":    in.Enabled,
		"source":     strings.TrimSpace(in.Source),
	}
	if err := h.cases.Update(c.Request.Context(), id, fields); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "case.update", "case", id.Hex(), "ok")
	response.OK(c, gin.H{"id": id.Hex(), "ok": true})
}

func (h *AdminHandler) DeleteCase(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	if err := h.cases.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "case.delete", "case", id.Hex(), "ok")
	response.OK(c, gin.H{"id": id.Hex(), "ok": true})
}

func (h *AdminHandler) TogglePin(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	existing, err := h.cases.GetAny(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	want := !existing.Pinned
	if want && existing.Enabled {
		n, err := h.cases.CountPinned(c.Request.Context())
		if err != nil {
			response.ServerError(c, err)
			return
		}
		if n >= maxPinnedCase {
			response.Conflict(c, fmt.Sprintf("置顶案例最多 %d 个", maxPinnedCase))
			return
		}
	}
	if err := h.cases.Update(c.Request.Context(), id, bson.M{"pinned": want}); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "case not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "case.pin", "case", id.Hex(), fmt.Sprintf("pinned=%v", want))
	response.OK(c, gin.H{"id": id.Hex(), "pinned": want})
}

// ---------- 账号管理 ----------

func (h *AdminHandler) ListUsers(c *gin.Context) {
	coll := h.store.Collection(model.CollUser)
	cur, err := coll.Find(c.Request.Context(), bson.M{},
		options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(200))
	if err != nil {
		response.ServerError(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var list []model.User
	if err := cur.All(c.Request.Context(), &list); err != nil {
		response.ServerError(c, err)
		return
	}
	if list == nil {
		list = []model.User{}
	}
	response.OK(c, list)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	id, err := PrimitiveFromParam(c)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}
	var in struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, "invalid body")
		return
	}
	if !validRole(in.Role) {
		response.BadRequest(c, "role invalid")
		return
	}
	if err := h.users.SetRole(c.Request.Context(), id, in.Role); err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.ServerError(c, err)
		return
	}
	logOp(c, "user.role", "user", id.Hex(), "role="+in.Role)
	response.OK(c, gin.H{"id": id.Hex(), "role": in.Role})
}

// ---------- 操作日志 ----------

func (h *AdminHandler) ListOpLogs(c *gin.Context) {
	coll := h.store.Collection(model.CollOpLog)
	cur, err := coll.Find(c.Request.Context(), bson.M{},
		options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(100))
	if err != nil {
		response.ServerError(c, err)
		return
	}
	defer cur.Close(c.Request.Context())
	var list []model.OpLog
	if err := cur.All(c.Request.Context(), &list); err != nil {
		response.ServerError(c, err)
		return
	}
	if list == nil {
		list = []model.OpLog{}
	}
	response.OK(c, list)
}

// ---------- 辅助函数 ----------

func (h *AdminHandler) validateCase(in model.Case) error {
	if strings.TrimSpace(in.Title) == "" {
		return errCase("title is required")
	}
	if len([]rune(in.Title)) > caseTitleMax {
		return errCase("title too long")
	}
	if strings.TrimSpace(in.Style) == "" {
		return errCase("style is required")
	}
	if strings.TrimSpace(in.Space) == "" {
		return errCase("space is required")
	}
	if len(in.Colors) > caseColorsMax {
		return errCase("colors too many")
	}
	if len(in.Images) > caseImagesMax {
		return errCase("images too many")
	}
	if len(in.Highlights) > caseHighlightsMax {
		return errCase("highlights too many")
	}
	if len(in.Materials) > caseMaterialsMax {
		return errCase("materials too many")
	}
	if len(in.Hardware) > caseHardwareMax {
		return errCase("hardware too many")
	}
	for _, m := range in.Materials {
		if len([]rune(m)) > caseDescMax {
			return errCase("material too long")
		}
	}
	if in.Price < 0 {
		return errCase("price must be >= 0")
	}
	if !validImageURL(in.Cover) {
		return errCase("cover must be http(s) or /img-pool/...")
	}
	for _, im := range in.Images {
		if !validImageURL(im) {
			return errCase("image must be http(s) or /img-pool/...")
		}
	}
	return nil
}

func normalizeCase(in *model.Case) {
	in.Title = strings.TrimSpace(in.Title)
	in.Style = strings.TrimSpace(in.Style)
	in.Space = strings.TrimSpace(in.Space)
	in.Colors = cleanStrings(in.Colors)
	in.Size = strings.TrimSpace(in.Size)
	in.Area = strings.TrimSpace(in.Area)
	in.Estate = strings.TrimSpace(in.Estate)
	in.Rooms = strings.TrimSpace(in.Rooms)
	in.Tone = strings.TrimSpace(in.Tone)
	in.Cover = strings.TrimSpace(in.Cover)
	in.Images = cleanStrings(in.Images)
	in.Highlights = cleanStrings(in.Highlights)
	in.Materials = cleanStrings(in.Materials)
	in.Hardware = cleanStrings(in.Hardware)
	in.Source = strings.TrimSpace(in.Source)
}

func (h *AdminHandler) checkPinLimit(c *gin.Context, in *model.Case) error {
	if !in.Pinned || !in.Enabled {
		return nil
	}
	n, err := h.cases.CountPinned(c.Request.Context())
	if err != nil {
		return err
	}
	if n >= maxPinnedCase {
		return errCase(fmt.Sprintf("置顶案例最多 %d 个", maxPinnedCase))
	}
	return nil
}

func logOp(c *gin.Context, action, resource, target, status string) {
	uid, _ := c.Get("uid")
	role, _ := c.Get("role")
	id, _ := uid.(primitive.ObjectID)
	roleStr, _ := role.(string)
	log.Printf("[op] %s %s target=%s ip=%s actor=%s role=%s status=%s",
		action, resource, target, c.ClientIP(), id.Hex(), roleStr, status)
}

// ---------- 校验工具 ----------

func validTagType(t string) bool {
	switch t {
	case model.TagStyle, model.TagSpace, model.TagColor, model.TagSize, model.TagPrice:
		return true
	}
	return false
}

func validRole(r string) bool {
	switch r {
	case model.RoleUser, model.RoleSales, model.RoleSupplier, model.RoleAdmin:
		return true
	}
	return false
}

func validImageURL(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "/img-pool/") {
		return true
	}
	if strings.HasPrefix(s, "/uploads/") {
		return true
	}
	u, err := url.Parse(s)
	if err != nil {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return u.Host != ""
}

func labelForPrice(p int) string {
	switch {
	case p <= 0:
		return "请询价"
	case p < 10000:
		return "1万以下"
	case p < 30000:
		return "1-3万"
	case p < 50000:
		return "3-5万"
	case p < 100000:
		return "5-10万"
	default:
		return "10万+"
	}
}

type caseErr struct{ s string }

func (e *caseErr) Error() string { return e.s }
func errCase(s string) error     { return &caseErr{s} }
