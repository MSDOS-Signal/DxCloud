// Package service：PaaS 应用/项目/部署（蓝绿 + 健康检查 + 回滚 + Traefik 域名路由）。
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"go.uber.org/zap"
)

var defaultEnvs = []string{"development", "testing", "staging", "production"}

// CreateAppReq 应用创建/更新请求。
type CreateAppReq struct {
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

// ---------- 项目服务 ----------

type ProjectService struct {
	repo   *repository.Repos
	iamSvc *iam.Service
}

func NewProjectService(repo *repository.Repos, iamSvc *iam.Service) *ProjectService {
	return &ProjectService{repo: repo, iamSvc: iamSvc}
}

func (s *ProjectService) List(ctx context.Context, ac AccessCtx) ([]model.Project, error) {
	return s.repo.ProjectListForContext(ac.OrgID)
}

func (s *ProjectService) Create(ctx context.Context, ac AccessCtx, name, code, description, ip, requestID string, orgID uint64) (*model.Project, error) {
	if name == "" || len(name) > 127 {
		return nil, errors.New("项目名需 1-127 位")
	}
	// 组织内项目名唯一（多租户隔离）；单租户模式 orgID=0 全局唯一。
	if _, err := s.repo.ProjectGetByNameOrg(name, orgID); err == nil {
		return nil, errors.New("project name already exists in this organization")
	}
	if code == "" {
		code = name
	}
	p := &model.Project{Name: name, Code: code, Description: description, Status: 1, CreatedBy: &ac.UserID, OrgID: orgID}
	if err := s.repo.ProjectCreate(p); err != nil {
		return nil, err
	}
	for i, env := range defaultEnvs {
		_ = s.repo.EnvCreate(&model.ProjectEnvironment{ProjectID: p.ID, Name: env, Seq: i + 1})
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "project.create", "project", name, ip, requestID, 1, nil)
	return p, nil
}

func (s *ProjectService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	p, err := s.repo.ProjectGetByID(id)
	if err != nil {
		return err
	}
	if p.OrgID != ac.OrgID {
		return ErrForbidden
	}
	apps, err := s.repo.AppList(&id, "")
	if err == nil && len(apps) > 0 {
		return errors.New("项目下存在应用，请先删除应用")
	}
	if err := s.repo.ProjectSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "project.delete", "project", p.Name, ip, requestID, 1, nil)
	return nil
}

func (s *ProjectService) Environments(ac AccessCtx, projectID uint64) ([]model.ProjectEnvironment, error) {
	p, err := s.repo.ProjectGetByID(projectID)
	if err != nil {
		return nil, err
	}
	if p.OrgID != ac.OrgID {
		return nil, ErrForbidden
	}
	return s.repo.EnvList(projectID)
}

// ---------- 应用 + 部署服务（蓝绿） ----------

type AppService struct {
	repo       *repository.Repos
	compute    docker.Provider
	iamSvc     *iam.Service
	appNetwork string // 应用容器网络名（如 dxcloud_edge）
	netID      string // 解析后的 Docker 网络 ID（懒加载缓存）
	log        *zap.Logger
}

func NewAppService(repo *repository.Repos, compute docker.Provider, iamSvc *iam.Service, appNetwork string, log *zap.Logger) *AppService {
	return &AppService{repo: repo, compute: compute, iamSvc: iamSvc, appNetwork: appNetwork, log: log}
}

// resolveNetwork 解析应用网络名 → Docker 网络 ID（懒加载）。
func (s *AppService) resolveNetwork(ctx context.Context) string {
	if s.netID != "" {
		return s.netID
	}
	if s.appNetwork == "" {
		return ""
	}
	networks, err := s.compute.ListNetworks(ctx)
	if err != nil {
		return ""
	}
	for _, n := range networks {
		if n.Name == s.appNetwork {
			s.netID = n.ID
			return n.ID
		}
	}
	return ""
}

func (s *AppService) List(ctx context.Context, ac AccessCtx, projectID *uint64, keyword string) ([]model.Application, error) {
	if ac.OrgID > 0 {
		return s.repo.AppListOrg(ac.OrgID, projectID, keyword)
	}
	return s.repo.AppListOrg(0, projectID, keyword)
}

