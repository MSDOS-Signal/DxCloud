package database

import (
	"errors"
	"os"

	"github.com/dxcloud/cloud-api/internal/config"
	"github.com/dxcloud/cloud-api/internal/model"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 全部权限点（Phase 2 种子；后续 Phase 只增不改，新增权限走新迁移/种子追加）。
// Name 为中文展示名（控制台权限列表/角色授权直接使用，避免用户看到 ecs:xx 这类原始码）。
var permissions = []model.Permission{
	{Code: "ecs:list", Name: "查看云主机列表", Module: "ecs"}, {Code: "ecs:get", Name: "查看云主机详情", Module: "ecs"},
	{Code: "ecs:create", Name: "创建云主机", Module: "ecs"}, {Code: "ecs:update", Name: "更新云主机配置", Module: "ecs"},
	{Code: "ecs:delete", Name: "删除云主机", Module: "ecs"}, {Code: "ecs:start", Name: "启动云主机", Module: "ecs"},
	{Code: "ecs:stop", Name: "停止云主机", Module: "ecs"}, {Code: "ecs:restart", Name: "重启云主机", Module: "ecs"},
	{Code: "ecs:force-stop", Name: "强制停止云主机", Module: "ecs"}, {Code: "ecs:console", Name: "连接远程终端", Module: "ecs"},
	{Code: "ecs:logs", Name: "查看云主机日志", Module: "ecs"}, {Code: "ecs:stats", Name: "查看云主机监控", Module: "ecs"},
	{Code: "ecs:recreate", Name: "重建云主机", Module: "ecs"},

	{Code: "image:list", Name: "查看镜像列表", Module: "image"}, {Code: "image:pull", Name: "拉取镜像", Module: "image"},
	{Code: "image:delete", Name: "删除镜像", Module: "image"}, {Code: "image:build", Name: "构建镜像", Module: "image"},
	{Code: "image:tag", Name: "镜像打标签", Module: "image"}, {Code: "image:push", Name: "推送镜像", Module: "image"},

	{Code: "network:list", Name: "查看网络列表", Module: "network"}, {Code: "network:create", Name: "创建网络", Module: "network"},
	{Code: "network:delete", Name: "删除网络", Module: "network"}, {Code: "network:connect", Name: "连接/断开网络", Module: "network"},

	{Code: "volume:list", Name: "查看存储列表", Module: "volume"}, {Code: "volume:create", Name: "创建云磁盘", Module: "volume"},
	{Code: "volume:delete", Name: "删除云磁盘", Module: "volume"}, {Code: "volume:attach", Name: "挂载/卸载云磁盘", Module: "volume"},

	{Code: "registry:list", Name: "查看镜像仓库", Module: "registry"}, {Code: "registry:create", Name: "添加镜像仓库", Module: "registry"},
	{Code: "registry:delete", Name: "管理镜像仓库", Module: "registry"}, {Code: "registry:push", Name: "推送至镜像仓库", Module: "registry"},
	{Code: "registry:pull", Name: "从镜像仓库拉取", Module: "registry"},

	{Code: "app:list", Name: "查看应用列表", Module: "app"}, {Code: "app:create", Name: "创建应用", Module: "app"},
	{Code: "app:update", Name: "更新应用", Module: "app"}, {Code: "app:delete", Name: "删除应用", Module: "app"},
	{Code: "app:deploy", Name: "部署应用", Module: "app"}, {Code: "app:rollback", Name: "回滚应用版本", Module: "app"},

	{Code: "pipeline:list", Name: "查看流水线", Module: "pipeline"}, {Code: "pipeline:create", Name: "创建流水线", Module: "pipeline"},
	{Code: "pipeline:update", Name: "更新流水线", Module: "pipeline"}, {Code: "pipeline:delete", Name: "删除流水线", Module: "pipeline"},
	{Code: "pipeline:run", Name: "运行流水线", Module: "pipeline"}, {Code: "pipeline:cancel", Name: "取消流水线", Module: "pipeline"},

	{Code: "project:list", Name: "查看项目列表", Module: "project"}, {Code: "project:create", Name: "创建项目", Module: "project"},
	{Code: "project:update", Name: "更新项目", Module: "project"}, {Code: "project:delete", Name: "删除项目", Module: "project"},
	{Code: "project:deploy", Name: "项目部署", Module: "project"},

	{Code: "domain:list", Name: "查看域名列表", Module: "domain"}, {Code: "domain:create", Name: "添加域名", Module: "domain"},
	{Code: "domain:delete", Name: "删除域名", Module: "domain"}, {Code: "domain:bind", Name: "绑定域名", Module: "domain"},

	{Code: "user:list", Name: "查看用户", Module: "user"}, {Code: "user:create", Name: "创建用户", Module: "user"},
	{Code: "user:update", Name: "更新用户", Module: "user"}, {Code: "user:delete", Name: "删除用户", Module: "user"},
	{Code: "user:grant", Name: "分配角色权限", Module: "user"},

	{Code: "org:list", Name: "查看组织", Module: "org"}, {Code: "org:create", Name: "创建组织", Module: "org"},
	{Code: "org:update", Name: "更新组织", Module: "org"}, {Code: "org:delete", Name: "删除组织", Module: "org"},
	{Code: "org:member:manage", Name: "管理组织成员", Module: "org"},

	{Code: "quota:view", Name: "查看资源配额", Module: "quota"}, {Code: "quota:update", Name: "管理资源配额", Module: "quota"},
	{Code: "billing:view", Name: "查看计费账单", Module: "billing"},
	{Code: "audit:view", Name: "查看审计日志", Module: "audit"},
	{Code: "settings:view", Name: "查看系统设置", Module: "settings"}, {Code: "settings:update", Name: "更新系统设置", Module: "settings"},

	{Code: "security:view", Name: "查看安全中心", Module: "security"}, {Code: "security:scan", Name: "执行安全扫描", Module: "security"},
	{Code: "secret:list", Name: "查看密钥列表", Module: "secret"}, {Code: "secret:create", Name: "创建密钥", Module: "secret"},
	{Code: "secret:delete", Name: "删除密钥", Module: "secret"}, {Code: "secret:reveal", Name: "查看密钥明文", Module: "secret"},
}

var systemRoles = []model.Role{
	{Code: "superadmin", Name: "超级管理员", Description: "平台全局管理员，拥有全部权限", IsSystem: true},
	{Code: "admin", Name: "管理员", Description: "组织管理员（用户管理/配额/计费）", IsSystem: true},
	{Code: "developer", Name: "开发者", Description: "项目内全量开发权限", IsSystem: true},
	{Code: "operator", Name: "运维", Description: "实例运维操作与流水线执行", IsSystem: true},
	{Code: "user", Name: "普通用户", Description: "自服务用户（仅自己资源）", IsSystem: true},
	{Code: "viewer", Name: "只读", Description: "全局只读", IsSystem: true},
}

func allCodes(prefix string, names []string) []string {
	var out []string
	for _, n := range names {
		out = append(out, prefix+":"+n)
	}
	return out
}

var ecsAll = allCodes("ecs", []string{"list", "get", "create", "update", "delete", "start", "stop", "restart", "force-stop", "console", "logs", "stats", "recreate"})
var imageAll = allCodes("image", []string{"list", "pull", "delete", "build", "tag", "push"})
var networkAll = allCodes("network", []string{"list", "create", "delete", "connect"})
var volumeAll = allCodes("volume", []string{"list", "create", "delete", "attach"})
var registryAll = allCodes("registry", []string{"list", "create", "delete", "push", "pull"})
var appAll = allCodes("app", []string{"list", "create", "update", "delete", "deploy", "rollback"})
var pipelineAll = allCodes("pipeline", []string{"list", "create", "update", "delete", "run", "cancel"})
var projectAll = allCodes("project", []string{"list", "create", "update", "delete", "deploy"})
var domainAll = allCodes("domain", []string{"list", "create", "delete", "bind"})
var orgAll = allCodes("org", []string{"list", "create", "update", "delete", "member:manage"})

// rolePermMap 角色 → 权限码（对应架构文档 §13.1 角色矩阵）
var rolePermMap = map[string][]string{
	"superadmin": {},
	"admin": concat(ecsAll, imageAll, networkAll, volumeAll, registryAll, appAll, pipelineAll, projectAll, domainAll,
		[]string{"user:list", "user:create", "user:update", "user:grant"},
		orgAll,
		[]string{"quota:view", "quota:update", "billing:view", "audit:view", "settings:view"},
		[]string{"security:view", "security:scan", "secret:list", "secret:create", "secret:delete", "secret:reveal"}),
	"developer": concat(ecsAll, imageAll, networkAll, volumeAll, registryAll, appAll, pipelineAll, projectAll, domainAll,
		[]string{"secret:list", "secret:create", "secret:reveal"}),
	"operator": concat(
		[]string{"ecs:list", "ecs:get", "ecs:start", "ecs:stop", "ecs:restart", "ecs:force-stop", "ecs:console", "ecs:logs", "ecs:stats"},
		[]string{"pipeline:list", "pipeline:run", "pipeline:cancel"},
		[]string{"app:list", "image:list", "network:list", "volume:list", "registry:list", "project:list", "domain:list"}),
	"user": concat(
		[]string{"ecs:list", "ecs:get", "ecs:create", "ecs:update", "ecs:delete", "ecs:start", "ecs:stop", "ecs:restart", "ecs:force-stop", "ecs:console", "ecs:logs", "ecs:stats", "ecs:recreate"},
		[]string{"app:list", "app:create", "app:update", "app:delete", "app:deploy", "app:rollback"},
		[]string{"pipeline:list", "pipeline:create", "pipeline:update", "pipeline:delete", "pipeline:run", "pipeline:cancel"},
		[]string{"image:list", "image:pull", "network:list", "volume:list", "registry:list", "project:list", "project:create", "domain:list"}),
	"viewer": concat(
		[]string{"ecs:list", "ecs:get", "ecs:logs", "ecs:stats"},
		[]string{"image:list", "network:list", "volume:list", "registry:list", "app:list", "pipeline:list", "project:list", "domain:list"},
		[]string{"quota:view", "billing:view", "audit:view"}),
}

func concat(slices ...[]string) []string {
	var out []string
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

// Seed 幂等种子：权限点、系统角色、角色-权限映射、初始管理员。
func Seed(db *gorm.DB, cfg *config.Config, log *zap.Logger) error {
	// 1. 权限点（中文展示名同步：旧库中 name=code 的记录会被更新为中文名）
	for i := range permissions {
		p := permissions[i]
		if p.Name == "" {
			p.Name = p.Code
		}
		var existing model.Permission
		err := db.Where("code = ?", p.Code).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&p).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if existing.Name != p.Name {
			if err := db.Model(&existing).Update("name", p.Name).Error; err != nil {
				return err
			}
		}
	}

	// 2. 系统角色
	for _, r := range systemRoles {
		if err := db.Where("code = ?", r.Code).FirstOrCreate(&r).Error; err != nil {
			return err
		}
	}

	// 3. 角色-权限（superadmin = 全部权限码；其余按 rolePermMap）
	allPermCodes := make([]string, 0, len(permissions))
	for _, p := range permissions {
		allPermCodes = append(allPermCodes, p.Code)
	}
	rolePermMap["superadmin"] = allPermCodes

	for roleCode, codes := range rolePermMap {
		var role model.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			return err
		}
		for _, code := range codes {
			var perm model.Permission
			if err := db.Where("code = ?", code).First(&perm).Error; err != nil {
				continue
			}
			rp := model.RolePermission{RoleID: role.ID, PermissionID: perm.ID}
			if err := db.Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).FirstOrCreate(&rp).Error; err != nil {
				return err
			}
		}
	}

	// 4. 初始管理员（admin / ADMIN_INIT_PASSWORD，默认 Admin@123456）
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		password := os.Getenv("ADMIN_INIT_PASSWORD")
		if password == "" {
			password = "Admin@123456"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		admin := model.User{
			Username:     "admin",
			Email:        "admin@dxcloud.local",
			PasswordHash: string(hash),
			Nickname:     "Administrator",
			Status:       model.UserStatusActive,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
		var sa model.Role
		if err := db.Where("code = ?", "superadmin").First(&sa).Error; err == nil {
			if err := db.Create(&model.UserRole{UserID: admin.ID, RoleID: sa.ID}).Error; err != nil {
				return err
			}
		}
		if os.Getenv("ADMIN_INIT_PASSWORD") == "" {
			log.Warn("initial admin created with default password Admin@123456, please change it immediately",
				zap.String("username", admin.Username))
		}
	}

	// 5. 默认私有 Registry（url 来自 REGISTRY_URL 环境变量）
	var rc int64
	if err := db.Model(&model.Registry{}).Count(&rc).Error; err != nil {
		return err
	}
	if rc == 0 {
		reg := model.Registry{
			Name: "内置 Registry", URL: cfg.RegistryURL, Type: "self", Status: 1,
		}
		if err := db.Create(&reg).Error; err != nil {
			return err
		}
		log.Info("default registry seeded", zap.String("url", cfg.RegistryURL))
	}

	return nil
}
