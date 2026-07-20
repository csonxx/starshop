// Package config 集中管理所有运行时配置项
//
// 通过 STAR_* 环境变量注入, 启动时一次性 Load() 加载, 之后跨模块只读使用.
// 见 README "环境变量" 章节
package config

import (
	"os"
	"strings"
)

// Config 不可变运行时配置 (启动时 Load, 之后不允许修改)
type Config struct {
	HTTPPort       string // HTTP 监听端口
	MongoURI       string // MongoDB 连接 URI
	DBName         string // 库名
	JWTSecret      string // JWT 签名密钥 (生产必须改)
	CORSOrigins    string // CORS 白名单, * 或逗号分隔域
	AdminPhone     string // 管理员手机号白名单
	StaticCode     string // 开发期固定验证码
	SalesPhones    string // 销售角色白名单 (csv)
	SupplierPhones string // 供应商角色白名单 (csv)
}

// Load 一次性构造 Config, 任何缺失的环境变量回退到默认值
func Load() *Config {
	return &Config{
		HTTPPort:       env("STAR_HTTP_PORT", "8080"),
		MongoURI:       env("STAR_MONGO_URI", "mongodb://localhost:27017"),
		DBName:         env("STAR_DB_NAME", "star"),
		JWTSecret:      env("STAR_JWT_SECRET", "star-dev-secret-please-change"),
		CORSOrigins:    env("STAR_CORS_ORIGINS", "*"),
		AdminPhone:     env("STAR_ADMIN_PHONE", "13800138000"),
		StaticCode:     env("STAR_STATIC_CODE", "1234"),
		SalesPhones:    env("STAR_SALES_PHONES", "13900000001,13900000002"),
		SupplierPhones: env("STAR_SUPPLIER_PHONES", "13700000001,13700000002"),
	}
}

// IsAdminPhone 手机号是否在管理员白名单 (仅一个)
func (c *Config) IsAdminPhone(p string) bool { return p == c.AdminPhone }

// IsSalesPhone 手机号是否在销售白名单
func (c *Config) IsSalesPhone(p string) bool { return containsCSV(c.SalesPhones, p) }

// IsSupplierPhone 手机号是否在供应商白名单
func (c *Config) IsSupplierPhone(p string) bool { return containsCSV(c.SupplierPhones, p) }

// containsCSV 在逗号分隔字符串中查找精确匹配
func containsCSV(csv, p string) bool {
	for _, s := range splitCSV(csv) {
		if s == p {
			return true
		}
	}
	return false
}

// splitCSV 把 "a, b,c" 切为 ["a","b","c"] (去空格)
func splitCSV(s string) []string {
	out := []string{}
	for _, seg := range strings.Split(s, ",") {
		seg = strings.TrimSpace(seg)
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

// env 读取环境变量, 不存在或为空则返回 fallback
func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}