package repository

import (
	"time"

	"github.com/dxcloud/cloud-api/internal/model"
)

// ---------- 指标 ----------

func (r *Repos) MetricInsert(samples []model.MetricSample) error {
	if len(samples) == 0 {
		return nil
	}
	return r.DB.Create(&samples).Error
}

// MetricLatest 每个资源的最近一次采样（kind+ref_id 分组）。
func (r *Repos) MetricLatest(kind string, within time.Duration) ([]model.MetricSample, error) {
	var items []model.MetricSample
	sub := r.DB.Model(&model.MetricSample{}).
		Select("MAX(ts) AS ts, ref_id").
		Where("kind = ? AND ts > ?", kind, time.Now().Add(-within)).
		Group("ref_id")
	err := r.DB.Model(&model.MetricSample{}).
		Joins("JOIN (?) t ON metric_samples.ref_id = t.ref_id AND metric_samples.ts = t.ts", sub).
		Where("metric_samples.kind = ?", kind).
		Find(&items).Error
	return items, err
}

func (r *Repos) MetricLatestForRefs(kind string, within time.Duration, refIDs []uint64) ([]model.MetricSample, error) {
	if len(refIDs) == 0 {
		return []model.MetricSample{}, nil
	}
	sub := r.DB.Model(&model.MetricSample{}).
		Select("MAX(ts) AS ts, ref_id").
		Where("kind = ? AND ts > ? AND ref_id IN ?", kind, time.Now().Add(-within), refIDs).
		Group("ref_id")
	var items []model.MetricSample
	err := r.DB.Model(&model.MetricSample{}).
		Joins("JOIN (?) t ON metric_samples.ref_id = t.ref_id AND metric_samples.ts = t.ts", sub).
		Where("metric_samples.kind = ?", kind).
		Find(&items).Error
	return items, err
}

// MetricSeriesRow 分钟级聚合行（显式类型，避免 map[string]any 导致 []byte → base64）。
type MetricSeriesRow struct {
	Bucket string  `json:"bucket"`
	CPU    float64 `json:"cpu"`
	MemPct float64 `json:"mem_pct"`
	NetRx  int64   `json:"net_rx"`
	NetTx  int64   `json:"net_tx"`
}

// MetricSeries 分钟级聚合序列。
func (r *Repos) MetricSeries(kind string, refID *uint64, minutes int) ([]MetricSeriesRow, error) {
	if minutes <= 0 || minutes > 24*60 {
		minutes = 60
	}
	q := r.DB.Model(&model.MetricSample{}).
		Select("DATE_FORMAT(ts, '%Y-%m-%dT%H:%i:00Z') AS bucket, ROUND(AVG(cpu_pct),2) AS cpu, ROUND(AVG(mem_used/mem_limit*100),2) AS mem_pct, SUM(net_rx) AS net_rx, SUM(net_tx) AS net_tx").
		Where("kind = ? AND ts > ?", kind, time.Now().Add(-time.Duration(minutes)*time.Minute))
	if refID != nil {
		q = q.Where("ref_id = ?", *refID)
	}
	q = q.Group("bucket").Order("bucket ASC")
	var rows []MetricSeriesRow
	err := q.Find(&rows).Error
	return rows, err
}

func (r *Repos) MetricSeriesForRefs(kind string, refIDs []uint64, minutes int) ([]MetricSeriesRow, error) {
	if len(refIDs) == 0 {
		return []MetricSeriesRow{}, nil
	}
	if minutes <= 0 || minutes > 24*60 {
		minutes = 60
	}
	q := r.DB.Model(&model.MetricSample{}).
		Select("DATE_FORMAT(ts, '%Y-%m-%dT%H:%i:00Z') AS bucket, ROUND(AVG(cpu_pct),2) AS cpu, ROUND(AVG(mem_used/mem_limit*100),2) AS mem_pct, SUM(net_rx) AS net_rx, SUM(net_tx) AS net_tx").
		Where("kind = ? AND ts > ? AND ref_id IN ?", kind, time.Now().Add(-time.Duration(minutes)*time.Minute), refIDs).
		Group("bucket").Order("bucket ASC")
	var rows []MetricSeriesRow
	err := q.Find(&rows).Error
	return rows, err
}

// MetricCleanup 删除过期采样。
func (r *Repos) MetricCleanup(before time.Time) (int64, error) {
	res := r.DB.Where("ts < ?", before).Delete(&model.MetricSample{})
	return res.RowsAffected, res.Error
}

// ---------- 操作日志 ----------

func (r *Repos) OpLogCreate(l *model.OperationLog) error {
	return r.DB.Create(l).Error
}

func (r *Repos) OpLogList(module, keyword string, page, size int, userID, orgID uint64, canManage bool) ([]model.OperationLog, int64, error) {
	q := r.DB.Model(&model.OperationLog{})
	if !canManage {
		q = q.Where("user_id = ?", userID)
	} else if orgID > 0 {
		q = q.Where("org_id = ?", orgID)
	}
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("action LIKE ? OR target_name LIKE ?", like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.OperationLog
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

// AuditLogList 审计日志检索。
func (r *Repos) AuditLogList(action, keyword string, page, size int, userID, orgID uint64, canManage bool) ([]model.AuditLog, int64, error) {
	q := r.DB.Model(&model.AuditLog{})
	if !canManage {
		q = q.Where("user_id = ?", userID)
	} else if orgID > 0 {
		q = q.Where("org_id = ?", orgID)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("action LIKE ? OR resource_id LIKE ? OR ip LIKE ?", like, like, like)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.AuditLog
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

// LoginLogList 登录日志。
func (r *Repos) LoginLogList(page, size int, userID uint64, canManage bool) ([]model.LoginLog, int64, error) {
	q := r.DB.Model(&model.LoginLog{})
	if !canManage {
		q = q.Where("user_id = ?", userID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.LoginLog
	err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error
	return items, total, err
}

// ---------- 通知 ----------

func (r *Repos) NotifyCreate(n *model.Notification) error {
	return r.DB.Create(n).Error
}

func (r *Repos) NotifyList(userID uint64, unreadOnly bool, limit int) ([]model.Notification, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := r.DB.Model(&model.Notification{}).Where("user_id = ?", userID)
	if unreadOnly {
		q = q.Where("read_at IS NULL")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Notification
	err := q.Order("id DESC").Limit(limit).Find(&items).Error
	return items, total, err
}

func (r *Repos) NotifyMarkRead(userID, id uint64) error {
	res := r.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("read_at", time.Now())
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return res.Error
}

func (r *Repos) NotifyMarkAllRead(userID uint64) error {
	return r.DB.Model(&model.Notification{}).
		Where("user_id = ? AND read_at IS NULL", userID).
		Update("read_at", time.Now()).Error
}
