// Package config 读取环境变量并提供默认值.
package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

// Config 启动时使用的配置.
type Config struct {
	HTTPPort       string
	MongoURI       string
	DBName         string
	JWTSecret      string
	CORSOrigins    string
	AdminPhone     string
	StaticCode     string
	SalesPhones    []string
	SupplierPhones []string
	SeedForce      bool
	SeedAllowList  []string
}

// Load 从环境变量加载配置. 默认值仅用于本地开发.
func LoadValidated() (*Config, error) {
	cfg := &Config{
		HTTPPort:       env("STAR_HTTP_PORT", "8181"),
		MongoURI:       env("STAR_MONGO_URI", "mongodb://localhost:27017"),
		DBName:         env("STAR_DB_NAME", "star"),
		JWTSecret:      env("STAR_JWT_SECRET", ""),
		CORSOrigins:    env("STAR_CORS_ORIGINS", "http://localhost:5173,http://localhost:4173"),
		AdminPhone:     env("STAR_ADMIN_PHONE", "13800138000"),
		StaticCode:     env("STAR_STATIC_CODE", ""),
		SalesPhones:    splitPhones(env("STAR_SALES_PHONES", "")),
		SupplierPhones: splitPhones(env("STAR_SUPPLIER_PHONES", "")),
		SeedForce:      env("STAR_ALLOW_DESTRUCTIVE_SEED", "") == "true",
		SeedAllowList:  splitCSV(env("STAR_SEED_ALLOW_DBS", "star_test,star_dev")),
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func Load() *Config {
	cfg, err := LoadValidated()
	if err != nil {
		panic(err)
	}
	if err := cfg.RequireProdCheck(); err != nil {
		panic(err)
	}
	return cfg
}

// IsProd 判断是否生产: 简单根据 HTTP_PORT != 默认 8181 判定不到, 这里只看 JWT 是否设置.
// 生产强制要求 JWT secret.
func (c *Config) RequireProdCheck() error {
	if c.JWTSecret == "" || len(c.JWTSecret) < 16 {
		// 自动生成一个临时 secret 在本地开发用, 但拒绝生产模式.
		if env("STAR_ENV", "dev") == "prod" {
			return errConfig("STAR_JWT_SECRET must be set (>=16 chars) in production")
		}
		c.JWTSecret = env("STAR_JWT_SECRET", "dev-insecure-secret-please-change")
	}
	if c.StaticCode != "" && env("STAR_ENV", "dev") == "prod" {
		return errConfig("STAR_STATIC_CODE is not allowed in production")
	}
	return nil
}

func (c *Config) RoleForPhone(phone string) string {
	if phone == c.AdminPhone {
		return "admin"
	}
	if contains(c.SalesPhones, phone) {
		return "sales"
	}
	if contains(c.SupplierPhones, phone) {
		return "supplier"
	}
	return "user"
}

func (c *Config) CanSeed() bool {
	return c.SeedForce || contains(c.SeedAllowList, c.DBName)
}

func (c *Config) validate() error {
	port, err := strconv.Atoi(c.HTTPPort)
	if err != nil || port < 1 || port > 65535 {
		return errConfig("STAR_HTTP_PORT must be a valid port")
	}
	if strings.TrimSpace(c.MongoURI) == "" {
		return errConfig("STAR_MONGO_URI must not be empty")
	}
	if strings.TrimSpace(c.DBName) == "" {
		return errConfig("STAR_DB_NAME must not be empty")
	}
	if strings.ContainsAny(c.DBName, " /\\") {
		return errConfig("STAR_DB_NAME contains invalid characters")
	}
	if net.ParseIP(c.HTTPPort) != nil {
		return errConfig("STAR_HTTP_PORT must be a port, not an address")
	}
	return nil
}

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return def
}

func splitPhones(s string) []string {
	return splitCSV(s)
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

type configError struct{ s string }

func (e *configError) Error() string { return fmt.Sprintf("invalid config: %s", e.s) }

func errConfig(s string) error { return &configError{s: s} }
