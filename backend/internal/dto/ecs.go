package dto

// ---------- ECS ----------

type PortMappingReq struct {
	ContainerPort int    `json:"container_port" binding:"required"`
	HostPort      int    `json:"host_port" binding:"required"`
	Protocol      string `json:"protocol"`
}

type MountReq struct {
	VolumeID uint64 `json:"volume_id" binding:"required"`
	Target   string `json:"target" binding:"required"`
	ReadOnly bool   `json:"read_only"`
}

type CreateEcsReq struct {
	Name           string           `json:"name" binding:"required"`
	Description    string           `json:"description"`
	Image          string           `json:"image" binding:"required"`
	CPU            float64          `json:"cpu"`
	MemoryMB       int64            `json:"memory_mb"`
	DiskGB         int64            `json:"disk_gb"`
	Ports          []PortMappingReq `json:"ports"`
	Env            []string         `json:"env"`
	Command        []string         `json:"command"`
	RestartPolicy  string           `json:"restart_policy"`
	ReadonlyRootfs bool             `json:"readonly_rootfs"`
	NetworkID      string           `json:"network_id"`
	FixedIP        string           `json:"fixed_ip"`
	Mounts         []MountReq       `json:"mounts"`
}

type UpdateEcsReq struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type PortMappingResp struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"`
}

type EcsInfo struct {
	ID             uint64            `json:"id"`
	InstanceNo     string            `json:"instance_no"`
	OrgID          *uint64           `json:"org_id"`
	OwnerID        uint64            `json:"owner_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Image          string            `json:"image"`
	CPU            float64           `json:"cpu"`
	MemoryMB       int64             `json:"memory_mb"`
	DiskGB         int64             `json:"disk_gb"`
	Ports          []PortMappingResp `json:"ports"`
	Env            []string          `json:"env"`
	Command        []string          `json:"command"`
	RestartPolicy  string            `json:"restart_policy"`
	ReadonlyRootfs bool              `json:"readonly_rootfs"`
	NetworkID      string            `json:"network_id"`
	FixedIP        string            `json:"fixed_ip"`
	Mounts         []MountResp       `json:"mounts"`
	DesiredState   string            `json:"desired_state"`
	ObservedState  string            `json:"observed_state"`
	ContainerID    string            `json:"container_id"`
	ContainerName  string            `json:"container_name"`
	LastError      string            `json:"last_error"`
	CreatedAt      string            `json:"created_at"`
}

type MountResp struct {
	VolumeName string `json:"volume_name"`
	Target     string `json:"target"`
	ReadOnly   bool   `json:"read_only"`
}

type EcsStatsResp struct {
	CPUPercent    float64 `json:"cpu_percent"`
	MemUsed       uint64  `json:"mem_used"`
	MemLimit      uint64  `json:"mem_limit"`
	MemPercent    float64 `json:"mem_percent"`
	NetRxBytes    uint64  `json:"net_rx"`
	NetTxBytes    uint64  `json:"net_tx"`
	DiskReadBytes uint64  `json:"disk_read"`
	DiskWrite     uint64  `json:"disk_write"`
	PIDs          uint64  `json:"pids"`
}

type EcsLogsResp struct {
	Logs string `json:"logs"`
	Tail int    `json:"tail"`
}
