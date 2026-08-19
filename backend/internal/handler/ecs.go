package handler

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/dxcloud/cloud-api/internal/dto"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type EcsHandler struct {
	svc    *service.EcsService
	iamSvc *iam.Service
}

func NewEcsHandler(svc *service.EcsService, iamSvc *iam.Service) *EcsHandler {
	return &EcsHandler{svc: svc, iamSvc: iamSvc}
}

// accessCtx 从 JWT 上下文 + 角色 + 租户组装访问者信息。
func (h *EcsHandler) accessCtx(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

func (h *EcsHandler) mapError(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return errcode.CodeNotFound, "实例不存在"
	case errors.Is(err, service.ErrForbidden):
		return errcode.CodeForbidden, "没有访问权限"
	case errors.Is(err, service.ErrQuotaExceed):
		return errcode.CodeBadRequest, err.Error()
	case errors.Is(err, service.ErrNoCredit):
		return errcode.CodeBadRequest, err.Error()
	case errors.Is(err, service.ErrPortConflict):
		return errcode.CodeConflict, err.Error()
	case errors.Is(err, service.ErrBadState):
		return errcode.CodeBadRequest, err.Error()
	default:
		return errcode.CodeInternal, err.Error()
	}
}

func toEcsInfo(inst *model.EcsInstance) dto.EcsInfo {
	info := dto.EcsInfo{
		ID: inst.ID, InstanceNo: inst.InstanceNo, OrgID: inst.OrgID, OwnerID: inst.OwnerID,
		Name: inst.Name, Description: inst.Description, Image: inst.Image,
		CPU: inst.CPU, MemoryMB: inst.MemoryMB, DiskGB: inst.DiskGB,
		RestartPolicy: inst.RestartPolicy, ReadonlyRootfs: inst.ReadonlyRootfs,
		NetworkID: inst.NetworkID,
		FixedIP:   inst.FixedIP, DesiredState: inst.DesiredState, ObservedState: inst.ObservedState,
		ContainerID: inst.ContainerID, ContainerName: inst.ContainerName,
		LastError: inst.LastError, CreatedAt: inst.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	_ = json.Unmarshal([]byte(inst.Ports), &info.Ports)
	_ = json.Unmarshal([]byte(inst.Env), &info.Env)
	_ = json.Unmarshal([]byte(inst.Command), &info.Command)
	_ = json.Unmarshal([]byte(inst.Mounts), &info.Mounts)
	if info.Mounts == nil {
		info.Mounts = []dto.MountResp{}
	}
	if info.Ports == nil {
		info.Ports = []dto.PortMappingResp{}
	}
	if info.Env == nil {
		info.Env = []string{}
	}
	if info.Command == nil {
		info.Command = []string{}
	}
	return info
}

func (h *EcsHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	items, total, err := h.svc.List(c.Request.Context(), h.accessCtx(c), repository.EcsFilter{
		Status: c.Query("status"), Keyword: c.Query("keyword"), Page: page, Size: size,
	})
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	rows := make([]dto.EcsInfo, 0, len(items))
	for i := range items {
		rows = append(rows, toEcsInfo(&items[i]))
	}
	resp.OK(c, dto.PageResult{Total: total, Items: rows})
}

func (h *EcsHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	inst, err := h.svc.Get(c.Request.Context(), h.accessCtx(c), id)
	if err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, toEcsInfo(inst))
}

func (h *EcsHandler) Create(c *gin.Context) {
	var req dto.CreateEcsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request: "+err.Error())
		return
	}
	if req.CPU == 0 {
		req.CPU = 1
	}
	if req.MemoryMB == 0 {
		req.MemoryMB = 512
	}
	if req.DiskGB == 0 {
		req.DiskGB = 10
	}
	ports := make([]service.PortMapping, 0, len(req.Ports))
	for _, p := range req.Ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		ports = append(ports, service.PortMapping{ContainerPort: p.ContainerPort, HostPort: p.HostPort, Protocol: proto})
	}
	inst, err := h.svc.Create(c.Request.Context(), h.accessCtx(c), service.CreateReq{
		Name: req.Name, Description: req.Description, Image: req.Image,
		CPU: req.CPU, MemoryMB: req.MemoryMB, DiskGB: req.DiskGB,
		Ports: ports, Env: req.Env, Command: req.Command,
		RestartPolicy: req.RestartPolicy, ReadonlyRootfs: req.ReadonlyRootfs,
		NetworkID: req.NetworkID, FixedIP: req.FixedIP,
		OrgID: func() *uint64 {
			if o := middleware.GetOrgID(c); o > 0 {
				return &o
			}
			return nil
		}(),
		Mounts: func() []service.MountReq {
			out := make([]service.MountReq, 0, len(req.Mounts))
			for _, m := range req.Mounts {
				out = append(out, service.MountReq{VolumeID: m.VolumeID, Target: m.Target, ReadOnly: m.ReadOnly})
			}
			return out
		}(),
	}, c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, toEcsInfo(inst))
}

