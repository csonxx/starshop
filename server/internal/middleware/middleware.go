// Package middleware 统一中间件: CORS, JWT, 日志等.
package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

// Claims JWT 声明. 仅放非敏感字段.
type Claims struct {
	UID   string `json:"uid"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

func IssueToken(secret, userID, phone, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UID:   userID,
		Phone: phone,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now),
			Issuer:    "star-server",
			Subject:   userID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// CORS 解析逗号分隔白名单, Origin 命中才回写.
func CORS(allow string) gin.HandlerFunc {
	allowList := splitAndTrim(allow)
	if len(allowList) == 1 && allowList[0] == "*" {
		return func(c *gin.Context) {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
			if c.Request.Method == http.MethodOptions {
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
			c.Next()
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		c.Header("Vary", "Origin")
		if origin != "" && contains(allowList, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Auth 强制 JWT 中间件, 同时把用户最新角色刷入 context.
func Auth(secret string, users *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := authenticate(c, secret, users)
		if err != nil {
			if errors.Is(err, repo.ErrNotFound) {
				response.Unauthorized(c, "user not found")
			} else if errors.Is(err, errInvalidToken) {
				response.Unauthorized(c, "invalid token")
			} else {
				response.ServerError(c, err)
			}
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}

// OptionalAuth 公开接口可选鉴权. Token 缺失或无效时不报错, 但要刷新角色.
func OptionalAuth(secret string, users *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		if claims, err := authenticate(c, secret, users); err == nil {
			c.Set("claims", claims)
		}
		c.Next()
	}
}

var errInvalidToken = errors.New("invalid token")

func authenticate(c *gin.Context, secret string, users *repo.UserRepo) (*Claims, error) {
	claims, err := parseToken(c, secret)
	if err != nil {
		return nil, errInvalidToken
	}
	oid, err := primitive.ObjectIDFromHex(claims.UID)
	if err != nil {
		return nil, errInvalidToken
	}
	user, err := users.Get(c.Request.Context(), oid)
	if err != nil {
		return nil, err
	}
	claims.Role = user.Role
	c.Set("uid", oid)
	c.Set("role", user.Role)
	return claims, nil
}

// Admin 必须管理员.
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if r, ok := role.(string); !ok || r != model.RoleAdmin {
			response.Forbidden(c, "admin role required")
			c.Abort()
			return
		}
		c.Next()
	}
}

// parseToken 解析 Bearer Token.
func parseToken(c *gin.Context, secret string) (*Claims, error) {
	tok := tokenFromHeader(c)
	if tok == "" {
		return nil, errors.New("missing token")
	}
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tok, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithIssuer("star-server"), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if claims.UID == "" || claims.Phone == "" || claims.Subject != claims.UID {
		return nil, errInvalidToken
	}
	return claims, nil
}

func tokenFromHeader(c *gin.Context) string {
	auth := c.GetHeader("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// AccessLog 简单访问日志, 不打印敏感 token.
func AccessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		gin.DefaultWriter.Write([]byte(
			"[access] " + c.Request.Method + " " + c.Request.URL.RequestURI() +
				" status=" + statusText(c.Writer.Status()) +
				" dur=" + dur.String() +
				" ip=" + c.ClientIP() + "\n",
		))
	}
}

func statusText(s int) string {
	switch {
	case s >= 500:
		return "5xx"
	case s >= 400:
		return "4xx"
	case s >= 300:
		return "3xx"
	}
	return "2xx"
}

func splitAndTrim(s string) []string {
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

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}
