package docker

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/dxcloud/cloud-api/internal/config"
)

// DockerProvider 基于 Docker Engine API 的计算实现。
// 安全基线（架构文档 §12 L4，代码级强制）：
//
//	非特权、no-new-privileges、CapDrop ALL + 最小 CapAdd、PidsLimit、CPU/内存上限（配额钳制）、swap=0。
type DockerProvider struct {
	cli *client.Client
	// seccompProfile 自定义 seccomp profile 名（空 = 守护进程默认 profile）
	seccompProfile string
}

func NewDockerProvider(cfg *config.Config) (*DockerProvider, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("docker engine unreachable (DOCKER_HOST=%s): %w", cfg.DockerHost, err)
	}
	return &DockerProvider{cli: cli, seccompProfile: cfg.SeccompProfile}, nil
}

var defaultCapAdd = []string{"NET_BIND_SERVICE", "CHOWN", "SETUID", "SETGID", "DAC_OVERRIDE"}

// securityOpts 安全选项：no-new-privileges 恒开；配置了自定义 seccomp profile 时附加（绝不 unconfined）。
func (d *DockerProvider) securityOpts() []string {
	opts := []string{"no-new-privileges"}
	if d.seccompProfile != "" {
		opts = append(opts, "seccomp="+d.seccompProfile)
	}
	return opts
}

func restartPolicyFrom(p string) container.RestartPolicy {
	// 默认禁止自动重启；允许 no/always/unless-stopped/on-failure
	switch p {
	case "always":
		return container.RestartPolicy{Name: "always"}
	case "unless-stopped":
		return container.RestartPolicy{Name: "unless-stopped"}
	case "on-failure":
		return container.RestartPolicy{Name: "on-failure", MaximumRetryCount: 5}
	default:
		return container.RestartPolicy{Name: "no"}
	}
}

func buildPortBindings(ports []PortMapping) nat.PortMap {
	pm := nat.PortMap{}
	for _, p := range ports {
		proto := p.Protocol
		if proto == "" {
			proto = "tcp"
		}
		key := nat.Port(fmt.Sprintf("%d/%s", p.ContainerPort, proto))
		pm[key] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: fmt.Sprintf("%d", p.HostPort)}}
	}
	return pm
}

// Create 创建并启动容器（docker create + docker start）。
func (d *DockerProvider) Create(ctx context.Context, spec CreateSpec) (Info, error) {
	if spec.CPU <= 0 {
		spec.CPU = 1
	}
	if spec.MemoryMB <= 0 {
		spec.MemoryMB = 512
	}
	memBytes := spec.MemoryMB * 1024 * 1024

	cfg := &container.Config{
		Image:      spec.Image,
		Env:        spec.Env,
		Cmd:        strslice.StrSlice(spec.Cmd),
		Entrypoint: strslice.StrSlice(spec.Entrypoint),
		WorkingDir: spec.WorkingDir,
		Labels:     spec.Labels,
	}

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			NanoCPUs:   int64(spec.CPU * 1e9),
			Memory:     memBytes,
			MemorySwap: memBytes, // swap = 0，防内存偷跑
			PidsLimit:  intPtr(256),
		},
		CapDrop:       []string{"ALL"},
		CapAdd:        defaultCapAdd,
		SecurityOpt:   d.securityOpts(),
		RestartPolicy: restartPolicyFrom(spec.RestartPolicy),
		PortBindings:  buildPortBindings(spec.Ports),
	}
	if spec.NetworkID != "" {
		hostCfg.NetworkMode = container.NetworkMode(spec.NetworkID)
	}
	if spec.ReadonlyRootfs {
		hostCfg.ReadonlyRootfs = true
		hostCfg.Tmpfs = map[string]string{"/tmp": "rw"}
	}
	for _, m := range spec.Mounts {
		hostCfg.Mounts = append(hostCfg.Mounts, mount.Mount{
			Type:     mount.TypeVolume,
			Source:   m.VolumeName,
			Target:   m.Target,
			ReadOnly: m.ReadOnly,
		})
	}

	// 静态 IP：加入自定义网络时通过 networkingConfig 指定 IPAM 地址
	var networkingConfig *network.NetworkingConfig
	if spec.NetworkID != "" && spec.FixedIP != "" {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				spec.NetworkID: {
					IPAMConfig: &network.EndpointIPAMConfig{IPv4Address: spec.FixedIP},
				},
			},
		}
	}

	created, err := d.cli.ContainerCreate(ctx, cfg, hostCfg, networkingConfig, nil, spec.Name)
	if err != nil {
		return Info{}, fmt.Errorf("container create: %w", err)
	}
	if err := d.cli.ContainerStart(ctx, created.ID, container.StartOptions{}); err != nil {
		// 启动失败：清理残留，保持资源一致
		_ = d.cli.ContainerRemove(ctx, created.ID, container.RemoveOptions{Force: true})
		return Info{}, fmt.Errorf("container start: %w", err)
	}
	return d.Inspect(ctx, created.ID)
}

