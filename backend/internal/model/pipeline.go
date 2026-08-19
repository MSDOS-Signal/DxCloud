package model

import (
	"time"

	"gorm.io/gorm"
)

// Pipeline / Run / Job 状态
const (
	PipePending  = "pending"
	PipeRunning  = "running"
	PipeSuccess  = "success"
	PipeFailed   = "failed"
	PipeCanceled = "canceled"

	JobPending  = "pending"
	JobRunning  = "running"
	JobSuccess  = "success"
	JobFailed   = "failed"
	JobSkipped  = "skipped"
)

type Pipeline struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	OrgID       *uint64        `json:"org_id"`
	ProjectID   *uint64        `json:"project_id"`
	OwnerID     uint64         `json:"owner_id"`
	Name        string         `gorm:"size:64" json:"name"`
	Description string         `gorm:"size:255" json:"description"`
	Definition  string         `gorm:"type:mediumtext" json:"definition"`
	Status      int8           `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Pipeline) TableName() string { return "pipelines" }

type PipelineStep struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	PipelineID uint64    `json:"pipeline_id"`
	Name       string    `gorm:"size:64" json:"name"`
	Type       string    `gorm:"size:32" json:"type"`
	Seq        int       `json:"seq"`
	ConfigJSON string    `gorm:"type:text" json:"config"`
	CreatedAt  time.Time `json:"created_at"`
}

func (PipelineStep) TableName() string { return "pipeline_steps" }

type PipelineRun struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	PipelineID  uint64    `json:"pipeline_id"`
	RunNo       int       `json:"run_no"`
	TriggerType string    `gorm:"column:trigger_type;size:16" json:"trigger"`
	Ref         string    `gorm:"size:64" json:"ref"`
	CommitSHA   string    `gorm:"size:64" json:"commit_sha"`
	Status      string    `gorm:"size:16;index" json:"status"`
	StartedAt   *time.Time `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at"`
	DurationMs  int64     `json:"duration_ms"`
	TriggeredBy *uint64   `json:"triggered_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (PipelineRun) TableName() string { return "pipeline_runs" }

type PipelineJobRun struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	PipelineRunID uint64    `json:"pipeline_run_id"`
	StepID        uint64    `json:"step_id"`
	Name          string    `gorm:"size:64" json:"name"`
	Type          string    `gorm:"size:32" json:"type"`
	Status        string    `gorm:"size:16" json:"status"`
	ExitCode      int       `json:"exit_code"`
	ContainerID   string    `gorm:"size:128" json:"container_id"`
	LogPath       string    `gorm:"size:255" json:"log_path"`
	StartedAt     *time.Time `json:"started_at"`
	FinishedAt    *time.Time `json:"finished_at"`
	CreatedAt     time.Time `json:"created_at"`
}

func (PipelineJobRun) TableName() string { return "pipeline_job_runs" }
