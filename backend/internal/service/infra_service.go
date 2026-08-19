// Package service：基础设施服务（镜像/网络/存储/Registry）。
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dxcloud/cloud-api/internal/docker"
	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrInUse     = errors.New("resource in use")
	ErrNotFound2 = errors.New("not found")
)

// ---------- 镜像服务 ----------

type ImageService struct {
	repo         *repository.Repos
	compute      docker.ImageProvider
	iamSvc       *iam.Service
	settings     *SettingsService // 区域/镜像加速源（中国大陆自动改写官方镜像拉取地址）
	log          *zap.Logger
	pullCtx      context.Context // 镜像拉取上下文（可被 Stop 取消，服务关闭时停止所有 in-flight 拉取）
	pullCancel   context.CancelFunc
	pullMu       sync.Mutex
	pullLogs     map[uint64]string
	pullLines    map[uint64]string
	noisySeen    map[uint64]map[string]bool
	activePulls  map[uint64]bool
	pullProgress map[uint64]float64
}

func NewImageService(repo *repository.Repos, compute docker.ImageProvider, iamSvc *iam.Service, settings *SettingsService, log *zap.Logger) *ImageService {
	pullCtx, pullCancel := context.WithCancel(context.Background())
	return &ImageService{
		repo: repo, compute: compute, iamSvc: iamSvc, settings: settings, log: log,
		pullCtx: pullCtx, pullCancel: pullCancel,
		pullLogs: make(map[uint64]string), pullLines: make(map[uint64]string),
		noisySeen:    make(map[uint64]map[string]bool),
		activePulls:  make(map[uint64]bool),
		pullProgress: make(map[uint64]float64),
	}
}

// Stop 取消所有进行中的镜像拉取（服务关闭时调用）。
func (s *ImageService) Stop() {
	s.pullCancel()
}

func (s *ImageService) List(page, size int, keyword string, orgID uint64) ([]model.DockerImage, int64, error) {
	return s.repo.ImageList(page, size, keyword, orgID)
}

// Pull 异步拉取镜像：DB 先行（pulling）→ goroutine 拉取 → 更新 ready/failed。
// 中国大陆区域下，官方 Docker Hub 镜像自动经加速源拉取（用户侧仍展示原始引用）。
func (s *ImageService) Pull(ctx context.Context, ac AccessCtx, imageRef string, ip, requestID string) (*model.DockerImage, error) {
	repo, tag := splitImageRef(imageRef)
	if repo == "" {
		return nil, errors.New("镜像地址格式不正确")
	}
	pullRefs, mirrored := s.resolvePullRefs(imageRef)
	if existing, err := s.repo.ImageGetByRepoTag(repo, tag, ac.OrgID); err == nil && s.imageInScope(ac.OrgID, existing) {
		if existing.Status == model.ImageStatusPulling {
			if s.isPullActive(existing.ID) {
				return existing, nil // 幂等：正在拉取
			}
			// 服务重启后恢复遗留的 pulling 任务
			s.appendPullLog(existing.ID, "检测到未完成的历史任务，正在重新开始\n")
		}
		existing.Status = model.ImageStatusPulling
		existing.PullError = ""
		existing.PullLog = ""
		_ = s.repo.ImageUpdate(existing)
		uid := ac.UserID
		s.iamSvc.Audit(ctx, &uid, "image.pull", "image", imageRef, ip, requestID, 1, nil)
		s.setPullActive(existing.ID, true)
		go func() {
			defer s.setPullActive(existing.ID, false)
			s.doPull(existing.ID, imageRef, pullRefs, mirrored, ac.UserID, ac.OrgID)
		}()
		return existing, nil
	}
	img := &model.DockerImage{
		Repo: repo, Tag: tag, Status: model.ImageStatusPulling, Source: "pull",
	}
	if ac.OrgID > 0 {
		v := ac.OrgID
		img.OrgID = &v
	}
	if err := s.repo.ImageCreate(img); err != nil {
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "image.pull", "image", imageRef, ip, requestID, 1, nil)
	s.setPullActive(img.ID, true)
	go func() {
		defer s.setPullActive(img.ID, false)
		s.doPull(img.ID, imageRef, pullRefs, mirrored, ac.UserID, ac.OrgID)
	}()
	return img, nil
}

func (s *ImageService) isPullActive(id uint64) bool {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	return s.activePulls[id]
}

func (s *ImageService) setPullActive(id uint64, active bool) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	s.activePulls[id] = active
}

func (s *ImageService) setPullProgress(id uint64, percent float64) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	s.pullProgress[id] = percent
}

