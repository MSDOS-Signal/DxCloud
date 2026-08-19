// Package handler 存放 HTTP 处理器（按模块拆分，本阶段仅 health）。
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db      *gorm.DB
	rdb     *redis.Client
	compute docker.ComputeProvider
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client, compute docker.ComputeProvider) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb, compute: compute}
}

type dependencyStatus struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthData struct {
	Status           string           `json:"status"`
	DB               dependencyStatus `json:"db"`
	Redis            dependencyStatus `json:"redis"`
	Time             string           `json:"time"`
	Version          string           `json:"version"`
	DockerVersion    string           `json:"docker_version"`
	DockerAPIVersion string           `json:"docker_api_version"`
	OS               string           `json:"os"`
	Arch             string           `json:"arch"`
	Kernel           string           `json:"kernel"`
	CPUCount         int              `json:"cpu_count"`
	MemTotal         int64            `json:"mem_total"`
}

// collect 探测 MySQL 与 Redis。
func (h *HealthHandler) collect() HealthData {
	data := HealthData{Status: "ok", Time: time.Now().Format(time.RFC3339), Version: "DxCloud v1.0"}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if sqlDB, err := h.db.DB(); err != nil {
		data.DB = dependencyStatus{Status: "down", Error: err.Error()}
	} else if err := sqlDB.PingContext(ctx); err != nil {
		data.DB = dependencyStatus{Status: "down", Error: err.Error()}
	} else {
		data.DB = dependencyStatus{Status: "up"}
	}

	if err := h.rdb.Ping(ctx).Err(); err != nil {
		data.Redis = dependencyStatus{Status: "down", Error: err.Error()}
	} else {
		data.Redis = dependencyStatus{Status: "up"}
	}

	if data.DB.Status != "up" || data.Redis.Status != "up" {
		data.Status = "degraded"
	}
	if h.compute != nil {
		if info, err := h.compute.SystemInfo(ctx); err == nil {
			data.DockerVersion = info.DockerVersion
			data.DockerAPIVersion = info.DockerAPIVersion
			data.OS = info.OS
			data.Arch = info.Arch
			data.Kernel = info.Kernel
			data.CPUCount = info.CPUCount
			data.MemTotal = info.MemTotal
		}
	}
	return data
}

// Livez 运维探针（无统一响应包裹），供 compose / traefik 健康检查使用。
func (h *HealthHandler) Livez(c *gin.Context) {
	data := h.collect()
	status := http.StatusOK
	if data.Status != "ok" {
		status = http.StatusServiceUnavailable
	}
	c.JSON(status, data)
}

// Health 业务探针（统一响应结构）。
func (h *HealthHandler) Health(c *gin.Context) {
	resp.OK(c, h.collect())
}