func intPtr(v int64) *int64 { return &v }

func (d *DockerProvider) Start(ctx context.Context, id string) error {
	return d.cli.ContainerStart(ctx, id, container.StartOptions{})
}

func (d *DockerProvider) Stop(ctx context.Context, id string, force bool) error {
	if force {
		return d.cli.ContainerKill(ctx, id, "SIGKILL")
	}
	timeout := 10
	return d.cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
}

func (d *DockerProvider) Restart(ctx context.Context, id string) error {
	return d.cli.ContainerRestart(ctx, id, container.StopOptions{})
}

func (d *DockerProvider) Remove(ctx context.Context, id string, force bool) error {
	return d.cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: force})
}

func (d *DockerProvider) Exists(ctx context.Context, id string) (bool, error) {
	_, err := d.cli.ContainerInspect(ctx, id)
	if err == nil {
		return true, nil
	}
	if client.IsErrNotFound(err) {
		return false, nil
	}
	return false, err
}

func (d *DockerProvider) Inspect(ctx context.Context, id string) (Info, error) {
	j, err := d.cli.ContainerInspect(ctx, id)
	if err != nil {
		return Info{}, err
	}
	return infoFromJSON(j), nil
}

func infoFromJSON(j types.ContainerJSON) Info {
	info := Info{
		ID:         j.ID,
		Name:       trimName(j.Name),
		Status:     j.State.Status,
		Running:    j.State.Running,
		ExitCode:   j.State.ExitCode,
		Image:      j.Config.Image,
		Env:        j.Config.Env,
		Cmd:        j.Config.Cmd,
		Entrypoint: j.Config.Entrypoint,
		Memory:     j.HostConfig.Memory,
		Labels:     j.Config.Labels,
	}
	// 自定义网络下 IPAddress 为空，从第一个网络端点取
	info.IP = j.NetworkSettings.IPAddress
	if info.IP == "" {
		for _, ep := range j.NetworkSettings.Networks {
			if ep.IPAddress != "" {
				info.IP = ep.IPAddress
				break
			}
		}
	}
	if j.HostConfig.NanoCPUs > 0 {
		info.NanoCPUs = j.HostConfig.NanoCPUs
	}
	if j.HostConfig.PidsLimit != nil {
		info.PidsLimit = int(*j.HostConfig.PidsLimit)
	}
	info.RestartPolicy = string(j.HostConfig.RestartPolicy.Name)
	info.Ports = map[string][]PortBinding{}
	for k, bindings := range j.HostConfig.PortBindings {
		list := make([]PortBinding, 0, len(bindings))
		for _, b := range bindings {
			list = append(list, PortBinding{HostIP: b.HostIP, HostPort: b.HostPort})
		}
		info.Ports[string(k)] = list
	}
	if t, err := time.Parse(time.RFC3339Nano, j.Created); err == nil {
		info.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339Nano, j.State.StartedAt); err == nil {
		info.StartedAt = t
	}
	return info
}

func trimName(name string) string {
	if len(name) > 0 && name[0] == '/' {
		return name[1:]
	}
	return name
}

