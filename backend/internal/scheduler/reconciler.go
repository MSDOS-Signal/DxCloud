// Package scheduler：后台任务（Reconciler 对账循环、后续 metrics/计费 tick）。
package scheduler

import (
	"context"
	"time"

	"github.com/dxcloud/cloud-api/internal/service"
	"go.uber.org/zap"
)

// Reconciler 每 interval 触发一轮资源对账（简化版 Controller / Reconciliation Loop）。
type Reconciler struct {
	svc      *service.EcsService
	log      *zap.Logger
	interval time.Duration
}

func NewReconciler(svc *service.EcsService, log *zap.Logger, interval time.Duration) *Reconciler {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &Reconciler{svc: svc, log: log, interval: interval}
}

// Run 阻塞运行，ctx 取消后退出。
func (r *Reconciler) Run(ctx context.Context) {
	r.log.Info("reconciler started", zap.Duration("interval", r.interval))
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			r.log.Info("reconciler stopped")
			return
		case <-ticker.C:
			r.svc.Reconcile(ctx)
		}
	}
}
