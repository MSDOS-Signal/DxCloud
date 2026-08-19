package handler

import (
	"strconv"
	"strings"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

// ---------- 镜像 ----------

type ImageHandler struct {
	svc    *service.ImageService
	iamSvc *iam.Service
}

func NewImageHandler(svc *service.ImageService, iamSvc *iam.Service) *ImageHandler {
	return &ImageHandler{svc: svc, iamSvc: iamSvc}
}

func (h *ImageHandler) accessCtx(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *ImageHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	items, total, err := h.svc.List(page, size, c.Query("keyword"), h.accessCtx(c).OrgID)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, dtoPage(total, items))
}

func (h *ImageHandler) PullLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	status, pullError, logs, progress, err := h.svc.PullLogs(h.accessCtx(c).OrgID, id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "image task not found")
		return
	}
	resp.OK(c, map[string]any{"status": status, "pull_error": pullError, "logs": logs, "progress": progress})
}

func (h *ImageHandler) Pull(c *gin.Context) {
	var req struct {
		Image string `json:"image" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	img, err := h.svc.Pull(c.Request.Context(), h.accessCtx(c), req.Image, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, map[string]any{"id": img.ID, "repo": img.Repo, "tag": img.Tag, "status": img.Status, "async": true})
}

func (h *ImageHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), h.accessCtx(c), id, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *ImageHandler) Tag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		Repo string `json:"repo" binding:"required"`
		Tag  string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	img, err := h.svc.Tag(c.Request.Context(), h.accessCtx(c), id, req.Repo, req.Tag, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, map[string]any{"id": img.ID, "repo": img.Repo, "tag": img.Tag})
}

// Search GET /images/search?q=xxx —— 镜像名称自动补全建议（Docker Hub 在线搜索 + 国内兜底目录）
func (h *ImageHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		resp.OK(c, []service.ImageSearchResult{})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	resp.OK(c, h.svc.Search(c.Request.Context(), q, limit))
}

// ---------- 网络 ----------

type NetworkHandler struct {
	svc    *service.NetworkService
	iamSvc *iam.Service
}

func NewNetworkHandler(svc *service.NetworkService, iamSvc *iam.Service) *NetworkHandler {
	return &NetworkHandler{svc: svc, iamSvc: iamSvc}
}

func (h *NetworkHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *NetworkHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *NetworkHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	info, err := h.svc.Inspect(c.Request.Context(), h.ac(c), id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "network not found")
		return
	}
	resp.OK(c, info)
}

func (h *NetworkHandler) Create(c *gin.Context) {
	var req struct {
		Name     string `json:"name" binding:"required"`
		Subnet   string `json:"subnet"`
		Gateway  string `json:"gateway"`
		IPRange  string `json:"ip_range"`
		Internal bool   `json:"internal"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	n, err := h.svc.Create(c.Request.Context(), h.ac(c), req.Name, req.Subnet, req.Gateway, req.IPRange, req.Internal, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, n)
}

func (h *NetworkHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), h.ac(c), id, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *NetworkHandler) Connect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		InstanceID uint64 `json:"instance_id" binding:"required"`
		FixedIP    string `json:"fixed_ip"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.Connect(c.Request.Context(), h.ac(c), id, req.InstanceID, req.FixedIP, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *NetworkHandler) Disconnect(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		InstanceID uint64 `json:"instance_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.Disconnect(c.Request.Context(), h.ac(c), id, req.InstanceID, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

// ---------- 存储 ----------

type VolumeHandler struct {
	svc    *service.VolumeService
	iamSvc *iam.Service
}

func NewVolumeHandler(svc *service.VolumeService, iamSvc *iam.Service) *VolumeHandler {
	return &VolumeHandler{svc: svc, iamSvc: iamSvc}
}

func (h *VolumeHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *VolumeHandler) List(c *gin.Context) {
	items, err := h.svc.List(h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *VolumeHandler) Create(c *gin.Context) {
	var req struct {
		Name       string `json:"name" binding:"required"`
		CapacityGB int    `json:"capacity_gb"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	v, err := h.svc.Create(c.Request.Context(), h.ac(c), req.Name, req.CapacityGB, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, v)
}

func (h *VolumeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), h.ac(c), id, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

// ---------- Registry ----------

type RegistryHandler struct {
	svc    *service.RegistryService
	iamSvc *iam.Service
}

func NewRegistryHandler(svc *service.RegistryService, iamSvc *iam.Service) *RegistryHandler {
	return &RegistryHandler{svc: svc, iamSvc: iamSvc}
}

func (h *RegistryHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *RegistryHandler) List(c *gin.Context) {
	items, err := h.svc.ListRegistries()
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *RegistryHandler) Repositories(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	repos, err := h.svc.Repositories(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, repos)
}

func (h *RegistryHandler) Pull(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		Name string `json:"name" binding:"required"`
		Tag  string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.Pull(c.Request.Context(), h.ac(c), id, req.Name, req.Tag, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *RegistryHandler) DeleteTag(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	// 仓库名可含斜杠（namespace/name），Gin :param 无法匹配，故用 POST + body 传参
	var req struct {
		Name string `json:"name" binding:"required"`
		Tag  string `json:"tag" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.DeleteTag(c.Request.Context(), id, req.Name, req.Tag); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "registry.delete_tag", "registry", req.Name+":"+req.Tag,
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func dtoPage(total int64, items any) any {
	return map[string]any{"total": total, "items": items}
}
