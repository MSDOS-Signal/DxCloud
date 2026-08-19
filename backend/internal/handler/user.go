package handler

import (
	"fmt"
	"strconv"

	"github.com/dxcloud/cloud-api/internal/dto"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *iam.Service
}

func NewUserHandler(svc *iam.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

type userRow struct {
	ID        uint64   `json:"id"`
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Nickname  string   `json:"nickname"`
	AvatarURL string   `json:"avatar_url"`
	Status    int8     `json:"status"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	items, total, err := h.svc.UserList(repository.UserFilter{Keyword: c.Query("keyword"), Page: page, Size: size})
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	rows := make([]userRow, 0, len(items))
	for _, u := range items {
		roles, _ := h.svc.GetUserRoleCodes(c.Request.Context(), u.ID)
		rows = append(rows, userRow{
			ID: u.ID, Username: u.Username, Email: u.Email, Nickname: u.Nickname,
			AvatarURL: u.AvatarURL, Status: u.Status, Roles: roles,
			CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	resp.OK(c, dto.PageResult{Total: total, Items: rows})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	user, err := h.svc.UserByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "user not found")
		return
	}
	roles, _ := h.svc.GetUserRoleCodes(c.Request.Context(), user.ID)
	resp.OK(c, userRow{
		ID: user.ID, Username: user.Username, Email: user.Email, Nickname: user.Nickname,
		AvatarURL: user.AvatarURL, Status: user.Status, Roles: roles,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
	})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request: "+err.Error())
		return
	}
	user, err := h.svc.CreateUser(c.Request.Context(), req.Username, req.Email, req.Password, req.Nickname, req.RoleCodes)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "user.create", "user", fmt.Sprintf("%d", user.ID),
		c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"username": user.Username})
	resp.OK(c, map[string]any{"id": user.ID, "username": user.Username})
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req dto.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.UpdateUser(c.Request.Context(), id, req.Nickname, req.Status); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "user.update", "user", fmt.Sprintf("%d", id),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	uid := middleware.GetUserID(c)
	if err := h.svc.DeleteUser(c.Request.Context(), uid, id); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	h.svc.Audit(c.Request.Context(), &uid, "user.delete", "user", fmt.Sprintf("%d", id),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *UserHandler) GrantRoles(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req dto.GrantRolesReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.GrantRoles(c.Request.Context(), id, req.RoleCodes); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.svc.Audit(c.Request.Context(), &uid, "user.grant_roles", "user", fmt.Sprintf("%d", id),
		c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"roles": req.RoleCodes})
	resp.OK(c, nil)
}

// 保持 model 引用（状态常量在后续禁用/启用 UI 使用）
var _ = model.UserStatusActive
