package model

import (
	"time"

	"gorm.io/gorm"
)

// 部署状态
const (
	DeployPending    = "pending"
	DeployDeploying  = "deploying"
	DeploySuccess    = "success"
	DeployFailed     = "failed"
	DeployRolledBack = "rolled-back"
)

type Application struct {
	ID                  uint64         `gorm:"primaryKey" json:"id"`
	OrgID               *uint64        `json:"org_id"`
	ProjectID           *uint64        `json:"project_id"`
	OwnerID             uint64         `json:"owner_id"`
	Name                string         `gorm:"size:64" json:"name"`
	Type                string         `gorm:"size:32" json:"type"`
	Image               string         `gorm:"size:255" json:"image"`
	GitURL              string         `gorm:"size:512" json:"git_url"`
	GitBranch           string         `gorm:"size:64" json:"git_branch"`
	Port                int            `json:"port"`
	HealthCheckPath     string         `gorm:"size:255" json:"health_check_path"`
	Env                 string         `gorm:"type:text" json:"env"`
	Domain              string         `gorm:"size:255" json:"domain"`
	ActiveDeploymentID  *uint64        `json:"active_deployment_id"`
	Status              int8           `json:"status"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Application) TableName() string { return "applications" }

type AppVersion struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	ApplicationID uint64    `json:"application_id"`
	Version       string    `gorm:"size:64" json:"version"`
	ImageRef      string    `gorm:"size:255" json:"image_ref"`
	CommitSHA     string    `gorm:"size:64" json:"commit_sha"`
	Status        string    `gorm:"size:16" json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

func (AppVersion) TableName() string { return "application_versions" }

type Deployment struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	OrgID         *uint64   `json:"org_id"`
	ProjectID     *uint64   `json:"project_id"`
	ApplicationID uint64    `json:"application_id"`
	EnvironmentID *uint64   `json:"environment_id"`
	VersionID     *uint64   `json:"version_id"`
	Version       string    `gorm:"size:64" json:"version"`
	ImageRef      string    `gorm:"size:255" json:"image_ref"`
	Strategy      string    `gorm:"size:16" json:"strategy"`
	Status        string    `gorm:"size:16" json:"status"`
	HealthStatus  string    `gorm:"size:16" json:"health_status"`
	Trigger       string    `gorm:"column:trigger_type;size:16" json:"trigger"`
	PipelineRunID *uint64   `json:"pipeline_run_id"`
	ContainerID   string    `gorm:"size:128" json:"container_id"`
	ContainerName string    `gorm:"size:128" json:"container_name"`
	ConfigJSON    string    `gorm:"type:text" json:"-"`
	Note          string    `gorm:"size:255" json:"note"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Deployment) TableName() string { return "deployments" }

type Domain struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	OrgID         *uint64        `json:"org_id"`
	ProjectID     *uint64        `json:"project_id"`
	ApplicationID *uint64        `json:"application_id"`
	Domain        string         `gorm:"size:255" json:"domain"`
	TargetPort    int            `json:"target_port"`
	TLS           bool           `json:"tls"`
	CertID        *uint64        `json:"cert_id"`
	Status        int8           `json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Domain) TableName() string { return "domains" }
