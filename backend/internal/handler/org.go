package handler

import (
	"strconv"
	"time"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

func startOfMonth() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
}

func timeNow() time.Time {
	return time.Now()
}

// ---------- 组织 ----------

type OrgHandler struct {
	svc    *service.OrgService
	iamSvc *iam.Service
}

func NewOrgHandler(svc *service.OrgService, iamSvc *iam.Service) *OrgHandler {
	return &OrgHandler{svc: svc, iamSvc: iamSvc}
}

func (h *OrgHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *OrgHandler) List(c *gin.Context) {
	items, err := h.svc.ListOrgs()
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *OrgHandler) MyOrgs(c *gin.Context) {
	ids, err := h.svc.MemberOrgs(middleware.GetUserID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, ids)
}

func (h *OrgHandler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Code string `json:"code"`
		Plan string `json:"plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	org, err := h.svc.Create(c.Request.Context(), h.ac(c), req.Name, req.Code, req.Plan, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, org)
}

func (h *OrgHandler) Delete(c *gin.Context) {
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

func (h *OrgHandler) Members(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	items, err := h.svc.Members(c.Request.Context(), h.ac(c), id, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		resp.Fail(c, errcode.CodeForbidden, "没有访问权限")
		return
	}
	resp.OK(c, items)
}

func (h *OrgHandler) AddMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		Username string `json:"username" binding:"required"`
		OrgRole  string `json:"org_role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.AddMember(c.Request.Context(), h.ac(c), id, req.Username, req.OrgRole, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

func (h *OrgHandler) RemoveMember(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	uid, err := strconv.ParseUint(c.Param("uid"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid uid")
		return
	}
	if err := h.svc.RemoveMember(c.Request.Context(), h.ac(c), id, uid, c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

// ---------- 配额 ----------

type QuotaHandler struct {
	svc    *service.QuotaService
	iamSvc *iam.Service
}

func NewQuotaHandler(svc *service.QuotaService, iamSvc *iam.Service) *QuotaHandler {
	return &QuotaHandler{svc: svc, iamSvc: iamSvc}
}

func (h *QuotaHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *QuotaHandler) requestedOrg(c *gin.Context) uint64 {
	ac := h.ac(c)
	if v := c.Query("org_id"); v != "" && ac.CanManage() {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return id
		}
	}
	return ac.OrgID
}

func (h *QuotaHandler) List(c *gin.Context) {
	orgID := h.requestedOrg(c)
	items, err := h.svc.List(orgID)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *QuotaHandler) Update(c *gin.Context) {
	orgID := h.requestedOrg(c)
	var req struct {
		ResourceType string `json:"resource_type" binding:"required"`
		LimitValue   int64  `json:"limit_value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.Update(orgID, req.ResourceType, req.LimitValue); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	resp.OK(c, nil)
}

// ---------- 计费 ----------

type BillingHandler struct {
	svc    *service.BillingService
	iamSvc *iam.Service
}

func NewBillingHandler(svc *service.BillingService, iamSvc *iam.Service) *BillingHandler {
	return &BillingHandler{svc: svc, iamSvc: iamSvc}
}

func (h *BillingHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *BillingHandler) requestedOrg(c *gin.Context) uint64 {
	ac := h.ac(c)
	if v := c.Query("org_id"); v != "" && ac.CanManage() {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			return id
		}
	}
	return ac.OrgID
}

func (h *BillingHandler) Summary(c *gin.Context) {
	orgID := h.requestedOrg(c)
	org, err := h.svc.OrgByID(orgID)
	credit := 0.0
	if err == nil {
		credit = org.Credit
	}
	from := startOfMonth()
	sums, _ := h.svc.UsageSum(orgID, from, timeNow())
	resp.OK(c, map[string]any{
		"credit":      credit,
		"usage_month": sums,
		"price":       map[string]float64{"cpu_hour": 0.1, "mem_gb_hour": 0.05, "disk_gb_hour": 0.01},
	})
}

func (h *BillingHandler) Records(c *gin.Context) {
	orgID := h.requestedOrg(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	items, err := h.svc.UsageList(orgID, limit)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

// Tick 手动触发一次用量结算（测试/运维用，admin）。
func (h *BillingHandler) Tick(c *gin.Context) {
	if err := h.svc.Collect(c.Request.Context()); err != nil {
		resp.Fail(c, errcode.CodeInternal, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "billing.tick", "billing", "-", c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

func (h *BillingHandler) Recharge(c *gin.Context) {
	var req struct {
		OrgID  uint64  `json:"org_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	ac := h.ac(c)
	if !ac.CanManage() && req.OrgID != ac.OrgID {
		resp.Fail(c, errcode.CodeForbidden, "只能充值当前组织")
		return
	}
	if err := h.svc.Recharge(c.Request.Context(), req.OrgID, req.Amount); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "billing.recharge", "billing", "-", c.ClientIP(), middleware.GetRequestID(c), 1, map[string]any{"org": req.OrgID, "amount": req.Amount})
	resp.OK(c, nil)
}
