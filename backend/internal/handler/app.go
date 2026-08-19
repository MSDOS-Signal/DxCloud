package handler

import (
	"errors"
	"strconv"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

// ---------- 项目 ----------

type ProjectHandler struct {
	svc    *service.ProjectService
	iamSvc *iam.Service
}

func NewProjectHandler(svc *service.ProjectService, iamSvc *iam.Service) *ProjectHandler {
	return &ProjectHandler{svc: svc, iamSvc: iamSvc}
}

func (h *ProjectHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *ProjectHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Code        string `json:"code"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	p, err := h.svc.Create(c.Request.Context(), h.ac(c), req.Name, req.Code, req.Description, c.ClientIP(), middleware.GetRequestID(c), middleware.GetOrgID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, p)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
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

func (h *ProjectHandler) Environments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	envs, err := h.svc.Environments(h.ac(c), id)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, envs)
}

// ---------- 应用 ----------

type AppHandler struct {
	svc    *service.AppService
	iamSvc *iam.Service
}

func NewAppHandler(svc *service.AppService, iamSvc *iam.Service) *AppHandler {
	return &AppHandler{svc: svc, iamSvc: iamSvc}
}

func (h *AppHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func parseAppReq(c *gin.Context) (service.CreateAppReq, error) {
	var req struct {
		ProjectID       uint64   `json:"project_id"`
		Name            string   `json:"name"`
		Type            string   `json:"type"`
		Image           string   `json:"image"`
		GitURL          string   `json:"git_url"`
		GitBranch       string   `json:"git_branch"`
		Port            int      `json:"port"`
		HealthCheckPath string   `json:"health_check_path"`
		Env             []string `json:"env"`
		Domain          string   `json:"domain"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return service.CreateAppReq{}, err
	}
	return service.CreateAppReq{
		ProjectID: req.ProjectID, Name: req.Name, Type: req.Type, Image: req.Image,
		GitURL: req.GitURL, GitBranch: req.GitBranch, Port: req.Port,
		HealthCheckPath: req.HealthCheckPath, Env: req.Env, Domain: req.Domain,
	}, nil
}

func (h *AppHandler) List(c *gin.Context) {
	var projectID *uint64
	if v := c.Query("project_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			projectID = &id
		}
	}
	items, err := h.svc.List(c.Request.Context(), h.ac(c), projectID, c.Query("keyword"))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *AppHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	app, err := h.svc.Get(c.Request.Context(), h.ac(c), id)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			resp.Fail(c, errcode.CodeForbidden, "no access to this app")
			return
		}
		resp.Fail(c, errcode.CodeNotFound, "app not found")
		return
	}
	resp.OK(c, app)
}

func (h *AppHandler) Create(c *gin.Context) {
	req, err := parseAppReq(c)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	app, err := h.svc.Create(c.Request.Context(), h.ac(c), req, c.ClientIP(), middleware.GetRequestID(c), middleware.GetOrgID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, app)
}

func (h *AppHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	req, err := parseAppReq(c)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.Update(c.Request.Context(), h.ac(c), id, req, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *AppHandler) Delete(c *gin.Context) {
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

// Deploy 手动部署（蓝绿）。
func (h *AppHandler) Deploy(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		Image    string            `json:"image"`
		Env      map[string]string `json:"env"`
		HostPort int               `json:"host_port"`
		Note     string            `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	d, err := h.svc.Deploy(c.Request.Context(), h.ac(c), id, service.DeployReq{
		Image: req.Image, Env: req.Env, HostPort: req.HostPort, Note: req.Note, Trigger: "manual",
	}, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, d)
}

func (h *AppHandler) Rollback(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	vid, err := strconv.ParseUint(c.Param("vid"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid version id")
		return
	}
	d, err := h.svc.Rollback(c.Request.Context(), h.ac(c), id, vid, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, d)
}

func (h *AppHandler) Versions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	items, err := h.svc.VersionsFor(h.ac(c), id)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *AppHandler) Deployments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := h.svc.DeploymentsFor(h.ac(c), id, limit)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

// ---------- 域名 ----------

type DomainHandler struct {
	svc    *service.DomainService
	iamSvc *iam.Service
}

func NewDomainHandler(svc *service.DomainService, iamSvc *iam.Service) *DomainHandler {
	return &DomainHandler{svc: svc, iamSvc: iamSvc}
}

func (h *DomainHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *DomainHandler) List(c *gin.Context) {
	items, err := h.svc.List(h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *DomainHandler) Bind(c *gin.Context) {
	var req struct {
		Domain     string `json:"domain" binding:"required"`
		AppID      uint64 `json:"application_id"`
		TargetPort uint64 `json:"target_port"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	d, err := h.svc.Bind(c.Request.Context(), h.ac(c), req.Domain, req.AppID, req.TargetPort, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, d)
}

func (h *DomainHandler) Unbind(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.Unbind(c.Request.Context(), h.ac(c), id, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}
