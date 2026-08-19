package model

import (
	"time"

	"gorm.io/gorm"
)

// Secret 托管密钥（value_cipher 为 AES-256-GCM 密文，永不落明文）。
type Secret struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	OrgID       uint64         `gorm:"index" json:"org_id"`
	Name        string         `gorm:"size:128" json:"name"`
	ValueCipher string         `gorm:"type:text" json:"-"`
	CreatedBy   *uint64        `json:"created_by"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Secret) TableName() string { return "secrets" }

// SecurityReport 安全扫描报告（summary 为发现项 JSON 数组）。
type SecurityReport struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Kind         string    `gorm:"size:32" json:"kind"`
	Score        int       `json:"score"`
	FindingCount int       `json:"finding_count"`
	Summary      string    `gorm:"type:text" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

func (SecurityReport) TableName() string { return "security_reports" }
