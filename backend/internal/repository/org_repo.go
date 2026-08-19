package repository

import (
	"errors"
	"time"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- 组织 ----------

func (r *Repos) OrgList() ([]model.Organization, error) {
	var items []model.Organization
	err := r.DB.Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repos) OrgGetByID(id uint64) (*model.Organization, error) {
	var o model.Organization
	err := r.DB.First(&o, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &o, err
}

func (r *Repos) OrgCreate(o *model.Organization) error {
	return r.DB.Create(o).Error
}

func (r *Repos) OrgUpdate(o *model.Organization) error {
	return r.DB.Model(o).Select("name", "plan", "credit").Updates(o).Error
}

// OrgAdjustCredit 原子调整组织余额（正数充值，负数扣费），避免 read-then-write 竞态。
func (r *Repos) OrgAdjustCredit(orgID uint64, delta float64) error {
	return r.DB.Model(&model.Organization{}).Where("id = ?", orgID).
		UpdateColumn("credit", gorm.Expr("credit + ?", delta)).Error
}

func (r *Repos) OrgSoftDelete(id uint64) error {
	return r.DB.Delete(&model.Organization{}, id).Error
}

// ---------- 成员 ----------

func (r *Repos) OrgMemberList(orgID uint64) ([]model.OrganizationMember, error) {
	var items []model.OrganizationMember
	err := r.DB.Where("org_id = ?", orgID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repos) OrgMemberGet(orgID, userID uint64) (*model.OrganizationMember, error) {
	var m model.OrganizationMember
	err := r.DB.Where("org_id = ? AND user_id = ?", orgID, userID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &m, err
}

func (r *Repos) OrgMemberCreate(m *model.OrganizationMember) error {
	return r.DB.Create(m).Error
}

func (r *Repos) OrgMemberUpdate(m *model.OrganizationMember) error {
	return r.DB.Model(m).Select("org_role", "status").Updates(m).Error
}

func (r *Repos) OrgMemberDelete(orgID, userID uint64) error {
	return r.DB.Where("org_id = ? AND user_id = ?", orgID, userID).Delete(&model.OrganizationMember{}).Error
}

// OrgMemberOrgs 用户所属组织。
func (r *Repos) OrgMemberOrgs(userID uint64) ([]uint64, error) {
	var ids []uint64
	err := r.DB.Model(&model.OrganizationMember{}).Where("user_id = ?", userID).Pluck("org_id", &ids).Error
	return ids, err
}

// OrgListByMember 用户所属组织（完整对象，含角色/余额）。
func (r *Repos) OrgListByMember(userID uint64) ([]model.Organization, error) {
	var items []model.Organization
	err := r.DB.Model(&model.Organization{}).
		Joins("JOIN organization_members m ON m.org_id = organizations.id AND m.user_id = ? AND m.status = 1", userID).
		Order("organizations.id ASC").
		Find(&items).Error
	return items, err
}

// ---------- 配额 ----------

func (r *Repos) QuotaList(orgID uint64) ([]model.ResourceQuota, error) {
	var items []model.ResourceQuota
	err := r.DB.Where("org_id = ?", orgID).Order("resource_type ASC").Find(&items).Error
	return items, err
}

func (r *Repos) QuotaGet(orgID uint64, resourceType string) (*model.ResourceQuota, error) {
	var q model.ResourceQuota
	err := r.DB.Where("org_id = ? AND resource_type = ?", orgID, resourceType).First(&q).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &q, err
}

func (r *Repos) QuotaUpsert(q *model.ResourceQuota) error {
	var existing model.ResourceQuota
	err := r.DB.Where("org_id = ? AND resource_type = ?", q.OrgID, q.ResourceType).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.DB.Create(q).Error
	}
	if err != nil {
		return err
	}
	return r.DB.Model(&existing).Update("limit_value", q.LimitValue).Error
}

// ---------- 用量 ----------

func (r *Repos) UsageCreate(u *model.ResourceUsage) error {
	return r.DB.Create(u).Error
}

func (r *Repos) UsageList(orgID uint64, from, to time.Time, limit int) ([]model.ResourceUsage, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	q := r.DB.Model(&model.ResourceUsage{}).Where("org_id = ?", orgID)
	if !from.IsZero() {
		q = q.Where("period >= ?", from)
	}
	if !to.IsZero() {
		q = q.Where("period <= ?", to)
	}
	var items []model.ResourceUsage
	err := q.Order("id DESC").Limit(limit).Find(&items).Error
	return items, err
}

// UsageSum 区间汇总。
func (r *Repos) UsageSum(orgID uint64, from, to time.Time) (map[string]float64, error) {
	var rows []struct {
		ResourceType string  `gorm:"column:resource_type"`
		Sum          float64 `gorm:"column:s"`
	}
	err := r.DB.Model(&model.ResourceUsage{}).
		Select("resource_type, SUM(used_value) AS s").
		Where("org_id = ? AND period >= ? AND period <= ?", orgID, from, to).
		Group("resource_type").Scan(&rows).Error
	out := map[string]float64{}
	for _, row := range rows {
		out[row.ResourceType] = row.Sum
	}
	return out, err
}
