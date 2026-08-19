// Package docker 是 Compute Provider 抽象层：
// 业务层只依赖 Provider 接口，DockerProvider 是当前实现；
// 未来 Kubernetes/Podman/OpenStack 只需新增实现（架构文档 §3.2）。
package docker

import (
	"bufio"
	"context"
	"io"
	"net"
	"time"
)

// PortMapping 端口映射（云概念：公网入口 = 宿主端口发布）。
type PortMapping struct {
	ContainerPort int    `json:"container_port"`
	HostPort      int    `json:"host_port"`
	Protocol      string `json:"protocol"` // tcp/udp，默认 tcp
}

// Mount 挂载（云磁盘 → Docker Volume，Phase 5 接入 UI）。
type Mount struct {
	VolumeName string `json:"volume_name"`
	Target     string `json:"target"`
	ReadOnly   bool   `json:"read_only"`
}

// CreateSpec 创建规格（云规格 → Docker HostConfig 映射）。
type CreateSpec struct {
	Name           string
	Image          string
	CPU            float64 // 核
	MemoryMB       int64
	Env            []string
	Cmd            []string
	Entrypoint     []string
	WorkingDir     string
	Ports          []PortMapping
	Mounts         []Mount
	RestartPolicy  string // no/always/unless-stopped/on-failure
	NetworkID      string // 空 = 默认 bridge
	FixedIP        string
	ReadonlyRootfs bool
	Labels         map[string]string
}

// Info 容器实况（对账用）。
type Info struct {
	ID            string
	Name          string
	Status        string // created/running/exited/restarting/...
	Running       bool
	ExitCode      int
	Image         string
	Env           []string
	Cmd           []string
	Entrypoint    []string
	NanoCPUs      int64
	Memory        int64
	IP            string
	Ports         map[string][]PortBinding
	Labels        map[string]string
	RestartPolicy string
	CreatedAt     time.Time
	StartedAt     time.Time
	PidsLimit     int
}

type PortBinding struct {
	HostIP   string `json:"host_ip"`
	HostPort string `json:"host_port"`
}

// Stats 单次采样统计（两帧差值计算 CPU%）。
type Stats struct {
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

// SystemInfo Docker Engine 与宿主机的版本/资源信息。
type SystemInfo struct {
	DockerVersion    string
	DockerAPIVersion string
	OS               string
	Arch             string
	Kernel           string
	CPUCount         int
	MemTotal         int64
}

// ContainerSecurity 容器安全态势（基线审计）。
type ContainerSecurity struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Image           string   `json:"image"`
	Status          string   `json:"status"`
	Kind            string   `json:"kind"`
	Privileged      bool     `json:"privileged"`
	NoNewPrivileges bool     `json:"no_new_privileges"`
	CapAdd          []string `json:"cap_add"`
	CapDrop         []string `json:"cap_drop"`
	ReadonlyRootfs  bool     `json:"readonly_rootfs"`
	PidsLimit       int      `json:"pids_limit"`
	MemoryLimit     int64    `json:"memory_limit"`
	NanoCPUs        int64    `json:"nano_cpus"`
	User            string   `json:"user"`
	SecurityOpt     []string `json:"security_opt"`
}

// ExecSession 容器内交互会话（PTY 双向流 + 关闭函数）。
type ExecSession struct {
	Conn   net.Conn
	Reader *bufio.Reader
	Close  func() error
}

// ImageInfo 镜像信息。
type ImageInfo struct {
	ID          string
	RepoTags    []string
	RepoDigests []string
	Size        int64
	CreatedAt   time.Time
}

// ImagePullProgress 镜像拉取进度行。Docker Engine 返回的 JSON Lines
// 中 status 表示阶段，progress 表示当前层/总量的进度。
type ImagePullProgress struct {
	Status   string
	Progress string
	Error    string
}

