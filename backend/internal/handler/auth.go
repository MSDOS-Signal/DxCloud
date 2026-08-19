package handler

import (
	"errors"
	"fmt"

	"github.com/dxcloud/cloud-api/internal/dto"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *iam.Service
}

func NewAuthHandler(svc *iam.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request: "+err.Error())
		return
	}
	user, err := h.svc.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	access, refresh, exp, err := h.svc.IssueTokens(c.Request.Context(), user)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "issue token failed")
		return
	}
	h.svc.Audit(c.Request.Context(), &user.ID, "auth.register", "user", fmt.Sprintf("%d", user.ID),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, dto.TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: exp, TokenType: "Bearer"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request: "+err.Error())
		return
	}
	user, err := h.svc.Login(c.Request.Context(), req.Username, req.Password, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		h.svc.Audit(c.Request.Context(), nil, "auth.login", "user", req.Username,
			c.ClientIP(), middleware.GetRequestID(c), 2, map[string]any{"username": req.Username})
		if errors.Is(err, iam.ErrAccountLocked) {
			resp.Fail(c, errcode.CodeUnauthorized, "账号已锁定：连续失败 5 次，请 15 分钟后重试")
			return
		}
		resp.Fail(c, errcode.CodeUnauthorized, "用户名或密码错误")
		return
	}
	access, refresh, exp, err := h.svc.IssueTokens(c.Request.Context(), user)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "issue token failed")
		return
	}
	h.svc.Audit(c.Request.Context(), &user.ID, "auth.login", "user", fmt.Sprintf("%d", user.ID),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, dto.TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: exp, TokenType: "Bearer"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	access, refresh, exp, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		resp.Fail(c, errcode.CodeUnauthorized, "登录已过期，请重新登录")
		return
	}
	resp.OK(c, dto.TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: exp, TokenType: "Bearer"})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutReq
	_ = c.ShouldBindJSON(&req)
	uid := middleware.GetUserID(c)
	h.svc.Logout(c.Request.Context(), middleware.GetJTI(c), req.RefreshToken)
	h.svc.Audit(c.Request.Context(), &uid, "auth.logout", "user", fmt.Sprintf("%d", uid),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, roles, roleNames, perms, err := h.svc.Me(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeUnauthorized, "unauthorized")
		return
	}
	resp.OK(c, dto.UserInfo{
		ID: user.ID, Username: user.Username, Email: user.Email,
		Nickname: user.Nickname, AvatarURL: user.AvatarURL, Status: user.Status,
		Roles: roles, RoleNames: roleNames, Permissions: perms,
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req struct {
		Nickname  string `json:"nickname"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "请求格式错误")
		return
	}
	uid := middleware.GetUserID(c)
	if err := h.svc.UpdateProfile(c.Request.Context(), uid, req.Nickname, req.AvatarURL); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	h.svc.Audit(c.Request.Context(), &uid, "auth.update_profile", "user", fmt.Sprintf("%d", uid),
		c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"nickname": req.Nickname})
	resp.OK(c, nil)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	uid := middleware.GetUserID(c)
	if err := h.svc.ChangePassword(c.Request.Context(), uid, req.OldPassword, req.NewPassword); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	// 改密后当前会话立即失效（需重新登录），同时撤销其余全部会话
	h.svc.Logout(c.Request.Context(), middleware.GetJTI(c), "")
	h.svc.Audit(c.Request.Context(), &uid, "auth.change_password", "user", fmt.Sprintf("%d", uid),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *AuthHandler) Sessions(c *gin.Context) {
	list, err := h.svc.Sessions(c.Request.Context(), middleware.GetUserID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "list sessions failed")
		return
	}
	resp.OK(c, list)
}

func (h *AuthHandler) DeleteSession(c *gin.Context) {
	if err := h.svc.DeleteSession(c.Request.Context(), middleware.GetUserID(c), c.Param("id")); err != nil {
		resp.Fail(c, errcode.CodeNotFound, "session not found")
		return
	}
	resp.OK(c, nil)
}
