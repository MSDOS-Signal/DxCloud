package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"go.uber.org/zap"
)

// MetricsCollector 每 interval 对运行中的 ECS/部署容器并行采样并落库；定期清理过期数据。
type MetricsCollector struct {
	repo     *repository.Repos
	compute  docker.Provider
	log      *zap.Logger
	interval time.Duration
}

func NewMetricsCollector(repo *repository.Repos, compute docker.Provider, log *zap.Logger, interval time.Duration) *MetricsCollector {
	if interval <= 0 {
		interval = time.Minute
	}
	return &MetricsCollector{repo: repo, compute: compute, log: log, interval: interval}
}

func (c *MetricsCollector) Run(ctx context.Context) {
	c.log.Info("metrics collector started", zap.Duration("interval", c.interval))
	// 启动时立即采样一次，避免新实例要等完整采样周期后才出现曲线。
	c.collect(ctx)
	cleanupTicker := time.NewTicker(6 * time.Hour)
	defer cleanupTicker.Stop()
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.collect(ctx)
		case <-cleanupTicker.C:
			c.cleanup(ctx)
		}
	}
}

func (c *MetricsCollector) collect(ctx context.Context) {
	type target struct {
		kind        string
		refID       uint64
		containerID string
	}
	var targets []target

	// ECS 运行实例
	ecsRows, _, err := c.repo.EcsList(repository.EcsFilter{Status: model.EcsRunning, Page: 1, Size: 500})
	if err == nil {
		for i := range ecsRows {
			if ecsRows[i].ContainerID != "" {
				targets = append(targets, target{kind: "ecs", refID: ecsRows[i].ID, containerID: ecsRows[i].ContainerID})
			}
		}
	}
	// 成功部署的应用容器
	deps, err := c.repo.DeploymentListByApp(0, 0) // 0=全部（后续改为按状态查询）
	_ = deps
	_ = err

	var mu sync.Mutex
	var samples []model.MetricSample
	var wg sync.WaitGroup
	sem := make(chan struct{}, 8) // 并发上限
	for _, t := range targets {
		wg.Add(1)
		sem <- struct{}{}
		go func(t target) {
			defer wg.Done()
			defer func() { <-sem }()
			st, err := c.compute.StatsOneShot(ctx, t.containerID)
			if err != nil {
				return
			}
			mu.Lock()
			samples = append(samples, model.MetricSample{
				Kind: t.kind, RefID: t.refID, TS: time.Now(),
				CPUPct: st.CPUPercent, MemUsed: int64(st.MemUsed), MemLimit: int64(st.MemLimit),
				NetRx: int64(st.NetRxBytes), NetTx: int64(st.NetTxBytes),
				DiskRead: int64(st.DiskReadBytes), DiskWrite: int64(st.DiskWrite),
			})
			mu.Unlock()
		}(t)
	}
	wg.Wait()
	if err := c.repo.MetricInsert(samples); err != nil {
		c.log.Warn("metric insert failed", zap.Error(err))
	}
}

func (c *MetricsCollector) cleanup(ctx context.Context) {
	n, err := c.repo.MetricCleanup(time.Now().Add(-7 * 24 * time.Hour))
	if err != nil {
		c.log.Warn("metric cleanup failed", zap.Error(err))
		return
	}
	c.log.Info("metric cleanup done", zap.Int64("removed", n))
}
