package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- 项目 ----------

func (r *Repos) ProjectList() ([]model.Project, error) {
	var items []model.Project
	err := r.DB.Order("id DESC").Find(&items).Error
	return items, err
}

// ProjectListByOrg 按组织过滤项目（多租户隔离）。
func (r *Repos) ProjectListByOrg(orgID uint64) ([]model.Project, error) {
	var items []model.Project
	err := r.DB.Where("org_id = ?", orgID).Order("id DESC").Find(&items).Error
	return items, err
}

// ProjectListForContext 0=默认空间（org_id=0），大于 0=指定组织。
func (r *Repos) ProjectListForContext(orgID uint64) ([]model.Project, error) {
	q := r.DB.Model(&model.Project{})
	if orgID == 0 {
		q = q.Where("org_id IS NULL OR org_id = 0")
	} else {
		q = q.Where("org_id = ?", orgID)
	}
	var items []model.Project
	err := q.Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) ProjectGetByID(id uint64) (*model.Project, error) {
	var p model.Project
	err := r.DB.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *Repos) ProjectGetByName(name string) (*model.Project, error) {
	var p model.Project
	err := r.DB.Where("name = ?", name).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// ProjectGetByNameOrg 组织内项目名唯一性检查。
func (r *Repos) ProjectGetByNameOrg(name string, orgID uint64) (*model.Project, error) {
	var p model.Project
	err := r.DB.Where("name = ? AND org_id = ?", name, orgID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *Repos) ProjectCreate(p *model.Project) error {
	return r.DB.Create(p).Error
}

func (r *Repos) ProjectSoftDelete(id uint64) error {
	return r.DB.Delete(&model.Project{}, id).Error
}

func (r *Repos) EnvCreate(e *model.ProjectEnvironment) error {
	return r.DB.Create(e).Error
}

func (r *Repos) EnvList(projectID uint64) ([]model.ProjectEnvironment, error) {
	var items []model.ProjectEnvironment
	err := r.DB.Where("project_id = ?", projectID).Order("seq ASC").Find(&items).Error
	return items, err
}

// ---------- 应用 ----------

func (r *Repos) AppList(projectID *uint64, keyword string) ([]model.Application, error) {
	q := r.DB.Model(&model.Application{})
	if projectID != nil {
		q = q.Where("project_id = ?", *projectID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ?", like)
	}
	var items []model.Application
	err := q.Order("id DESC").Find(&items).Error
	return items, err
}

// AppListOrg 组织维度应用列表（多租户隔离）。
func (r *Repos) AppListOrg(orgID uint64, projectID *uint64, keyword string) ([]model.Application, error) {
	q := r.DB.Model(&model.Application{}).Where("org_id = ?", orgID)
	if orgID == 0 {
		q = r.DB.Model(&model.Application{}).Where("org_id IS NULL OR org_id = 0")
	}
	if projectID != nil {
		q = q.Where("project_id = ?", *projectID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("name LIKE ?", like)
	}
	var items []model.Application
	err := q.Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) AppGetByID(id uint64) (*model.Application, error) {
	var a model.Application
	err := r.DB.First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *Repos) AppGetByName(name string) (*model.Application, error) {
	var a model.Application
	err := r.DB.Where("name = ?", name).First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *Repos) AppGetByNameOrg(name string, orgID uint64) (*model.Application, error) {
	var a model.Application
	q := r.DB.Where("name = ?", name)
	if orgID == 0 {
		q = q.Where("org_id IS NULL OR org_id = 0")
	} else {
		q = q.Where("org_id = ?", orgID)
	}
	err := q.First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &a, err
}

func (r *Repos) AppCreate(a *model.Application) error {
	return r.DB.Create(a).Error
}

func (r *Repos) AppUpdate(a *model.Application) error {
	return r.DB.Model(a).Select("name", "type", "image", "git_url", "git_branch", "port", "health_check_path", "env", "domain", "active_deployment_id").Updates(a).Error
}

func (r *Repos) AppSoftDelete(id uint64) error {
	return r.DB.Delete(&model.Application{}, id).Error
}

// ---------- 版本 ----------

func (r *Repos) VersionCreate(v *model.AppVersion) error {
	return r.DB.Create(v).Error
}

func (r *Repos) VersionGetByID(id uint64) (*model.AppVersion, error) {
	var v model.AppVersion
	err := r.DB.First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, err
}

func (r *Repos) VersionList(appID uint64) ([]model.AppVersion, error) {
	var items []model.AppVersion
	err := r.DB.Where("application_id = ?", appID).Order("id DESC").Find(&items).Error
	return items, err
}

// ---------- 部署 ----------

func (r *Repos) DeploymentCreate(d *model.Deployment) error {
	return r.DB.Create(d).Error
}

func (r *Repos) DeploymentUpdate(d *model.Deployment) error {
	return r.DB.Model(d).Select("status", "health_status", "container_id", "container_name", "config_json", "note", "started_at", "finished_at").Updates(d).Error
}

func (r *Repos) DeploymentGetByID(id uint64) (*model.Deployment, error) {
	var d model.Deployment
	err := r.DB.First(&d, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

// DeploymentListByApp 应用部署历史。
func (r *Repos) DeploymentListByApp(appID uint64, limit int) ([]model.Deployment, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var items []model.Deployment
	err := r.DB.Where("application_id = ?", appID).Order("id DESC").Limit(limit).Find(&items).Error
	return items, err
}

// DeploymentListAllByApp 取应用全部部署记录（用于删除时清理容器，不受 limit 上限约束）。
func (r *Repos) DeploymentListAllByApp(appID uint64) ([]model.Deployment, error) {
	var items []model.Deployment
	err := r.DB.Where("application_id = ?", appID).Order("id DESC").Find(&items).Error
	return items, err
}

// DeploymentActive 当前生效部署。
func (r *Repos) DeploymentActive(appID uint64) (*model.Deployment, error) {
	var d model.Deployment
	err := r.DB.Where("application_id = ? AND status = ?", appID, model.DeploySuccess).Order("id DESC").First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

// ---------- 域名 ----------

func (r *Repos) DomainList() ([]model.Domain, error) {
	var items []model.Domain
	err := r.DB.Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) DomainListForContext(orgID uint64) ([]model.Domain, error) {
	q := r.DB.Model(&model.Domain{})
	if orgID == 0 {
		q = q.Where("org_id IS NULL OR org_id = 0")
	} else {
		q = q.Where("org_id = ?", orgID)
	}
	var items []model.Domain
	err := q.Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) DomainGetByName(domain string) (*model.Domain, error) {
	var d model.Domain
	err := r.DB.Where("domain = ?", domain).First(&d).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *Repos) DomainGetByID(id uint64) (*model.Domain, error) {
	var d model.Domain
	err := r.DB.First(&d, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &d, err
}

func (r *Repos) DomainCreate(d *model.Domain) error {
	return r.DB.Create(d).Error
}

func (r *Repos) DomainUpdate(d *model.Domain) error {
	return r.DB.Model(d).Select("application_id", "target_port", "tls").Updates(d).Error
}

func (r *Repos) DomainSoftDelete(id uint64) error {
	return r.DB.Delete(&model.Domain{}, id).Error
}
