// Package service：ECS 业务编排层（状态机 + 配额 + 端口分配 + 审计 + 事件）。
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/pkg/redisx"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrNotFound       = errors.New("ecs instance not found")
	ErrForbidden      = errors.New("no access to this instance")
	ErrQuotaExceed    = errors.New("配额不足")
	ErrNoCredit       = errors.New("组织虚拟余额不足，请联系管理员充值")
	ErrPortConflict   = errors.New("宿主端口已被占用，请更换端口")
	ErrBadState       = errors.New("当前状态不支持该操作，请稍后重试")
	transitionTimeout = 90 * time.Second
)

// AccessCtx 请求者上下文（handler 从 JWT 与角色组装）。
type AccessCtx struct {
	UserID uint64
	Roles  []string
	OrgID  uint64 // 当前租户上下文（0=未指定，单租户模式）
}

func (ac AccessCtx) hasRole(role string) bool {
	for _, r := range ac.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (ac AccessCtx) HasRole(role string) bool {
	return ac.hasRole(role)
}

// canManage：superadmin/admin/developer 可管理全部实例。
func (ac AccessCtx) canManage() bool {
	return ac.hasRole("superadmin") || ac.hasRole("admin") || ac.hasRole("developer")
}

// CanManage 导出版本，供 handler 层做资源归属判定（IDOR 防护）。
func (ac AccessCtx) CanManage() bool {
	return ac.canManage()
}

// canOperate：operator 可运维（启停/重启/日志/监控），不可增删改。
func (ac AccessCtx) canOperate() bool {
	return ac.canManage() || ac.hasRole("operator")
}

type EcsService struct {
	repo       *repository.Repos
	compute    docker.ComputeProvider
	iamSvc     *iam.Service
	rdb        *redis.Client
	quotaSvc   *QuotaService
	billingSvc *BillingService
	log        *zap.Logger
}

func NewEcsService(repo *repository.Repos, compute docker.ComputeProvider, iamSvc *iam.Service, rdb *redis.Client, quotaSvc *QuotaService, billingSvc *BillingService, log *zap.Logger) *EcsService {
	return &EcsService{repo: repo, compute: compute, iamSvc: iamSvc, rdb: rdb, quotaSvc: quotaSvc, billingSvc: billingSvc, log: log}
}

func genInstanceNo() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return "i-" + hex.EncodeToString(b)
}

// event 写实例事件。
func (s *EcsService) event(instanceID uint64, typ, level, message string, actorID *uint64, requestID string) {
	_ = s.repo.EcsEventCreate(&model.EcsEvent{
		InstanceID: instanceID, EventType: typ, Level: level, Message: message,
		ActorID: actorID, RequestID: requestID,
	})
}

func (s *EcsService) audit(ctx context.Context, ac AccessCtx, action, resourceID, ip, requestID string, status int8, detail any) {
	u := ac.UserID
	s.iamSvc.Audit(ctx, &u, action, "ecs", resourceID, ip, requestID, status, detail)
}

// opLog 写操作日志（用户操作流水，Phase 9 日志中心）。
func (s *EcsService) opLog(ac AccessCtx, module, action, targetID, targetName, ip string, result int8, durationMs int64) {
	uid := ac.UserID
	_ = s.repo.OpLogCreate(&model.OperationLog{
		UserID: &uid, Module: module, Action: action,
		TargetType: "ecs", TargetID: targetID, TargetName: targetName,
		Result: result, DurationMs: durationMs, IP: ip,
	})
}

// accessCheck 资源归属校验（user 角色仅限自己资源；IDOR 在 SQL 层已被 owner/org 过滤兜底）。
func accessAllowed(ac AccessCtx, inst *model.EcsInstance, operate bool) bool {
	// 租户上下文：同组织成员可访问组织资源（写操作由 RBAC 权限另行把关）
	if ac.OrgID > 0 && inst.OrgID != nil && *inst.OrgID == ac.OrgID {
		return true
	}
	if ac.OrgID == 0 && inst.OrgID == nil {
		if ac.canManage() || inst.OwnerID == ac.UserID || (operate && ac.canOperate()) {
			return true
		}
		return false
	}
	// 当前组织上下文中的资源必须归属当前组织；未指定组织时不能借 ID 访问组织资源
	if ac.hasRole("superadmin") {
		return true
	}
	if ac.OrgID > 0 && inst.OrgID != nil && *inst.OrgID == ac.OrgID {
		if operate && ac.canOperate() {
			return true
		}
		if ac.canManage() || inst.OwnerID == ac.UserID {
			return true
		}
	}
	if orgMatches(ac.OrgID, inst.OrgID) && inst.OwnerID == ac.UserID {
		return true
	}
	return false
}