var dockerProgressPattern = regexp.MustCompile(`(?i)([\d.]+)\s*(B|KB|MB|GB)?\s*/\s*([\d.]+)\s*(B|KB|MB|GB)`)

func parseDockerProgress(progress string) (float64, bool) {
	m := dockerProgressPattern.FindStringSubmatch(progress)
	if len(m) < 5 {
		return 0, false
	}
	done, err1 := strconv.ParseFloat(m[1], 64)
	total, err2 := strconv.ParseFloat(m[3], 64)
	if err1 != nil || err2 != nil || total <= 0 {
		return 0, false
	}
	done = applyByteUnit(done, strings.ToUpper(m[2]))
	total = applyByteUnit(total, strings.ToUpper(m[4]))
	if total <= 0 {
		return 0, false
	}
	return done / total * 100, true
}

func applyByteUnit(value float64, unit string) float64 {
	switch unit {
	case "GB":
		return value * 1024 * 1024 * 1024
	case "MB":
		return value * 1024 * 1024
	case "KB":
		return value * 1024
	default:
		return value
	}
}

func (s *ImageService) imageInScope(orgID uint64, img *model.DockerImage) bool {
	if orgID == 0 {
		return img.OrgID == nil
	}
	return img.OrgID != nil && *img.OrgID == orgID
}

func (s *ImageService) resolvePullRefs(ref string) ([]string, bool) {
	if s.settings == nil {
		return []string{ref}, false
	}
	return s.settings.ResolvePullRefs(ref)
}

func (s *ImageService) doPull(id uint64, imageRef string, pullRefs []string, mirrored bool, userID, orgID uint64) {
	ctx, cancel := context.WithTimeout(s.pullCtx, 30*time.Minute)
	defer cancel()
	img, err := s.repo.ImageGetByID(id)
	if err != nil {
		return
	}
	s.setPullLog(id, fmt.Sprintf("开始拉取 %s（配置了 %d 个候选源）\n", imageRef, len(pullRefs)))
	var usedRef string
	var lastErr error
	for index, pullRef := range pullRefs {
		if index > 0 {
			s.appendPullLog(id, fmt.Sprintf("当前源失败，自动切换备用源 %s\n", pullRef))
			s.persistPullLog(id)
		}
		s.log.Info("image pull attempt", zap.String("ref", imageRef), zap.String("pull_ref", pullRef), zap.Int("attempt", index+1))
		attemptCtx, attemptCancel := context.WithTimeout(ctx, 90*time.Second)
		var lastLogSync time.Time
		err := s.compute.PullImageWithProgress(attemptCtx, pullRef, func(p docker.ImagePullProgress) {
			line := strings.TrimSpace(p.Status)
			if p.Progress != "" {
				line += " " + strings.TrimSpace(p.Progress)
			}
			if percent, ok := parseDockerProgress(p.Progress); ok {
				s.setPullProgress(id, percent)
			}
			if line != "" {
				s.appendPullLogLine(id, line)
			}
			if time.Since(lastLogSync) >= time.Second {
				lastLogSync = time.Now()
				s.persistPullLog(id)
			}
		})
		attemptCancel()
		if err == nil {
			s.setPullProgress(id, 100)
			usedRef = pullRef
			break
		}
		lastErr = err
		s.appendPullLog(id, fmt.Sprintf("[失败] %s：%v\n", pullRef, err))
		s.persistPullLog(id)
	}
	if usedRef == "" {
		s.setPullProgress(id, 0)
		img.Status = model.ImageStatusFailed
		if mirrored {
			img.PullError = fmt.Sprintf("所有国内加速源均拉取失败：%v（可在 设置 → 区域与镜像源 中更换源或开启代理后重试）", lastErr)
		} else {
			img.PullError = lastErr.Error()
		}
		s.appendPullLog(id, "[失败] 拉取任务已结束\n")
		s.pullMu.Lock()
		img.PullLog = s.pullLogs[id]
		s.pullMu.Unlock()
		_ = s.repo.ImageUpdate(img)
		notify(s.repo, userID, "image", "镜像拉取失败", imageRef+"："+img.PullError, "/images")
		return
	}
	// 加速源拉取的镜像打回原始引用，保证后续按原始 ref 使用/展示
	if mirrored {
		if err := s.compute.TagImage(ctx, usedRef, imageRef); err != nil {
			s.log.Warn("retag mirrored image failed", zap.String("pull_ref", usedRef), zap.String("ref", imageRef), zap.Error(err))
		}
	}
	info, err := s.compute.InspectImage(ctx, imageRef)
	if err != nil && mirrored {
		info, err = s.compute.InspectImage(ctx, usedRef)
	}
	if err == nil {
		img.ImageID = info.ID
		img.SizeBytes = info.Size
		img.DockerCreatedAt = &info.CreatedAt
	}
	img.Status = model.ImageStatusReady
	img.PullError = ""
	s.appendPullLog(id, "拉取完成，正在写入镜像索引\n")
	s.pullMu.Lock()
	img.PullLog = s.pullLogs[id]
	s.pullMu.Unlock()
	_ = s.repo.ImageUpdate(img)
	s.log.Info("image pulled", zap.String("ref", imageRef), zap.Bool("mirrored", mirrored))
	sizeTxt := ""
	if img.SizeBytes > 0 {
		sizeTxt = fmt.Sprintf("，大小 %.1f MB", float64(img.SizeBytes)/1024/1024)
	}
	notify(s.repo, userID, "image", "镜像拉取成功", imageRef+" 已就绪"+sizeTxt+"，可直接用于创建实例", "/images")
}

