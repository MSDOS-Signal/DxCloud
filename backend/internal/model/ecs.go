package model

import (
	"time"

	"gorm.io/gorm"
)

// ECS 生命周期状态（desired = 用户意图，observed = Docker 实况；双状态模型供 Reconciler 对账）
const (
	EcsCreating   = "creating"
	EcsRunning    = "running"
	EcsStopped    = "stopped"
	EcsStarting   = "starting"
	EcsStopping   = "stopping"
	EcsRestarting = "restarting"
	EcsDeleting   = "deleting"
	EcsFailed     = "failed"
	EcsUnknown    = "unknown"
)

type EcsInstance struct {
	ID             uint64         `gorm:"primaryKey" json:"id"`
	InstanceNo     string         `gorm:"size:32;uniqueIndex" json:"instance_no"`
	OrgID          *uint64        `json:"org_id"`
	ProjectID      *uint64        `json:"project_id"`
	OwnerID        uint64         `gorm:"index" json:"owner_id"`
	Name           string         `gorm:"size:64" json:"name"`
	Description    string         `gorm:"size:255" json:"description"`
	Image          string         `gorm:"size:255" json:"image"`
	CPU            float64        `json:"cpu"`
	MemoryMB       int64          `json:"memory_mb"`
	DiskGB         int64          `json:"disk_gb"`
	NetworkID      string         `gorm:"size:64" json:"network_id"`
	FixedIP        string         `gorm:"size:64" json:"fixed_ip"`
	Ports          string         `gorm:"type:text" json:"ports"`
	Env            string         `gorm:"type:text" json:"env"`
	Command        string         `gorm:"type:text" json:"command"`
	Mounts         string         `gorm:"type:text" json:"mounts"`
	RestartPolicy  string         `gorm:"size:32" json:"restart_policy"`
	ReadonlyRootfs bool           `json:"readonly_rootfs"`
	DesiredState   string         `gorm:"size:16;index" json:"desired_state"`
	ObservedState  string         `gorm:"size:16;index" json:"observed_state"`
	ContainerID    string         `gorm:"size:128;uniqueIndex" json:"container_id"`
	ContainerName  string         `gorm:"size:128" json:"container_name"`
	LastError      string         `gorm:"type:text" json:"last_error"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (EcsInstance) TableName() string { return "ecs_instances" }

type EcsEvent struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	InstanceID uint64    `gorm:"index" json:"instance_id"`
	EventType  string    `gorm:"size:32" json:"event_type"`
	Level      string    `gorm:"size:16" json:"level"`
	Message    string    `gorm:"size:512" json:"message"`
	ActorID    *uint64   `json:"actor_id"`
	RequestID  string    `gorm:"size:64" json:"request_id"`
	CreatedAt  time.Time `json:"created_at"`
}

func (EcsEvent) TableName() string { return "ecs_instance_events" }
