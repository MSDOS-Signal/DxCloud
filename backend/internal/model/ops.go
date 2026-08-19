package model

import "time"

type MetricSample struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Kind      string    `gorm:"size:8" json:"kind"`
	RefID     uint64    `json:"ref_id"`
	TS        time.Time `json:"ts"`
	CPUPct    float64   `json:"cpu_pct"`
	MemUsed   int64     `json:"mem_used"`
	MemLimit  int64     `json:"mem_limit"`
	NetRx     int64     `json:"net_rx"`
	NetTx     int64     `json:"net_tx"`
	DiskRead  int64     `json:"disk_read"`
	DiskWrite int64     `json:"disk_write"`
}

func (MetricSample) TableName() string { return "metric_samples" }

type OperationLog struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	OrgID      *uint64   `json:"org_id"`
	UserID     *uint64   `json:"user_id"`
	Module     string    `gorm:"size:32" json:"module"`
	Action     string    `gorm:"size:64" json:"action"`
	TargetType string    `gorm:"size:32" json:"target_type"`
	TargetID   string    `gorm:"size:64" json:"target_id"`
	TargetName string    `gorm:"size:128" json:"target_name"`
	Result     int8      `json:"result"`
	DurationMs int64     `json:"duration_ms"`
	IP         string    `gorm:"size:64" json:"ip"`
	CreatedAt  time.Time `json:"created_at"`
}

func (OperationLog) TableName() string { return "operation_logs" }

type Notification struct {
	ID        uint64     `gorm:"primaryKey" json:"id"`
	UserID    uint64     `json:"user_id"`
	OrgID     *uint64    `json:"org_id"`
	Type      string     `gorm:"size:32" json:"type"`
	Title     string     `gorm:"size:128" json:"title"`
	Content   string     `gorm:"size:512" json:"content"`
	Link      string     `gorm:"size:255" json:"link"`
	ReadAt    *time.Time `json:"read_at"`
	CreatedAt time.Time  `json:"created_at"`
}

func (Notification) TableName() string { return "notifications" }
