package service

import (
	"context"
	"time"

	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
)

type MonitorService struct {
	repo *repository.Repos
}

func NewMonitorService(repo *repository.Repos) *MonitorService {
	return &MonitorService{repo: repo}
}

func (s *MonitorService) ecsFilter(ac AccessCtx) repository.EcsFilter {
	f := repository.EcsFilter{Page: 1, Size: 10000}
	if ac.OrgID > 0 {
		v := ac.OrgID
		f.OrgID = &v
		return f
	}
	f.DefaultSpace = true
	if !ac.canManage() {
		v := ac.UserID
		f.OwnerID = &v
	}
	return f
}

func (s *MonitorService) accessibleECS(ctx context.Context, ac AccessCtx) ([]model.EcsInstance, error) {
	rows, _, err := s.repo.EcsList(s.ecsFilter(ac))
	return rows, err
}

// Dashboard 首页聚合（实例/应用/Pipeline/部署/成功率/资源水位）。
func (s *MonitorService) Dashboard(ctx context.Context, ac AccessCtx) (map[string]any, error) {
	out := map[string]any{}
	rows, err := s.accessibleECS(ctx, ac)
	if err != nil {
		return nil, err
	}
	var ecsRunning, ecsStopped int64
	for i := range rows {
		switch rows[i].ObservedState {
		case model.EcsRunning:
			ecsRunning++
		case model.EcsStopped:
			ecsStopped++
		}
	}
	out["ecs_total"] = int64(len(rows))
	out["ecs_running"] = ecsRunning
	out["ecs_stopped"] = ecsStopped

	var appCount, pipeCount int64
	s.repo.DB.Model(&model.Application{}).Scopes(repository.ScopeByTenantOrNull(ac.OrgID)).Count(&appCount)
	s.repo.DB.Model(&model.Pipeline{}).Scopes(repository.ScopeByTenantOrNull(ac.OrgID)).Count(&pipeCount)
	out["app_count"] = appCount
	out["pipeline_count"] = pipeCount

	var deployToday int64
	s.repo.DB.Model(&model.Deployment{}).
		Scopes(repository.ScopeByTenantOrNull(ac.OrgID)).
		Where("created_at > ?", time.Now().Truncate(24*time.Hour)).
		Count(&deployToday)
	out["deploy_today"] = deployToday

	var runOK, runFail int64
	dayAgo := time.Now().Add(-24 * time.Hour)
	pipelineIDs := s.repo.DB.Model(&model.Pipeline{}).Select("id").Scopes(repository.ScopeByTenantOrNull(ac.OrgID))
	s.repo.DB.Model(&model.PipelineRun{}).
		Where("pipeline_id IN (?) AND created_at > ? AND status = ?", pipelineIDs, dayAgo, model.PipeSuccess).
		Count(&runOK)
	s.repo.DB.Model(&model.PipelineRun{}).
		Where("pipeline_id IN (?) AND created_at > ? AND status = ?", pipelineIDs, dayAgo, model.PipeFailed).
		Count(&runFail)
	rate := 100.0
	if runOK+runFail > 0 {
		rate = float64(runOK) / float64(runOK+runFail) * 100
	}
	out["pipeline_success_rate"] = rate
	out["pipeline_runs_24h"] = runOK + runFail

	refIDs := make([]uint64, 0, len(rows))
	for i := range rows {
		refIDs = append(refIDs, rows[i].ID)
	}
	cpuAvg, memAvg := s.currentUsage(ac, refIDs)
	out["cpu_avg"] = cpuAvg
	out["mem_avg"] = memAvg
	return out, nil
}

func (s *MonitorService) currentUsage(ac AccessCtx, refIDs []uint64) (float64, float64) {
	samples, err := s.repo.MetricLatestForRefs("ecs", 10*time.Minute, refIDs)
	if err != nil || len(samples) == 0 {
		return 0, 0
	}
	var cpuSum, memSum float64
	for _, sm := range samples {
		cpuSum += sm.CPUPct
		if sm.MemLimit > 0 {
			memSum += float64(sm.MemUsed) / float64(sm.MemLimit) * 100
		}
	}
	return cpuSum / float64(len(samples)), memSum / float64(len(samples))
}

// Series 分钟级聚合曲线；指定实例时先做资源归属校验。
func (s *MonitorService) Series(ctx context.Context, kind string, refID *uint64, minutes int, ac AccessCtx) ([]repository.MetricSeriesRow, error) {
	if refID != nil {
		inst, err := s.repo.EcsGetByID(*refID)
		if err != nil {
			return nil, err
		}
		if !accessAllowed(ac, inst, true) {
			return nil, ErrForbidden
		}
		return s.repo.MetricSeries(kind, refID, minutes)
	}
	rows, err := s.accessibleECS(ctx, ac)
	if err != nil {
		return nil, err
	}
	refIDs := make([]uint64, 0, len(rows))
	for i := range rows {
		refIDs = append(refIDs, rows[i].ID)
	}
	return s.repo.MetricSeriesForRefs(kind, refIDs, minutes)
}
