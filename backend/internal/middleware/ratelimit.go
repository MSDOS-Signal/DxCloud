package middleware

import (
	"time"

	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/ratelimit"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

// RateLimit：固定窗口限流（Redis 计数；故障时放行避免锁死）。
func RateLimit(l *ratelimit.Limiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := "rate:" + c.FullPath() + ":" + c.ClientIP()
		ok, err := l.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			c.Next()
			return
		}
		if !ok {
			resp.Fail(c, errcode.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}
