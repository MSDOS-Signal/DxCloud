// Package runner：CI Job 执行器（Docker 一次性容器，隔离执行用户脚本）。
// 铁律：不挂 docker.sock；资源限制强制；超时 kill；workspace 独立卷；非特权。
package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"go.uber.org/zap"
)

type JobSpec struct {
	Image           string
	Cmd             []string
	WorkspaceVolume string // 挂载到 /workspace
	Env             []string
	CPU             float64
	MemoryMB        int64
	Timeout         time.Duration
	NetworkID       string // 应用网络（需外网时使用）
	Labels          map[string]string
}

type DockerJobRunner struct {
	compute docker.Provider
	log     *zap.Logger
}

func NewDockerJobRunner(compute docker.Provider, log *zap.Logger) *DockerJobRunner {
	return &DockerJobRunner{compute: compute, log: log}
}

// RunJob 在隔离容器中执行任务：
//   - 镜像缺失自动拉取（10 分钟超时）
//   - 容器限制：CPU/内存/PIDs/非特权/no-new-privileges（Provider 基线）
//   - 日志流式写入 logWriter；onContainer 在容器创建后回调（用于取消 kill）
//   - 完成后容器自动清理（workspace 卷保留）
func (r *DockerJobRunner) RunJob(ctx context.Context, name string, spec JobSpec, logWriter io.Writer, onContainer func(containerID string)) (exitCode int, err error) {
	// 1) 镜像就绪
	if _, err := r.compute.InspectImage(ctx, spec.Image); err != nil {
		pullCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
		defer cancel()
		if perr := r.compute.PullImage(pullCtx, spec.Image); perr != nil {
			return -1, fmt.Errorf("pull image %s: %w", spec.Image, perr)
		}
	}
	if spec.CPU <= 0 {
		spec.CPU = 2
	}
	if spec.MemoryMB <= 0 {
		spec.MemoryMB = 2048
	}
	if spec.Timeout <= 0 {
		spec.Timeout = 30 * time.Minute
	}

	// 2) 创建 + 启动
	createSpec := docker.CreateSpec{
		Name:     name,
		Image:    spec.Image,
		CPU:      spec.CPU,
		MemoryMB: spec.MemoryMB,
		Env:      spec.Env,
		Cmd:      spec.Cmd,
		WorkingDir: "/workspace",
		Mounts: []docker.Mount{{
			VolumeName: spec.WorkspaceVolume,
			Target:     "/workspace",
		}},
		NetworkID:     spec.NetworkID,
		RestartPolicy: "no",
		Labels:        spec.Labels,
	}
	info, err := r.compute.Create(ctx, createSpec)
	if err != nil {
		return -1, fmt.Errorf("create job container: %w", err)
	}
	if onContainer != nil {
		onContainer(info.ID)
	}

	// 3) 日志流式落盘
	rc, err := r.compute.LogsFollow(ctx, info.ID)
	if err == nil {
		go func() {
			_, _ = io.Copy(logWriter, rc)
			rc.Close()
		}()
	} else {
		r.log.Warn("job log follow failed", zap.Error(err))
	}

	// 4) 等待退出（超时 kill）
	waitCtx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	type waitResult struct {
		code int
		err  error
	}
	ch := make(chan waitResult, 1)
	go func() {
		code, werr := r.compute.WaitContainer(waitCtx, info.ID)
		ch <- waitResult{code: code, err: werr}
	}()

	select {
	case res := <-ch:
		exitCode = res.code
		err = res.err
	case <-waitCtx.Done():
		_ = r.compute.Stop(waitCtx, info.ID, true)
		<-ch
		return -1, waitCtx.Err() // context.Canceled（取消）或 DeadlineExceeded（超时）
	}

	// 5) 清理容器（保留 workspace 卷）
	_ = r.compute.Remove(context.Background(), info.ID, true)
	return exitCode, err
}

// Kill 取消时强杀运行中的 Job 容器。
func (r *DockerJobRunner) Kill(containerID string) {
	if containerID == "" {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = r.compute.Stop(ctx, containerID, true)
}

// LogFile 创建日志文件（parent 目录自动创建）。
func LogFile(path string) (*os.File, error) {
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return nil, err
	}
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}
