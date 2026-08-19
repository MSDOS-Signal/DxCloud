// Package security：Phase 11 安全加固 —— 密钥托管、容器安全基线审计、镜像策略扫描。
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/pkg/crypto"
	"go.uber.org/zap"
)

// ---------- 发现项 ----------

type Finding struct {
	Severity string `json:"severity"` // high / medium / low / info
	Kind     string `json:"kind"`     // baseline / image
	Target   string `json:"target"`
	Message  string `json:"message"`
}

var (
	ErrSecretNotFound = fmt.Errorf("secret not found")
	ErrSecretForbid   = fmt.Errorf("no access to this secret")
)

// ---------- 密钥托管 ----------

type SecretService struct {
	repo *repository.Repos
	key  []byte
	log  *zap.Logger
}

func NewSecretService(repo *repository.Repos, key []byte, log *zap.Logger) *SecretService {
	return &SecretService{repo: repo, key: key, log: log}
}

// Create 加密存储密钥（组织维度隔离：ac.OrgID>0 归组织，否则默认空间全局）。
func (s *SecretService) Create(ctx context.Context, ac AccessCtx, name, value string) (*model.Secret, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(name) > 127 {
		return nil, fmt.Errorf("密钥名需 1-127 位")
	}
	if value == "" {
		return nil, fmt.Errorf("密钥值不能为空")
	}
	if _, err := s.repo.SecretGetByName(ac.OrgID, name); err == nil {
		return nil, fmt.Errorf("密钥名已存在（组织内唯一）")
	}
	cipher, err := crypto.Encrypt(s.key, value)
	if err != nil {
		return nil, err
	}
	uid := ac.UserID
	sec := &model.Secret{OrgID: ac.OrgID, Name: name, ValueCipher: cipher, CreatedBy: &uid}
	if err := s.repo.SecretCreate(sec); err != nil {
		return nil, err
	}
	return sec, nil
}

// List 当前租户上下文下的密钥（永不返回明文）。
func (s *SecretService) List(ctx context.Context, ac AccessCtx) ([]model.Secret, error) {
	return s.repo.SecretList(ac.OrgID)
}

// Reveal 解密读取（后端校验租户归属；调用方需 secret:reveal 权限）。
func (s *SecretService) Reveal(ctx context.Context, ac AccessCtx, id uint64) (string, error) {
	sec, err := s.repo.SecretGetByID(id)
	if err != nil {
		return "", ErrSecretNotFound
	}
	if sec.OrgID != ac.OrgID {
		return "", ErrSecretForbid
	}
	return crypto.Decrypt(s.key, sec.ValueCipher)
}

// Delete 删除密钥（组织归属校验）。
func (s *SecretService) Delete(ctx context.Context, ac AccessCtx, id uint64) error {
	sec, err := s.repo.SecretGetByID(id)
	if err != nil {
		return ErrSecretNotFound
	}
	if sec.OrgID != ac.OrgID {
		return ErrSecretForbid
	}
	return s.repo.SecretDelete(id)
}

// ---------- 安全扫描（容器基线 + 镜像策略） ----------

type SecurityService struct {
	repo    *repository.Repos
	compute docker.ComputeProvider
	images  docker.ImageProvider
	log     *zap.Logger
}

func NewSecurityService(repo *repository.Repos, compute docker.ComputeProvider, images docker.ImageProvider, log *zap.Logger) *SecurityService {
	return &SecurityService{repo: repo, compute: compute, images: images, log: log}
}

// ScanBaseline 容器安全基线审计（仅平台托管容器 com.dxcloud.kind 参与计分）。
func (s *SecurityService) ScanBaseline(ctx context.Context) (*model.SecurityReport, error) {
	containers, err := s.compute.SecurityAudit(ctx)
	if err != nil {
		return nil, err
	}
	findings := make([]Finding, 0)
	for _, c := range containers {
		if c.Kind == "" {
			// 非平台托管（MySQL/Redis/Traefik 等基础设施），仅提示不计分
			findings = append(findings, Finding{Severity: "info", Kind: "baseline", Target: c.Name, Message: "非平台托管容器（com.dxcloud.kind 缺失），跳过基线检查"})
			continue
		}
		if c.Privileged {
			findings = append(findings, Finding{Severity: "high", Kind: "baseline", Target: c.Name, Message: "容器以特权模式运行，禁止"})
		}
		if !c.NoNewPrivileges {
			findings = append(findings, Finding{Severity: "high", Kind: "baseline", Target: c.Name, Message: "缺少 no-new-privileges，禁止提权逃逸失效"})
		}
		if !capDropAll(c.CapDrop) {
			findings = append(findings, Finding{Severity: "high", Kind: "baseline", Target: c.Name, Message: fmt.Sprintf("未 CapDrop ALL（当前 drop=%v）", c.CapDrop)})
		}
		for _, cap := range c.CapAdd {
			if cap == "ALL" || cap == "SYS_ADMIN" || cap == "NET_ADMIN" || cap == "SYS_PTRACE" || cap == "SYS_MODULE" {
				findings = append(findings, Finding{Severity: "high", Kind: "baseline", Target: c.Name, Message: "危险能力 " + cap})
			}
		}
		if c.PidsLimit <= 0 || c.PidsLimit > 1024 {
			findings = append(findings, Finding{Severity: "low", Kind: "baseline", Target: c.Name, Message: fmt.Sprintf("PidsLimit=%d（建议 1-1024）", c.PidsLimit)})
		}
		if c.MemoryLimit == 0 || c.NanoCPUs == 0 {
			findings = append(findings, Finding{Severity: "low", Kind: "baseline", Target: c.Name, Message: "缺少 CPU/内存上限（MemoryLimit=0 或 NanoCPUs=0）"})
		}
		if c.User == "" || c.User == "root" {
			findings = append(findings, Finding{Severity: "medium", Kind: "baseline", Target: c.Name, Message: "以 root 运行（建议指定非 root 用户）"})
		}
	}
	return s.saveReport(ctx, "baseline", findings)
}

