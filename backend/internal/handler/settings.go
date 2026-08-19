// Package handler：系统设置（区域/镜像加速源）。
package handler

import (
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	svc    *service.SettingsService
	iamSvc *iam.Service
}

func NewSettingsHandler(svc *service.SettingsService, iamSvc *iam.Service) *SettingsHandler {
	return &SettingsHandler{svc: svc, iamSvc: iamSvc}
}

// Get GET /api/v1/settings —— 读取区域与镜像源配置
func (h *SettingsHandler) Get(c *gin.Context) {
	resp.OK(c, h.svc.Overview())
}

// Update PUT /api/v1/settings —— 保存区域与镜像源配置
func (h *SettingsHandler) Update(c *gin.Context) {
	var req struct {
		Region string `json:"region"`
		Mirror string `json:"registry_mirror"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "请求格式错误")
		return
	}
	uid := middleware.GetUserID(c)
	if err := h.svc.SetRegionAndMirror(req.Region, req.Mirror, &uid); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	h.iamSvc.Audit(c.Request.Context(), &uid, "settings.update", "settings", "region/mirror", c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"region": req.Region, "mirror": req.Mirror})
	resp.OK(c, h.svc.Overview())
}

// Test POST /api/v1/settings/test-mirror —— 测试当前或候选镜像源可达性。
func (h *SettingsHandler) TestMirror(c *gin.Context) {
	var req struct {
		Mirror string `json:"mirror" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "请求格式错误")
		return
	}
	resp.OK(c, h.svc.TestMirror(req.Mirror))
}
