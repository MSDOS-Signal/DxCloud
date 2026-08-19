package handler

import (
	"strconv"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/pipeline"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type PipelineHandler struct {
	engine *pipeline.Engine
	repo   *repository.Repos
	iamSvc *iam.Service
}

func NewPipelineHandler(engine *pipeline.Engine, repo *repository.Repos, iamSvc *iam.Service) *PipelineHandler {
	return &PipelineHandler{engine: engine, repo: repo, iamSvc: iamSvc}
}

func (h *PipelineHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

// canAccessPipeline 资源归属校验（与 ECS accessCheck 同策略：管理员免检，同组织可访问，否则仅属主）。
func (h *PipelineHandler) canAccessPipeline(ac service.AccessCtx, p *model.Pipeline) bool {
	inContext := (ac.OrgID == 0 && (p.OrgID == nil || *p.OrgID == 0)) ||
		(ac.OrgID > 0 && p.OrgID != nil && *p.OrgID == ac.OrgID)
	if inContext && ac.CanManage() {
		return true
	}
	if inContext && p.OwnerID == ac.UserID {
		return true
	}
	return ac.HasRole("superadmin")
}

func (h *PipelineHandler) List(c *gin.Context) {
	ac := h.ac(c)
	var ownerID *uint64
	if !ac.CanManage() {
		ownerID = &ac.UserID
	}
	items, err := h.repo.PipelineList(c.Query("keyword"), ownerID, ac.OrgID)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *PipelineHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	p, err := h.repo.PipelineGetByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
		return
	}
	if !h.canAccessPipeline(h.ac(c), p) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	resp.OK(c, p)
}

func (h *PipelineHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Definition  string `json:"definition" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if _, err := pipeline.ParseDefinition(req.Definition); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	if _, err := h.repo.PipelineGetByName(req.Name, h.ac(c).OrgID); err == nil {
		resp.Fail(c, errcode.CodeConflict, "pipeline name already exists")
		return
	}
	var orgID *uint64
	if h.ac(c).OrgID > 0 {
		v := h.ac(c).OrgID
		orgID = &v
	}
	p := &model.Pipeline{Name: req.Name, Description: req.Description, Definition: req.Definition, Status: 1, OwnerID: middleware.GetUserID(c), OrgID: orgID}
	if err := h.repo.PipelineCreate(p); err != nil {
		resp.Fail(c, errcode.CodeInternal, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "pipeline.create", "pipeline", req.Name, c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, p)
}

func (h *PipelineHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Definition  string `json:"definition" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if _, err := pipeline.ParseDefinition(req.Definition); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	p, err := h.repo.PipelineGetByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
		return
	}
	if !h.canAccessPipeline(h.ac(c), p) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	p.Name = req.Name
	p.Description = req.Description
	p.Definition = req.Definition
	if err := h.repo.PipelineUpdate(p); err != nil {
		resp.Fail(c, errcode.CodeInternal, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "pipeline.update", "pipeline", req.Name, c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *PipelineHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	p, err := h.repo.PipelineGetByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
		return
	}
	if !h.canAccessPipeline(h.ac(c), p) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	if err := h.repo.PipelineSoftDelete(id); err != nil {
		resp.Fail(c, errcode.CodeInternal, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "pipeline.delete", "pipeline", strconv.FormatUint(id, 10), c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

// Run 触发一次执行。
func (h *PipelineHandler) Run(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		Ref string `json:"ref"`
	}
	_ = c.ShouldBindJSON(&req)
	// 归属校验：非属主/管理员不得触发他人 pipeline
	pipe, err := h.repo.PipelineGetByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
		return
	}
	if !h.canAccessPipeline(h.ac(c), pipe) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	run, err := h.engine.CreateRun(c.Request.Context(), h.ac(c), id, req.Ref, "", "manual", c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, run)
}

// canAccessRun 通过 run → pipeline 链校验归属。
func (h *PipelineHandler) canAccessRun(ac service.AccessCtx, runID uint64) bool {
	run, err := h.repo.RunGetByID(runID)
	if err != nil {
		return false
	}
	pipe, err := h.repo.PipelineGetByID(run.PipelineID)
	if err != nil {
		return false
	}
	return h.canAccessPipeline(ac, pipe)
}

// Cancel 取消运行。
func (h *PipelineHandler) Cancel(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if !h.canAccessRun(h.ac(c), id) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	if err := h.engine.Cancel(c.Request.Context(), id); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "pipeline.cancel", "pipeline-run", strconv.FormatUint(id, 10), c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *PipelineHandler) RunList(c *gin.Context) {
	ac := h.ac(c)
	var pipelineID *uint64
	if v := c.Query("pipeline_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			pipelineID = &id
		}
	}
	// 指定了 pipeline_id 时校验归属
	if pipelineID != nil {
		pipe, err := h.repo.PipelineGetByID(*pipelineID)
		if err != nil {
			resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
			return
		}
		if !h.canAccessPipeline(ac, pipe) {
			resp.Fail(c, errcode.CodeForbidden, "forbidden")
			return
		}
	}
	var ownerID *uint64
	if !ac.CanManage() {
		ownerID = &ac.UserID
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := h.repo.RunList(pipelineID, limit, ownerID, ac.OrgID)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *PipelineHandler) RunGet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if !h.canAccessRun(h.ac(c), id) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	run, err := h.repo.RunGetByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "run not found")
		return
	}
	resp.OK(c, run)
}

func (h *PipelineHandler) RunJobs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if !h.canAccessRun(h.ac(c), id) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	jobs, err := h.repo.JobListByRun(id)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, jobs)
}

// RunLogs 读取单个 job 的日志尾部。
func (h *PipelineHandler) RunLogs(c *gin.Context) {
	runID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if !h.canAccessRun(h.ac(c), runID) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	jobID, err := strconv.ParseUint(c.Query("job_id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid job_id")
		return
	}
	logs, err := h.engine.Logs(runID, jobID, 256*1024)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, err.Error())
		return
	}
	resp.OK(c, map[string]any{"logs": logs})
}