func (s *EcsService) accessCheck(ac AccessCtx, inst *model.EcsInstance, operate bool) error {
	if accessAllowed(ac, inst, operate) {
		return nil
	}
	return ErrForbidden
}

// ---------- 端口 ----------

type PortMapping = docker.PortMapping

func parsePorts(inst *model.EcsInstance) []PortMapping {
	if inst.Ports == "" {
		return nil
	}
	var out []PortMapping
	_ = json.Unmarshal([]byte(inst.Ports), &out)
	return out
}

// checkPortsFree 双重校验：Docker 运行时占用 + DB 已分配（本实例自己占用的端口除外）。
func (s *EcsService) checkPortsFree(ctx context.Context, ports []PortMapping, selfID uint64) error {
	if len(ports) == 0 {
		return nil
	}
	used, err := s.compute.UsedHostPorts(ctx)
	if err != nil {
		s.log.Warn("query docker ports failed", zap.Error(err))
		used = map[int]bool{}
	}
	// DB 已分配
	rows, _, err := s.repo.EcsList(repository.EcsFilter{Page: 1, Size: 10000})
	if err != nil {
		return err
	}
	dbUsed := map[int]uint64{}
	for i := range rows {
		if rows[i].ID == selfID {
			continue
		}
		for _, p := range parsePorts(&rows[i]) {
			dbUsed[p.HostPort] = rows[i].ID
		}
	}
	for _, p := range ports {
		if p.HostPort <= 0 || p.HostPort > 65535 {
			return fmt.Errorf("宿主端口无效：%d（需 1-65535）", p.HostPort)
		}
		if used[p.HostPort] {
			return fmt.Errorf("%w：%d（被运行中容器占用）", ErrPortConflict, p.HostPort)
		}
		if _, ok := dbUsed[p.HostPort]; ok {
			return fmt.Errorf("%w：%d（已被其他实例分配）", ErrPortConflict, p.HostPort)
		}
	}
	return nil
}

// ---------- 创建 ----------