// RepoGetApp 供 handler 读取应用详情。
func (s *AppService) RepoGetApp(id uint64) (*model.Application, error) {
	return s.repo.AppGetByID(id)
}

// appAccessible 组织归属校验（多租户隔离）：应用 org 或所属项目 org 命中才放行。
func (s *AppService) appAccessible(ac AccessCtx, app *model.Application) error {
	if ac.OrgID == 0 {
		return nil // 单租户兼容模式
	}
	oid := uint64(0)
	if app.OrgID != nil {
		oid = *app.OrgID
	}
	if oid == ac.OrgID {
		return nil
	}
	if app.ProjectID != nil {
		if p, err := s.repo.ProjectGetByID(*app.ProjectID); err == nil && p.OrgID == ac.OrgID {
			return nil
		}
	}
	return ErrForbidden
}

// Get 应用详情（带组织隔离校验）。
func (s *AppService) Get(ctx context.Context, ac AccessCtx, id uint64) (*model.Application, error) {
	app, err := s.repo.AppGetByID(id)
	if err != nil {
		return nil, err
	}
	if err := s.appAccessible(ac, app); err != nil {
		return nil, err
	}
	return app, nil
}

func (s *AppService) Create(ctx context.Context, ac AccessCtx, req CreateAppReq, ip, requestID string, orgID uint64) (*model.Application, error) {
	if req.Name == "" || len(req.Name) > 63 {
		return nil, errors.New("应用名需 1-63 位")
	}
	if req.Port <= 0 || req.Port > 65535 {
		req.Port = 80
	}
	// 项目归属校验：租户上下文下不允许引用其它组织的项目
	if ac.OrgID > 0 && req.ProjectID != 0 {
		p, err := s.repo.ProjectGetByID(req.ProjectID)
		if err != nil || p.OrgID != ac.OrgID {
			return nil, errors.New("项目不存在或不属于当前组织")
		}
	}
	if ac.OrgID == 0 && req.ProjectID != 0 {
		p, err := s.repo.ProjectGetByID(req.ProjectID)
		if err != nil || p.OrgID != 0 {
			return nil, errors.New("项目不存在或不属于当前组织")
		}
	}
	envJSON, _ := json.Marshal(req.Env)
	app := &model.Application{
		OwnerID: ac.UserID, Name: req.Name, Type: req.Type, Image: req.Image,
		GitURL: req.GitURL, GitBranch: req.GitBranch, Port: req.Port,
		HealthCheckPath: req.HealthCheckPath, Env: string(envJSON),
		Domain: req.Domain, Status: 1,
	}
	if orgID > 0 {
		app.OrgID = &orgID
	}
	if req.ProjectID != 0 {
		v := req.ProjectID
		app.ProjectID = &v
	}
	if ac.OrgID == 0 {
		app.OrgID = nil
	}
	if err := s.repo.AppCreate(app); err != nil {
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "app.create", "app", req.Name, ip, requestID, 1, nil)
	return app, nil
}

func (s *AppService) Update(ctx context.Context, ac AccessCtx, id uint64, req CreateAppReq, ip, requestID string) error {
	app, err := s.repo.AppGetByID(id)
	if err != nil {
		return err
	}
	if err := s.appAccessible(ac, app); err != nil {
		return err
	}
	if req.ProjectID != 0 {
		p, err := s.repo.ProjectGetByID(req.ProjectID)
		if err != nil || p.OrgID != ac.OrgID {
			return errors.New("项目不存在或不属于当前组织")
		}
	}
	if req.Name != "" {
		app.Name = req.Name
	}
	app.Type = req.Type
	app.Image = req.Image
	app.GitURL = req.GitURL
	app.GitBranch = req.GitBranch
	app.Port = req.Port
	app.HealthCheckPath = req.HealthCheckPath
	envJSON, _ := json.Marshal(req.Env)
	app.Env = string(envJSON)
	if err := s.repo.AppUpdate(app); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "app.update", "app", app.Name, ip, requestID, 1, nil)
	return nil
}

