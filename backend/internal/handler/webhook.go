package handler

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/model"
	"github.com/dxcloud/cloud-api/internal/pipeline"
	"github.com/dxcloud/cloud-api/internal/repository"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/crypto"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type WebhookHandler struct {
	repo   *repository.Repos
	iamSvc *iam.Service
	engine *pipeline.Engine
	key    []byte
}

func NewWebhookHandler(repo *repository.Repos, iamSvc *iam.Service, engine *pipeline.Engine, cryptoKey []byte) *WebhookHandler {
	return &WebhookHandler{repo: repo, iamSvc: iamSvc, engine: engine, key: cryptoKey}
}

func (h *WebhookHandler) ac(c *gin.Context) service.AccessCtx {
	uid := middleware.GetUserID(c)
	roles, _ := h.iamSvc.GetUserRoleCodes(c.Request.Context(), uid)
	return service.AccessCtx{UserID: uid, Roles: roles, OrgID: middleware.GetOrgID(c)}
}

// canAccessPipeline 资源归属校验（与 PipelineHandler 同策略）。
func (h *WebhookHandler) canAccessPipeline(ac service.AccessCtx, p *model.Pipeline) bool {
	inContext := (ac.OrgID == 0 && (p.OrgID == nil || *p.OrgID == 0)) ||
		(ac.OrgID > 0 && p.OrgID != nil && *p.OrgID == ac.OrgID)
	if inContext && ac.CanManage() {
		return true
	}
	if inContext && p.OwnerID == ac.UserID {
		return true
	}
	return ac.HasRole("superadmin")
}

func (h *WebhookHandler) List(c *gin.Context) {
	ac := h.ac(c)
	var pipelineID *uint64
	if v := c.Query("pipeline_id"); v != "" {
		if id, err := strconv.ParseUint(v, 10, 64); err == nil {
			pipelineID = &id
		}
	}
	// 指定了 pipeline_id 时校验归属
	if pipelineID != nil {
		pipe, err := h.repo.PipelineGetByID(*pipelineID)
		if err != nil {
			resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
			return
		}
		if !h.canAccessPipeline(ac, pipe) {
			resp.Fail(c, errcode.CodeForbidden, "forbidden")
			return
		}
	}
	var ownerID *uint64
	if !ac.CanManage() {
		ownerID = &ac.UserID
	}
	items, err := h.repo.WebhookList(pipelineID, ownerID, ac.OrgID)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "query failed")
		return
	}
	resp.OK(c, items)
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var req struct {
		PipelineID   uint64 `json:"pipeline_id" binding:"required"`
		Provider     string `json:"provider" binding:"required"`
		BranchFilter string `json:"branch_filter"`
		Secret       string `json:"secret"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid request")
		return
	}
	if req.Provider != "github" && req.Provider != "gitlab" && req.Provider != "gitee" {
		resp.Fail(c, errcode.CodeBadRequest, "provider 需为 github/gitlab/gitee")
		return
	}
	pipe, err := h.repo.PipelineGetByID(req.PipelineID)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
		return
	}
	if !h.canAccessPipeline(h.ac(c), pipe) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	secret := req.Secret
	if secret == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		secret = hex.EncodeToString(b)
	}
	enc, err := crypto.Encrypt(h.key, secret)
	if err != nil {
		resp.Fail(c, errcode.CodeInternal, "encrypt secret failed")
		return
	}
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	code := hex.EncodeToString(b)
	w := &model.Webhook{
		PipelineID: req.PipelineID, Provider: req.Provider,
		SecretEnc: enc, BranchFilter: req.BranchFilter, Events: "push", Status: 1, HookCode: code,
	}
	if err := h.repo.WebhookCreate(w); err != nil {
		resp.Fail(c, errcode.CodeInternal, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "webhook.create", "webhook", fmt.Sprintf("%s/%s", req.Provider, code),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	// secret 仅创建时返回一次
	resp.OK(c, gin.H{
		"id": w.ID, "hook_code": code, "provider": req.Provider,
		"url":    fmt.Sprintf("/api/v1/webhooks/%s/%s", req.Provider, code),
		"secret": secret,
	})
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "invalid id")
		return
	}
	// 归属校验：通过 webhook → pipeline 链确认属主
	w, err := h.repo.WebhookGetByID(id)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "webhook not found")
		return
	}
	pipe, err := h.repo.PipelineGetByID(w.PipelineID)
	if err != nil {
		resp.Fail(c, errcode.CodeNotFound, "pipeline not found")
		return
	}
	if !h.canAccessPipeline(h.ac(c), pipe) {
		resp.Fail(c, errcode.CodeForbidden, "forbidden")
		return
	}
	if err := h.repo.WebhookSoftDelete(id); err != nil {
		resp.Fail(c, errcode.CodeInternal, err.Error())
		return
	}
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "webhook.delete", "webhook", strconv.FormatUint(id, 10),
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
	resp.OK(c, nil)
}

// Receive 公开 Webhook 接收端：HMAC 签名/令牌校验 + 分支过滤 + 触发 Pipeline。
// POST /api/v1/webhooks/:provider/:code
func (h *WebhookHandler) Receive(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Param("code")

	w, err := h.repo.WebhookGetByCode(code)
	if err != nil || w.Provider != provider || w.Status != 1 {
		c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "webhook not found"})
		return
	}
	secret, err := crypto.Decrypt(h.key, w.SecretEnc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 50000, "message": "decrypt failed"})
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": "read body failed"})
		return
	}

	// 签名校验（按 provider 协议）
	switch provider {
	case "github":
		got := c.GetHeader("X-Hub-Signature-256")
		if !strings.HasPrefix(got, "sha256=") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "missing signature"})
			return
		}
		mac := hmac.New(sha256.New, []byte(secret))
		_, _ = mac.Write(body)
		want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(got), []byte(want)) {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "invalid signature"})
			return
		}
	case "gitlab":
		if c.GetHeader("X-Gitlab-Token") != secret {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "invalid token"})
			return
		}
	case "gitee":
		if c.GetHeader("X-Gitee-Token") != secret {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "invalid token"})
			return
		}
	default:
		c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "unknown provider"})
		return
	}

	// 解析 ref / commit
	var payload struct {
		Ref   string `json:"ref"`
		After string `json:"after"`
	}
	_ = json.Unmarshal(body, &payload)
	if payload.Ref == "" {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ignored (no ref)", "data": nil})
		return
	}
	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")
	if !matchBranch(w.BranchFilter, branch) {
		c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ignored (branch filter)", "data": nil})
		return
	}

	// 触发 Pipeline（以 pipeline 属主身份）
	pipe, err := h.repo.PipelineGetByID(w.PipelineID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 40400, "message": "pipeline not found"})
		return
	}
	ac := service.AccessCtx{UserID: pipe.OwnerID, OrgID: func() uint64 {
		if pipe.OrgID == nil {
			return 0
		}
		return *pipe.OrgID
	}()}
	run, err := h.engine.CreateRun(c.Request.Context(), ac, w.PipelineID, branch, payload.After, "webhook", c.ClientIP(), middleware.GetRequestID(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"run_id": run.ID}, "request_id": middleware.GetRequestID(c)})
}

// matchBranch 分支过滤（空=全部；支持 * 通配，如 release/*）。
func matchBranch(filter, branch string) bool {
	if filter == "" {
		return true
	}
	if strings.Contains(filter, "*") {
		ok, err := path.Match(filter, branch)
		return err == nil && ok
	}
	return filter == branch
}