// NetworkSpec 云网络规格（子网/IPAM）。
type NetworkSpec struct {
	Name     string
	Driver   string
	Subnet   string
	Gateway  string
	IPRange  string
	Internal bool
	Labels   map[string]string
}

// NetworkInfo 网络实况。
type NetworkInfo struct {
	ID         string
	Name       string
	Driver     string
	Subnet     string
	Gateway    string
	IPRange    string
	Containers map[string]string // 容器名 → IP
}

// VolumeInfo 卷实况。
type VolumeInfo struct {
	Name       string
	Driver     string
	Mountpoint string
	Labels     map[string]string
}

// ComputeProvider 计算能力接口（未来 KubernetesProvider/PodmanProvider 实现同一接口）。
type ComputeProvider interface {
	Create(ctx context.Context, spec CreateSpec) (Info, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string, force bool) error
	Restart(ctx context.Context, id string) error
	Remove(ctx context.Context, id string, force bool) error
	Inspect(ctx context.Context, id string) (Info, error)
	Exists(ctx context.Context, id string) (bool, error)
	Logs(ctx context.Context, id string, tail int) (string, error)
	LogsRaw(ctx context.Context, id string) (string, error)
	LogsFollow(ctx context.Context, id string) (io.ReadCloser, error)
	WaitContainer(ctx context.Context, id string) (int, error)
	StatsOneShot(ctx context.Context, id string) (Stats, error)
	SystemInfo(ctx context.Context) (SystemInfo, error)
	ListByLabels(ctx context.Context, labels map[string]string) ([]Info, error)
	UsedHostPorts(ctx context.Context) (map[int]bool, error)
	SecurityAudit(ctx context.Context) ([]ContainerSecurity, error)
	PullImage(ctx context.Context, ref string) error
	// Exec（Web Terminal 底层：bash → sh 回退）
	ExecCreate(ctx context.Context, containerID string) (execID string, err error)
	ExecAttach(ctx context.Context, containerID, execID string) (*ExecSession, error)
	ExecResize(ctx context.Context, execID string, cols, rows int) error
}

// ImageProvider 镜像能力。
type ImageProvider interface {
	ListImages(ctx context.Context) ([]ImageInfo, error)
	RemoveImage(ctx context.Context, ref string, force bool) error
	TagImage(ctx context.Context, src, dst string) error
	InspectImage(ctx context.Context, ref string) (*ImageInfo, error)
	PullImage(ctx context.Context, ref string) error
	PullImageWithProgress(ctx context.Context, ref string, progress func(ImagePullProgress)) error
	ImageBuild(ctx context.Context, contextTar io.Reader, dockerfile string, tags []string) (io.ReadCloser, error)
	ImagePush(ctx context.Context, ref string) error
}

// NetworkProvider 网络能力（VPC 模拟：bridge + 自定义子网 IPAM）。
type NetworkProvider interface {
	CreateNetwork(ctx context.Context, spec NetworkSpec) (NetworkInfo, error)
	RemoveNetwork(ctx context.Context, id string) error
	InspectNetwork(ctx context.Context, id string) (*NetworkInfo, error)
	ListNetworks(ctx context.Context) ([]NetworkInfo, error)
	ConnectNetwork(ctx context.Context, netID, containerID, fixedIP string) error
	DisconnectNetwork(ctx context.Context, netID, containerID string) error
}

// StorageProvider 存储能力（云磁盘 → named volume）。
type StorageProvider interface {
	CreateVolume(ctx context.Context, name string, labels map[string]string) (VolumeInfo, error)
	RemoveVolume(ctx context.Context, name string) error
	InspectVolume(ctx context.Context, name string) (*VolumeInfo, error)
	ListVolumes(ctx context.Context) ([]VolumeInfo, error)
}

// Provider 聚合接口：DockerProvider 全部实现。
type Provider interface {
	ComputeProvider
	ImageProvider
	NetworkProvider
	StorageProvider
}