func (s *AppService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	app, err := s.repo.AppGetByID(id)
	if err != nil {
		return err
	}
	if err := s.appAccessible(ac, app); err != nil {
		return err
	}
	// 清理该应用全部容器（取全部部署记录，不受 limit 限制）
	deps, err := s.repo.DeploymentListAllByApp(id)
	if err != nil {
		return fmt.Errorf("list deployments for cleanup: %w", err)
	}
	for i := range deps {
		if deps[i].ContainerID != "" {
			_ = s.compute.Remove(ctx, deps[i].ContainerID, true)
		}
	}
	if err := s.repo.AppSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "app.delete", "app", app.Name, ip, requestID, 1, nil)
	notify(s.repo, ac.UserID, "deploy", "应用已删除", "「"+app.Name+"」及其全部部署容器已清理", "/apps")
	return nil
}

// ---------- 部署（蓝绿 + 健康检查 + Traefik 优先级切换） ----------

type DeployReq struct {
	Image    string            // 覆盖应用默认镜像
	Env      map[string]string // 环境变量覆盖
	HostPort int               // 可选：直接发布宿主端口（无域名时）
	Note     string
	Trigger  string // manual/webhook/pipeline
}

// Deploy 执行一次蓝绿部署：
//  1. 校验镜像存在 → 2) 创建候选容器（router 优先级 10）→ 3) 健康检查（HTTP/TCP，60s）
//  4. 通过：旧容器重建为优先级 0（零中断切换）并停止 → Success
//  5. 失败：清理候选容器 → Failed
func (s *AppService) Deploy(ctx context.Context, ac AccessCtx, appID uint64, req DeployReq, ip, requestID string) (*model.Deployment, error) {
	app, err := s.repo.AppGetByID(appID)
	if err != nil {
		return nil, ErrNotFound
	}
	if err := s.appAccessible(ac, app); err != nil {
		return nil, err
	}
	imageRef := req.Image
	if imageRef == "" {
		imageRef = app.Image
	}
	if imageRef == "" {
		return nil, errors.New("镜像不能为空")
	}
	// 镜像必须已在本机引擎（镜像中心拉取）
	if _, err := s.compute.InspectImage(ctx, imageRef); err != nil {
		return nil, fmt.Errorf("镜像不存在：%s（请先在镜像中心拉取）", imageRef)
	}
	if req.Trigger == "" {
		req.Trigger = "manual"
	}

	// 版本号：镜像 tag + 部署序号（同镜像重复部署也唯一）
	version := ""
	if _, tag := splitImageRef(imageRef); tag != "latest" && tag != "" {
		version = tag
	} else {
		version = fmt.Sprintf("r%d", time.Now().Unix())
	}

	// 部署记录
	now := time.Now()
	d := &model.Deployment{
		ApplicationID: app.ID, ProjectID: app.ProjectID,
		Version: version, ImageRef: imageRef, Strategy: "blue-green",
		Status: model.DeployDeploying, Trigger: req.Trigger, Note: req.Note, StartedAt: &now,
	}
	if err := s.repo.DeploymentCreate(d); err != nil {
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "deploy.start", "app", app.Name, ip, requestID, 1, map[string]any{"deployment": d.ID, "image": imageRef})

	// 候选容器规格
	env := mergeEnv(app.Env, req.Env)
	containerName := fmt.Sprintf("dx-app-%d-v%d", app.ID, d.ID)
	labels := map[string]string{
		"com.dxcloud.kind":      "app",
		"com.dxcloud.app-id":    fmt.Sprintf("%d", app.ID),
		"com.dxcloud.deploy-id": fmt.Sprintf("%d", d.ID),
	}
	ports := []docker.PortMapping{}
	if req.HostPort > 0 {
		ports = append(ports, docker.PortMapping{ContainerPort: app.Port, HostPort: req.HostPort, Protocol: "tcp"})
	}
	routerKey := fmt.Sprintf("app%d-v%d", app.ID, d.ID)
	if app.Domain != "" {
		labels["traefik.enable"] = "true"
		labels["traefik.http.routers."+routerKey+".rule"] = "Host(`" + app.Domain + "`)"
		labels["traefik.http.routers."+routerKey+".entrypoints"] = "web"
		labels["traefik.http.routers."+routerKey+".priority"] = "20"
		labels["traefik.http.routers."+routerKey+".service"] = routerKey
		labels["traefik.http.services."+routerKey+".loadbalancer.server.port"] = strconv.Itoa(app.Port)
	}

	spec := docker.CreateSpec{
		Name:     containerName,
		Image:    imageRef,
		CPU:      1,
		MemoryMB: 512,
		Env:      env,
		Ports:    ports,
		Labels:   labels,
		// 应用容器统一加入平台应用网络（自定义 bridge → IPAM 分配 IP：健康检查可探测、Traefik 可路由）
		NetworkID: s.resolveNetwork(ctx),
	}

	// 快照配置（用于降级重建）
	cfgJSON, _ := json.Marshal(map[string]any{
		"name": containerName, "image": imageRef, "env": env, "ports": ports,
		"labels": labels, "router_key": routerKey, "port": app.Port,
	})
	d.ConfigJSON = string(cfgJSON)

	// 创建 + 启动候选
	info, err := s.compute.Create(ctx, spec)
	if err != nil {
		d.Status = model.DeployFailed
		d.Note = err.Error()
		_ = s.repo.DeploymentUpdate(d)
		notify(s.repo, ac.UserID, "deploy", "应用部署失败", "「"+app.Name+"」部署失败："+err.Error(), "/apps/"+strconv.FormatUint(app.ID, 10))
		return d, err
	}
	d.ContainerID = info.ID
	d.ContainerName = info.Name

	// 健康检查（容器内网 IP + 端口）
	healthy := s.waitHealthy(ctx, info.ID, app.Port, app.HealthCheckPath, 60*time.Second)
	if !healthy {
		_ = s.compute.Remove(ctx, info.ID, true)
		d.Status = model.DeployFailed
		d.HealthStatus = "unhealthy"
		d.Note = "健康检查未通过"
		fin := time.Now()
		d.FinishedAt = &fin
		_ = s.repo.DeploymentUpdate(d)
		s.iamSvc.Audit(ctx, &uid, "deploy.failed", "app", app.Name, ip, requestID, 2, map[string]any{"deployment": d.ID})
		notify(s.repo, ac.UserID, "deploy", "应用部署失败", "「"+app.Name+"」健康检查未通过，部署已回退", "/apps/"+strconv.FormatUint(app.ID, 10))
		return d, errors.New("health check failed")
	}
	d.HealthStatus = "healthy"

	// 切换：旧容器降级（重建为优先级 0）+ 停止；有 host_port 的旧容器直接停止（端口切换，秒级中断）
	if old, err := s.repo.DeploymentActive(app.ID); err == nil && old.ID != d.ID && old.ContainerID != "" {
		s.demoteOld(ctx, old, routerKey, app)
	}

	// 版本记录（version 附带部署 ID 保证唯一）
	ver := &model.AppVersion{ApplicationID: app.ID, Version: fmt.Sprintf("%s.%d", version, d.ID), ImageRef: imageRef, Status: "active"}
	if err := s.repo.VersionCreate(ver); err == nil {
		d.VersionID = &ver.ID
	}
	d.Status = model.DeploySuccess
	fin := time.Now()
	d.FinishedAt = &fin
	_ = s.repo.DeploymentUpdate(d)

	app.ActiveDeploymentID = &d.ID
	if app.Domain == "" && req.HostPort > 0 {
		// 无域名应用：记录端口信息在 note
	}
	_ = s.repo.AppUpdate(app)

	s.iamSvc.Audit(ctx, &uid, "deploy.success", "app", app.Name, ip, requestID, 1,
		map[string]any{"deployment": d.ID, "version": version, "image": imageRef})
	if req.Trigger == "pipeline" {
		notify(s.repo, ac.UserID, "deploy", "Pipeline 部署成功", "「"+app.Name+"」已通过流水线自动部署，版本 "+version+"（镜像 "+imageRef+"）", "/apps/"+strconv.FormatUint(app.ID, 10))
	} else {
		notify(s.repo, ac.UserID, "deploy", "应用部署成功", "「"+app.Name+"」部署成功，版本 "+version+"（镜像 "+imageRef+"）", "/apps/"+strconv.FormatUint(app.ID, 10))
	}
	return d, nil
}