func (s *ImageService) appendPullLog(id uint64, text string) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	s.pullLogs[id] += text
	if len(s.pullLogs[id]) > 64*1024 {
		s.pullLogs[id] = s.pullLogs[id][len(s.pullLogs[id])-48*1024:]
	}
}

func (s *ImageService) appendPullLogLine(id uint64, line string) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()

	if s.pullLines[id] == line {
		return
	}
	s.pullLines[id] = line

	switch strings.TrimSpace(strings.SplitN(line, " ", 2)[0]) {
	case "Downloading", "Already exists", "Extracting", "Pulling fs layer":
		if s.noisySeen[id] == nil {
			s.noisySeen[id] = make(map[string]bool)
		}
		key := strings.SplitN(line, " ", 2)[0]
		if s.noisySeen[id][key] {
			return
		}
		s.noisySeen[id][key] = true
		s.pullLogs[id] += line + "（后续重复进度已折叠）\n"
	default:
		s.pullLogs[id] += line + "\n"
	}

	if len(s.pullLogs[id]) > 64*1024 {
		s.pullLogs[id] = s.pullLogs[id][len(s.pullLogs[id])-48*1024:]
	}
}

func (s *ImageService) setPullLog(id uint64, text string) {
	s.pullMu.Lock()
	defer s.pullMu.Unlock()
	s.pullLogs[id] = text
	delete(s.pullLines, id)
	delete(s.noisySeen, id)
}

func (s *ImageService) persistPullLog(id uint64) {
	s.pullMu.Lock()
	text := s.pullLogs[id]
	s.pullMu.Unlock()
	img, err := s.repo.ImageGetByID(id)
	if err == nil {
		img.PullLog = text
		_ = s.repo.ImageUpdate(img)
	}
}

// PullLogs 返回镜像任务的当前状态与日志。
func (s *ImageService) PullLogs(orgID, id uint64) (status, pullError, logs string, progress *float64, err error) {
	img, err := s.repo.ImageGetByID(id)
	if err != nil {
		return "", "", "", nil, err
	}
	if !s.imageInScope(orgID, img) {
		return "", "", "", nil, ErrNotFound
	}
	s.pullMu.Lock()
	logs = s.pullLogs[id]
	s.pullMu.Unlock()
	if logs == "" {
		logs = img.PullLog
	}
	s.pullMu.Lock()
	if p, ok := s.pullProgress[id]; ok {
		progress = &p
	}
	s.pullMu.Unlock()
	return img.Status, img.PullError, logs, progress, nil
}

func (s *ImageService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	img, err := s.repo.ImageGetByID(id)
	if err != nil {
		return err
	}
	if !s.imageInScope(ac.OrgID, img) {
		return ErrNotFound
	}
	ref := img.Repo + ":" + img.Tag
	if err := s.compute.RemoveImage(ctx, ref, false); err != nil && !strings.Contains(err.Error(), "No such image") {
		return err
	}
	if err := s.repo.ImageSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "image.delete", "image", ref, ip, requestID, 1, nil)
	return nil
}

func (s *ImageService) Tag(ctx context.Context, ac AccessCtx, id uint64, newRepo, newTag, ip, requestID string) (*model.DockerImage, error) {
	img, err := s.repo.ImageGetByID(id)
	if err != nil {
		return nil, err
	}
	if !s.imageInScope(ac.OrgID, img) {
		return nil, ErrNotFound
	}
	src := img.Repo + ":" + img.Tag
	dst := newRepo + ":" + newTag
	if err := s.compute.TagImage(ctx, src, dst); err != nil {
		return nil, err
	}
	info, err := s.compute.InspectImage(ctx, dst)
	row := &model.DockerImage{
		Repo: newRepo, Tag: newTag, Source: "import",
	}
	if ac.OrgID > 0 {
		v := ac.OrgID
		row.OrgID = &v
	}
	if err == nil {
		row.ImageID = info.ID
		row.SizeBytes = info.Size
		row.DockerCreatedAt = &info.CreatedAt
	}
	row.Status = model.ImageStatusReady
	if err := s.repo.ImageCreate(row); err != nil {
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "image.tag", "image", src+" -> "+dst, ip, requestID, 1, nil)
	return row, nil
}

