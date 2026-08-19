package middleware

import (
	"runtime/debug"

	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery 捕获 panic，记录堆栈并返回统一 500 结构（替代 Gin 默认纯文本 Recovery）。
func Recovery(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.String("request_id", GetRequestID(c)),
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())),
				)
				resp.Fail(c, errcode.CodeInternal, "internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
