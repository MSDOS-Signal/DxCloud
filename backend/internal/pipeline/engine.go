// Package pipeline：轻量 CI/CD Pipeline 引擎（Phase 7）。
// 队列：Redis LIST dx:pipe:queue（3.0 兼容）；执行：内嵌 Worker + 隔离 Job 容器；
// 状态机：pending → running → success/failed/canceled；步骤：pending/running/success/failed/skipped。
package pipeline

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/runner"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

const (
	queueKey     = "dx:pipe:queue"
	cancelPrefix = "dx:cancel:"
	logDir       = "/tmp/dxlogs"
)

var (
	ErrStepTypeUnsupported = errors.New("step type not supported in this phase")
	stepTypeWhitelist      = map[string]bool{
		"git": true, "shell": true,
		"docker-build": true, "docker-push": true, "docker-deploy": true, "wait-health": true,
	}
)

// PipelineDef YAML 定义。
type PipelineDef struct {
	Name    string            `yaml:"name"`
	Timeout string            `yaml:"timeout"`
	Env     map[string]string `yaml:"env"`
	Steps   []StepDef         `yaml:"steps"`
}

type StepDef struct {
	Name         string   `yaml:"name"`
	Type         string   `yaml:"type"`
	Script       string   `yaml:"script"`
	URL          string   `yaml:"url"`
	Branch       string   `yaml:"branch"`
	AllowFailure bool     `yaml:"allow_failure"`
	Timeout      string   `yaml:"timeout"`
	Dockerfile   string   `yaml:"dockerfile"`
	Tags         []string `yaml:"tags"`
	Image        string   `yaml:"image"`
	Application  string   `yaml:"application"`
	Environment  string   `yaml:"environment"`
}

// ParseDefinition 解析并校验 YAML。
func ParseDefinition(def string) (*PipelineDef, error) {
	if strings.TrimSpace(def) == "" {
		return nil, errors.New("定义不能为空")
	}
	var p PipelineDef
	if err := yaml.Unmarshal([]byte(def), &p); err != nil {
		return nil, fmt.Errorf("YAML 解析失败: %w", err)
	}
	if len(p.Steps) == 0 {
		return nil, errors.New("至少需要一个步骤")
	}
	for i, s := range p.Steps {
		if s.Name == "" {
			return nil, fmt.Errorf("第 %d 步缺少 name", i+1)
		}
		if !stepTypeWhitelist[s.Type] {
			return nil, fmt.Errorf("第 %d 步类型 %q 不在白名单（git/shell/docker-build/docker-push/docker-deploy/wait-health）", i+1, s.Type)
		}
		if s.Type == "shell" && strings.TrimSpace(s.Script) == "" {
			return nil, fmt.Errorf("第 %d 步 shell 需要 script", i+1)
		}
		if s.Type == "git" && strings.TrimSpace(s.URL) == "" {
			return nil, fmt.Errorf("第 %d 步 git 需要 url", i+1)
		}
	}
	return &p, nil
}

// Deployer 由应用服务实现（路由装配时注入），用于 docker-deploy 步骤的平台侧部署。
type Deployer interface {
	DeployByName(ctx context.Context, ac service.AccessCtx, appName, imageRef, note string) (*model.Deployment, error)
}

type ImageResolver interface {
	ResolvePullRef(ref string) (pullRef string, mirrored bool)
}

type Engine struct {
	repo              *repository.Repos
	runner            *runner.DockerJobRunner
	compute           docker.Provider
	rdb               *redis.Client
	iamSvc            *iam.Service
	appNetwork        string
	registryEngineURL string
	deployer          Deployer
	imageResolver     ImageResolver
	workers           int
	log               *zap.Logger
}

func NewEngine(repo *repository.Repos, jobRunner *runner.DockerJobRunner, compute docker.Provider, rdb *redis.Client, iamSvc *iam.Service, appNetwork string, workers int, log *zap.Logger) *Engine {
	if workers <= 0 {
		workers = 2
	}
	return &Engine{repo: repo, runner: jobRunner, compute: compute, rdb: rdb, iamSvc: iamSvc, appNetwork: appNetwork, workers: workers, log: log}
}

// SetDeployer / SetRegistryEngineURL：路由装配后注入（避免构造顺序耦合）。
func (e *Engine) SetDeployer(d Deployer)           { e.deployer = d }
func (e *Engine) SetRegistryEngineURL(u string)    { e.registryEngineURL = u }
func (e *Engine) SetImageResolver(r ImageResolver) { e.imageResolver = r }

// Start 崩溃恢复 + 启动 N 个 Worker。
func (e *Engine) Start(ctx context.Context) {
	// 恢复：进行中的运行标记失败（引擎重启导致中断），同时清理残留 Job 容器
	if runs, err := e.repo.RunsInFlight(); err == nil {
		for i := range runs {
			if runs[i].Status == model.PipeRunning {
				// 清理崩溃残留的 Job 容器（按 run-id 标签查找并 stop/remove）
				e.cleanupRunContainers(ctx, runs[i].ID)
				runs[i].Status = model.PipeFailed
				fin := time.Now()
				runs[i].FinishedAt = &fin
				_ = e.repo.RunUpdate(&runs[i])
			}
		}
	}
	for i := 0; i < e.workers; i++ {
		go e.worker(ctx, i)
	}
	e.log.Info("pipeline engine started", zap.Int("workers", e.workers))
}

// cleanupRunContainers 清理引擎崩溃时残留的 Job 容器（按 Docker 标签查找）。
func (e *Engine) cleanupRunContainers(ctx context.Context, runID uint64) {
	containers, err := e.compute.ListByLabels(ctx, map[string]string{
		"com.dxcloud.kind":   "pipeline-job",
		"com.dxcloud.run-id": strconv.FormatUint(runID, 10),
	})
	if err != nil {
		e.log.Warn("cleanup: list orphan containers failed", zap.Uint64("run", runID), zap.Error(err))
		return
	}
	for _, c := range containers {
		_ = e.compute.Stop(ctx, c.ID, true)
		_ = e.compute.Remove(ctx, c.ID, true)
		e.log.Info("cleanup: removed orphaned job container", zap.String("container", c.ID), zap.Uint64("run", runID))
	}
}

func (e *Engine) worker(ctx context.Context, id int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		res, err := e.rdb.BLPop(ctx, 5*time.Second, queueKey).Result()
		if err != nil || len(res) < 2 {
			continue
		}
		runID, _ := strconv.ParseUint(res[1], 10, 64)
		if err := e.executeRun(ctx, runID); err != nil {
			e.log.Error("execute run failed", zap.Uint64("run", runID), zap.Error(err))
		}
	}
}

// Enqueue 入队（创建 run + jobs 后调用）。
func (e *Engine) Enqueue(runID uint64) error {
	return e.rdb.LPush(context.Background(), queueKey, strconv.FormatUint(runID, 10)).Err()
}

// CreateRun 解析定义、快照步骤、创建 run/jobs 并入队。
func (e *Engine) CreateRun(ctx context.Context, ac service.AccessCtx, pipelineID uint64, ref, commitSHA, triggerType, ip, requestID string) (*model.PipelineRun, error) {
	pipe, err := e.repo.PipelineGetByID(pipelineID)
	if err != nil {
		return nil, err
	}
	def, err := ParseDefinition(pipe.Definition)
	if err != nil {
		return nil, err
	}
	// 步骤快照
	steps := make([]model.PipelineStep, 0, len(def.Steps))
	for i, s := range def.Steps {
		cfg, _ := jsonMarshal(s)
		steps = append(steps, model.PipelineStep{
			PipelineID: pipe.ID, Name: s.Name, Type: s.Type, Seq: i + 1, ConfigJSON: cfg,
		})
	}
	if err := e.repo.StepReplace(pipe.ID, steps); err != nil {
		return nil, err
	}
	runNo := e.repo.RunLastNo(pipe.ID) + 1
	run := &model.PipelineRun{
		PipelineID: pipe.ID, RunNo: runNo, TriggerType: triggerType,
		Ref: ref, CommitSHA: commitSHA, Status: model.PipePending, TriggeredBy: &ac.UserID,
	}
	if err := e.repo.RunCreate(run); err != nil {
		return nil, err
	}
	// job 行
	for i, s := range def.Steps {
		j := &model.PipelineJobRun{
			PipelineRunID: run.ID, StepID: steps[i].ID, Name: s.Name, Type: s.Type,
			Status: model.JobPending, LogPath: fmt.Sprintf("%s/%d.log", logDir, time.Now().UnixNano()),
		}
		if err := e.repo.JobCreate(j); err != nil {
			return nil, err
		}
	}
	if err := e.Enqueue(run.ID); err != nil {
		return nil, err
	}
	uid := ac.UserID
	e.iamSvc.Audit(ctx, &uid, "pipeline.run", "pipeline", pipe.Name, ip, requestID, 1, map[string]any{"run": run.ID})
	return run, nil
}