func capDropAll(drop []string) bool {
	for _, d := range drop {
		if d == "ALL" {
			return true
		}
	}
	return false
}

// ScanImages 镜像策略扫描（生产禁忌：latest 标签、悬空镜像、超大镜像）。
func (s *SecurityService) ScanImages(ctx context.Context) (*model.SecurityReport, error) {
	images, err := s.images.ListImages(ctx)
	if err != nil {
		return nil, err
	}
	findings := make([]Finding, 0)
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == "<none>:<none>" {
				findings = append(findings, Finding{Severity: "info", Kind: "image", Target: img.ID[:12], Message: "悬空镜像（建议清理）"})
				continue
			}
			if strings.HasSuffix(tag, ":latest") {
				findings = append(findings, Finding{Severity: "medium", Kind: "image", Target: tag, Message: "使用 latest 标签（生产环境应使用不可变版本号）"})
			}
		}
		if img.Size > 2*1024*1024*1024 {
			findings = append(findings, Finding{Severity: "low", Kind: "image", Target: img.ID[:12], Message: fmt.Sprintf("镜像体积 %.1f GB（建议瘦身）", float64(img.Size)/1024/1024/1024)})
		}
	}
	return s.saveReport(ctx, "image", findings)
}

// computeScore 发现项计分：high -10 / medium -5 / low -2 / info 0，下限 0。
func computeScore(findings []Finding) int {
	score := 100
	for _, f := range findings {
		switch f.Severity {
		case "high":
			score -= 10
		case "medium":
			score -= 5
		case "low":
			score -= 2
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}

func (s *SecurityService) saveReport(ctx context.Context, kind string, findings []Finding) (*model.SecurityReport, error) {
	score := computeScore(findings)
	summary, _ := json.Marshal(findings)
	report := &model.SecurityReport{Kind: kind, Score: score, FindingCount: len(findings), Summary: string(summary)}
	if err := s.repo.SecurityReportCreate(report); err != nil {
		return nil, err
	}
	return report, nil
}

// ScanAll 执行一次全量扫描（基线 + 镜像）。
func (s *SecurityService) ScanAll(ctx context.Context) ([]*model.SecurityReport, error) {
	reports := make([]*model.SecurityReport, 0, 2)
	if r, err := s.ScanBaseline(ctx); err == nil {
		reports = append(reports, r)
	} else {
		s.log.Warn("baseline scan failed", zap.Error(err))
	}
	if r, err := s.ScanImages(ctx); err == nil {
		reports = append(reports, r)
	} else {
		s.log.Warn("image scan failed", zap.Error(err))
	}
	return reports, nil
}

// Dashboard 聚合最新各维度报告。
func (s *SecurityService) Dashboard(ctx context.Context) (map[string]any, error) {
	kinds := []string{"baseline", "image"}
	reports := make([]map[string]any, 0, len(kinds))
	totalFindings := 0
	worst := 100
	for _, k := range kinds {
		r, err := s.repo.SecurityReportLatest(k)
		if err != nil {
			continue
		}
		var findings []Finding
		_ = json.Unmarshal([]byte(r.Summary), &findings)
		totalFindings += r.FindingCount
		if r.Score < worst {
			worst = r.Score
		}
		reports = append(reports, map[string]any{
			"kind": r.Kind, "score": r.Score, "finding_count": r.FindingCount,
			"findings": findings, "scanned_at": r.CreatedAt,
		})
	}
	return map[string]any{
		"score":          worst,
		"finding_count":  totalFindings,
		"reports":        reports,
		"baseline_rules": []string{"非特权", "no-new-privileges", "CapDrop ALL + 最小 CapAdd", "PidsLimit", "CPU/内存上限", "非 root 运行"},
		"image_rules":    []string{"禁止 latest 用于生产", "悬空镜像提示", "镜像体积 >2GB 提示"},
	}, nil
}

// Reports 扫描历史。
func (s *SecurityService) Reports(ctx context.Context, limit int) ([]model.SecurityReport, error) {
	return s.repo.SecurityReportList(limit)
}

// ReportFindings 单份报告发现项。
func (s *SecurityService) ReportFindings(ctx context.Context, id uint64) ([]Finding, *model.SecurityReport, error) {
	r, err := s.repo.SecurityReportGetByID(id)
	if err != nil {
		return nil, nil, err
	}
	var findings []Finding
	_ = json.Unmarshal([]byte(r.Summary), &findings)
	return findings, r, nil
}
