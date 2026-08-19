package model

import (
	"time"

	"gorm.io/gorm"
)

type Webhook struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	ProjectID    *uint64        `json:"project_id"`
	PipelineID   uint64         `json:"pipeline_id"`
	Provider     string         `gorm:"size:16" json:"provider"`
	SecretEnc    string         `gorm:"size:512" json:"-"`
	BranchFilter string         `gorm:"size:128" json:"branch_filter"`
	Events       string         `gorm:"size:64" json:"events"`
	Status       int8           `json:"status"`
	HookCode     string         `gorm:"size:32" json:"hook_code"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Webhook) TableName() string { return "webhooks" }