func (d *DockerProvider) Logs(ctx context.Context, id string, tail int) (string, error) {
	if tail <= 0 {
		tail = 200
	}
	if tail > 2000 {
		tail = 2000
	}
	rc, err := d.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
		Timestamps: true,
	})
	if err != nil {
		return "", err
	}
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := stdcopy.StdCopy(&buf, &buf, rc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// LogsRaw 无时间戳的原始日志（二进制流场景，如 workspace tar 打包）。
func (d *DockerProvider) LogsRaw(ctx context.Context, id string) (string, error) {
	rc, err := d.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: false,
		Tail:       "all",
	})
	if err != nil {
		return "", err
	}
	defer rc.Close()
	var buf bytes.Buffer
	if _, err := stdcopy.StdCopy(&buf, &buf, rc); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// LogsFollow 流式日志（demux 后的合并流，调用方负责 Close）。
func (d *DockerProvider) LogsFollow(ctx context.Context, id string) (io.ReadCloser, error) {
	rc, err := d.cli.ContainerLogs(ctx, id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       "all",
		Timestamps: true,
	})
	if err != nil {
		return nil, err
	}
	pr, pw := io.Pipe()
	go func() {
		defer rc.Close()
		defer pw.Close()
		_, _ = stdcopy.StdCopy(pw, pw, rc)
	}()
	return pr, nil
}

// WaitContainer 等待容器退出并返回退出码。
func (d *DockerProvider) WaitContainer(ctx context.Context, id string) (int, error) {
	statusCh, errCh := d.cli.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		return -1, err
	case status := <-statusCh:
		return int(status.StatusCode), nil
	case <-ctx.Done():
		return -1, ctx.Err()
	}
}

