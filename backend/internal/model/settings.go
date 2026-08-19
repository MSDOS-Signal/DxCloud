package model

import "time"

// SystemSetting 系统设置（system_settings 表，key 唯一）。
// value 以 JSON 字符串存储，保持与 MySQL JSON 列兼容。
type SystemSetting struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Key         string    `gorm:"column:key;size:128;uniqueIndex" json:"key"`
	Value       string    `gorm:"column:value;type:json" json:"value"`
	Description string    `gorm:"size:255" json:"description"`
	UpdatedBy   *uint64   `gorm:"column:updated_by" json:"updated_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SystemSetting) TableName() string { return "system_settings" }

// 设置键常量
const (
	// SettingRegion 区域：cn=中国大陆（拉取官方镜像自动走加速镜像源）；global=非中国大陆（直连 Docker Hub）
	SettingRegion = "region"
	// SettingRegistryMirror 中国大陆区域使用的镜像加速源域名（如 hub.rat.dev）
	SettingRegistryMirror = "registry_mirror"
)

// 区域取值
const (
	RegionCN     = "cn"
	RegionGlobal = "global"
)
