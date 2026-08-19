package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- 密钥托管 ----------

func (r *Repos) SecretCreate(s *model.Secret) error {
	return r.DB.Create(s).Error
}

func (r *Repos) SecretList(orgID uint64) ([]model.Secret, error) {
	var items []model.Secret
	err := r.DB.Where("org_id = ?", orgID).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) SecretGetByID(id uint64) (*model.Secret, error) {
	var s model.Secret
	err := r.DB.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *Repos) SecretGetByName(orgID uint64, name string) (*model.Secret, error) {
	var s model.Secret
	err := r.DB.Where("org_id = ? AND name = ?", orgID, name).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

func (r *Repos) SecretDelete(id uint64) error {
	return r.DB.Delete(&model.Secret{}, id).Error
}

// ---------- 安全报告 ----------

func (r *Repos) SecurityReportCreate(report *model.SecurityReport) error {
	return r.DB.Create(report).Error
}

func (r *Repos) SecurityReportList(limit int) ([]model.SecurityReport, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var items []model.SecurityReport
	err := r.DB.Order("id DESC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *Repos) SecurityReportGetByID(id uint64) (*model.SecurityReport, error) {
	var s model.SecurityReport
	err := r.DB.First(&s, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

// SecurityReportLatest 每种 kind 最新一份。
func (r *Repos) SecurityReportLatest(kind string) (*model.SecurityReport, error) {
	var s model.SecurityReport
	err := r.DB.Where("kind = ?", kind).Order("id DESC").First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}