// statsJSON 自定义镜像 Docker StatsJSON 的关键字段，避免 SDK 版本间类型迁移问题。
type statsJSON struct {
	CPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemUsage uint64 `json:"system_usage"`
		OnlineCPUs  uint32 `json:"online_cpus"`
	} `json:"cpu_stats"`
	PreCPUStats struct {
		CPUUsage struct {
			TotalUsage uint64 `json:"total_usage"`
		} `json:"cpu_usage"`
		SystemUsage uint64 `json:"system_usage"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage uint64 `json:"usage"`
		Limit uint64 `json:"limit"`
	} `json:"memory_stats"`
	PidsStats struct {
		Current uint64 `json:"current"`
	} `json:"pids_stats"`
	Networks map[string]struct {
		RxBytes uint64 `json:"rx_bytes"`
		TxBytes uint64 `json:"tx_bytes"`
	} `json:"networks"`
	BlkioStats struct {
		IoServiceBytesRecursive []struct {
			Op    string `json:"op"`
			Value uint64 `json:"value"`
		} `json:"io_service_bytes_recursive"`
	} `json:"blkio_stats"`
}

func (d *DockerProvider) readStats(ctx context.Context, id string) (statsJSON, error) {
	r, err := d.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return statsJSON{}, err
	}
	defer r.Body.Close()
	var s statsJSON
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		return statsJSON{}, err
	}
	return s, nil
}

// StatsOneShot 两帧采样（间隔 1s）计算 CPU%。
func (d *DockerProvider) StatsOneShot(ctx context.Context, id string) (Stats, error) {
	s1, err := d.readStats(ctx, id)
	if err != nil {
		return Stats{}, err
	}
	time.Sleep(time.Second)
	s2, err := d.readStats(ctx, id)
	if err != nil {
		return Stats{}, err
	}

	out := Stats{
		MemUsed:  s2.MemoryStats.Usage,
		MemLimit: s2.MemoryStats.Limit,
		PIDs:     s2.PidsStats.Current,
	}
	if s2.MemoryStats.Limit > 0 {
		out.MemPercent = float64(s2.MemoryStats.Usage) / float64(s2.MemoryStats.Limit) * 100
	}
	cpuDelta := s2.CPUStats.CPUUsage.TotalUsage - s1.CPUStats.CPUUsage.TotalUsage
	sysDelta := s2.CPUStats.SystemUsage - s1.CPUStats.SystemUsage
	if sysDelta > 0 && cpuDelta > 0 {
		online := s2.CPUStats.OnlineCPUs
		if online == 0 {
			online = 1
		}
		out.CPUPercent = float64(cpuDelta) / float64(sysDelta) * float64(online) * 100
	}
	for _, n := range s2.Networks {
		out.NetRxBytes += n.RxBytes
		out.NetTxBytes += n.TxBytes
	}
	for _, io := range s2.BlkioStats.IoServiceBytesRecursive {
		switch io.Op {
		case "Read":
			out.DiskReadBytes += io.Value
		case "Write":
			out.DiskWrite += io.Value
		}
	}
	return out, nil
}

// SystemInfo 返回 Docker Engine 的版本、API 版本与宿主机资源信息。
func (d *DockerProvider) SystemInfo(ctx context.Context) (SystemInfo, error) {
	version, err := d.cli.ServerVersion(ctx)
	if err != nil {
		return SystemInfo{}, err
	}
	info, err := d.cli.Info(ctx)
	if err != nil {
		return SystemInfo{}, err
	}
	return SystemInfo{
		DockerVersion:    version.Version,
		DockerAPIVersion: version.APIVersion,
		OS:               info.OperatingSystem,
		Arch:             info.Architecture,
		Kernel:           info.KernelVersion,
		CPUCount:         info.NCPU,
		MemTotal:         info.MemTotal,
	}, nil
}

func (d *DockerProvider) ListByLabels(ctx context.Context, labels map[string]string) ([]Info, error) {
	flt := filters.NewArgs()
	for k, v := range labels {
		flt.Add("label", k+"="+v)
	}
	list, err := d.cli.ContainerList(ctx, container.ListOptions{All: true, Filters: flt})
	if err != nil {
		return nil, err
	}
	out := make([]Info, 0, len(list))
	for _, c := range list {
		info := Info{
			ID:     c.ID,
			Name:   stringsJoin(c.Names, ","),
			Status: c.State,
			Image:  c.Image,
			Labels: c.Labels,
		}
		out = append(out, info)
	}
	return out, nil
}

func stringsJoin(ss []string, sep string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += sep
		}
		out += trimName(s)
	}
	return out
}

// UsedHostPorts 汇总当前全部容器占用的宿主端口（端口冲突检测的运行时事实源）。
func (d *DockerProvider) UsedHostPorts(ctx context.Context) (map[int]bool, error) {
	list, err := d.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	used := map[int]bool{}
	for _, c := range list {
		j, err := d.cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		for _, bindings := range j.HostConfig.PortBindings {
			for _, b := range bindings {
				var p int
				if _, err := fmt.Sscanf(b.HostPort, "%d", &p); err == nil && p > 0 {
					used[p] = true
				}
			}
		}
	}
	return used, nil
}

// SecurityAudit 全部容器安全态势（基线审计：特权/能力/只读根/资源上限等）。
func (d *DockerProvider) SecurityAudit(ctx context.Context) ([]ContainerSecurity, error) {
	list, err := d.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	out := make([]ContainerSecurity, 0, len(list))
	for _, c := range list {
		j, err := d.cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		cs := ContainerSecurity{
			ID:             c.ID,
			Name:           stringsJoin(c.Names, ","),
			Image:          c.Image,
			Status:         c.State,
			Kind:           c.Labels["com.dxcloud.kind"],
			Privileged:     j.HostConfig.Privileged,
			CapAdd:         j.HostConfig.CapAdd,
			CapDrop:        j.HostConfig.CapDrop,
			ReadonlyRootfs: j.HostConfig.ReadonlyRootfs,
			User:           j.Config.User,
			SecurityOpt:    j.HostConfig.SecurityOpt,
			MemoryLimit:    j.HostConfig.Memory,
			NanoCPUs:       j.HostConfig.NanoCPUs,
		}
		if j.HostConfig.PidsLimit != nil {
			cs.PidsLimit = int(*j.HostConfig.PidsLimit)
		}
		for _, so := range j.HostConfig.SecurityOpt {
			if so == "no-new-privileges" {
				cs.NoNewPrivileges = true
				break
			}
		}
		out = append(out, cs)
	}
	return out, nil
}

func (d *DockerProvider) PullImage(ctx context.Context, ref string) error {
	return d.PullImageWithProgress(ctx, ref, nil)
}

// PullImageWithProgress 拉取镜像并解析 Docker Engine 的 JSON Lines 进度。
func (d *DockerProvider) PullImageWithProgress(ctx context.Context, ref string, progress func(ImagePullProgress)) error {
	rc, err := d.cli.ImagePull(ctx, ref, image.PullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var lastErr error
	for scanner.Scan() {
		var line struct {
			Status      string `json:"status"`
			Progress    string `json:"progress"`
			Error       string `json:"error"`
			ErrorDetail struct {
				Message string `json:"message"`
			} `json:"errorDetail"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}
		if progress != nil {
			progress(ImagePullProgress{
				Status:   line.Status,
				Progress: line.Progress,
				Error:    line.Error,
			})
		}
		if line.Error != "" {
			lastErr = errors.New(line.Error)
		}
		if line.ErrorDetail.Message != "" {
			lastErr = errors.New(line.ErrorDetail.Message)
		}
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	return lastErr
}

