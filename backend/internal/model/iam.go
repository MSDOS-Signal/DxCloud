// Package model 定义 GORM 实体（IAM 域）。
package model

import (
	"time"

	"gorm.io/gorm"
)

// 用户状态
const (
	UserStatusActive   int8 = 1
	UserStatusDisabled int8 = 2
	UserStatusLocked   int8 = 3
)

type User struct {
	ID           uint64         `gorm:"primaryKey" json:"id"`
	Username     string         `gorm:"size:64;uniqueIndex" json:"username"`
	Email        string         `gorm:"size:128;uniqueIndex" json:"email"`
	PasswordHash string         `gorm:"size:255" json:"-"`
	Nickname     string         `gorm:"size:64" json:"nickname"`
	AvatarURL    string         `gorm:"type:text" json:"avatar_url"`
	Status       int8           `gorm:"default:1" json:"status"`
	LastLoginAt  *time.Time     `json:"last_login_at"`
	LastLoginIP  string         `gorm:"size:64" json:"last_login_ip"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string { return "users" }

type Role struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"size:64;uniqueIndex" json:"code"`
	Name        string         `gorm:"size:64" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	IsSystem    bool           `gorm:"default:false" json:"is_system"`
	Scope       string         `gorm:"size:16;default:global" json:"scope"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Role) TableName() string { return "roles" }

type Permission struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Code        string    `gorm:"size:64;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:64" json:"name"`
	Module      string    `gorm:"size:32" json:"module"`
	Description string    `gorm:"size:255" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Permission) TableName() string { return "permissions" }

type UserRole struct {
	ID        uint64    `gorm:"primaryKey"`
	UserID    uint64    `gorm:"uniqueIndex:uk_user_roles" json:"user_id"`
	RoleID    uint64    `gorm:"uniqueIndex:uk_user_roles" json:"role_id"`
	OrgID     *uint64   `gorm:"uniqueIndex:uk_user_roles" json:"org_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (UserRole) TableName() string { return "user_roles" }

type RolePermission struct {
	ID           uint64    `gorm:"primaryKey"`
	RoleID       uint64    `gorm:"uniqueIndex:uk_role_permissions" json:"role_id"`
	PermissionID uint64    `gorm:"uniqueIndex:uk_role_permissions" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func (RolePermission) TableName() string { return "role_permissions" }

type LoginLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	UserID    *uint64   `json:"user_id"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:512" json:"user_agent"`
	Status    int8      `gorm:"default:0" json:"status"` // 1=成功 2=失败
	Message   string    `gorm:"size:255" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (LoginLog) TableName() string { return "login_logs" }

type AuditLog struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	OrgID        *uint64   `json:"org_id"`
	UserID       *uint64   `json:"user_id"`
	Action       string    `gorm:"size:64" json:"action"`
	ResourceType string    `gorm:"size:32" json:"resource_type"`
	ResourceID   string    `gorm:"size:64" json:"resource_id"`
	IP           string    `gorm:"size:64" json:"ip"`
	RequestID    string    `gorm:"size:64" json:"request_id"`
	Detail       string    `gorm:"type:text" json:"detail"`
	Status       int8      `gorm:"default:1" json:"status"` // 1=成功 2=拒绝
	CreatedAt    time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }
