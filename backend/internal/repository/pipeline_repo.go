package repository

import (
	"errors"

	"github.com/dxcloud/cloud-api/internal/model"
	"gorm.io/gorm"
)

// ---------- Pipeline ----------

func (r *Repos) PipelineList(keyword string, ownerID *uint64, orgID uint64) ([]model.Pipeline, error) {
	q := r.DB.Model(&model.Pipeline{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if ownerID != nil {
		q = q.Scopes(ScopeByOwner(*ownerID))
	}
	q = q.Scopes(ScopeByTenantOrNull(orgID))
	var items []model.Pipeline
	err := q.Order("id DESC").Find(&items).Error
	return items, err
}

func (r *Repos) PipelineGetByID(id uint64) (*model.Pipeline, error) {
	var p model.Pipeline
	err := r.DB.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *Repos) PipelineGetByName(name string, orgID uint64) (*model.Pipeline, error) {
	var p model.Pipeline
	err := r.DB.Where("name = ?", name).Scopes(ScopeByTenantOrNull(orgID)).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

func (r *Repos) PipelineCreate(p *model.Pipeline) error {
	return r.DB.Create(p).Error
}

func (r *Repos) PipelineUpdate(p *model.Pipeline) error {
	return r.DB.Model(p).Select("name", "description", "definition", "status").Updates(p).Error
}

func (r *Repos) PipelineSoftDelete(id uint64) error {
	return r.DB.Delete(&model.Pipeline{}, id).Error
}

// ---------- Steps（解析快照） ----------

func (r *Repos) StepReplace(pipelineID uint64, steps []model.PipelineStep) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("pipeline_id = ?", pipelineID).Delete(&model.PipelineStep{}).Error; err != nil {
			return err
		}
		for i := range steps {
			if err := tx.Create(&steps[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repos) StepList(pipelineID uint64) ([]model.PipelineStep, error) {
	var items []model.PipelineStep
	err := r.DB.Where("pipeline_id = ?", pipelineID).Order("seq ASC").Find(&items).Error
	return items, err
}

// ---------- Runs ----------

func (r *Repos) RunCreate(run *model.PipelineRun) error {
	return r.DB.Create(run).Error
}

func (r *Repos) RunUpdate(run *model.PipelineRun) error {
	return r.DB.Model(run).Select("status", "started_at", "finished_at", "duration_ms", "commit_sha").Updates(run).Error
}

func (r *Repos) RunGetByID(id uint64) (*model.PipelineRun, error) {
	var run model.PipelineRun
	err := r.DB.First(&run, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &run, err
}

func (r *Repos) RunList(pipelineID *uint64, limit int, ownerID *uint64, orgID uint64) ([]model.PipelineRun, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.DB.Model(&model.PipelineRun{})
	if pipelineID != nil {
		q = q.Where("pipeline_id = ?", *pipelineID)
	}
	if ownerID != nil {
		q = q.Where("pipeline_id IN (SELECT id FROM pipelines WHERE owner_id = ?)", *ownerID)
	}
	if orgID == 0 {
		q = q.Where("pipeline_id IN (SELECT id FROM pipelines WHERE org_id IS NULL OR org_id = 0)")
	} else {
		q = q.Where("pipeline_id IN (SELECT id FROM pipelines WHERE org_id = ?)", orgID)
	}
	var items []model.PipelineRun
	err := q.Order("id DESC").Limit(limit).Find(&items).Error
	return items, err
}

// RunLastNo 计算下一次 run_no。
func (r *Repos) RunLastNo(pipelineID uint64) int {
	var max int
	r.DB.Model(&model.PipelineRun{}).Where("pipeline_id = ?", pipelineID).
		Select("COALESCE(MAX(run_no),0)").Scan(&max)
	return max
}

// RunsInFlight 引擎重启恢复：处于进行中的运行。
func (r *Repos) RunsInFlight() ([]model.PipelineRun, error) {
	var items []model.PipelineRun
	err := r.DB.Where("status IN ?", []string{model.PipePending, model.PipeRunning}).Find(&items).Error
	return items, err
}

// ---------- Jobs ----------

func (r *Repos) JobCreate(j *model.PipelineJobRun) error {
	return r.DB.Create(j).Error
}

func (r *Repos) JobUpdate(j *model.PipelineJobRun) error {
	return r.DB.Model(j).Select("status", "exit_code", "container_id", "log_path", "started_at", "finished_at").Updates(j).Error
}

func (r *Repos) JobGetByID(id uint64) (*model.PipelineJobRun, error) {
	var j model.PipelineJobRun
	err := r.DB.First(&j, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &j, err
}

func (r *Repos) JobListByRun(runID uint64) ([]model.PipelineJobRun, error) {
	var items []model.PipelineJobRun
	err := r.DB.Where("pipeline_run_id = ?", runID).Order("id ASC").Find(&items).Error
	return items, err
}
