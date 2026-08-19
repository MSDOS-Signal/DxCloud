// Package service - AI 助手：代理智谱 GLM（免费 glm-4-flash），SSE 流式转发。
// API Key 只存在于服务端，前端零接触。
package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dxcloud/cloud-api/internal/config"
	"go.uber.org/zap"
)

// ChatMessage 前端提交的对话消息（role: user/assistant，仅透传白名单角色）。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type AIService struct {
	cfg *config.Config
	log *zap.Logger
}

func NewAIService(cfg *config.Config, log *zap.Logger) *AIService {
	return &AIService{cfg: cfg, log: log}
}

// systemPrompt 内置 DxCloud 全量知识库，保证回答贴合本平台实际功能。
func systemPrompt() string {
	return `你是「多晓云」DxCloud 平台的内置 AI 助手「多晓」（英文名 DuoXiao，与平台同名——多晓即 Dx 的谐音，寓意通晓云上万事），运行在平台控制台的悬浮窗里。你的唯一职责是解答用户关于多晓云 DxCloud 平台本身的使用问题。请严格依据以下平台知识回答；知识未覆盖的内容如实说明并给出通用建议，禁止编造平台不存在的功能。

【平台定位】
多晓云（DxCloud，中文名与英文名同用）是一个基于 Docker 的轻量级一体化云平台，模拟公有云 ECS 控制台体验，整合 IaaS（云主机/网络/存储）、PaaS（应用托管）、SaaS 与自研 CI/CD。所有资源都是真实的 Docker 容器/卷/网络，非模拟数据。

【技术架构】
- 前端：Nuxt 3 + Vue 3 + TypeScript + Naive UI（SPA 单页应用），ECharts 图表
- 后端：Go + Gin + GORM + MySQL + Redis，JWT 认证 + RBAC 权限
- 运行时：Docker 引擎（Linux 容器），私有 Registry（dx-registry，端口 5000/15000）
- 边缘网关：Traefik（80 端口路由到前端 cloud-web:3000 与后端 cloud-api:8080）
- 部署方式：docker compose 一键启动（mysql/redis/registry/api/web/proxy）

【功能模块地图】
1. 概览 Dashboard：资源统计卡（运行实例/vCPU/内存/磁盘）、状态分布环形图、资源用量进度条、实例列表（每 5s 自动刷新）
2. ECS 云主机：创建实例（选镜像、规格 CPU/内存/磁盘、端口映射如 18888→80）、启动/停止/强停/重启、删除；实例详情页有 VNC Web 终端（进入容器执行 Linux 命令）、日志、CPU/内存/网络监控图表、事件记录、云磁盘挂载/卸载
3. 容器实例：Docker 容器级管理与部署历史
4. 镜像中心：Docker Hub 搜索拉取（支持国内加速源）、私有 Registry 仓库、镜像打标签/删除
5. 网络：创建自定义 Docker bridge 网络（自动分配子网如 10.8.8.0/24），容器连接/断开；容器在该网络获得内网 IP（如 10.8.8.2）
6. 存储：Docker volume 云磁盘，创建后挂载到 ECS 实例
7. PaaS 应用：项目→应用→版本→部署→回滚；应用部署为容器并接入 dxcloud_edge 网络由 Traefik 路由；域名绑定
8. CI/CD Pipeline：自研流水线引擎，阶段为 build→test→docker build→push registry→deploy→health check；Git 仓库 + Webhook（GitHub/Gitee）触发，Job 在隔离容器中执行，可查看每阶段日志
9. 监控：平台级 CPU/内存/网络时序图表（30s 自动刷新）；日志审计（操作记录）
10. 组织与多租户：默认空间（单租户模式）+ 组织空间隔离资源；成员管理、资源配额（vCPU/内存/磁盘/实例数上限）
11. 计费中心：按量计费模拟，充值、消费记录、每日 tick 结算
12. 安全中心：容器/镜像漏洞扫描、安全报告；密钥托管（加密存储）
13. 设置：区域选择（中国大陆/非大陆，决定镜像加速源）、镜像源连通性测试

【用户与权限】
- 管理员（Administrator）：用户/角色/权限管理（IAM）、组织管理、配额调整、计费充值、系统设置
- 普通用户：管理自己的 ECS/应用/流水线等资源；资源按用户与组织隔离，互相不可见
- 权限点如 ecs:create、app:deploy 等，后端中间件强制校验

【高频问题知识】
- 容器内网 IP（如 10.8.8.2、172.17.0.2）无法从宿主机浏览器直接访问——这是 Docker 内部虚拟网络，宿主机没有路由。正确方式：创建实例时配置端口映射（如 18888→80），然后浏览器访问 http://localhost:18888
- 实例列表端口列显示 "—" 表示未配置端口映射，外部完全无法访问，需重建实例并配置映射
- Web 终端是真实容器 shell，但精简镜像（如 nginx:alpine）可能缺少 ip/ifconfig 等命令，属正常现象
- 实例状态由 Reconciler 每 15s 与 Docker 引擎对账同步，页面每 5s 刷新
- 镜像拉取慢：设置中切换中国大陆区域并选择加速源（如南京大学、中科大镜像站）
- 登录态过期会自动跳转登录页；JWT Access Token 15 分钟 + Refresh Token 7 天自动续期

【回答风格】
- 用简体中文，简洁准确，直接给操作步骤（菜单路径 > 按钮名）
- 与平台相关时优先给「在哪个页面、点什么、填什么」的路径指引
- 适度使用列表和代码块，不啰嗦`
}

