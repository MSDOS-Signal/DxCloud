package handler

import (
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type PermissionHandler struct {
	svc *iam.Service
}

func NewPermissionHandler(svc *iam.Service) *PermissionHandler {
	return &PermissionHandler{svc: svc}
}

func (h *PermissionHandler) List(c *gin.Context) {
	perms, err := h.svc.PermissionList()
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, perms)
}