func splitImageRef(ref string) (repo, tag string) {
	parts := strings.SplitN(ref, ":", 2)
	repo = parts[0]
	tag = "latest"
	if len(parts) == 2 && parts[1] != "" {
		tag = parts[1]
	}
	return repo, tag
}

// ---------- 镜像搜索建议（拉取输入自动补全） ----------

// ImageSearchResult 镜像搜索建议项。
type ImageSearchResult struct {
	Name        string `json:"name"`        // 镜像名（可直接用于拉取，如 nginx、bitnami/redis）
	Description string `json:"description"` // 描述（常用目录为中文）
	Official    bool   `json:"official"`    // 是否官方镜像
	Source      string `json:"source"`      // hub=Docker Hub 在线搜索 / local=内置常用目录
}

// commonImageCatalog 内置常用镜像目录：Docker Hub 搜索接口在中国大陆网络不可达时的兜底，
// 同时保证离线/弱网环境下输入补全依然可用。
var commonImageCatalog = []ImageSearchResult{
	{Name: "nginx", Description: "高性能 Web 服务器 / 反向代理", Official: true},
	{Name: "redis", Description: "内存数据库 / 缓存 / 消息队列", Official: true},
	{Name: "mysql", Description: "MySQL 关系型数据库", Official: true},
	{Name: "mariadb", Description: "MariaDB 数据库（MySQL 分支）", Official: true},
	{Name: "postgres", Description: "PostgreSQL 关系型数据库", Official: true},
	{Name: "mongo", Description: "MongoDB 文档数据库", Official: true},
	{Name: "ubuntu", Description: "Ubuntu 基础系统镜像", Official: true},
	{Name: "debian", Description: "Debian 基础系统镜像", Official: true},
	{Name: "alpine", Description: "极简 Linux 基础镜像（约 5MB）", Official: true},
	{Name: "centos", Description: "CentOS 基础系统镜像", Official: true},
	{Name: "python", Description: "Python 运行环境", Official: true},
	{Name: "node", Description: "Node.js 运行环境", Official: true},
	{Name: "golang", Description: "Go 语言编译运行环境", Official: true},
	{Name: "openjdk", Description: "OpenJDK Java 运行环境", Official: true},
	{Name: "eclipse-temurin", Description: "Eclipse Temurin JDK（Spring Boot 推荐）", Official: true},
	{Name: "maven", Description: "Maven 构建环境（Java 项目编译）", Official: true},
	{Name: "tomcat", Description: "Tomcat Java Web 容器", Official: true},
	{Name: "httpd", Description: "Apache HTTP Server", Official: true},
	{Name: "php", Description: "PHP 运行环境", Official: true},
	{Name: "ruby", Description: "Ruby 运行环境", Official: true},
	{Name: "rust", Description: "Rust 编译环境", Official: true},
	{Name: "rabbitmq", Description: "RabbitMQ 消息队列", Official: true},
	{Name: "kafka", Description: "Apache Kafka 消息流平台", Official: true},
	{Name: "zookeeper", Description: "ZooKeeper 分布式协调服务", Official: true},
	{Name: "elasticsearch", Description: "Elasticsearch 搜索分析引擎", Official: true},
	{Name: "kibana", Description: "Kibana 可视化分析平台", Official: true},
	{Name: "grafana", Description: "Grafana 监控可视化面板", Official: true},
	{Name: "prometheus", Description: "Prometheus 监控系统", Official: true},
	{Name: "minio", Description: "MinIO 对象存储（S3 兼容）", Official: true},
	{Name: "traefik", Description: "Traefik 云原生反向代理/网关", Official: true},
	{Name: "caddy", Description: "Caddy Web 服务器（自动 HTTPS）", Official: true},
	{Name: "memcached", Description: "Memcached 分布式缓存", Official: true},
	{Name: "influxdb", Description: "InfluxDB 时序数据库", Official: true},
	{Name: "clickhouse", Description: "ClickHouse 列式分析数据库", Official: true},
	{Name: "neo4j", Description: "Neo4j 图数据库", Official: true},
	{Name: "nacos/nacos-server", Description: "Nacos 注册配置中心（微服务）", Official: false},
	{Name: "bitnami/redis", Description: "Bitnami Redis 发行版", Official: false},
	{Name: "portainer/portainer-ce", Description: "Portainer Docker 可视化管理", Official: false},
	{Name: "jenkins/jenkins", Description: "Jenkins CI/CD 自动化服务器", Official: false},
	{Name: "sonarqube", Description: "SonarQube 代码质量扫描", Official: true},
	{Name: "emqx/emqx", Description: "EMQX MQTT 消息服务器", Official: false},
	{Name: "apache/rocketmq", Description: "RocketMQ 分布式消息中间件", Official: false},
	{Name: "registry", Description: "Docker Registry 私有镜像仓库", Official: true},
	{Name: "docker", Description: "Docker CLI/Engine 镜像", Official: true},
	{Name: "busybox", Description: "BusyBox 极简工具镜像", Official: true},
	{Name: "hello-world", Description: "Docker 连通性测试镜像", Official: true},
}