// StreamChat 把对话转发给智谱 GLM 并以 SSE 原样转发增量内容。
// writeChunk(data string) 由 handler 提供，负责写回客户端。
func (s *AIService) StreamChat(ctx context.Context, messages []ChatMessage, pageContext string, writeChunk func(delta string) bool) error {
	if s.cfg.AI.APIKey == "" {
		return fmt.Errorf("AI 服务未配置 ZHIPU_API_KEY")
	}

	// 组装消息：system 知识库 + 页面上下文 + 历史对话（限制长度防滥用）
	sys := systemPrompt()
	if pageContext != "" {
		sys += "\n\n【用户当前所在页面】" + pageContext + "（回答时优先结合该页面语境）"
	}
	history := sanitizeMessages(messages, 20)
	payload := map[string]any{
		"model":       s.cfg.AI.Model,
		"stream":      true,
		"temperature": 0.6,
		"max_tokens":  2048,
		"messages":    append([]ChatMessage{{Role: "system", Content: sys}}, history...),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("请求序列化失败: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.AI.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("构造上游请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.AI.APIKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.log.Warn("zhipu api unreachable", zap.Error(err))
		return fmt.Errorf("AI 服务连接失败，请稍后重试")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		s.log.Warn("zhipu api error", zap.Int("status", resp.StatusCode), zap.String("body", string(b)))
		return fmt.Errorf("AI 服务返回错误（HTTP %d）", resp.StatusCode)
	}

	// 逐行解析 SSE：data: {"choices":[{"delta":{"content":"..."}}]} / data: [DONE]
	br := bufio.NewReaderSize(resp.Body, 32*1024)
	for {
		line, err := br.ReadString('\n')
		if line != "" {
			trimmed := strings.TrimRight(line, "\r\n")
			if strings.HasPrefix(trimmed, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(trimmed, "data:"))
				if data == "[DONE]" {
					return nil
				}
				if delta := extractDelta(data); delta != "" {
					if !writeChunk(delta) {
						return nil // 客户端断开
					}
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return nil // 上游中断时尽量保住已生成内容
		}
	}
}

// extractDelta 从 OpenAI 兼容 chunk 中提取增量文本。
func extractDelta(data string) string {
	var chunk struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return ""
	}
	if len(chunk.Choices) == 0 {
		return ""
	}
	return chunk.Choices[0].Delta.Content
}

// sanitizeMessages 只放行 user/assistant 角色，并截断超长内容。
func sanitizeMessages(msgs []ChatMessage, limit int) []ChatMessage {
	out := make([]ChatMessage, 0, len(msgs))
	for _, m := range msgs {
		if m.Role != "user" && m.Role != "assistant" {
			continue
		}
		content := m.Content
		if len(content) > 4000 {
			content = content[:4000]
		}
		out = append(out, ChatMessage{Role: m.Role, Content: content})
	}
	if len(out) > limit {
		out = out[len(out)-limit:]
	}
	if len(out) == 0 {
		out = []ChatMessage{{Role: "user", Content: "你好"}}
	}
	return out
}