// demoteOld 将旧版本容器降级：domain 应用重建为 priority 0（流量切换，零中断）；
// host_port 应用直接停止并释放端口（秒级中断，架构已知限制）。
func (s *AppService) demoteOld(ctx context.Context, old *model.Deployment, newRouterKey string, app *model.Application) {
	if old.ConfigJSON == "" {
		_ = s.compute.Stop(ctx, old.ContainerID, false)
		return
	}
	var cfg struct {
		Name      string
		Image     string
		Env       []string
		Ports     []docker.PortMapping
		Labels    map[string]string
		RouterKey string
		Port      int
	}
	if err := json.Unmarshal([]byte(old.ConfigJSON), &cfg); err != nil {
		_ = s.compute.Stop(ctx, old.ContainerID, false)
		return
	}
	// 重建旧容器：优先级 0（不再承接流量）并停止
	_ = s.compute.Remove(ctx, old.ContainerID, true)
	if cfg.RouterKey != "" {
		if v, ok := cfg.Labels["traefik.http.routers."+cfg.RouterKey+".priority"]; ok && v == "20" {
			cfg.Labels["traefik.http.routers."+cfg.RouterKey+".priority"] = "0"
		}
	}
	spec := docker.CreateSpec{
		Name: cfg.Name, Image: cfg.Image, CPU: 1, MemoryMB: 512,
		Env: cfg.Env, Ports: cfg.Ports, Labels: cfg.Labels,
		NetworkID: s.resolveNetwork(ctx),
	}
	info, err := s.compute.Create(ctx, spec)
	if err == nil {
		_ = s.compute.Stop(ctx, info.ID, false)
		old.ContainerID = info.ID
		_ = s.repo.DeploymentUpdate(old)
	}
}

