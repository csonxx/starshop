// Package response 提供统一响应封装.
package response

import (
	"github.com/gin-gonic/gin"
)

// 业务 code 常量.
const (
	CodeOK         = 0
	CodeBadRequest = 400
	CodeUnauthorized = 401
	CodeForbidden  = 403
	CodeNotFound   = 404
	CodeConflict   = 409
	CodeServer     = 500
)

// Body 统一响应.
type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK 成功响应.
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Body{Code: CodeOK, Message: "ok", Data: data})
}

// BadRequest 客户端参数错误.
func BadRequest(c *gin.Context, msg string) {
	if msg == "" {
		msg = "bad request"
	}
	c.JSON(400, Body{Code: CodeBadRequest, Message: msg})
}

// Unauthorized 未登录或登录失效.
func Unauthorized(c *gin.Context, msg string) {
	if msg == "" {
		msg = "unauthorized"
	}
	c.JSON(401, Body{Code: CodeUnauthorized, Message: msg})
}

// Forbidden 已登录但无权限.
func Forbidden(c *gin.Context, msg string) {
	if msg == "" {
		msg = "forbidden"
	}
	c.JSON(403, Body{Code: CodeForbidden, Message: msg})
}

// NotFound 资源不存在.
func NotFound(c *gin.Context, msg string) {
	if msg == "" {
		msg = "not found"
	}
	c.JSON(404, Body{Code: CodeNotFound, Message: msg})
}

// Conflict 资源冲突, 例如数量超限.
func Conflict(c *gin.Context, msg string) {
	if msg == "" {
		msg = "conflict"
	}
	c.JSON(409, Body{Code: CodeConflict, Message: msg})
}

// ServerError 内部错误. 不向客户端泄露原始字符串.
func ServerError(c *gin.Context, _ error) {
	c.JSON(500, Body{Code: CodeServer, Message: "internal error"})
}