func (s *EcsService) Create(ctx context.Context, ac AccessCtx, req CreateReq, ip, requestID string) (*model.EcsInstance, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if ac.OrgID > 0 && req.OrgID != nil && *req.OrgID != ac.OrgID {
		return nil, ErrForbidden
	}
	// 配额：组织配额优先；虚拟余额门禁
	orgID := uint64(0)
	if req.OrgID != nil {
		orgID = *req.OrgID
	}
	if err := s.quotaSvc.CheckEcsQuota(orgID, ac.UserID, req.CPU, req.MemoryMB); err != nil {
		return nil, err
	}
	if !s.billingSvc.HasCredit(orgID) {
		return nil, ErrNoCredit
	}
	if err := s.checkPortsFree(ctx, req.Ports, 0); err != nil {
		return nil, err
	}

	envJSON, _ := json.Marshal(req.Env)
	cmdJSON, _ := json.Marshal(req.Command)
	portsJSON, _ := json.Marshal(req.Ports)

	// 解析网络与挂载（云 ID → Docker 资源名）
	dockerNetID := ""
	if req.NetworkID != "" {
		var netDBID uint64
		if _, err := fmt.Sscanf(req.NetworkID, "%d", &netDBID); err != nil {
			return nil, errors.New("invalid network_id")
		}
		netRow, err := s.repo.NetworkGetByID(netDBID)
		if err != nil {
			return nil, errors.New("network not found")
		}
		if !orgMatches(ac.OrgID, netRow.OrgID) {
			return nil, errors.New("network not found")
		}
		dockerNetID = netRow.DockerNetID
	}
	mounts := make([]docker.Mount, 0, len(req.Mounts))
	for _, m := range req.Mounts {
		vol, err := s.repo.VolumeGetByID(m.VolumeID)
		if err != nil {
			return nil, fmt.Errorf("volume %d not found", m.VolumeID)
		}
		if !orgMatches(ac.OrgID, vol.OrgID) {
			return nil, fmt.Errorf("volume %d not found", m.VolumeID)
		}
		mounts = append(mounts, docker.Mount{VolumeName: vol.DockerName, Target: m.Target, ReadOnly: m.ReadOnly})
	}
	mountsJSON, _ := json.Marshal(mounts)

	inst := &model.EcsInstance{
		InstanceNo:     genInstanceNo(),
		OrgID:          req.OrgID,
		OwnerID:        ac.UserID,
		Name:           req.Name,
		Description:    req.Description,
		Image:          req.Image,
		CPU:            req.CPU,
		MemoryMB:       req.MemoryMB,
		DiskGB:         req.DiskGB,
		NetworkID:      dockerNetID,
		FixedIP:        req.FixedIP,
		Ports:          string(portsJSON),
		Env:            string(envJSON),
		Command:        string(cmdJSON),
		Mounts:         string(mountsJSON),
		RestartPolicy:  req.RestartPolicy,
		ReadonlyRootfs: req.ReadonlyRootfs,
		DesiredState:   model.EcsCreating,
		ObservedState:  model.EcsCreating,
	}
	if err := s.repo.EcsCreate(inst); err != nil {
		return nil, fmt.Errorf("db create: %w", err)
	}
	actor := ac.UserID
	s.event(inst.ID, "create", "info", "创建实例请求已受理", &actor, requestID)
	s.audit(ctx, ac, "ecs.create", inst.InstanceNo, ip, requestID, 1, map[string]any{"name": req.Name, "image": req.Image})
	s.opLog(ac, "ecs", "create", inst.InstanceNo, req.Name, ip, 1, 0)

	// Docker 创建 + 启动
	spec := docker.CreateSpec{
		Name:           "dx-" + inst.InstanceNo,
		Image:          req.Image,
		CPU:            req.CPU,
		MemoryMB:       req.MemoryMB,
		Env:            req.Env,
		Cmd:            req.Command,
		Ports:          req.Ports,
		Mounts:         mounts,
		RestartPolicy:  req.RestartPolicy,
		NetworkID:      dockerNetID,
		FixedIP:        req.FixedIP,
		ReadonlyRootfs: req.ReadonlyRootfs,
		Labels: map[string]string{
			"com.dxcloud.kind":        "ecs",
			"com.dxcloud.instance-id": inst.InstanceNo,
			"com.dxcloud.owner-id":    fmt.Sprintf("%d", ac.UserID),
			"com.dxcloud.org-id":      orgLabel(req.OrgID),
		},
	}
	info, err := s.compute.Create(ctx, spec)
	if err != nil {
		inst.DesiredState = model.EcsFailed
		inst.ObservedState = model.EcsFailed
		if strings.Contains(err.Error(), "No such image") || strings.Contains(err.Error(), "pull access denied") {
			inst.LastError = "镜像不存在，请先在镜像中心拉取：" + req.Image
		} else {
			inst.LastError = err.Error()
		}
		_ = s.repo.EcsUpdateState(inst)
		s.event(inst.ID, "create", "error", "创建失败："+inst.LastError, &actor, requestID)
		s.audit(ctx, ac, "ecs.create", inst.InstanceNo, ip, requestID, 2, map[string]any{"error": inst.LastError})
		notify(s.repo, ac.UserID, "ecs", "实例创建失败", "「"+req.Name+"」（"+inst.InstanceNo+"）创建失败："+inst.LastError, "/ecs/"+fmt.Sprintf("%d", inst.ID))
		return inst, errors.New(inst.LastError)
	}

	inst.ContainerID = info.ID
	inst.ContainerName = info.Name
	inst.FixedIP = info.IP
	inst.DesiredState = model.EcsRunning
	inst.ObservedState = model.EcsRunning
	inst.LastError = ""
	if err := s.repo.EcsUpdateState(inst); err != nil {
		s.log.Error("update ecs state failed", zap.Error(err))
	}
	s.event(inst.ID, "create", "info", "实例创建成功并已启动", &actor, requestID)
	notify(s.repo, ac.UserID, "ecs", "实例创建成功", "「"+req.Name+"」（"+inst.InstanceNo+"）已启动，镜像 "+req.Image, "/ecs/"+fmt.Sprintf("%d", inst.ID))
	return inst, nil
}

