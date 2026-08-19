package middleware

import (
	"strings"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

const (
	CtxUserID   = "auth_user_id"
	CtxUsername = "auth_username"
	CtxJTI      = "auth_jti"
)

// AuthRequired：JWT 鉴权（签名 → 黑名单 → 用户状态）。
func AuthRequired(svc *iam.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			resp.Fail(c, errcode.CodeUnauthorized, "未登录或登录已过期")
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(h, "Bearer ")
		user, claims, err := svc.Authenticate(c.Request.Context(), tokenStr)
		if err != nil {
			resp.Fail(c, errcode.CodeUnauthorized, "登录状态无效，请重新登录")
			c.Abort()
			return
		}
		c.Set(CtxUserID, user.ID)
		c.Set(CtxUsername, user.Username)
		c.Set(CtxJTI, claims.ID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) uint64 {
	if v, ok := c.Get(CtxUserID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}

func GetUsername(c *gin.Context) string {
	if v, ok := c.Get(CtxUsername); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func GetJTI(c *gin.Context) string {
	if v, ok := c.Get(CtxJTI); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
