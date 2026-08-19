package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- 镜像 ----------

func (r *Repos) ImageList(page, size int, keyword string, orgID uint64) ([]model.DockerImage, int64, error) {
	q := r.DB.Model(&model.DockerImage{}).Scopes(ScopeByTenantOrNull(orgID))
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("repo LIKE ? OR tag LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.DockerImage
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

func (r *Repos) ImageGetByID(id uint64) (*model.DockerImage, error) {
	var img model.DockerImage
	err := r.DB.First(&img, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &img, err
}

func (r *Repos) ImageGetByRepoTag(repo, tag string, orgID uint64) (*model.DockerImage, error) {
	var img model.DockerImage
	err := r.DB.Where("repo = ? AND tag = ?", repo, tag).Scopes(ScopeByTenantOrNull(orgID)).Order("id DESC").First(&img).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &img, err
}

func (r *Repos) ImageCreate(img *model.DockerImage) error {
	return r.DB.Create(img).Error
}

func (r *Repos) ImageUpdate(img *model.DockerImage) error {
	return r.DB.Model(img).Select("image_id", "size_bytes", "docker_created_at", "status", "pull_error", "pull_log").Updates(img).Error
}

func (r *Repos) ImageSoftDelete(id uint64) error {
	return r.DB.Delete(&model.DockerImage{}, id).Error
}

// ---------- 网络 ----------

func (r *Repos) NetworkList(orgID uint64) ([]model.DockerNetwork, error) {
	var items []model.DockerNetwork
	err := r.DB.Scopes(ScopeByTenantOrNull(orgID)).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) NetworkGetByID(id uint64) (*model.DockerNetwork, error) {
	var n model.DockerNetwork
	err := r.DB.First(&n, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &n, err
}

func (r *Repos) NetworkGetByName(name string, orgID uint64) (*model.DockerNetwork, error) {
	var n model.DockerNetwork
	err := r.DB.Where("name = ?", name).Scopes(ScopeByTenantOrNull(orgID)).First(&n).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &n, err
}

func (r *Repos) NetworkCreate(n *model.DockerNetwork) error {
	return r.DB.Create(n).Error
}

func (r *Repos) NetworkUpdate(n *model.DockerNetwork) error {
	return r.DB.Model(n).Select("docker_network_id", "driver", "subnet", "gateway", "ip_range").Updates(n).Error
}

func (r *Repos) NetworkSoftDelete(id uint64) error {
	return r.DB.Delete(&model.DockerNetwork{}, id).Error
}

// ---------- 存储 ----------

func (r *Repos) VolumeList(orgID uint64) ([]model.DockerVolume, error) {
	var items []model.DockerVolume
	err := r.DB.Scopes(ScopeByTenantOrNull(orgID)).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) VolumeGetByID(id uint64) (*model.DockerVolume, error) {
	var v model.DockerVolume
	err := r.DB.First(&v, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, err
}

func (r *Repos) VolumeGetByName(name string, orgID uint64) (*model.DockerVolume, error) {
	var v model.DockerVolume
	err := r.DB.Where("name = ?", name).Scopes(ScopeByTenantOrNull(orgID)).First(&v).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &v, err
}

func (r *Repos) VolumeCreate(v *model.DockerVolume) error {
	return r.DB.Create(v).Error
}

func (r *Repos) VolumeUpdate(v *model.DockerVolume) error {
	return r.DB.Model(v).Select("mountpoint", "used_mb").Updates(v).Error
}

func (r *Repos) VolumeSoftDelete(id uint64) error {
	return r.DB.Delete(&model.DockerVolume{}, id).Error
}

// ---------- Registry ----------

func (r *Repos) RegistryList() ([]model.Registry, error) {
	var items []model.Registry
	err := r.DB.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repos) RegistryGetByID(id uint64) (*model.Registry, error) {
	var reg model.Registry
	err := r.DB.First(&reg, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &reg, err
}

func (r *Repos) RegistryCreate(reg *model.Registry) error {
	return r.DB.Create(reg).Error
}

func (r *Repos) RegistrySoftDelete(id uint64) error {
	return r.DB.Delete(&model.Registry{}, id).Error
}

func (r *Repos) RegistryRepoUpsert(repo *model.RegistryRepository) error {
	var existing model.RegistryRepository
	err := r.DB.Where("registry_id = ? AND namespace = ? AND name = ?", repo.RegistryID, repo.Namespace, repo.Name).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.DB.Create(repo).Error
	}
	return err
}

func (r *Repos) RegistryRepoList(registryID uint64) ([]model.RegistryRepository, error) {
	var items []model.RegistryRepository
	err := r.DB.Where("registry_id = ?", registryID).Order("name ASC").Find(&items).Error
	return items, err
}
