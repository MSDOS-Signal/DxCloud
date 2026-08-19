package handler

import (
	"fmt"
	"strconv"

	"github.com/dxcloud/cloud-api/internal/dto"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	svc *iam.Service
}

func NewRoleHandler(svc *iam.Service) *RoleHandler {
	return &RoleHandler{svc: svc}
}

type roleRow struct {
	ID          uint64   `json:"id"`
	Code        string   `json:"code"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	IsSystem    bool     `json:"is_system"`
	Scope       string   `json:"scope"`
	Permissions []string `json:"permissions"`
}

func (h *RoleHandler) List(c *gin.Context) {
	roles, err := h.svc.RoleList()
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	rows := make([]roleRow, 0, len(roles))
	for _, r := range roles {
		perms, _ := h.svc.RolePermCodes(r.ID)
		rows = append(rows, roleRow{
			ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description,
			IsSystem: r.IsSystem, Scope: r.Scope, Permissions: perms,
		})
	}
	resp.OK(c, rows)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req dto.CreateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	role, err := h.svc.CreateRole(c.Request.Context(), req.Code, req.Name, req.Description, req.Scope)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "role.create", "role", fmt.Sprintf("%d", role.ID),
		c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"code": role.Code})
	resp.OK(c, map[string]any{"id": role.ID, "code": role.Code})
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req dto.UpdateRoleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.UpdateRole(c.Request.Context(), id, req.Name, req.Description, req.Scope); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "role.update", "role", fmt.Sprintf("%d", id),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.DeleteRole(c.Request.Context(), id); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "role.delete", "role", fmt.Sprintf("%d", id),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *RoleHandler) GrantPermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req dto.GrantPermsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.GrantPerms(c.Request.Context(), id, req.PermCodes); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "role.grant_perms", "role", fmt.Sprintf("%d", id),
		c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"count": len(req.PermCodes)})
	resp.OK(c, nil)
}
