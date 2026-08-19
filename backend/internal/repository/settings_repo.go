package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- System Settings ----------

func (r *Repos) SettingsGet(key string) (*model.SystemSetting, error) {
	var s model.SystemSetting
	err := r.DB.Where("`key` = ?", key).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &s, err
}

// SettingsUpsert 存在则更新 value/description/updated_by，不存在则创建。
func (r *Repos) SettingsUpsert(s *model.SystemSetting) error {
	var existing model.SystemSetting
	err := r.DB.Where("`key` = ?", s.Key).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.DB.Create(s).Error
	}
	if err != nil {
		return err
	}
	return r.DB.Model(&existing).Updates(map[string]any{
		"value":       s.Value,
		"description": s.Description,
		"updated_by":  s.UpdatedBy,
	}).Error
}