// waitHealthy HTTP/TCP 健康检查（2s 间隔）。
func (s *AppService) waitHealthy(ctx context.Context, containerID string, port int, path string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		info, err := s.compute.Inspect(ctx, containerID)
		if err != nil {
			return false
		}
		if info.IP == "" {
			time.Sleep(2 * time.Second)
			continue
		}
		if ok := probeOnce(info.IP, port, path); ok {
			return true
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func probeOnce(ip string, port int, path string) bool {
	addr := net.JoinHostPort(ip, strconv.Itoa(port))
	client := &http.Client{Timeout: 2 * time.Second}
	if path != "" {
		url := "http://" + addr + path
		resp, err := client.Get(url)
		if err != nil {
			return false
		}
		resp.Body.Close()
		// 2xx/3xx 视为健康（404 等说明路径错误，判不健康）
		return resp.StatusCode < 400
	}
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func mergeEnv(appEnvJSON string, overrides map[string]string) []string {
	base := map[string]string{}
	if appEnvJSON != "" {
		var arr []string
		if json.Unmarshal([]byte(appEnvJSON), &arr) == nil {
			for _, kv := range arr {
				for i := 0; i < len(kv); i++ {
					if kv[i] == '=' {
						base[kv[:i]] = kv[i+1:]
						break
					}
				}
			}
		}
	}
	for k, v := range overrides {
		base[k] = v
	}
	out := make([]string, 0, len(base))
	for k, v := range base {
		out = append(out, k+"="+v)
	}
	return out
}

// DeployByName 按应用名部署（Pipeline docker-deploy 步骤调用）。
func (s *AppService) DeployByName(ctx context.Context, ac AccessCtx, appName, imageRef, note string) (*model.Deployment, error) {
	app, err := s.repo.AppGetByNameOrg(appName, ac.OrgID)
	if err != nil {
		return nil, fmt.Errorf("application %q not found", appName)
	}
	return s.Deploy(ctx, ac, app.ID, DeployReq{Image: imageRef, Note: note, Trigger: "pipeline"}, "", "")
}

// Rollback 用历史版本镜像重新执行部署（回滚 = 再部署一次旧版本）。
func (s *AppService) Rollback(ctx context.Context, ac AccessCtx, appID, versionID uint64, ip, requestID string) (*model.Deployment, error) {
	if _, err := s.repo.AppGetByID(appID); err != nil {
		return nil, ErrNotFound
	}
	ver, err := s.repo.VersionGetByID(versionID)
	if err != nil || ver.ApplicationID != appID {
		return nil, errors.New("version not found")
	}
	return s.Deploy(ctx, ac, appID, DeployReq{
		Image: ver.ImageRef, Note: "rollback to " + ver.Version, Trigger: "manual",
	}, ip, requestID)
}

// Versions / Deployments 查询。
func (s *AppService) Versions(appID uint64) ([]model.AppVersion, error) {
	return s.repo.VersionList(appID)
}

func (s *AppService) VersionsFor(ac AccessCtx, appID uint64) ([]model.AppVersion, error) {
	app, err := s.repo.AppGetByID(appID)
	if err != nil {
		return nil, err
	}
	if err := s.appAccessible(ac, app); err != nil {
		return nil, err
	}
	return s.repo.VersionList(appID)
}

func (s *AppService) Deployments(appID uint64, limit int) ([]model.Deployment, error) {
	return s.repo.DeploymentListByApp(appID, limit)
}

func (s *AppService) DeploymentsFor(ac AccessCtx, appID uint64, limit int) ([]model.Deployment, error) {
	app, err := s.repo.AppGetByID(appID)
	if err != nil {
		return nil, err
	}
	if err := s.appAccessible(ac, app); err != nil {
		return nil, err
	}
	return s.repo.DeploymentListByApp(appID, limit)
}

// ---------- 域名服务 ----------

type DomainService struct {
	repo   *repository.Repos
	iamSvc *iam.Service
}

func NewDomainService(repo *repository.Repos, iamSvc *iam.Service) *DomainService {
	return &DomainService{repo: repo, iamSvc: iamSvc}
}

func (s *DomainService) appAccessible(ac AccessCtx, app *model.Application) error {
	if ac.OrgID == 0 {
		return nil
	}
	if app.OrgID != nil && *app.OrgID == ac.OrgID {
		return nil
	}
	if app.ProjectID != nil {
		if p, err := s.repo.ProjectGetByID(*app.ProjectID); err == nil && p.OrgID == ac.OrgID {
			return nil
		}
	}
	return ErrForbidden
}

func (s *DomainService) List(ac AccessCtx) ([]model.Domain, error) {
	return s.repo.DomainListForContext(ac.OrgID)
}

// Bind 绑定域名到应用：写入 domains 表并同步应用的 domain 字段（部署时生成 Traefik 路由）。
func (s *DomainService) Bind(ctx context.Context, ac AccessCtx, domain string, appID, targetPort uint64, ip, requestID string) (*model.Domain, error) {
	if domain == "" || len(domain) > 253 {
		return nil, errors.New("域名非法")
	}
	if _, err := s.repo.DomainGetByName(domain); err == nil {
		return nil, errors.New("域名已被占用")
	}
	var appIDPtr *uint64
	if appID > 0 {
		appIDPtr = &appID
		app, err := s.repo.AppGetByID(appID)
		if err != nil {
			return nil, errors.New("应用不存在")
		}
		if err := s.appAccessible(ac, app); err != nil {
			return nil, ErrForbidden
		}
		app.Domain = domain
		_ = s.repo.AppUpdate(app)
	}
	port := targetPort
	if port == 0 {
		port = 80
	}
	d := &model.Domain{Domain: domain, ApplicationID: appIDPtr, TargetPort: int(port), Status: 1}
	if ac.OrgID > 0 {
		orgID := ac.OrgID
		d.OrgID = &orgID
	}
	if err := s.repo.DomainCreate(d); err != nil {
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "domain.bind", "domain", domain, ip, requestID, 1, map[string]any{"app": appID})
	return d, nil
}

func (s *DomainService) Unbind(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	d, err := s.repo.DomainGetByID(id)
	if err != nil {
		return err
	}
	if d.OrgID != nil && *d.OrgID != ac.OrgID {
		return ErrForbidden
	}
	if d.OrgID == nil && ac.OrgID != 0 {
		return ErrForbidden
	}
	if d.ApplicationID != nil {
		if app, err := s.repo.AppGetByID(*d.ApplicationID); err == nil && app.Domain == d.Domain {
			app.Domain = ""
			_ = s.repo.AppUpdate(app)
		}
	}
	if err := s.repo.DomainSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "domain.unbind", "domain", d.Domain, ip, requestID, 1, nil)
	return nil
}