// executeRun 执行一次运行。
func (e *Engine) executeRun(ctx context.Context, runID uint64) error {
	run, err := e.repo.RunGetByID(runID)
	if err != nil {
		return err
	}
	if run.Status != model.PipePending {
		return nil
	}
	pipe, err := e.repo.PipelineGetByID(run.PipelineID)
	if err != nil {
		return err
	}
	def, err := ParseDefinition(pipe.Definition)
	if err != nil {
		e.finishRun(run, model.PipeFailed)
		return err
	}

	now := time.Now()
	run.Status = model.PipeRunning
	run.StartedAt = &now
	_ = e.repo.RunUpdate(run)

	// 全局超时
	runTimeout := 2 * time.Hour
	if def.Timeout != "" {
		if d, err := time.ParseDuration(def.Timeout); err == nil && d > 0 {
			runTimeout = d
		}
	}
	runCtx, cancel := context.WithTimeout(ctx, runTimeout)
	defer cancel()

	// 取消监视
	stopWatch := make(chan struct{})
	defer close(stopWatch)
	go func() {
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-stopWatch:
				return
			case <-t.C:
				if v, err := e.rdb.Get(runCtx, cancelPrefix+strconv.FormatUint(runID, 10)).Result(); err == nil && v == "1" {
					cancel()
					return
				}
			}
		}
	}()

	// workspace 卷
	wsVolume := fmt.Sprintf("dxw-run-%d", runID)
	if _, err := e.compute.CreateVolume(ctx, wsVolume, map[string]string{"com.dxcloud.kind": "pipeline-workspace"}); err != nil {
		e.log.Warn("create workspace volume failed", zap.Error(err))
	}
	defer func() { _ = e.compute.RemoveVolume(context.Background(), wsVolume) }()

	steps := def.Steps
	jobs, err := e.repo.JobListByRun(runID)
	if err != nil {
		return err
	}
	failed := false
	for i := range steps {
		if runCtx.Err() != nil {
			e.markRemainingSkipped(jobs, i)
			e.finishRun(run, model.PipeCanceled)
			return nil
		}
		if i >= len(jobs) {
			break
		}
		job := &jobs[i]
		jobErr := e.executeStep(runCtx, run, job, steps[i], def.Env, wsVolume)
		if jobErr != nil {
			// 取消优先于失败判定（runCtx 取消 或 Redis 取消标志——kill 容器比 ticker 更快）
			if runCtx.Err() != nil || e.isCanceled(run.ID) {
				e.markRemainingSkipped(jobs, i+1)
				e.finishRun(run, model.PipeCanceled)
				return nil
			}
			if steps[i].AllowFailure {
				job.Status = model.JobSkipped
				fin := time.Now()
				job.FinishedAt = &fin
				_ = e.repo.JobUpdate(job)
				continue
			}
			e.markRemainingSkipped(jobs, i+1)
			failed = true
			break
		}
	}
	if failed {
		e.finishRun(run, model.PipeFailed)
		return nil
	}
	e.finishRun(run, model.PipeSuccess)
	return nil
}

