package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- ECS ----------

type EcsFilter struct {
	OwnerID      *uint64 // user 角色：强制属主过滤（IDOR 免疫）
	OrgID        *uint64 // Phase 10 启用
	ProjID       *uint64 // Phase 10 启用
	DefaultSpace bool    // 未指定组织时只返回默认空间（org_id IS NULL）
	Status       string
	Keyword      string
	Page         int
	Size         int
}

func (r *Repos) EcsCreate(inst *model.EcsInstance) error {
	return r.DB.Create(inst).Error
}

func (r *Repos) EcsGetByID(id uint64) (*model.EcsInstance, error) {
	var inst model.EcsInstance
	err := r.DB.First(&inst, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &inst, err
}

func (r *Repos) EcsGetByNo(instanceNo string) (*model.EcsInstance, error) {
	var inst model.EcsInstance
	err := r.DB.Where("instance_no = ?", instanceNo).First(&inst).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &inst, err
}

func (r *Repos) EcsList(f EcsFilter) ([]model.EcsInstance, int64, error) {
	q := r.DB.Model(&model.EcsInstance{})
	if f.OwnerID != nil {
		q = q.Scopes(ScopeByOwner(*f.OwnerID))
	}
	if f.OrgID != nil {
		q = q.Scopes(ScopeByOrg(*f.OrgID))
	} else if f.DefaultSpace {
		q = q.Where("org_id IS NULL")
	}
	if f.ProjID != nil {
		q = q.Scopes(ScopeByProject(*f.ProjID))
	}
	if f.Status != "" {
		q = q.Where("observed_state = ?", f.Status)
	}
	if f.Keyword != "" {
		like := "%" + f.Keyword + "%"
		q = q.Where("name LIKE ? OR instance_no LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.EcsInstance
	err := q.Order("id DESC").Offset((f.Page - 1) * f.Size).Limit(f.Size).Find(&items).Error
	return items, total, err
}

// EcsUpdateState 更新状态与容器信息（部分字段，避免覆盖业务字段）。
func (r *Repos) EcsUpdateState(inst *model.EcsInstance) error {
	return r.DB.Model(inst).Select(
		"desired_state", "observed_state", "container_id", "container_name", "last_error", "fixed_ip", "mounts", "network_id",
	).Updates(inst).Error
}

func (r *Repos) EcsUpdateInfo(inst *model.EcsInstance) error {
	return r.DB.Model(inst).Select("name", "description", "restart_policy").Updates(inst).Error
}

func (r *Repos) EcsSoftDelete(id uint64) error {
	return r.DB.Delete(&model.EcsInstance{}, id).Error
}

// EcsListByStates 对账用：按 observed/desired 状态批量取。
func (r *Repos) EcsListByStates(states []string, fields ...string) ([]model.EcsInstance, error) {
	var items []model.EcsInstance
	err := r.DB.Where("observed_state IN ? OR desired_state IN ?", states, states).Find(&items).Error
	return items, err
}

// EcsQuotaUsage 属主配额占用（实例数/CPU/内存，排除已删除与 Failed 状态）。
func (r *Repos) EcsQuotaUsage(ownerID uint64) (count int64, cpu float64, mem int64, err error) {
	row := struct {
		Count int64   `gorm:"column:cnt"`
		CPU   float64 `gorm:"column:cpu"`
		Mem   int64   `gorm:"column:mem"`
	}{}
	err = r.DB.Model(&model.EcsInstance{}).
		Select("COUNT(*) AS cnt, COALESCE(SUM(cpu),0) AS cpu, COALESCE(SUM(memory_mb),0) AS mem").
		Where("owner_id = ? AND observed_state <> ?", ownerID, model.EcsFailed).
		Scan(&row).Error
	return row.Count, row.CPU, row.Mem, err
}

// EcsOrgQuotaUsage 组织维度配额占用（多租户模式，按 org_id 汇总，排除已删除与 Failed 状态）。
func (r *Repos) EcsOrgQuotaUsage(orgID uint64) (count int64, cpu float64, mem int64, err error) {
	row := struct {
		Count int64   `gorm:"column:cnt"`
		CPU   float64 `gorm:"column:cpu"`
		Mem   int64   `gorm:"column:mem"`
	}{}
	err = r.DB.Model(&model.EcsInstance{}).
		Select("COUNT(*) AS cnt, COALESCE(SUM(cpu),0) AS cpu, COALESCE(SUM(memory_mb),0) AS mem").
		Where("org_id = ? AND observed_state <> ?", orgID, model.EcsFailed).
		Scan(&row).Error
	return row.Count, row.CPU, row.Mem, err
}

// EcsEventCreate 写事件。
func (r *Repos) EcsEventCreate(e *model.EcsEvent) error {
	return r.DB.Create(e).Error
}

func (r *Repos) EcsEventList(instanceID uint64, limit int) ([]model.EcsEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	var items []model.EcsEvent
	err := r.DB.Where("instance_id = ?", instanceID).Order("id DESC").Limit(limit).Find(&items).Error
	return items, err
}