// ---------- 查询 ----------

func (s *EcsService) Get(ctx context.Context, ac AccessCtx, id uint64) (*model.EcsInstance, error) {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.accessCheck(ac, inst, true); err != nil {
		return nil, ErrForbidden
	}
	// 详情页实时对账（单次 inspect，代价低）
	if inst.ContainerID != "" {
		if info, err := s.compute.Inspect(ctx, inst.ContainerID); err == nil {
			if mapped := stateFromDocker(info); mapped != inst.ObservedState {
				inst.ObservedState = mapped
				_ = s.repo.EcsUpdateState(inst)
				s.event(inst.ID, "reconcile", "warn", "状态对账：Docker 实际状态为 "+mapped, nil, "")
			}
		}
	}
	return inst, nil
}

func (s *EcsService) List(ctx context.Context, ac AccessCtx, f repository.EcsFilter) ([]model.EcsInstance, int64, error) {
	if ac.OrgID > 0 {
		oid := ac.OrgID
		f.OrgID = &oid // 租户上下文：列表强制组织过滤
	} else {
		f.DefaultSpace = true
	}
	if !ac.canManage() && ac.OrgID == 0 {
		uid := ac.UserID
		f.OwnerID = &uid
	}
	return s.repo.EcsList(f)
}

func (s *EcsService) Update(ctx context.Context, ac AccessCtx, id uint64, name, description, ip, requestID string) error {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return ErrNotFound
	}
	if err := s.accessCheck(ac, inst, false); err != nil {
		return err
	}
	inst.Name = name
	inst.Description = description
	if err := s.repo.EcsUpdateInfo(inst); err != nil {
		return err
	}
	s.event(inst.ID, "update", "info", "实例信息已更新", &ac.UserID, requestID)
	s.audit(ctx, ac, "ecs.update", inst.InstanceNo, ip, requestID, 1, nil)
	return nil
}

// ---------- 生命周期 ----------

func (s *EcsService) action(ctx context.Context, ac AccessCtx, id uint64, operate bool,
	transition string, fn func(ctx context.Context, inst *model.EcsInstance) error, actionName, ip, requestID string) error {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return ErrNotFound
	}
	if err := s.accessCheck(ac, inst, operate); err != nil {
		return err
	}
	if inst.ContainerID == "" {
		return ErrBadState
	}
	inst.DesiredState = transition
	_ = s.repo.EcsUpdateState(inst)
	s.event(inst.ID, actionName, "info", "执行 "+actionName, &ac.UserID, requestID)
	start := time.Now()

	if err := fn(ctx, inst); err != nil {
		inst.DesiredState = model.EcsFailed
		inst.ObservedState = model.EcsFailed
		inst.LastError = err.Error()
		_ = s.repo.EcsUpdateState(inst)
		s.event(inst.ID, actionName, "error", actionName+" 失败："+err.Error(), &ac.UserID, requestID)
		s.audit(ctx, ac, "ecs."+actionName, inst.InstanceNo, ip, requestID, 2, map[string]any{"error": err.Error()})
		s.opLog(ac, "ecs", actionName, inst.InstanceNo, inst.Name, ip, 2, time.Since(start).Milliseconds())
		notify(s.repo, ac.UserID, "ecs", "实例"+ecsActionName(actionName)+"失败", "「"+inst.Name+"」（"+inst.InstanceNo+"）"+ecsActionName(actionName)+"失败："+err.Error(), "/ecs/"+fmt.Sprintf("%d", inst.ID))
		return err
	}

	// 操作成功后按 Docker 实况更新
	info, err := s.compute.Inspect(ctx, inst.ContainerID)
	if err != nil {
		inst.DesiredState = model.EcsUnknown
		inst.ObservedState = model.EcsUnknown
		inst.LastError = err.Error()
	} else {
		inst.DesiredState = stateFromDocker(info)
		inst.ObservedState = inst.DesiredState
		inst.LastError = ""
	}
	_ = s.repo.EcsUpdateState(inst)
	s.event(inst.ID, actionName, "info", actionName+" 完成，状态 "+inst.ObservedState, &ac.UserID, requestID)
	s.audit(ctx, ac, "ecs."+actionName, inst.InstanceNo, ip, requestID, 1, nil)
	s.opLog(ac, "ecs", actionName, inst.InstanceNo, inst.Name, ip, 1, time.Since(start).Milliseconds())
	return nil
}