// executeStep 在隔离 Job 容器中执行单个步骤。
func (e *Engine) executeStep(ctx context.Context, run *model.PipelineRun, job *model.PipelineJobRun, step StepDef, env map[string]string, wsVolume string) error {
	now := time.Now()
	job.Status = model.JobRunning
	job.StartedAt = &now
	_ = e.repo.JobUpdate(job)

	// 打开日志文件（实时可读）
	logPath := job.LogPath
	f, err := runner.LogFile(logPath)
	if err != nil {
		f = nil
	}
	if f != nil {
		defer f.Close()
	}
	var writer io.Writer = io.Discard
	if f != nil {
		writer = f
	}

	// docker-* / wait-health：服务端执行（BuildKit 构建/推送/平台部署），不进 Job 容器
	if step.Type == "docker-build" || step.Type == "docker-push" || step.Type == "docker-deploy" || step.Type == "wait-health" {
		fin := time.Now()
		job.FinishedAt = &fin
		if err := e.executeServerStep(ctx, run, step, wsVolume, writer); err != nil {
			job.Status = model.JobFailed
			job.ExitCode = -1
			_ = e.repo.JobUpdate(job)
			return err
		}
		job.Status = model.JobSuccess
		job.ExitCode = 0
		_ = e.repo.JobUpdate(job)
		return nil
	}

	spec, err := e.buildJobSpec(step, env, wsVolume, run)
	if err != nil {
		job.Status = model.JobFailed
		job.ExitCode = -1
		fin := time.Now()
		job.FinishedAt = &fin
		_ = e.repo.JobUpdate(job)
		if f != nil {
			_, _ = f.WriteString(err.Error() + "\n")
		}
		return err
	}

	stepTimeout := 30 * time.Minute
	if step.Timeout != "" {
		if d, err := time.ParseDuration(step.Timeout); err == nil && d > 0 {
			stepTimeout = d
		}
	}
	spec.Timeout = stepTimeout

	code, runErr := e.runner.RunJob(ctx, fmt.Sprintf("dxj-%d-%d", run.ID, job.ID), spec, writer, func(id string) {
		job.ContainerID = id
		_ = e.repo.JobUpdate(job)
	})

	fin := time.Now()
	job.FinishedAt = &fin
	job.ExitCode = code
	if runErr != nil {
		job.Status = model.JobFailed
		if ctx.Err() != nil {
			job.Status = model.JobFailed
		}
		_ = e.repo.JobUpdate(job)
		if f != nil {
			_, _ = f.WriteString("\n[job error] " + runErr.Error() + "\n")
		}
		return runErr
	}
	if code != 0 {
		job.Status = model.JobFailed
		_ = e.repo.JobUpdate(job)
		if f != nil {
			_, _ = f.WriteString(fmt.Sprintf("\n[job exited with code %d]\n", code))
		}
		return fmt.Errorf("step %s exited with code %d", step.Name, code)
	}
	job.Status = model.JobSuccess
	_ = e.repo.JobUpdate(job)
	return nil
}

// buildJobSpec 步骤类型 → Job 规格（git/shell 走隔离 Job 容器；docker-* 由 executeStep 服务端处理）。
func (e *Engine) buildJobSpec(step StepDef, env map[string]string, wsVolume string, run *model.PipelineRun) (runner.JobSpec, error) {
	envList := make([]string, 0, len(env)+2)
	for k, v := range env {
		envList = append(envList, k+"="+v)
	}
	envList = append(envList, "PIPELINE_RUN_ID="+strconv.FormatUint(run.ID, 10))

	labels := map[string]string{
		"com.dxcloud.kind":   "pipeline-job",
		"com.dxcloud.run-id": strconv.FormatUint(run.ID, 10),
		"com.dxcloud.job-id": "job",
	}

	spec := runner.JobSpec{
		WorkspaceVolume: wsVolume,
		Env:             envList,
		CPU:             2,
		MemoryMB:        2048,
		NetworkID:       e.appNetwork, // 步骤可能需外网（git clone）
		Labels:          labels,
	}
	switch step.Type {
	case "shell":
		spec.Image = "alpine:3.20"
		spec.Cmd = []string{"sh", "-c", step.Script}
	case "git":
		spec.Image = "alpine/git:latest"
		cmd := []string{"git", "clone", "--depth", "1"}
		if step.Branch != "" {
			cmd = append(cmd, "--branch", step.Branch)
		}
		cmd = append(cmd, step.URL, ".")
		spec.Cmd = cmd
	case "docker-build", "docker-push", "docker-deploy", "wait-health":
		return spec, fmt.Errorf("%w: %s（服务端执行，走 executeServerStep）", ErrStepTypeUnsupported, step.Type)
	default:
		return spec, fmt.Errorf("unknown step type %q", step.Type)
	}
	if e.imageResolver != nil {
		if ref, ok := e.imageResolver.ResolvePullRef(spec.Image); ok {
			spec.Image = ref
		}
	}
	return spec, nil
}