func (h *EcsHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req dto.UpdateEcsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.Update(c.Request.Context(), h.accessCtx(c), id, req.Name, req.Description,
		c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, nil)
}

func (h *EcsHandler) Start(c *gin.Context)     { h.lifecycle(c, "start") }
func (h *EcsHandler) Stop(c *gin.Context)      { h.lifecycle(c, "stop") }
func (h *EcsHandler) ForceStop(c *gin.Context) { h.lifecycle(c, "force-stop") }
func (h *EcsHandler) Restart(c *gin.Context)   { h.lifecycle(c, "restart") }

func (h *EcsHandler) lifecycle(c *gin.Context, op string) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	ctx := c.Request.Context()
	ac := h.accessCtx(c)
	ip := c.ClientIP()
	rid := middleware.GetRequestID(c)
	var opErr error
	switch op {
	case "start":
		opErr = h.svc.Start(ctx, ac, id, ip, rid)
	case "stop":
		opErr = h.svc.Stop(ctx, ac, id, false, ip, rid)
	case "force-stop":
		opErr = h.svc.Stop(ctx, ac, id, true, ip, rid)
	case "restart":
		opErr = h.svc.Restart(ctx, ac, id, ip, rid)
	}
	if opErr != nil {
		code, msg := h.mapError(opErr)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, nil)
}

func (h *EcsHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(c.Request.Context(), h.accessCtx(c), id,
		c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, nil)
}

func (h *EcsHandler) Logs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	tail, _ := strconv.Atoi(c.DefaultQuery("tail", "200"))
	logs, err := h.svc.Logs(c.Request.Context(), h.accessCtx(c), id, tail)
	if err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, dto.EcsLogsResp{Logs: logs, Tail: tail})
}

func (h *EcsHandler) Stats(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	st, err := h.svc.Stats(c.Request.Context(), h.accessCtx(c), id)
	if err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, dto.EcsStatsResp{
		CPUPercent: st.CPUPercent, MemUsed: st.MemUsed, MemLimit: st.MemLimit,
		MemPercent: st.MemPercent, NetRxBytes: st.NetRxBytes, NetTxBytes: st.NetTxBytes,
		DiskReadBytes: st.DiskReadBytes, DiskWrite: st.DiskWrite, PIDs: st.PIDs,
	})
}

func (h *EcsHandler) Events(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	events, err := h.svc.Events(c.Request.Context(), h.accessCtx(c), id, limit)
	if err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, events)
}

// Exec 换取 Web Terminal 一次性令牌（60s，用后即焚）。
func (h *EcsHandler) Exec(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	token, instanceNo, err := h.svc.IssueConsoleToken(c.Request.Context(), h.accessCtx(c), id)
	if err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "console.token", "ecs", instanceNo,
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, gin.H{"token": token, "expires_in": 60})
}

// AttachVolume 挂载云磁盘（重建容器实现，数据卷保留）。
func (h *EcsHandler) AttachVolume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		VolumeID uint64 `json:"volume_id" binding:"required"`
		Target   string `json:"target" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.AttachVolume(c.Request.Context(), h.accessCtx(c), id, req.VolumeID, req.Target,
		c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, nil)
}

// DetachVolume 卸载云磁盘。
func (h *EcsHandler) DetachVolume(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	var req struct {
		VolumeID uint64 `json:"volume_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if err := h.svc.DetachVolume(c.Request.Context(), h.accessCtx(c), id, req.VolumeID,
		c.ClientIP(), middleware.GetRequestID(c)); err != nil {
		code, msg := h.mapError(err)
		resp.Fail(c, code, msg)
		return
	}
	resp.OK(c, nil)
}
