package model

import (
	"time"

	"gorm.io/gorm"
)

// ---------- 镜像 ----------

const (
	ImageStatusReady   = "ready"
	ImageStatusPulling = "pulling"
	ImageStatusFailed  = "failed"
)

type DockerImage struct {
	ID              uint64         `gorm:"primaryKey" json:"id"`
	OrgID           *uint64        `json:"org_id"`
	ProjectID       *uint64        `json:"project_id"`
	Repo            string         `gorm:"size:255" json:"repo"`
	Tag             string         `gorm:"size:128" json:"tag"`
	ImageID         string         `gorm:"size:128" json:"image_id"`
	SizeBytes       int64          `json:"size_bytes"`
	DockerCreatedAt *time.Time     `json:"docker_created_at"`
	Source          string         `gorm:"size:16" json:"source"`
	Status          string         `gorm:"size:16" json:"status"`
	PullError       string         `gorm:"size:512" json:"pull_error"`
	PullLog         string         `gorm:"type:text" json:"pull_log"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DockerImage) TableName() string { return "docker_images" }

// ---------- 网络 ----------

type DockerNetwork struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	OrgID       *uint64        `json:"org_id"`
	ProjectID   *uint64        `json:"project_id"`
	OwnerID     uint64         `json:"owner_id"`
	Name        string         `gorm:"size:64" json:"name"`
	DockerName  string         `gorm:"size:64" json:"docker_name"`
	DockerNetID string         `gorm:"column:docker_network_id;size:128" json:"docker_network_id"`
	Driver      string         `gorm:"size:32" json:"driver"`
	Subnet      string         `gorm:"size:64" json:"subnet"`
	Gateway     string         `gorm:"size:64" json:"gateway"`
	IPRange     string         `gorm:"size:64" json:"ip_range"`
	Internal    bool           `json:"internal"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DockerNetwork) TableName() string { return "docker_networks" }

// ---------- 存储 ----------

type DockerVolume struct {
	ID         uint64         `gorm:"primaryKey" json:"id"`
	OrgID      *uint64        `json:"org_id"`
	ProjectID  *uint64        `json:"project_id"`
	OwnerID    uint64         `json:"owner_id"`
	Name       string         `gorm:"size:64" json:"name"`
	DockerName string         `gorm:"size:64" json:"docker_name"`
	Driver     string         `gorm:"size:32" json:"driver"`
	Mountpoint string         `gorm:"size:255" json:"mountpoint"`
	CapacityGB int            `json:"capacity_gb"`
	UsedMB     int64          `json:"used_mb"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DockerVolume) TableName() string { return "docker_volumes" }

// ---------- Registry ----------

type Registry struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	OrgID       *uint64        `json:"org_id"`
	Name        string         `gorm:"size:64" json:"name"`
	URL         string         `gorm:"size:255" json:"url"`
	Username    string         `gorm:"size:128" json:"username"`
	PasswordEnc string         `gorm:"size:255" json:"-"`
	Type        string         `gorm:"size:16" json:"type"`
	Status      int8           `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Registry) TableName() string { return "registries" }

type RegistryRepository struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	RegistryID uint64    `json:"registry_id"`
	OrgID      *uint64   `json:"org_id"`
	ProjectID  *uint64   `json:"project_id"`
	Namespace  string    `gorm:"size:128" json:"namespace"`
	Name       string    `gorm:"size:255" json:"name"`
	Visibility string    `gorm:"size:16" json:"visibility"`
	PullCount  int64     `json:"pull_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (RegistryRepository) TableName() string { return "registry_repositories" }