func (s *EcsService) Start(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	return s.action(ctx, ac, id, true, model.EcsStarting,
		func(ctx context.Context, inst *model.EcsInstance) error {
			return s.compute.Start(ctx, inst.ContainerID)
		}, "start", ip, requestID)
}

func (s *EcsService) Stop(ctx context.Context, ac AccessCtx, id uint64, force bool, ip, requestID string) error {
	name := "stop"
	if force {
		name = "force-stop"
	}
	return s.action(ctx, ac, id, true, model.EcsStopping,
		func(ctx context.Context, inst *model.EcsInstance) error {
			return s.compute.Stop(ctx, inst.ContainerID, force)
		}, name, ip, requestID)
}

func (s *EcsService) Restart(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	return s.action(ctx, ac, id, true, model.EcsRestarting,
		func(ctx context.Context, inst *model.EcsInstance) error {
			return s.compute.Restart(ctx, inst.ContainerID)
		}, "restart", ip, requestID)
}

// ecsActionName 操作名转中文（通知标题用）。
func ecsActionName(action string) string {
	switch action {
	case "start":
		return "启动"
	case "stop":
		return "停止"
	case "force-stop":
		return "强制停止"
	case "restart":
		return "重启"
	default:
		return action
	}
}

// Delete 删除实例（Docker 强删 + DB 软删；镜像与卷保留）。
func (s *EcsService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return ErrNotFound
	}
	if err := s.accessCheck(ac, inst, false); err != nil {
		return err
	}
	inst.DesiredState = model.EcsDeleting
	_ = s.repo.EcsUpdateState(inst)
	s.event(inst.ID, "delete", "info", "删除实例", &ac.UserID, requestID)

	if inst.ContainerID != "" {
		if err := s.compute.Remove(ctx, inst.ContainerID, true); err != nil {
			// 容器移除失败不阻断删除流程，但记录事件与通知，避免容器静默残留
			s.log.Warn("remove container failed on delete", zap.String("instance", inst.InstanceNo), zap.Error(err))
			s.event(inst.ID, "delete", "warn", "容器移除失败（引擎侧）："+err.Error(), &ac.UserID, requestID)
		}
	}
	if err := s.repo.EcsSoftDelete(inst.ID); err != nil {
		return err
	}
	s.audit(ctx, ac, "ecs.delete", inst.InstanceNo, ip, requestID, 1, nil)
	s.opLog(ac, "ecs", "delete", inst.InstanceNo, inst.Name, ip, 1, 0)
	notify(s.repo, ac.UserID, "ecs", "实例已删除", "「"+inst.Name+"」（"+inst.InstanceNo+"）已删除，镜像与磁盘已保留", "/ecs")
	return nil
}

// ---------- 日志 / 监控 / 事件 ----------

func (s *EcsService) Logs(ctx context.Context, ac AccessCtx, id uint64, tail int) (string, error) {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return "", ErrNotFound
	}
	if err := s.accessCheck(ac, inst, true); err != nil {
		return "", err
	}
	if inst.ContainerID == "" {
		return "", ErrBadState
	}
	return s.compute.Logs(ctx, inst.ContainerID, tail)
}

func (s *EcsService) Stats(ctx context.Context, ac AccessCtx, id uint64) (*docker.Stats, error) {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.accessCheck(ac, inst, true); err != nil {
		return nil, err
	}
	if inst.ContainerID == "" {
		return nil, ErrBadState
	}
	st, err := s.compute.StatsOneShot(ctx, inst.ContainerID)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *EcsService) Events(ctx context.Context, ac AccessCtx, id uint64, limit int) ([]model.EcsEvent, error) {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.accessCheck(ac, inst, true); err != nil {
		return nil, err
	}
	return s.repo.EcsEventList(id, limit)
}

// ---------- Exec 透传（Web Terminal 底层） ----------

func (s *EcsService) ExecCreate(ctx context.Context, containerID string) (string, error) {
	return s.compute.ExecCreate(ctx, containerID)
}

