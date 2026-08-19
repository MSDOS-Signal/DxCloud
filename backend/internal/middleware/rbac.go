package middleware

import (
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

// RequirePerm：RBAC 权限点校验（后端强制，与前端隐藏按钮无关）。
// 拒绝时写审计日志（authz.deny）。
func RequirePerm(svc *iam.Service, perm string) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := GetUserID(c)
		ok, err := svc.HasPerm(c.Request.Context(), uid, perm)
		if err != nil {
			resp.Fail(c, errcode.CodeInternal, "internal error")
			c.Abort()
			return
		}
		if !ok {
			u := uid
			svc.Audit(c.Request.Context(), &u, "authz.deny", "api", c.FullPath(),
				c.ClientIP(), GetRequestID(c), 2, map[string]any{"perm": perm})
			resp.Fail(c, errcode.CodeForbidden, "没有操作权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