// Search 镜像搜索建议：内置常用目录优先，未命中时才尝试 Docker Hub；
// 两路结果按名称去重合并，最多返回 limit 条。
func (s *ImageService) Search(ctx context.Context, q string, limit int) []ImageSearchResult {
	q = strings.TrimSpace(strings.ToLower(q))
	if q == "" {
		return []ImageSearchResult{}
	}
	if limit <= 0 || limit > 30 {
		limit = 10
	}

	merged := make([]ImageSearchResult, 0, limit)
	seen := make(map[string]bool, limit)
	add := func(r ImageSearchResult) {
		key := strings.ToLower(r.Name)
		if seen[key] {
			return
		}
		seen[key] = true
		merged = append(merged, r)
	}

	// 中国大陆网络优先使用内置目录，常见关键词应立即返回，不等待 Docker Hub 超时。
	prefixHits := make([]ImageSearchResult, 0)
	substrHits := make([]ImageSearchResult, 0)
	for _, r := range commonImageCatalog {
		name := strings.ToLower(r.Name)
		switch {
		case strings.HasPrefix(name, q):
			prefixHits = append(prefixHits, r)
		case strings.Contains(name, q):
			substrHits = append(substrHits, r)
		}
	}
	localHits := append(prefixHits, substrHits...)
	for _, r := range localHits {
		r.Source = "local"
		add(r)
		if len(merged) >= limit {
			break
		}
	}
	if len(merged) > 0 {
		return merged
	}

	// 本地目录没有命中时再尝试 Docker Hub，超时缩短到 2.5 秒。
	onlineCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	for _, r := range s.searchDockerHub(onlineCtx, q) {
		add(r)
		if len(merged) >= limit {
			return merged
		}
	}
	return merged
}