func (s *EcsService) ExecAttach(ctx context.Context, containerID, execID string) (*docker.ExecSession, error) {
	return s.compute.ExecAttach(ctx, containerID, execID)
}

func (s *EcsService) ExecResize(ctx context.Context, execID string, cols, rows int) error {
	return s.compute.ExecResize(ctx, execID, cols, rows)
}

// ConsoleAccessCheck 控制台入口校验（属主 + 运行态），供 WS 与 REST 共用。
func (s *EcsService) ConsoleAccessCheck(ctx context.Context, ac AccessCtx, id uint64) (*model.EcsInstance, error) {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.accessCheck(ac, inst, true); err != nil {
		return nil, err
	}
	if inst.ContainerID == "" || inst.ObservedState != model.EcsRunning {
		return nil, errors.New("instance is not running")
	}
	return inst, nil
}

// ---------- Web Terminal 一次性令牌 ----------

const consoleTokenPrefix = "ws:token:"

// IssueConsoleToken 为实例签发一次性终端令牌（60s，绑定 user+instance），并返回实例号用于审计。
// 权限 ecs:console 已在中间件校验；此处再校验资源属主与运行态。
func (s *EcsService) IssueConsoleToken(ctx context.Context, ac AccessCtx, id uint64) (token, instanceNo string, err error) {
	inst, err := s.repo.EcsGetByID(id)
	if err != nil {
		return "", "", ErrNotFound
	}
	if err := s.accessCheck(ac, inst, true); err != nil {
		return "", "", err
	}
	if inst.ContainerID == "" {
		return "", "", ErrBadState
	}
	if inst.ObservedState != model.EcsRunning {
		return "", "", errors.New("instance is not running")
	}
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	token = hex.EncodeToString(b)
	key := consoleTokenPrefix + token
	if err := redisx.HSet(ctx, s.rdb, key, map[string]any{
		"user_id":     fmt.Sprintf("%d", ac.UserID),
		"instance_id": fmt.Sprintf("%d", inst.ID),
	}); err != nil {
		return "", "", err
	}
	if err := redisx.Expire(ctx, s.rdb, key, 60*time.Second); err != nil {
		return "", "", err
	}
	return token, inst.InstanceNo, nil
}

// ValidateConsoleToken 校验并消费一次性令牌（原子读取后立即删除，防止并发复用）。
func (s *EcsService) ValidateConsoleToken(ctx context.Context, token string) (userID, instanceID uint64, err error) {
	if token == "" {
		return 0, 0, errors.New("missing token")
	}
	key := consoleTokenPrefix + token
	// 原子 HGetAll + Del（Lua 脚本），消除并发窗口
	m, err := redisx.HGetAllAndDel(ctx, s.rdb, key)
	if err != nil || len(m) == 0 {
		return 0, 0, errors.New("invalid or expired token")
	}
	var uid, iid uint64
	if _, err := fmt.Sscanf(m["user_id"], "%d", &uid); err != nil {
		return 0, 0, errors.New("invalid token")
	}
	if _, err := fmt.Sscanf(m["instance_id"], "%d", &iid); err != nil {
		return 0, 0, errors.New("invalid token")
	}
	return uid, iid, nil
}

// ---------- 云磁盘挂载（Docker 限制：挂载变更需重建容器，数据卷保留） ----------

// AttachVolume 将云磁盘挂载到实例（重建容器实现）。
func (s *EcsService) AttachVolume(ctx context.Context, ac AccessCtx, instanceID, volumeID uint64, target string, ip, requestID string) error {
	inst, err := s.repo.EcsGetByID(instanceID)
	if err != nil {
		return ErrNotFound
	}
	if err := s.accessCheck(ac, inst, false); err != nil {
		return err
	}
	vol, err := s.repo.VolumeGetByID(volumeID)
	if err != nil {
		return errors.New("volume not found")
	}
	if !orgMatches(ac.OrgID, vol.OrgID) {
		return errors.New("volume not found")
	}
	mounts := parseMounts(inst)
	for _, m := range mounts {
		if m.VolumeName == vol.DockerName {
			return errors.New("volume already mounted")
		}
		if m.Target == target {
			return errors.New("mount target already used")
		}
	}
	mounts = append(mounts, docker.Mount{VolumeName: vol.DockerName, Target: target})
	if err := s.recreate(ctx, inst, mounts); err != nil {
		return err
	}
	s.event(inst.ID, "volume.attach", "info", "挂载云磁盘 "+vol.Name+" → "+target, &ac.UserID, requestID)
	s.audit(ctx, ac, "volume.attach", inst.InstanceNo, ip, requestID, 1, map[string]any{"volume": vol.Name, "target": target})
	return nil
}

