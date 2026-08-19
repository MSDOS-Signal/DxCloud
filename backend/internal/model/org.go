package model

import "time"

type OrganizationMember struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	OrgID     uint64    `gorm:"uniqueIndex:uk_org_members" json:"org_id"`
	UserID    uint64    `gorm:"uniqueIndex:uk_org_members" json:"user_id"`
	OrgRole   string    `gorm:"size:16" json:"org_role"`
	Status    int8      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (OrganizationMember) TableName() string { return "organization_members" }

type ResourceQuota struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	OrgID        uint64    `json:"org_id"`
	ProjectID    *uint64   `json:"project_id"`
	ResourceType string    `gorm:"size:32" json:"resource_type"`
	LimitValue   int64     `json:"limit_value"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (ResourceQuota) TableName() string { return "resource_quotas" }

type ResourceUsage struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	OrgID        uint64    `json:"org_id"`
	ProjectID    *uint64   `json:"project_id"`
	ResourceType string    `gorm:"size:32" json:"resource_type"`
	UsedValue    float64   `json:"used_value"`
	Period       time.Time `json:"period"`
	CreatedAt    time.Time `json:"created_at"`
}

func (ResourceUsage) TableName() string { return "resource_usage" }
