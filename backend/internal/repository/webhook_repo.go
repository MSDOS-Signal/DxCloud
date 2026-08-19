package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- Webhook ----------

func (r *Repos) WebhookList(pipelineID *uint64, ownerID *uint64, orgID uint64) ([]model.Webhook, error) {
	q := r.DB.Model(&model.Webhook{})
	if pipelineID != nil {
		q = q.Where("pipeline_id = ?", *pipelineID)
	}
	if ownerID != nil {
		q = q.Where("pipeline_id IN (SELECT id FROM pipelines WHERE owner_id = ?)", *ownerID)
	}
	if orgID == 0 {
		q = q.Where("pipeline_id IN (SELECT id FROM pipelines WHERE org_id IS NULL OR org_id = 0)")
	} else {
		q = q.Where("pipeline_id IN (SELECT id FROM pipelines WHERE org_id = ?)", orgID)
	}
	var items []model.Webhook
	err := q.Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) WebhookGetByID(id uint64) (*model.Webhook, error) {
	var w model.Webhook
	err := r.DB.First(&w, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &w, err
}

func (r *Repos) WebhookGetByCode(code string) (*model.Webhook, error) {
	var w model.Webhook
	err := r.DB.Where("hook_code = ?", code).First(&w).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &w, err
}

func (r *Repos) WebhookCreate(w *model.Webhook) error {
	return r.DB.Create(w).Error
}

func (r *Repos) WebhookSoftDelete(id uint64) error {
	return r.DB.Delete(&model.Webhook{}, id).Error
}
