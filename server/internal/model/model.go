// Package model 定义全部数据库模型.
package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// 集合名常量.
const (
	CollBanner = "banners"
	CollTag    = "tags"
	CollCase   = "cases"
	CollUser   = "users"
	CollOpLog  = "op_logs"
)

// TagType 定义前端展示用的标签维度.
const (
	TagStyle = "style"
	TagSpace = "space"
	TagColor = "color"
	TagSize  = "size"
	TagPrice = "price"
)

// RoleXxx 四种用户角色.
const (
	RoleUser     = "user"
	RoleSales    = "sales"
	RoleSupplier = "supplier"
	RoleAdmin    = "admin"
)

// CanSeePrice 角色是否可看精确价.
func CanSeePrice(role string) bool {
	return role == RoleAdmin || role == RoleSales || role == RoleSupplier
}

// User 用户.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Phone     string             `bson:"phone" json:"phone"`
	Nickname  string             `bson:"nickname,omitempty" json:"nickname"`
	Role      string             `bson:"role" json:"role"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Banner 头图轮播.
type Banner struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Image     string             `bson:"image" json:"image"`
	Link      string             `bson:"link" json:"link"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	Sort      int                `bson:"sort" json:"sort"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Tag 标签/筛选维度.
type Tag struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Type      string             `bson:"type" json:"type"`
	Name      string             `bson:"name" json:"name"`
	Value     string             `bson:"value" json:"value"`
	Enabled   bool               `bson:"enabled" json:"enabled"`
	Sort      int                `bson:"sort" json:"sort"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// Case 案例主表.
type Case struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title      string             `bson:"title" json:"title"`
	Style      string             `bson:"style" json:"style"`
	Space      string             `bson:"space" json:"space"`
	Colors     []string           `bson:"colors" json:"colors"`
	Size       string             `bson:"size" json:"size"`
	Area       string             `bson:"area" json:"area"`
	Estate     string             `bson:"estate" json:"estate"`
	Rooms      string             `bson:"rooms" json:"rooms"`
	Tone       string             `bson:"tone" json:"tone"`
	Price      int                `bson:"price" json:"price"`
	PriceLabel string             `bson:"priceLabel" json:"priceLabel"`
	Cover      string             `bson:"cover" json:"cover"`
	Images     []string           `bson:"images" json:"images"`
	Highlights []string           `bson:"highlights" json:"highlights"`
	Materials  []string           `bson:"materials" json:"materials"`
	Hardware   []string           `bson:"hardware" json:"hardware"`
	Pinned     bool               `bson:"pinned" json:"pinned"`
	Enabled    bool               `bson:"enabled" json:"enabled"`
	Source     string             `bson:"source" json:"source"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// OpLog 后台操作日志 (落库).
type OpLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ActorID   string             `bson:"actorId" json:"actorId"`
	ActorRole string             `bson:"actorRole" json:"actorRole"`
	Action    string             `bson:"action" json:"action"`
	Resource  string             `bson:"resource" json:"resource"`
	Target    string             `bson:"target" json:"target"`
	Payload   string             `bson:"payload" json:"payload"`
	Status    string             `bson:"status" json:"status"`
	IP        string             `bson:"ip" json:"ip"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}
