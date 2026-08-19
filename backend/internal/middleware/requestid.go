// Package middleware 存放 Gin 中间件：请求 ID、访问日志、恢复、CORS。
// Phase 2 起追加：JWT 鉴权、Tenant 上下文、RBAC、审计、限流。
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

// RequestID 为每个请求生成/透传请求 ID（响应头 + 上下文）。
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" {
			rid = "req-" + uuid.NewString()[:8]
		}
		c.Set(RequestIDKey, rid)
		c.Header("X-Request-Id", rid)
		c.Next()
	}
}

// GetRequestID 供其他中间件 / handler 读取当前请求 ID。
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
