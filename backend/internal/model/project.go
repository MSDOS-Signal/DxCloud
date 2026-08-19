package model

import (
	"time"

	"gorm.io/gorm"
)

type Organization struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:128" json:"name"`
	Code      string         `gorm:"size:64" json:"code"`
	Plan      string         `gorm:"size:16" json:"plan"`
	Credit    float64        `json:"credit"`
	Status    int8           `json:"status"`
	CreatedBy *uint64        `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Organization) TableName() string { return "organizations" }

type Project struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	OrgID       uint64         `json:"org_id"`
	Name        string         `gorm:"size:128" json:"name"`
	Code        string         `gorm:"size:64" json:"code"`
	Description string         `gorm:"size:255" json:"description"`
	Status      int8           `json:"status"`
	CreatedBy   *uint64        `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Project) TableName() string { return "projects" }

type ProjectEnvironment struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ProjectID uint64    `json:"project_id"`
	Name      string    `gorm:"size:16" json:"name"`
	Seq       int       `json:"seq"`
	CreatedAt time.Time `json:"created_at"`
}

func (ProjectEnvironment) TableName() string { return "project_environments" }
