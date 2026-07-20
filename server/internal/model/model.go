package model

// Package model 定义星仔高端定制官网的所有数据模型 + 角色权限常量
//
// 角色体系:
//
//	RoleUser     - 普通用户: 仅看价格区间 (priceLabel), price 字段被后端置 0
//	RoleSales    - 销售: 看精准价格数字 (price)
//	RoleSupplier - 供应商: 看精准价格数字 (price)
//	RoleAdmin    - 管理员: 全权限, 可进入运营后台
import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 角色常量
const (
	RoleUser     = "user"     // 普通用户: 仅看价格区间
	RoleSales    = "sales"    // 销售: 看精准价格
	RoleSupplier = "supplier" // 供应商: 看精准价格
	RoleAdmin    = "admin"    // 管理员

	TagStyle = "style" // 一级标签: 风格
	TagSpace = "space" // 二级标签: 空间
	TagColor = "color" // 二级标签: 颜色
	TagSize  = "size"  // 二级标签: 尺寸
	TagPrice = "price" // 二级标签: 价格区间
)

// CanSeePrice 是否查看精准价格 (role-based)
// 仅 sales / supplier / admin 可以看到 price 数字字段; 普通用户只能看 priceLabel 区间
func CanSeePrice(role string) bool {
	return role == RoleAdmin || role == RoleSales || role == RoleSupplier
}

// User 用户表 (手机号 + 角色)
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Phone     string             `bson:"phone" json:"phone"`
	Role      string             `bson:"role" json:"role"`
	Nickname  string             `bson:"nickname,omitempty" json:"nickname,omitempty"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// Banner 首页轮播图
type Banner struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Subtitle  string             `bson:"subtitle,omitempty" json:"subtitle,omitempty"`
	Image     string             `bson:"image" json:"image"`
	Link      string             `bson:"link,omitempty" json:"link,omitempty"`
	Sort      int                `bson:"sort" json:"sort"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// Tag 多级标签统一模型 (type 区分: style/space/color/size/price)
type Tag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`
	Name      string             `bson:"name" json:"name"`
	Value     string             `bson:"value" json:"value"`
	Color     string             `bson:"color,omitempty" json:"color,omitempty"`
	Icon      string             `bson:"icon,omitempty" json:"icon,omitempty"`
	Sort      int                `bson:"sort" json:"sort"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// Case 案例
// Price 字段在 API 返回时会按当前用户角色脱敏 (普通用户置 0)
// PriceLabel 始终返回 (区间: 1万以下/1-3万/3-5万/5-10万/10万+)
type Case struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Style      string             `bson:"style" json:"style"`           // 风格 value
	Space      string             `bson:"space" json:"space"`           // 空间 value
	Colors     []string           `bson:"colors" json:"colors"`         // 颜色 value 列表
	Size       string             `bson:"size" json:"size"`             // 尺寸 value
	Area       string             `bson:"area" json:"area"`             // 面积 (例如 "13㎡")
	Price      int                `bson:"price" json:"price"`           // 精准价格 (元), 普通用户脱敏为 0
	PriceLabel string             `bson:"priceLabel" json:"priceLabel"` // 价格区间 (恒可见)
	Cover      string             `bson:"cover" json:"cover"`           // 封面图 URL
	Images     []string           `bson:"images" json:"images"`         // 多图 URL
	Highlights []string           `bson:"highlights" json:"highlights"` // 设计亮点
	Materials  []string           `bson:"materials" json:"materials"`   // 主材
	Hardware   []string           `bson:"hardware" json:"hardware"`     // 五金
	Pinned     bool               `bson:"pinned" json:"pinned"`         // 是否置顶
	Enabled    bool               `bson:"enabled" json:"enabled"`       // 是否启用
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}