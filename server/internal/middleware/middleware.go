package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"star/server/internal/model"
	"star/server/internal/repo"
	"star/server/pkg/response"
)

func CORS(origins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origins)
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization,Accept")
		c.Header("Access-Control-Max-Age", "86400")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func RequestLog() gin.HandlerFunc {
	return gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: func(p gin.LogFormatterParams) string {
			return p.TimeStamp.Format("2006-01-02 15:04:05") + " " +
				p.Method + " " + p.Path + " " +
				p.ClientIP + " " + p.Latency.String() + " " +
				p.ErrorMessage + "\n"
		},
		Output: gin.DefaultWriter,
	})
}

type Claims struct {
	UserID string `json:"uid"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func IssueToken(secret, userID, phone, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Phone:  phone,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "star-server",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func Auth(secret string, users *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Unauthorized(c, "missing token")
			c.Abort()
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")
		parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !parsed.Valid {
			response.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}
		claims, ok := parsed.Claims.(*Claims)
		if !ok {
			response.Unauthorized(c, "invalid claims")
			c.Abort()
			return
		}
		c.Set("claims", claims)
		if claims.Role == model.RoleAdmin {
			c.Set("isAdmin", true)
		}
		c.Next()
	}
}

// OptionalAuth 公开接口的可选鉴权: 有 token 就解析, 无 token 也继续.
// 用于识别当前用户角色以便价格脱敏
func OptionalAuth(secret string, users *repo.UserRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Next()
			return
		}
		raw := strings.TrimPrefix(header, "Bearer ")
		parsed, err := jwt.ParseWithClaims(raw, &Claims{}, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !parsed.Valid {
			c.Next()
			return
		}
		if claims, ok := parsed.Claims.(*Claims); ok {
			c.Set("claims", claims)
			if claims.Role == model.RoleAdmin {
				c.Set("isAdmin", true)
			}
		}
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get("isAdmin")
		if !ok || v != true {
			response.Forbidden(c, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}