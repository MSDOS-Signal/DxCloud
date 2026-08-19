// Package api 负责路由注册与中间件装配。
package api

import (
	"strings"
	"time"

	"github.com/dxcloud/cloud-api/internal/config"
	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/handler"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/pipeline"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/runner"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/internal/websocket"
	"github.com/dxcloud/cloud-api/pkg/crypto"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/ratelimit"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func NewRouter(cfg *config.Config, log *zap.Logger, db *gorm.DB, rdb *redis.Client, compute docker.Provider) (*gin.Engine, *pipeline.Engine, func()) {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.AccessLog(log),
		middleware.Recovery(log),
		middleware.CORS(),
	)

	repos := repository.New(db)
	iamSvc := iam.NewService(cfg, log, db, rdb, repos)
	limiter := ratelimit.New(rdb)

	settingsSvc := service.NewSettingsService(repos, log)
	aiSvc := service.NewAIService(cfg, log)
	imageSvc := service.NewImageService(repos, compute, iamSvc, settingsSvc, log)
	networkSvc := service.NewNetworkService(repos, compute, iamSvc, log)
	volumeSvc := service.NewVolumeService(repos, compute, iamSvc, log)
	registrySvc := service.NewRegistryService(repos, compute, iamSvc, cfg.RegistryEngineURL, log)
	monitorSvc := service.NewMonitorService(repos)
	projectSvc := service.NewProjectService(repos, iamSvc)
	appSvc := service.NewAppService(repos, compute, iamSvc, cfg.AppNetwork, log)
	domainSvc := service.NewDomainService(repos, iamSvc)
	orgSvc := service.NewOrgService(repos, iamSvc)
	quotaSvc := service.NewQuotaService(repos)
	billingSvc := service.NewBillingService(repos, log)

	jobRunner := runner.NewDockerJobRunner(compute, log)
	pipeEngine := pipeline.NewEngine(repos, jobRunner, compute, rdb, iamSvc, cfg.AppNetwork, 2, log)
	pipeEngine.SetDeployer(appSvc)
	pipeEngine.SetRegistryEngineURL(cfg.RegistryEngineURL)
	pipeEngine.SetImageResolver(settingsSvc)

	ecsSvc := service.NewEcsService(repos, compute, iamSvc, rdb, quotaSvc, billingSvc, log)

	health := handler.NewHealthHandler(db, rdb, compute)
	authH := handler.NewAuthHandler(iamSvc)
	userH := handler.NewUserHandler(iamSvc)
	roleH := handler.NewRoleHandler(iamSvc)
	permH := handler.NewPermissionHandler(iamSvc)
	ecsH := handler.NewEcsHandler(ecsSvc, iamSvc)
	termH := websocket.NewTerminalHandler(ecsSvc, iamSvc, log)
	imageH := handler.NewImageHandler(imageSvc, iamSvc)
	networkH := handler.NewNetworkHandler(networkSvc, iamSvc)
	volumeH := handler.NewVolumeHandler(volumeSvc, iamSvc)
	registryH := handler.NewRegistryHandler(registrySvc, iamSvc)
	projectH := handler.NewProjectHandler(projectSvc, iamSvc)
	appH := handler.NewAppHandler(appSvc, iamSvc)
	domainH := handler.NewDomainHandler(domainSvc, iamSvc)
	pipeH := handler.NewPipelineHandler(pipeEngine, repos, iamSvc)
	whH := handler.NewWebhookHandler(repos, iamSvc, pipeEngine, crypto.NewKey(cfg.JWT.Secret))
	monH := handler.NewMonitorHandler(monitorSvc, repos, iamSvc)
	orgH := handler.NewOrgHandler(orgSvc, iamSvc)
	quotaH := handler.NewQuotaHandler(quotaSvc, iamSvc)
	billingH := handler.NewBillingHandler(billingSvc, iamSvc)
	securitySvc := service.NewSecurityService(repos, compute, compute, log)
	secretSvc := service.NewSecretService(repos, crypto.NewKey(cfg.JWT.Secret), log)
	securityH := handler.NewSecurityHandler(securitySvc, iamSvc)
	secretH := handler.NewSecretHandler(secretSvc, iamSvc)
	settingsH := handler.NewSettingsHandler(settingsSvc, iamSvc)
	aiH := handler.NewAIHandler(aiSvc, iamSvc)

	// 运维探针：GET /healthz
	r.GET("/healthz", health.Livez)

	v1 := r.Group("/api/v1")
	v1.GET("/health", health.Health)

	// ---------- 认证 ----------
	auth := v1.Group("/auth")
	auth.POST("/register", middleware.RateLimit(limiter, 20, time.Minute), authH.Register)
	auth.POST("/login", middleware.RateLimit(limiter, 5, time.Minute), authH.Login)
	auth.POST("/refresh", middleware.RateLimit(limiter, 60, time.Minute), authH.Refresh)

	authSec := auth.Group("", middleware.AuthRequired(iamSvc))
	authSec.POST("/logout", authH.Logout)
	authSec.GET("/me", authH.Me)
	authSec.PUT("/profile", authH.UpdateProfile)
	authSec.PUT("/password", authH.ChangePassword)
	authSec.GET("/sessions", authH.Sessions)
	authSec.DELETE("/sessions/:id", authH.DeleteSession)

	// ---------- IAM 管理（权限点在中间件强制，前端隐藏按钮只是体验） ----------
	iamGroup := v1.Group("", middleware.AuthRequired(iamSvc))
	iamGroup.GET("/users", middleware.RequirePerm(iamSvc, "user:list"), userH.List)
	iamGroup.GET("/users/:id", middleware.RequirePerm(iamSvc, "user:list"), userH.Get)
	iamGroup.POST("/users", middleware.RequirePerm(iamSvc, "user:create"), userH.Create)
	iamGroup.PUT("/users/:id", middleware.RequirePerm(iamSvc, "user:update"), userH.Update)
	iamGroup.DELETE("/users/:id", middleware.RequirePerm(iamSvc, "user:delete"), userH.Delete)
	iamGroup.PUT("/users/:id/roles", middleware.RequirePerm(iamSvc, "user:grant"), userH.GrantRoles)

	iamGroup.GET("/roles", middleware.RequirePerm(iamSvc, "user:list"), roleH.List)
	iamGroup.POST("/roles", middleware.RequirePerm(iamSvc, "user:create"), roleH.Create)
	iamGroup.PUT("/roles/:id", middleware.RequirePerm(iamSvc, "user:update"), roleH.Update)
	iamGroup.DELETE("/roles/:id", middleware.RequirePerm(iamSvc, "user:delete"), roleH.Delete)
	iamGroup.PUT("/roles/:id/permissions", middleware.RequirePerm(iamSvc, "user:grant"), roleH.GrantPermissions)

	iamGroup.GET("/permissions", middleware.RequirePerm(iamSvc, "user:list"), permH.List)

	// ---------- ECS（RBAC + 属主校验；底层 Docker 由 Provider 封装，前端零接触） ----------
	ecs := v1.Group("/ecs", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	ecs.GET("", middleware.RequirePerm(iamSvc, "ecs:list"), ecsH.List)
	ecs.POST("", middleware.RequirePerm(iamSvc, "ecs:create"), ecsH.Create)
	ecs.GET("/:id", middleware.RequirePerm(iamSvc, "ecs:get"), ecsH.Get)
	ecs.PUT("/:id", middleware.RequirePerm(iamSvc, "ecs:update"), ecsH.Update)
	ecs.DELETE("/:id", middleware.RequirePerm(iamSvc, "ecs:delete"), ecsH.Delete)
	ecs.POST("/:id/start", middleware.RequirePerm(iamSvc, "ecs:start"), ecsH.Start)
	ecs.POST("/:id/stop", middleware.RequirePerm(iamSvc, "ecs:stop"), ecsH.Stop)
	ecs.POST("/:id/force-stop", middleware.RequirePerm(iamSvc, "ecs:force-stop"), ecsH.ForceStop)
	ecs.POST("/:id/restart", middleware.RequirePerm(iamSvc, "ecs:restart"), ecsH.Restart)
	ecs.GET("/:id/logs", middleware.RequirePerm(iamSvc, "ecs:logs"), ecsH.Logs)
	ecs.GET("/:id/stats", middleware.RequirePerm(iamSvc, "ecs:stats"), ecsH.Stats)
	ecs.GET("/:id/events", middleware.RequirePerm(iamSvc, "ecs:get"), ecsH.Events)
	ecs.POST("/:id/exec", middleware.RequirePerm(iamSvc, "ecs:console"), ecsH.Exec)

	// ---------- WebSocket（一次性令牌鉴权，不依赖 Cookie/JWT 头） ----------
	r.GET("/ws/v1/ecs/:id/terminal", termH.Handle)

	// ---------- 镜像中心 ----------
	img := v1.Group("/images", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	img.GET("", middleware.RequirePerm(iamSvc, "image:list"), imageH.List)
	img.GET("/search", middleware.RequirePerm(iamSvc, "image:list"), imageH.Search)
	img.POST("/pull", middleware.RequirePerm(iamSvc, "image:pull"), imageH.Pull)
	img.GET("/:id/logs", middleware.RequirePerm(iamSvc, "image:list"), imageH.PullLogs)
	img.DELETE("/:id", middleware.RequirePerm(iamSvc, "image:delete"), imageH.Delete)
	img.POST("/:id/tag", middleware.RequirePerm(iamSvc, "image:tag"), imageH.Tag)

	// ---------- 网络 ----------
	net := v1.Group("/networks", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	net.GET("", middleware.RequirePerm(iamSvc, "network:list"), networkH.List)
	net.POST("", middleware.RequirePerm(iamSvc, "network:create"), networkH.Create)
	net.GET("/:id", middleware.RequirePerm(iamSvc, "network:list"), networkH.Get)
	net.DELETE("/:id", middleware.RequirePerm(iamSvc, "network:delete"), networkH.Delete)
	net.POST("/:id/connect", middleware.RequirePerm(iamSvc, "network:connect"), networkH.Connect)
	net.POST("/:id/disconnect", middleware.RequirePerm(iamSvc, "network:connect"), networkH.Disconnect)

	// ---------- 存储 ----------
	vol := v1.Group("/volumes", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	vol.GET("", middleware.RequirePerm(iamSvc, "volume:list"), volumeH.List)
	vol.POST("", middleware.RequirePerm(iamSvc, "volume:create"), volumeH.Create)
	vol.DELETE("/:id", middleware.RequirePerm(iamSvc, "volume:delete"), volumeH.Delete)

	// ECS 挂载/卸载（云磁盘）
	ecs.POST("/:id/volumes/attach", middleware.RequirePerm(iamSvc, "volume:attach"), ecsH.AttachVolume)
	ecs.POST("/:id/volumes/detach", middleware.RequirePerm(iamSvc, "volume:attach"), ecsH.DetachVolume)

	// ---------- Registry ----------
	reg := v1.Group("/registries", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	reg.GET("", middleware.RequirePerm(iamSvc, "registry:list"), registryH.List)
	reg.GET("/:id/repositories", middleware.RequirePerm(iamSvc, "registry:list"), registryH.Repositories)
	reg.POST("/:id/repositories/pull", middleware.RequirePerm(iamSvc, "registry:pull"), registryH.Pull)
	reg.POST("/:id/repositories/delete-tag", middleware.RequirePerm(iamSvc, "registry:delete"), registryH.DeleteTag)

	// ---------- 项目 / 应用 / 部署 / 域名（PaaS） ----------
	proj := v1.Group("/projects", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	proj.GET("", middleware.RequirePerm(iamSvc, "project:list"), projectH.List)
	proj.POST("", middleware.RequirePerm(iamSvc, "project:create"), projectH.Create)
	proj.DELETE("/:id", middleware.RequirePerm(iamSvc, "project:delete"), projectH.Delete)
	proj.GET("/:id/environments", middleware.RequirePerm(iamSvc, "project:list"), projectH.Environments)

	apps := v1.Group("/applications", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	apps.GET("", middleware.RequirePerm(iamSvc, "app:list"), appH.List)
	apps.POST("", middleware.RequirePerm(iamSvc, "app:create"), appH.Create)
	apps.GET("/:id", middleware.RequirePerm(iamSvc, "app:list"), appH.Get)
	apps.PUT("/:id", middleware.RequirePerm(iamSvc, "app:update"), appH.Update)
	apps.DELETE("/:id", middleware.RequirePerm(iamSvc, "app:delete"), appH.Delete)
	apps.POST("/:id/deploy", middleware.RequirePerm(iamSvc, "app:deploy"), appH.Deploy)
	apps.GET("/:id/versions", middleware.RequirePerm(iamSvc, "app:list"), appH.Versions)
	apps.POST("/:id/versions/:vid/rollback", middleware.RequirePerm(iamSvc, "app:rollback"), appH.Rollback)
	apps.GET("/:id/deployments", middleware.RequirePerm(iamSvc, "app:list"), appH.Deployments)

	dom := v1.Group("/domains", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	dom.GET("", middleware.RequirePerm(iamSvc, "domain:list"), domainH.List)
	dom.POST("", middleware.RequirePerm(iamSvc, "domain:bind"), domainH.Bind)
	dom.DELETE("/:id", middleware.RequirePerm(iamSvc, "domain:delete"), domainH.Unbind)

	// ---------- Pipeline（队列 + 内嵌 Worker + 隔离 Job 容器） ----------
	pipes := v1.Group("/pipelines", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	pipes.GET("", middleware.RequirePerm(iamSvc, "pipeline:list"), pipeH.List)
	pipes.POST("", middleware.RequirePerm(iamSvc, "pipeline:create"), pipeH.Create)
	pipes.GET("/:id", middleware.RequirePerm(iamSvc, "pipeline:list"), pipeH.Get)
	pipes.PUT("/:id", middleware.RequirePerm(iamSvc, "pipeline:update"), pipeH.Update)
	pipes.DELETE("/:id", middleware.RequirePerm(iamSvc, "pipeline:delete"), pipeH.Delete)
	pipes.POST("/:id/run", middleware.RequirePerm(iamSvc, "pipeline:run"), pipeH.Run)

	pruns := v1.Group("/pipeline-runs", middleware.AuthRequired(iamSvc))
	pruns.GET("", middleware.RequirePerm(iamSvc, "pipeline:list"), pipeH.RunList)
	pruns.GET("/:id", middleware.RequirePerm(iamSvc, "pipeline:list"), pipeH.RunGet)
	pruns.GET("/:id/jobs", middleware.RequirePerm(iamSvc, "pipeline:list"), pipeH.RunJobs)
	pruns.GET("/:id/logs", middleware.RequirePerm(iamSvc, "pipeline:list"), pipeH.RunLogs)
	pruns.POST("/:id/cancel", middleware.RequirePerm(iamSvc, "pipeline:cancel"), pipeH.Cancel)

	// ---------- Webhook（管理 + 公开接收端：签名校验，不走 JWT） ----------
	wh := v1.Group("/webhooks", middleware.AuthRequired(iamSvc))
	wh.GET("", middleware.RequirePerm(iamSvc, "pipeline:list"), whH.List)
	wh.POST("", middleware.RequirePerm(iamSvc, "pipeline:create"), whH.Create)
	wh.DELETE("/:id", middleware.RequirePerm(iamSvc, "pipeline:delete"), whH.Delete)
	r.POST("/api/v1/webhooks/:provider/:code", middleware.RateLimit(limiter, 60, time.Minute), whH.Receive)

	// ---------- 监控 / 日志 / 通知（Phase 9） ----------
	mon := v1.Group("/monitor", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	mon.GET("/dashboard", middleware.RequirePerm(iamSvc, "ecs:list"), monH.Dashboard)
	mon.GET("/series", middleware.RequirePerm(iamSvc, "ecs:stats"), monH.Series)

	logs := v1.Group("/logs", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	logs.GET("", middleware.RequirePerm(iamSvc, "audit:view"), monH.Logs)

	notif := v1.Group("/notifications", middleware.AuthRequired(iamSvc))
	notif.GET("", monH.Notifications)
	notif.PUT("/:id/read", monH.NotificationRead)
	notif.POST("/read-all", monH.NotificationReadAll)

	// ---------- 组织 / 配额 / 计费（Phase 10 Multi-Tenant） ----------
	orgs := v1.Group("/organizations", middleware.AuthRequired(iamSvc))
	orgs.GET("", middleware.RequirePerm(iamSvc, "org:list"), orgH.List)
	orgs.GET("/mine", orgH.MyOrgs)
	orgs.POST("", middleware.RequirePerm(iamSvc, "org:create"), orgH.Create)
	orgs.DELETE("/:id", middleware.RequirePerm(iamSvc, "org:delete"), orgH.Delete)
	orgs.GET("/:id/members", middleware.RequirePerm(iamSvc, "org:list"), orgH.Members)
	orgs.POST("/:id/members", middleware.RequirePerm(iamSvc, "org:member:manage"), orgH.AddMember)
	orgs.DELETE("/:id/members/:uid", middleware.RequirePerm(iamSvc, "org:member:manage"), orgH.RemoveMember)

	quotas := v1.Group("/quotas", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	quotas.GET("", middleware.RequirePerm(iamSvc, "quota:view"), quotaH.List)
	quotas.PUT("", middleware.RequirePerm(iamSvc, "quota:update"), quotaH.Update)

	billing := v1.Group("/billing", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	billing.GET("", middleware.RequirePerm(iamSvc, "billing:view"), billingH.Summary)
	billing.GET("/records", middleware.RequirePerm(iamSvc, "billing:view"), billingH.Records)
	billing.POST("/tick", middleware.RequirePerm(iamSvc, "quota:update"), billingH.Tick)
	billing.POST("/recharge", middleware.RequirePerm(iamSvc, "quota:update"), billingH.Recharge)

	// ---------- 安全中心 / 密钥托管（Phase 11 Security Hardening） ----------
	sec := v1.Group("/security", middleware.AuthRequired(iamSvc))
	sec.GET("/dashboard", middleware.RequirePerm(iamSvc, "security:view"), securityH.Dashboard)
	sec.POST("/scan", middleware.RequirePerm(iamSvc, "security:scan"), securityH.Scan)
	sec.GET("/reports", middleware.RequirePerm(iamSvc, "security:view"), securityH.Reports)
	sec.GET("/reports/:id", middleware.RequirePerm(iamSvc, "security:view"), securityH.ReportFindings)

	secrets := v1.Group("/secrets", middleware.AuthRequired(iamSvc), middleware.TenantContext(iamSvc, orgSvc))
	secrets.GET("", middleware.RequirePerm(iamSvc, "secret:list"), secretH.List)
	secrets.POST("", middleware.RequirePerm(iamSvc, "secret:create"), secretH.Create)
	secrets.GET("/:id/reveal", middleware.RequirePerm(iamSvc, "secret:reveal"), secretH.Reveal)
	secrets.DELETE("/:id", middleware.RequirePerm(iamSvc, "secret:delete"), secretH.Delete)

	// ---------- 系统设置（区域 / 镜像加速源；读取对所有登录用户开放，修改需权限） ----------
	set := v1.Group("/settings", middleware.AuthRequired(iamSvc))
	set.GET("", settingsH.Get)
	set.PUT("", middleware.RequirePerm(iamSvc, "settings:update"), settingsH.Update)
	set.POST("/test-mirror", middleware.RequirePerm(iamSvc, "settings:update"), settingsH.TestMirror)

	// ---------- AI 助手（智谱 GLM 代理；登录即可用，限流防滥用） ----------
	ai := v1.Group("/ai", middleware.AuthRequired(iamSvc))
	ai.POST("/chat", middleware.RateLimit(limiter, 30, time.Minute), aiH.Chat)

	// 未实现的 API 端点：统一结构提示（后续 Phase 逐步替换为真实 handler）
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			resp.Fail(c, errcode.CodeNotImplemented, "接口尚未实现")
			return
		}
		resp.Fail(c, errcode.CodeNotFound, "页面或接口不存在")
	})
	r.NoMethod(func(c *gin.Context) {
		resp.Fail(c, errcode.CodeNotImplemented, "请求方式不允许")
	})

	return r, pipeEngine, imageSvc.Stop
}
