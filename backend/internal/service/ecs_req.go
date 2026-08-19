package service

import (
	"fmt"
	"regexp"
	"strings"
)

// MountReq 云磁盘挂载请求（VolumeID = 磁盘 DB id）。
type MountReq struct {
	VolumeID uint64 `json:"volume_id"`
	Target   string `json:"target"`
	ReadOnly bool   `json:"read_only"`
}

// CreateReq ECS 创建请求（校验后进入 Docker CreateSpec）。
type CreateReq struct {
	Name           string
	Description    string
	Image          string
	CPU            float64
	MemoryMB       int64
	DiskGB         int64
	Ports          []PortMapping
	Env            []string
	Command        []string
	RestartPolicy  string
	ReadonlyRootfs bool
	NetworkID      string // 云网络 DB id（十进制字符串）
	FixedIP        string
	Mounts         []MountReq
	OrgID          *uint64 // 租户归属（nil = 单租户兼容模式）
}

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]{1,63}$`)

func (r *CreateReq) Validate() error {
	if !namePattern.MatchString(r.Name) {
		return fmt.Errorf("实例名称需 2-64 位，仅字母/数字/._-")
	}
	if strings.TrimSpace(r.Image) == "" {
		return fmt.Errorf("镜像不能为空")
	}
	if len(r.Image) > 255 {
		return fmt.Errorf("镜像名过长")
	}
	if r.CPU < 0.25 || r.CPU > 16 {
		return fmt.Errorf("CPU 需在 0.25-16 核之间")
	}
	if r.MemoryMB < 128 || r.MemoryMB > 32*1024 {
		return fmt.Errorf("内存需在 128MB-32GB 之间")
	}
	if r.DiskGB <= 0 || r.DiskGB > 500 {
		return fmt.Errorf("磁盘需在 1-500GB 之间（逻辑配额）")
	}
	for _, e := range r.Env {
		if !strings.Contains(e, "=") {
			return fmt.Errorf("环境变量需为 KEY=VALUE 格式：%s", e)
		}
	}
	switch r.RestartPolicy {
	case "", "no", "always", "unless-stopped", "on-failure":
	default:
		return fmt.Errorf("restart_policy 非法：%s", r.RestartPolicy)
	}
	if r.RestartPolicy == "" {
		r.RestartPolicy = "no"
	}
	for _, p := range r.Ports {
		if p.ContainerPort <= 0 || p.ContainerPort > 65535 {
			return fmt.Errorf("容器端口非法：%d", p.ContainerPort)
		}
		if p.Protocol != "" && p.Protocol != "tcp" && p.Protocol != "udp" {
			return fmt.Errorf("端口协议非法：%s", p.Protocol)
		}
	}
	for _, m := range r.Mounts {
		if m.VolumeID == 0 || m.Target == "" || !strings.HasPrefix(m.Target, "/") {
			return fmt.Errorf("挂载非法：volume_id=%d target=%q（target 需绝对路径）", m.VolumeID, m.Target)
		}
	}
	return nil
}
