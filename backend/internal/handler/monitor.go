package handler

import (
	"strconv"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type MonitorHandler struct {
	svc    *service.MonitorService
	repo   *repository.Repos
	iamSvc *iam.Service
}

func NewMonitorHandler(svc *service.MonitorService, repo *repository.Repos, iamSvc *iam.Service) *MonitorHandler {
	return &MonitorHandler{svc: svc, repo: repo, iamSvc: iamSvc}
}

func (h *MonitorHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

// Dashboard 首页聚合指标。
func (h *MonitorHandler) Dashboard(c *gin.Context) {
	data, err := h.svc.Dashboard(c.Request.Context(), h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, data)
}

// Series 监控曲线（?kind=ecs&ref_id=&minutes=60）。
func (h *MonitorHandler) Series(c *gin.Context) {
	kind := c.DefaultQuery("kind", "ecs")
	var refID *uint64
	if v := c.Query("ref_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			refID = &id
		}
	}
	minutes, _ := strconv.Atoi(c.DefaultQuery("minutes", "60"))
	rows, err := h.svc.Series(c.Request.Context(), kind, refID, minutes, h.ac(c))
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	if rows == nil {
		rows = []repository.MetricSeriesRow{} // 无采样时返回空数组而非 null（前端图表友好）
	}
	resp.OK(c, rows)
}

// Logs 统一日志检索（type=operation|audit|login）。
func (h *MonitorHandler) Logs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	keyword := c.Query("keyword")
	ac := h.ac(c)

	switch c.Query("type") {
	case "audit":
		items, total, err := h.repo.AuditLogList(c.Query("action"), keyword, page, size, ac.UserID, ac.OrgID, ac.CanManage())
		if err != nil {
			resp.Fail(c, errcode.CodeInternal, "query failed")
			return
		}
		resp.OK(c, map[string]any{"total": total, "items": items})
	case "login":
		items, total, err := h.repo.LoginLogList(page, size, ac.UserID, ac.CanManage())
		if err != nil {
			resp.Fail(c, errcode.CodeInternal, "query failed")
			return
		}
		resp.OK(c, map[string]any{"total": total, "items": items})
	default: // operation
		items, total, err := h.repo.OpLogList(c.Query("module"), keyword, page, size, ac.UserID, ac.OrgID, ac.CanManage())
		if err != nil {
			resp.Fail(c, errcode.CodeInternal, "query failed")
			return
		}
		resp.OK(c, map[string]any{"total": total, "items": items})
	}
}

// ---------- 通知 ----------

func (h *MonitorHandler) Notifications(c *gin.Context) {
	unreadOnly := c.Query("unread") == "1"
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	items, total, err := h.repo.NotifyList(middleware.GetUserID(c), unreadOnly, limit)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, map[string]any{"total": total, "items": items})
}

func (h *MonitorHandler) NotificationRead(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.repo.NotifyMarkRead(middleware.GetUserID(c), id); err != nil {
		resp.Fail(c, errcode.CodeNotFound, "notification not found")
		return
	}
	resp.OK(c, nil)
}

func (h *MonitorHandler) NotificationReadAll(c *gin.Context) {
	if err := h.repo.NotifyMarkAllRead(middleware.GetUserID(c)); err != nil {
		resp.Fail(c, errcode.CodeInternal, "failed")
		return
	}
	resp.OK(c, nil)
}
