// Package service：系统设置服务（区域/镜像加速源）。
package service

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"go.uber.org/zap"
)

// 默认镜像加速源（中国大陆）。公共加速源可用性会随时间变化，
// 用户可在「设置 → 区域与镜像源」中随时切换或自定义。
const DefaultCNMirror = "docker.1ms.run"

// MirrorCandidates 预置候选加速源（前端下拉展示，用户也可手填）。
var MirrorCandidates = []string{
	"docker.1ms.run",
	"docker.m.daocloud.io",
	"hub.rat.dev",
	"dockerproxy.net",
}

type SettingsService struct {
	repo  *repository.Repos
	log   *zap.Logger
	httpc *http.Client
}

func NewSettingsService(repo *repository.Repos, log *zap.Logger) *SettingsService {
	return &SettingsService{repo: repo, log: log, httpc: &http.Client{Timeout: 6 * time.Second}}
}

func jsonString(v string) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func parseStringSetting(raw string, fallback string) string {
	var v string
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		return fallback
	}
	return v
}

// Region 当前区域（默认 cn：面向中国大陆用户的开箱体验）。
func (s *SettingsService) Region() string {
	st, err := s.repo.SettingsGet(model.SettingRegion)
	if err != nil {
		return model.RegionCN
	}
	r := parseStringSetting(st.Value, model.RegionCN)
	if r != model.RegionCN && r != model.RegionGlobal {
		return model.RegionCN
	}
	return r
}

// Mirror 当前配置的加速源域名（默认 DefaultCNMirror）。
func (s *SettingsService) Mirror() string {
	st, err := s.repo.SettingsGet(model.SettingRegistryMirror)
	if err != nil {
		return DefaultCNMirror
	}
	m := strings.TrimSpace(parseStringSetting(st.Value, DefaultCNMirror))
	if m == "" {
		return DefaultCNMirror
	}
	return strings.TrimSuffix(m, "/")
}

// SetRegionAndMirror 保存区域与加速源（updatedBy 可为 nil）。
func (s *SettingsService) SetRegionAndMirror(region, mirror string, updatedBy *uint64) error {
	if region != "" {
		if region != model.RegionCN && region != model.RegionGlobal {
			return errInvalidRegion
		}
		if err := s.repo.SettingsUpsert(&model.SystemSetting{
			Key: model.SettingRegion, Value: jsonString(region),
			Description: "区域：cn=中国大陆 / global=非中国大陆", UpdatedBy: updatedBy,
		}); err != nil {
			return err
		}
	}
	if strings.TrimSpace(mirror) != "" {
		m := strings.TrimSuffix(strings.TrimSpace(mirror), "/")
		m = strings.TrimPrefix(strings.TrimPrefix(m, "https://"), "http://")
		if err := s.repo.SettingsUpsert(&model.SystemSetting{
			Key: model.SettingRegistryMirror, Value: jsonString(m),
			Description: "中国大陆区域拉取官方镜像使用的加速源域名", UpdatedBy: updatedBy,
		}); err != nil {
			return err
		}
	}
	s.log.Info("settings updated", zap.String("region", region), zap.String("mirror", mirror))
	return nil
}

// Overview 设置页展示数据。
func (s *SettingsService) Overview() map[string]any {
	return map[string]any{
		"region":            s.Region(),
		"registry_mirror":   s.Mirror(),
		"mirror_candidates": MirrorCandidates,
		"default_mirror":    DefaultCNMirror,
	}
}

// TestMirror 探测镜像加速源的 Registry v2 端点。公共源可能只允许
// /v2/ 健康检查，也可能对 HEAD 行为不一致，因此返回可达性而非承诺可用性。
func (s *SettingsService) TestMirror(mirror string) map[string]any {
	mirror = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(mirror, "https://"), "http://"))
	if mirror == "" {
		return map[string]any{"reachable": false, "message": "镜像源地址不能为空", "mirror": ""}
	}
	mirror = strings.TrimSuffix(mirror, "/")
	req, err := http.NewRequest(http.MethodGet, "https://"+mirror+"/v2/", nil)
	if err != nil {
		return map[string]any{"reachable": false, "message": "镜像源地址格式错误", "mirror": mirror}
	}
	resp, err := s.httpc.Do(req)
	if err != nil {
		return map[string]any{"reachable": false, "message": "连接超时或源不可达，请换一个源", "mirror": mirror}
	}
	defer resp.Body.Close()
	reachable := resp.StatusCode < 500
	return map[string]any{
		"reachable": reachable,
		"message": func() string {
			if reachable {
				return "镜像源可访问"
			}
			return "镜像源返回异常状态码 " + resp.Status
		}(),
		"mirror":      mirror,
		"status_code": resp.StatusCode,
	}
}

// ResolvePullRef 计算实际拉取引用：
// 中国大陆区域 + 官方 Docker Hub 镜像（引用首段不含域名）→ 加加速源前缀；
// 其余情况（私有 registry、已带域名、global 区域）原样返回。
// 返回值 mirrored=true 表示发生了改写，调用方拉取成功后应把镜像打回原始 ref。
func (s *SettingsService) ResolvePullRef(ref string) (pullRef string, mirrored bool) {
	return s.resolvePullRef(ref, s.Mirror())
}

func (s *SettingsService) resolvePullRef(ref, mirror string) (pullRef string, mirrored bool) {
	if s.Region() != model.RegionCN {
		return ref, false
	}
	if mirror == "" {
		return ref, false
	}
	repo, tag := splitImageRef(ref)
	if repo == "" {
		return ref, false
	}
	first := repo
	if i := strings.Index(repo, "/"); i > 0 {
		first = repo[:i]
	}
	// 首段含 . 或 : 视为自有 registry 域名（如 host.docker.internal:15000、registry.example.com）
	if strings.ContainsAny(first, ".:") || first == "localhost" {
		return ref, false
	}
	// 已经使用加速源前缀的不重复改写
	if strings.HasPrefix(repo, mirror+"/") || repo == mirror {
		return ref, false
	}
	path := repo
	if !strings.Contains(repo, "/") {
		path = "library/" + repo // Docker Hub 官方镜像位于 library 命名空间
	}
	return mirror + "/" + path + ":" + tag, true
}

// ResolvePullRefs 返回当前配置源和候选源的拉取引用，镜像服务会在前一个源
// 失败时自动尝试下一个源。global、私有 Registry 和带域名引用保持单候选。
func (s *SettingsService) ResolvePullRefs(ref string) (refs []string, mirrored bool) {
	refs = make([]string, 0, len(MirrorCandidates)+1)
	seen := make(map[string]bool)
	primary, mirrored := s.resolvePullRef(ref, s.Mirror())
	if !mirrored {
		return []string{ref}, false
	}
	refs = append(refs, primary)
	seen[primary] = true
	for _, mirror := range MirrorCandidates {
		if mirror == "" {
			continue
		}
		candidate, _ := s.resolvePullRef(ref, mirror)
		if candidate != "" && !seen[candidate] {
			seen[candidate] = true
			refs = append(refs, candidate)
		}
	}
	return refs, true
}

var errInvalidRegion = newBadRequestError("region 仅支持 cn 或 global")

type badRequestError struct{ msg string }

func (e *badRequestError) Error() string  { return e.msg }
func newBadRequestError(msg string) error { return &badRequestError{msg: msg} }
