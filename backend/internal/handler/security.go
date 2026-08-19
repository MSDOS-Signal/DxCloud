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

// ---------- 安全中心 ----------

type SecurityHandler struct {
	svc    *service.SecurityService
	iamSvc *iam.Service
}

func NewSecurityHandler(svc *service.SecurityService, iamSvc *iam.Service) *SecurityHandler {
	return &SecurityHandler{svc: svc, iamSvc: iamSvc}
}

func (h *SecurityHandler) Dashboard(c *gin.Context) {
	out, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "dashboard failed")
		return
	}
	resp.OK(c, out)
}

func (h *SecurityHandler) Scan(c *gin.Context) {
	reports, err := h.svc.ScanAll(c.Request.Context())
	if err != nil || len(reports) == 0 {
		resp.Fail(c, errcode.CodeInternal, "scan failed")
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "security.scan", "security", "-", c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, reports)
}

func (h *SecurityHandler) Reports(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, err := h.svc.Reports(c.Request.Context(), limit)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *SecurityHandler) ReportFindings(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	findings, report, err := h.svc.ReportFindings(c.Request.Context(), id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "report not found")
		return
	}
	resp.OK(c, map[string]any{
		"id": report.ID, "kind": report.Kind, "score": report.Score,
		"finding_count": report.FindingCount, "scanned_at": report.CreatedAt, "findings": findings,
	})
}

// ---------- 密钥托管 ----------

type SecretHandler struct {
	svc    *service.SecretService
	iamSvc *iam.Service
}

func NewSecretHandler(svc *service.SecretService, iamSvc *iam.Service) *SecretHandler {
	return &SecretHandler{svc: svc, iamSvc: iamSvc}
}

func (h *SecretHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *SecretHandler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context(), h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *SecretHandler) Create(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	sec, err := h.svc.Create(c.Request.Context(), h.ac(c), req.Name, req.Value)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "secret.create", "secret", sec.Name, c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, sec)
}

func (h *SecretHandler) Reveal(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	value, err := h.svc.Reveal(c.Request.Context(), h.ac(c), id)
	if err != nil {
		if errors.Is(err, service.ErrSecretForbid) {
			resp.Fail(c, errcode.CodeForbidden, "no access to this secret")
			return
		}
		resp.Fail(c, errcode.CodeNotFound, "secret not found")
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "secret.reveal", "secret", strconv.FormatUint(id, 10), c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, map[string]any{"id": id, "value": value})
}

func (h *SecretHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), h.ac(c), id); err != nil {
		if errors.Is(err, service.ErrSecretForbid) {
			resp.Fail(c, errcode.CodeForbidden, "no access to this secret")
			return
		}
		resp.Fail(c, errcode.CodeNotFound, "secret not found")
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "secret.delete", "secret", strconv.FormatUint(id, 10), c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}