// ImageBuild 服务端 BuildKit 构建（contextTar = 构建上下文 tar 流，调用方负责 Close 返回流）。
func (d *DockerProvider) ImageBuild(ctx context.Context, contextTar io.Reader, dockerfile string, tags []string) (io.ReadCloser, error) {
	resp, err := d.cli.ImageBuild(ctx, contextTar, types.ImageBuildOptions{
		Dockerfile:  dockerfile,
		Tags:        tags,
		Remove:      true,
		ForceRemove: true,
		NetworkMode: "default",
	})
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// ImagePush 推送镜像（无鉴权私有 Registry；外置源需 daemon 配置凭据）。
func (d *DockerProvider) ImagePush(ctx context.Context, ref string) error {
	rc, err := d.cli.ImagePush(ctx, ref, image.PushOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(rc)
	return err
}

// ---------- Exec（Web Terminal） ----------

// ExecCreate 在容器内创建交互式 exec。
// 注意：ContainerExecCreate 不校验二进制是否存在（启动时才报错），
// 因此先探测 /bin/bash，不存在则回退 /bin/sh（绝大多数 Linux 镜像均有 sh）。
func (d *DockerProvider) ExecCreate(ctx context.Context, containerID string) (string, error) {
	shell := "/bin/sh"
	if ok, _ := d.binaryExists(ctx, containerID, "/bin/bash"); ok {
		shell = "/bin/bash"
	}
	resp, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{shell},
	})
	if err != nil {
		return "", fmt.Errorf("exec create: %w", err)
	}
	return resp.ID, nil
}

// binaryExists 通过非交互 exec（test -x）探测容器内二进制是否存在。
func (d *DockerProvider) binaryExists(ctx context.Context, containerID, path string) (bool, error) {
	create, err := d.cli.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		Cmd: []string{"test", "-x", path},
	})
	if err != nil {
		return false, err
	}
	if err := d.cli.ContainerExecStart(ctx, create.ID, types.ExecStartCheck{}); err != nil {
		return false, err
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		info, err := d.cli.ContainerExecInspect(ctx, create.ID)
		if err != nil {
			return false, err
		}
		if !info.Running {
			return info.ExitCode == 0, nil
		}
		if time.Now().After(deadline) {
			return false, fmt.Errorf("timeout probing %s", path)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// ExecAttach 挂接 exec 会话（TTY：stdout/stderr 混流在 PTY，无需 stdcopy 解复用）。
func (d *DockerProvider) ExecAttach(ctx context.Context, containerID, execID string) (*ExecSession, error) {
	hj, err := d.cli.ContainerExecAttach(ctx, execID, types.ExecStartCheck{Tty: true})
	if err != nil {
		return nil, fmt.Errorf("exec attach: %w", err)
	}
	return &ExecSession{
		Conn:   hj.Conn,
		Reader: hj.Reader,
		Close: func() error {
			hj.Close()
			return nil
		},
	}, nil
}

func (d *DockerProvider) ExecResize(ctx context.Context, execID string, cols, rows int) error {
	return d.cli.ContainerExecResize(ctx, execID, container.ResizeOptions{Height: uint(rows), Width: uint(cols)})
}