// DetachVolume 卸载云磁盘（重建容器实现，数据保留在卷内）。
func (s *EcsService) DetachVolume(ctx context.Context, ac AccessCtx, instanceID, volumeID uint64, ip, requestID string) error {
	inst, err := s.repo.EcsGetByID(instanceID)
	if err != nil {
		return ErrNotFound
	}
	if err := s.accessCheck(ac, inst, false); err != nil {
		return err
	}
	vol, err := s.repo.VolumeGetByID(volumeID)
	if err != nil {
		return errors.New("volume not found")
	}
	if !orgMatches(ac.OrgID, vol.OrgID) {
		return errors.New("volume not found")
	}
	mounts := parseMounts(inst)
	kept := mounts[:0]
	found := false
	for _, m := range mounts {
		if m.VolumeName == vol.DockerName {
			found = true
			continue
		}
		kept = append(kept, m)
	}
	if !found {
		return errors.New("volume is not mounted")
	}
	if err := s.recreate(ctx, inst, kept); err != nil {
		return err
	}
	s.event(inst.ID, "volume.detach", "info", "卸载云磁盘 "+vol.Name, &ac.UserID, requestID)
	s.audit(ctx, ac, "volume.detach", inst.InstanceNo, ip, requestID, 1, map[string]any{"volume": vol.Name})
	return nil
}

// recreate 以相同规格重建容器（挂载变更用；端口/IP/环境变量等从 DB 规格还原）。
func (s *EcsService) recreate(ctx context.Context, inst *model.EcsInstance, mounts []docker.Mount) error {
	if inst.ContainerID == "" {
		return ErrBadState
	}
	// 1) 移除旧容器（卷数据保留）
	if err := s.compute.Remove(ctx, inst.ContainerID, true); err != nil {
		return fmt.Errorf("remove old container: %w", err)
	}
	// 2) 以相同规格重建
	mountsJSON, _ := json.Marshal(mounts)
	spec := docker.CreateSpec{
		Name:           inst.ContainerName,
		Image:          inst.Image,
		CPU:            inst.CPU,
		MemoryMB:       inst.MemoryMB,
		Env:            stringSlice(inst.Env),
		Cmd:            stringSlice(inst.Command),
		Ports:          parsePorts(inst),
		Mounts:         mounts,
		RestartPolicy:  inst.RestartPolicy,
		NetworkID:      inst.NetworkID,
		FixedIP:        inst.FixedIP,
		ReadonlyRootfs: inst.ReadonlyRootfs,
		Labels: map[string]string{
			"com.dxcloud.kind":        "ecs",
			"com.dxcloud.instance-id": inst.InstanceNo,
			"com.dxcloud.owner-id":    fmt.Sprintf("%d", inst.OwnerID),
			"com.dxcloud.org-id":      orgLabel(inst.OrgID),
		},
	}
	info, err := s.compute.Create(ctx, spec)
	if err != nil {
		inst.DesiredState = model.EcsFailed
		inst.ObservedState = model.EcsFailed
		inst.LastError = err.Error()
		inst.Mounts = string(mountsJSON)
		_ = s.repo.EcsUpdateState(inst)
		return err
	}
	inst.ContainerID = info.ID
	inst.ContainerName = info.Name
	inst.FixedIP = info.IP
	inst.Mounts = string(mountsJSON)
	inst.DesiredState = model.EcsRunning
	inst.ObservedState = model.EcsRunning
	inst.LastError = ""
	return s.repo.EcsUpdateState(inst)
}

func stringSlice(jsonArr string) []string {
	var out []string
	_ = json.Unmarshal([]byte(jsonArr), &out)
	return out
}

func orgLabel(orgID *uint64) string {
	if orgID == nil {
		return "0"
	}
	return fmt.Sprintf("%d", *orgID)
}

// ---------- 对账（Reconciler，简化版 Controller Loop） ----------