// executeServerStep 服务端步骤：docker-build（BuildKit）/ docker-push / docker-deploy / wait-health。
func (e *Engine) executeServerStep(ctx context.Context, run *model.PipelineRun, step StepDef, wsVolume string, writer io.Writer) error {
	switch step.Type {
	case "docker-build":
		if step.Dockerfile == "" {
			step.Dockerfile = "Dockerfile"
		}
		tags := make([]string, 0, len(step.Tags))
		for _, t := range step.Tags {
			t = strings.ReplaceAll(t, "${COMMIT_SHA}", run.CommitSHA)
			t = strings.ReplaceAll(t, "${REF}", run.Ref)
			tags = append(tags, t)
		}
		if len(tags) == 0 {
			return errors.New("docker-build 需要 tags")
		}
		tarData, err := e.workspaceTar(ctx, wsVolume)
		if err != nil {
			return fmt.Errorf("打包构建上下文失败: %w", err)
		}
		_, _ = fmt.Fprintf(writer, "[docker-build] tags=%v dockerfile=%s context=%d bytes\n", tags, step.Dockerfile, len(tarData))
		stream, err := e.compute.ImageBuild(ctx, bytes.NewReader(tarData), step.Dockerfile, tags)
		if err != nil {
			_, _ = fmt.Fprintf(writer, "[docker-build] ERROR: %v\n", err)
			return fmt.Errorf("build failed: %w", err)
		}
		defer stream.Close()
		return decodeBuildStream(stream, writer)
	case "docker-push":
		if len(step.Tags) == 0 {
			return errors.New("docker-push 需要 tags")
		}
		for _, t := range step.Tags {
			t = strings.ReplaceAll(t, "${COMMIT_SHA}", run.CommitSHA)
			_, _ = fmt.Fprintf(writer, "[docker-push] pushing %s ...\n", t)
			if err := e.compute.ImagePush(ctx, t); err != nil {
				return fmt.Errorf("push %s failed: %w", t, err)
			}
			_, _ = fmt.Fprintf(writer, "[docker-push] %s pushed\n", t)
		}
		return nil
	case "docker-deploy":
		if step.Image == "" {
			return errors.New("docker-deploy 需要 image 字段")
		}
		if e.deployer == nil {
			return errors.New("deployer not wired")
		}
		imageRef := strings.ReplaceAll(step.Image, "${COMMIT_SHA}", run.CommitSHA)
		// 以 pipeline 属主身份执行部署（webhook 触发时无登录用户）
		uid := uint64(0)
		if run.TriggeredBy != nil {
			uid = *run.TriggeredBy
		}
		roles, _ := e.iamSvc.GetUserRoleCodes(ctx, uid)
		ac := service.AccessCtx{UserID: uid, Roles: roles}
		if pipe, err := e.repo.PipelineGetByID(run.PipelineID); err == nil && pipe.OrgID != nil {
			ac.OrgID = *pipe.OrgID
		}
		_, _ = fmt.Fprintf(writer, "[docker-deploy] deploying %s to application %s ...\n", imageRef, step.Application)
		if _, err := e.deployer.DeployByName(ctx, ac, step.Application, imageRef, "pipeline run #"+strconv.FormatUint(run.ID, 10)); err != nil {
			_, _ = fmt.Fprintf(writer, "[docker-deploy] ERROR: %v\n", err)
			return fmt.Errorf("deploy failed: %w", err)
		}
		_, _ = fmt.Fprintf(writer, "[docker-deploy] done\n")
		return nil
	case "wait-health":
		return errors.New("wait-health 将在后续版本支持")
	default:
		return fmt.Errorf("unknown server step type %q", step.Type)
	}
}

