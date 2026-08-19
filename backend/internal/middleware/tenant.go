package middleware

import (
	"strconv"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

const CtxOrgID = "tenant_org_id"

// TenantContext 租户上下文：解析 X-Org-Id 并校验成员资格（superadmin 免检）。
// 无头时 org_id=0（单租户兼容模式，历史资源 org 为空）。
func TenantContext(iamSvc *iam.Service, orgSvc *service.OrgService) gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := uint64(0)
		if v := c.GetHeader("X-Org-Id"); v != "" {
			if id, err := strconv.ParseUint(v, 10, 64); err == nil && id > 0 {
				uid := GetUserID(c)
				roles, _ := iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
				isSA := false
				for _, r := range roles {
					if r == "superadmin" {
						isSA = true
						break
					}
				}
				if !isSA {
					if _, ok := orgSvc.IsMember(c.Request.Context(), id, uid); !ok {
						resp.Fail(c, errcode.CodeForbidden, "当前账号不属于该组织")
						c.Abort()
						return
					}
				}
				orgID = id
			}
		}
		c.Set(CtxOrgID, orgID)
		c.Next()
	}
}

// GetOrgID 读取租户上下文（0 = 未指定）。
func GetOrgID(c *gin.Context) uint64 {
	if v, ok := c.Get(CtxOrgID); ok {
		if id, ok := v.(uint64); ok {
			return id
		}
	}
	return 0
}