func stateFromDocker(info docker.Info) string {
	switch info.Status {
	case "running":
		return model.EcsRunning
	case "restarting":
		return model.EcsRestarting
	case "exited", "created":
		return model.EcsStopped
	case "removing", "dead", "paused":
		return model.EcsFailed
	default:
		return model.EcsUnknown
	}
}

// Reconcile 单轮对账：过渡态推进 + 稳态漂移检测 + 孤儿容器发现。
func (s *EcsService) Reconcile(ctx context.Context) {
	s.reconcileTransitions(ctx)
	s.reconcileDrift(ctx)
	s.reconcileOrphans(ctx)
}

func (s *EcsService) reconcileTransitions(ctx context.Context) {
	rows, err := s.repo.EcsListByStates([]string{
		model.EcsCreating, model.EcsStarting, model.EcsStopping, model.EcsRestarting, model.EcsDeleting,
	})
	if err != nil {
		s.log.Error("reconcile list failed", zap.Error(err))
		return
	}
	for i := range rows {
		inst := &rows[i]
		elapsed := time.Since(inst.UpdatedAt)
		if elapsed > transitionTimeout {
			inst.DesiredState = model.EcsFailed
			inst.ObservedState = model.EcsFailed
			inst.LastError = "transition timeout"
			_ = s.repo.EcsUpdateState(inst)
			s.event(inst.ID, "reconcile", "error", "状态转换超时，标记 Failed", nil, "")
			continue
		}
		if inst.ContainerID == "" {
			continue
		}
		info, err := s.compute.Inspect(ctx, inst.ContainerID)
		if err != nil {
			continue // 容器暂不可见，等下一轮/超时兜底
		}
		mapped := stateFromDocker(info)
		if mapped == model.EcsRunning || mapped == model.EcsStopped {
			inst.DesiredState = mapped
			inst.ObservedState = mapped
			_ = s.repo.EcsUpdateState(inst)
			s.event(inst.ID, "reconcile", "info", "状态转换完成："+mapped, nil, "")
		}
	}
}

func (s *EcsService) reconcileDrift(ctx context.Context) {
	rows, err := s.repo.EcsListByStates([]string{model.EcsRunning, model.EcsStopped, model.EcsRestarting})
	if err != nil {
		s.log.Error("reconcile list failed", zap.Error(err))
		return
	}
	for i := range rows {
		inst := &rows[i]
		if inst.ContainerID == "" {
			continue
		}
		exists, err := s.compute.Exists(ctx, inst.ContainerID)
		if err != nil {
			continue
		}
		if !exists {
			inst.ObservedState = model.EcsUnknown
			inst.DesiredState = model.EcsUnknown
			inst.LastError = "container missing (removed outside platform?)"
			_ = s.repo.EcsUpdateState(inst)
			s.event(inst.ID, "reconcile", "error", "容器消失，状态置 Unknown", nil, "")
			continue
		}
		info, err := s.compute.Inspect(ctx, inst.ContainerID)
		if err != nil {
			continue
		}
		mapped := stateFromDocker(info)
		if mapped != inst.ObservedState {
			inst.ObservedState = mapped
			if inst.DesiredState == model.EcsRunning || inst.DesiredState == model.EcsStopped {
				inst.DesiredState = mapped
			}
			_ = s.repo.EcsUpdateState(inst)
			s.event(inst.ID, "reconcile", "warn", "漂移对账：DB="+inst.ObservedState+" Docker="+mapped, nil, "")
		}
	}
}

// reconcileOrphans 发现带平台标签但 DB 无记录的容器（Phase 3：告警日志；Phase 9 增强为孤儿管理）。
func (s *EcsService) reconcileOrphans(ctx context.Context) {
	list, err := s.compute.ListByLabels(ctx, map[string]string{"com.dxcloud.kind": "ecs"})
	if err != nil {
		s.log.Error("orphan scan failed", zap.Error(err))
		return
	}
	for _, c := range list {
		no := c.Labels["com.dxcloud.instance-id"]
		if no == "" {
			continue
		}
		if _, err := s.repo.EcsGetByNo(no); err == nil {
			continue
		}
		s.log.Warn("orphan container detected",
			zap.String("container_id", c.ID), zap.String("name", c.Name), zap.String("instance_no", no))
	}
}