// workspaceTar 将 workspace 卷打包为 tar（一次性助手容器）。
func (e *Engine) workspaceTar(ctx context.Context, wsVolume string) ([]byte, error) {
	helperImage := "alpine:3.20"
	if e.imageResolver != nil {
		if ref, ok := e.imageResolver.ResolvePullRef(helperImage); ok {
			helperImage = ref
		}
	}
	info, err := e.compute.Create(ctx, docker.CreateSpec{
		Name:   fmt.Sprintf("dxj-tar-%d", time.Now().UnixNano()),
		Image:  helperImage,
		Cmd:    []string{"tar", "-cf", "-", "-C", "/workspace", "."},
		Mounts: []docker.Mount{{VolumeName: wsVolume, Target: "/workspace"}},
		CPU:    1, MemoryMB: 512,
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = e.compute.Remove(context.Background(), info.ID, true) }()
	if code, err := e.compute.WaitContainer(ctx, info.ID); err != nil || code != 0 {
		return nil, fmt.Errorf("tar helper exited with code %d: %v", code, err)
	}
	logs, err := e.compute.LogsRaw(ctx, info.ID)
	if err != nil {
		return nil, err
	}
	return []byte(logs), nil
}

// decodeBuildStream 解析 BuildKit 的 JSON 行流，提取 stream/error 文本写入日志。
func decodeBuildStream(stream io.Reader, writer io.Writer) error {
	sc := bufio.NewScanner(stream)
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	var lastErr string
	for sc.Scan() {
		line := sc.Bytes()
		var msg struct {
			Stream      string `json:"stream"`
			Error       string `json:"error"`
			ErrorDetail struct {
				Message string `json:"message"`
			} `json:"errorDetail"`
		}
		if err := json.Unmarshal(line, &msg); err == nil {
			if msg.Stream != "" {
				_, _ = writer.Write([]byte(msg.Stream))
			}
			if msg.Error != "" {
				lastErr = msg.Error
			}
			if msg.ErrorDetail.Message != "" {
				lastErr = msg.ErrorDetail.Message
			}
		} else {
			_, _ = writer.Write(line)
			_, _ = writer.Write([]byte("\n"))
		}
	}
	if lastErr != "" {
		return errors.New(lastErr)
	}
	return sc.Err()
}

// Cancel 取消运行：置取消标记 + 杀当前 Job 容器。
func (e *Engine) Cancel(ctx context.Context, runID uint64) error {
	run, err := e.repo.RunGetByID(runID)
	if err != nil {
		return err
	}
	if run.Status != model.PipePending && run.Status != model.PipeRunning {
		return errors.New("run is not cancellable")
	}
	if err := e.rdb.Set(ctx, cancelPrefix+strconv.FormatUint(runID, 10), "1", time.Hour).Err(); err != nil {
		return err
	}
	// 若已出队且正在执行：杀当前容器
	if jobs, err := e.repo.JobListByRun(runID); err == nil {
		for i := range jobs {
			if jobs[i].Status == model.JobRunning && jobs[i].ContainerID != "" {
				e.runner.Kill(jobs[i].ContainerID)
			}
		}
	}
	return nil
}

// isCanceled 查询 Redis 取消标志。
func (e *Engine) isCanceled(runID uint64) bool {
	v, err := e.rdb.Get(context.Background(), cancelPrefix+strconv.FormatUint(runID, 10)).Result()
	return err == nil && v == "1"
}

func (e *Engine) markRemainingSkipped(jobs []model.PipelineJobRun, from int) {
	for i := from; i < len(jobs); i++ {
		if jobs[i].Status == model.JobPending {
			jobs[i].Status = model.JobSkipped
			fin := time.Now()
			jobs[i].FinishedAt = &fin
			_ = e.repo.JobUpdate(&jobs[i])
		}
	}
}

func (e *Engine) finishRun(run *model.PipelineRun, status string) {
	fin := time.Now()
	run.Status = status
	run.FinishedAt = &fin
	if run.StartedAt != nil {
		run.DurationMs = fin.Sub(*run.StartedAt).Milliseconds()
	}
	_ = e.repo.RunUpdate(run)

	// 站内通知（触发者）
	if run.TriggeredBy != nil && *run.TriggeredBy != 0 {
		statusTxt := map[string]string{
			model.PipeSuccess: "成功", model.PipeFailed: "失败", model.PipeCanceled: "已取消",
		}[status]
		if statusTxt == "" {
			statusTxt = status
		}
		title := "Pipeline 运行" + statusTxt
		content := "运行 #" + strconv.Itoa(run.RunNo) + " " + statusTxt + "（耗时 " + (time.Duration(run.DurationMs) * time.Millisecond).String() + "），点击查看日志"
		_ = e.repo.NotifyCreate(&model.Notification{
			UserID: *run.TriggeredBy, Type: "pipeline",
			Title: title, Content: content,
			Link: "/pipeline-runs/" + strconv.FormatUint(run.ID, 10),
		})
	}
}

// Logs 读取 job 日志文件尾部（最多 256KB）。
func (e *Engine) Logs(runID, jobID uint64, maxBytes int64) (string, error) {
	job, err := e.repo.JobGetByID(jobID)
	if err != nil || job.PipelineRunID != runID {
		return "", errors.New("job not found")
	}
	if maxBytes <= 0 || maxBytes > 512*1024 {
		maxBytes = 256 * 1024
	}
	f, err := os.Open(job.LogPath)
	if err != nil {
		return "", nil // 尚无日志
	}
	defer f.Close()
	st, _ := f.Stat()
	offset := int64(0)
	if st.Size() > maxBytes {
		offset = st.Size() - maxBytes
	}
	_, _ = f.Seek(offset, 0)
	buf, _ := io.ReadAll(f)
	return string(buf), nil
}

// AccessCtx 复用 service.AccessCtx（handler 侧组装）。
type AccessCtx = service.AccessCtx

func jsonMarshal(v any) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}