// searchDockerHub 调用 Docker Hub 官方搜索 API（中国大陆网络可能超时，由调用方兜底）。
func (s *ImageService) searchDockerHub(ctx context.Context, q string) []ImageSearchResult {
	url := "https://hub.docker.com/v2/search/repositories/?query=" + urlQueryEscape(q) + "&page_size=10"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/json")
	respBody, err := http.DefaultClient.Do(req)
	if err != nil {
		s.log.Debug("docker hub search unreachable, fallback to local catalog", zap.Error(err))
		return nil
	}
	defer respBody.Body.Close()
	if respBody.StatusCode != http.StatusOK {
		return nil
	}
	body, err := io.ReadAll(io.LimitReader(respBody.Body, 1<<20))
	if err != nil {
		return nil
	}
	var parsed struct {
		Results []struct {
			RepoName         string `json:"repo_name"`
			ShortDescription string `json:"short_description"`
			IsOfficial       bool   `json:"is_official"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil
	}
	out := make([]ImageSearchResult, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		name := strings.TrimPrefix(r.RepoName, "library/")
		if name == "" {
			continue
		}
		out = append(out, ImageSearchResult{Name: name, Description: r.ShortDescription, Official: r.IsOfficial, Source: "hub"})
	}
	return out
}

func urlQueryEscape(q string) string {
	replacer := strings.NewReplacer(" ", "+", "&", "%26", "?", "%3F", "=", "%3D", "#", "%23")
	return replacer.Replace(q)
}

// ---------- 网络服务 ----------

type NetworkService struct {
	repo    *repository.Repos
	compute docker.NetworkProvider
	iamSvc  *iam.Service
	log     *zap.Logger
}

func NewNetworkService(repo *repository.Repos, compute docker.NetworkProvider, iamSvc *iam.Service, log *zap.Logger) *NetworkService {
	return &NetworkService{repo: repo, compute: compute, iamSvc: iamSvc, log: log}
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func orgMatches(orgID uint64, org *uint64) bool {
	if orgID == 0 {
		return org == nil
	}
	return org != nil && *org == orgID
}

func orgPtrForCtx(orgID uint64) *uint64 {
	if orgID == 0 {
		return nil
	}
	v := orgID
	return &v
}

func (s *NetworkService) List(ctx context.Context, ac AccessCtx) ([]model.DockerNetwork, error) {
	return s.repo.NetworkList(ac.OrgID)
}

// Inspect 网络详情（含容器连接关系，运行时实时）。
func (s *NetworkService) Inspect(ctx context.Context, ac AccessCtx, id uint64) (map[string]any, error) {
	n, err := s.repo.NetworkGetByID(id)
	if err != nil {
		return nil, err
	}
	if !orgMatches(ac.OrgID, n.OrgID) {
		return nil, ErrNotFound
	}
	out := map[string]any{
		"id": n.ID, "name": n.Name, "docker_name": n.DockerName,
		"subnet": n.Subnet, "gateway": n.Gateway, "ip_range": n.IPRange,
		"driver": n.Driver, "created_at": n.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if n.DockerNetID != "" {
		if info, err := s.compute.InspectNetwork(ctx, n.DockerNetID); err == nil {
			out["containers"] = info.Containers
		}
	}
	return out, nil
}

func (s *NetworkService) Create(ctx context.Context, ac AccessCtx, name, subnet, gateway, ipRange string, internal bool, ip, requestID string) (*model.DockerNetwork, error) {
	if name == "" || len(name) > 63 {
		return nil, errors.New("网络名需 1-63 位")
	}
	if _, err := s.repo.NetworkGetByName(name, ac.OrgID); err == nil {
		return nil, errors.New("network name already exists")
	}
	dockerName := "dxn-" + randHex(3)
	info, err := s.compute.CreateNetwork(ctx, docker.NetworkSpec{
		Name: dockerName, Driver: "bridge",
		Subnet: subnet, Gateway: gateway, IPRange: ipRange, Internal: internal,
		Labels: map[string]string{
			"com.dxcloud.kind":         "network",
			"com.dxcloud.network-name": name,
			"com.dxcloud.org-id":       orgLabel(orgPtrForCtx(ac.OrgID)),
		},
	})
	if err != nil {
		return nil, err
	}
	n := &model.DockerNetwork{
		OwnerID: ac.UserID, Name: name, DockerName: dockerName,
		DockerNetID: info.ID, Driver: "bridge",
		Subnet: subnet, Gateway: gateway, IPRange: ipRange, Internal: internal,
	}
	if ac.OrgID > 0 {
		v := ac.OrgID
		n.OrgID = &v
	}
	if err := s.repo.NetworkCreate(n); err != nil {
		_ = s.compute.RemoveNetwork(ctx, info.ID)
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "network.create", "network", name, ip, requestID, 1, map[string]any{"subnet": subnet})
	return n, nil
}

func (s *NetworkService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	n, err := s.repo.NetworkGetByID(id)
	if err != nil {
		return err
	}
	if !orgMatches(ac.OrgID, n.OrgID) {
		return ErrNotFound
	}
	if n.DockerNetID != "" {
		if info, err := s.compute.InspectNetwork(ctx, n.DockerNetID); err == nil && len(info.Containers) > 0 {
			return fmt.Errorf("%w: network has %d connected containers", ErrInUse, len(info.Containers))
		}
		if err := s.compute.RemoveNetwork(ctx, n.DockerNetID); err != nil && !strings.Contains(err.Error(), "not found") {
			return err
		}
	}
	if err := s.repo.NetworkSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "network.delete", "network", n.Name, ip, requestID, 1, nil)
	return nil
}

// Connect 容器加入网络（可选静态 IP）。
func (s *NetworkService) Connect(ctx context.Context, ac AccessCtx, netID uint64, instanceID uint64, fixedIP, ip, requestID string) error {
	n, err := s.repo.NetworkGetByID(netID)
	if err != nil {
		return err
	}
	if !orgMatches(ac.OrgID, n.OrgID) {
		return ErrNotFound
	}
	inst, err := s.repo.EcsGetByID(instanceID)
	if err != nil {
		return errors.New("instance not found")
	}
	if !orgMatches(ac.OrgID, inst.OrgID) {
		return ErrForbidden
	}
	if inst.ContainerID == "" {
		return errors.New("instance has no container")
	}
	if err := s.compute.ConnectNetwork(ctx, n.DockerNetID, inst.ContainerID, fixedIP); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "network.connect", "network", n.Name, ip, requestID, 1, map[string]any{"instance": inst.InstanceNo, "ip": fixedIP})
	return nil
}

func (s *NetworkService) Disconnect(ctx context.Context, ac AccessCtx, netID uint64, instanceID uint64, ip, requestID string) error {
	n, err := s.repo.NetworkGetByID(netID)
	if err != nil {
		return err
	}
	if !orgMatches(ac.OrgID, n.OrgID) {
		return ErrNotFound
	}
	inst, err := s.repo.EcsGetByID(instanceID)
	if err != nil {
		return errors.New("instance not found")
	}
	if !orgMatches(ac.OrgID, inst.OrgID) {
		return ErrForbidden
	}
	if inst.ContainerID == "" {
		return errors.New("instance has no container")
	}
	if err := s.compute.DisconnectNetwork(ctx, n.DockerNetID, inst.ContainerID); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "network.disconnect", "network", n.Name, ip, requestID, 1, map[string]any{"instance": inst.InstanceNo})
	return nil
}

// ---------- 存储服务 ----------

type VolumeService struct {
	repo    *repository.Repos
	compute docker.StorageProvider
	iamSvc  *iam.Service
	log     *zap.Logger
}

func NewVolumeService(repo *repository.Repos, compute docker.StorageProvider, iamSvc *iam.Service, log *zap.Logger) *VolumeService {
	return &VolumeService{repo: repo, compute: compute, iamSvc: iamSvc, log: log}
}

func (s *VolumeService) List(ac AccessCtx) ([]model.DockerVolume, error) {
	return s.repo.VolumeList(ac.OrgID)
}

func (s *VolumeService) Create(ctx context.Context, ac AccessCtx, name string, capacityGB int, ip, requestID string) (*model.DockerVolume, error) {
	if name == "" || len(name) > 63 {
		return nil, errors.New("磁盘名需 1-63 位")
	}
	if _, err := s.repo.VolumeGetByName(name, ac.OrgID); err == nil {
		return nil, errors.New("volume name already exists")
	}
	dockerName := "dxv-" + randHex(3)
	info, err := s.compute.CreateVolume(ctx, dockerName, map[string]string{
		"com.dxcloud.kind":        "volume",
		"com.dxcloud.volume-name": name,
		"com.dxcloud.org-id":      orgLabel(orgPtrForCtx(ac.OrgID)),
	})
	if err != nil {
		return nil, err
	}
	if capacityGB <= 0 {
		capacityGB = 10
	}
	v := &model.DockerVolume{
		OwnerID: ac.UserID, Name: name, DockerName: dockerName,
		Driver: info.Driver, Mountpoint: info.Mountpoint, CapacityGB: capacityGB,
	}
	if ac.OrgID > 0 {
		orgID := ac.OrgID
		v.OrgID = &orgID
	}
	if err := s.repo.VolumeCreate(v); err != nil {
		_ = s.compute.RemoveVolume(ctx, dockerName)
		return nil, err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "volume.create", "volume", name, ip, requestID, 1, map[string]any{"capacity_gb": capacityGB})
	return v, nil
}

func (s *VolumeService) Delete(ctx context.Context, ac AccessCtx, id uint64, ip, requestID string) error {
	v, err := s.repo.VolumeGetByID(id)
	if err != nil {
		return err
	}
	if !orgMatches(ac.OrgID, v.OrgID) {
		return ErrNotFound
	}
	// 使用中检查：扫描实例 mounts
	instances, _, err := s.repo.EcsList(repository.EcsFilter{Page: 1, Size: 10000})
	if err != nil {
		return err
	}
	for i := range instances {
		for _, m := range parseMounts(&instances[i]) {
			if m.VolumeName == v.DockerName {
				return fmt.Errorf("%w: mounted by instance %s", ErrInUse, instances[i].Name)
			}
		}
	}
	if err := s.compute.RemoveVolume(ctx, v.DockerName); err != nil && !strings.Contains(err.Error(), "No such volume") {
		return err
	}
	if err := s.repo.VolumeSoftDelete(id); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "volume.delete", "volume", v.Name, ip, requestID, 1, nil)
	return nil
}

// ---------- Registry 服务 ----------

type RegistryService struct {
	repo              *repository.Repos
	iamSvc            *iam.Service
	compute           docker.ImageProvider
	engineRegistryURL string // daemon 视角 registry 地址（宿主机 VM 可达）
	httpc             *http.Client
	log               *zap.Logger
}

func NewRegistryService(repo *repository.Repos, compute docker.ImageProvider, iamSvc *iam.Service, engineRegistryURL string, log *zap.Logger) *RegistryService {
	return &RegistryService{
		repo: repo, compute: compute, iamSvc: iamSvc,
		engineRegistryURL: engineRegistryURL,
		httpc:             &http.Client{Timeout: 15 * time.Second},
		log:               log,
	}
}

func (s *RegistryService) ListRegistries() ([]model.Registry, error) {
	return s.repo.RegistryList()
}

// restBase 归一化 registry 地址为可被 http 客户端使用的 URL。
func restBase(url string) string {
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return strings.TrimSuffix(url, "/")
	}
	return "http://" + strings.TrimSuffix(url, "/")
}

// Repositories 代理 registry REST API：catalog + tags。
func (s *RegistryService) Repositories(ctx context.Context, registryID uint64) ([]map[string]any, error) {
	reg, err := s.repo.RegistryGetByID(registryID)
	if err != nil {
		return nil, err
	}
	base := restBase(reg.URL)
	var catalog struct {
		Repositories []string `json:"repositories"`
	}
	if err := s.registryGet(ctx, base+"/v2/_catalog", &catalog); err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(catalog.Repositories))
	for _, name := range catalog.Repositories {
		var tags struct {
			Name string   `json:"name"`
			Tags []string `json:"tags"`
		}
		item := map[string]any{"name": name, "tags": []string{}}
		if err := s.registryGet(ctx, base+"/v2/"+name+"/tags/list", &tags); err == nil {
			item["tags"] = tags.Tags
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *RegistryService) registryGet(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := s.httpc.Do(req)
	if err != nil {
		return fmt.Errorf("registry unreachable (%s): %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("registry http %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

// Pull 从私有 Registry 拉取镜像到引擎（并登记镜像中心）。
// 注意：docker pull 由 daemon 执行，必须使用引擎可达地址（engineRegistryURL），而非容器内部地址。
func (s *RegistryService) Pull(ctx context.Context, ac AccessCtx, registryID uint64, name, tag, ip, requestID string) error {
	reg, err := s.repo.RegistryGetByID(registryID)
	if err != nil {
		return err
	}
	engineRef := s.engineRegistryURL + "/" + name + ":" + tag
	if err := s.compute.PullImage(ctx, engineRef); err != nil {
		return err
	}
	info, err := s.compute.InspectImage(ctx, engineRef)
	img := &model.DockerImage{Repo: reg.URL + "/" + name, Tag: tag, Source: "pull", Status: model.ImageStatusReady}
	if ac.OrgID > 0 {
		v := ac.OrgID
		img.OrgID = &v
	}
	if err == nil {
		img.ImageID = info.ID
		img.SizeBytes = info.Size
		img.DockerCreatedAt = &info.CreatedAt
	}
	if _, err := s.repo.ImageGetByRepoTag(img.Repo, img.Tag, ac.OrgID); err == nil {
		return nil
	}
	if err := s.repo.ImageCreate(img); err != nil {
		return err
	}
	uid := ac.UserID
	s.iamSvc.Audit(ctx, &uid, "registry.pull", "registry", name+":"+tag, ip, requestID, 1, nil)
	return nil
}

// DeleteTag 删除 registry 中指定 tag（HEAD 取 digest → DELETE manifest）。
func (s *RegistryService) DeleteTag(ctx context.Context, registryID uint64, name, tag string) error {
	reg, err := s.repo.RegistryGetByID(registryID)
	if err != nil {
		return err
	}
	base := restBase(reg.URL)
	manifestURL := base + "/v2/" + name + "/manifests/" + tag
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, manifestURL, nil)
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json, application/vnd.oci.image.manifest.v1+json")
	resp, err := s.httpc.Do(req)
	if err != nil {
		return err
	}
	digest := resp.Header.Get("Docker-Content-Digest")
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
	if digest == "" {
		return errors.New("manifest digest not found")
	}
	delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, base+"/v2/"+name+"/manifests/"+digest, nil)
	delResp, err := s.httpc.Do(delReq)
	if err != nil {
		return err
	}
	defer delResp.Body.Close()
	if delResp.StatusCode >= 300 {
		return fmt.Errorf("delete manifest http %d", delResp.StatusCode)
	}
	return nil
}

// parseMounts 解析实例挂载 JSON。
func parseMounts(inst *model.EcsInstance) []docker.Mount {
	if inst.Mounts == "" {
		return nil
	}
	var out []docker.Mount
	_ = json.Unmarshal([]byte(inst.Mounts), &out)
	return out
}
