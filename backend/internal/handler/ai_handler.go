// Package handler - AI 助手聊天：SSE 流式输出，前端 fetch reader 逐字渲染。
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/dxcloud/cloud-api/internal/iam"
	"github.com/dxcloud/cloud-api/internal/middleware"
	"github.com/dxcloud/cloud-api/internal/service"
	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/dxcloud/cloud-api/pkg/resp"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	svc    *service.AIService
	iamSvc *iam.Service
}

func NewAIHandler(svc *service.AIService, iamSvc *iam.Service) *AIHandler {
	return &AIHandler{svc: svc, iamSvc: iamSvc}
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req struct {
		Messages    []service.ChatMessage `json:"messages"`
		PageContext string                `json:"page_context"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.Fail(c, errcode.CodeBadRequest, "请求格式错误")
		return
	}
	if len(req.Messages) == 0 {
		resp.Fail(c, errcode.CodeBadRequest, "消息不能为空")
		return
	}

	// SSE 响应头：禁用代理缓冲，保证逐字推送
	hdr := c.Writer.Header()
	hdr.Set("Content-Type", "text/event-stream; charset=utf-8")
	hdr.Set("Cache-Control", "no-cache")
	hdr.Set("Connection", "keep-alive")
	hdr.Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, _ := c.Writer.(http.Flusher)
	writeSSE := func(payload any) bool {
		b, err := json.Marshal(payload)
		if err != nil {
			return true
		}
		if _, err := c.Writer.Write([]byte("data: " + string(b) + "\n\n")); err != nil {
			return false
		}
		if flusher != nil {
			flusher.Flush()
		}
		return true
	}

	streamed := false
	err := h.svc.StreamChat(c.Request.Context(), req.Messages, req.PageContext, func(delta string) bool {
		streamed = true
		return writeSSE(map[string]string{"delta": delta})
	})
	if err != nil {
		if !streamed {
			// 尚未输出任何内容：可安全回退为 JSON 错误（前端在流开始前解析 HTTP 错误）
			resp.Fail(c, errcode.CodeInternal, err.Error())
			return
		}
		writeSSE(map[string]string{"error": err.Error()})
	}
	writeSSE(map[string]string{"done": "1"})
	uid := middleware.GetUserID(c)
	h.iamSvc.Audit(c.Request.Context(), &uid, "ai.chat", "ai", "assistant",
		c.ClientIP(), middleware.GetRequestID(c), 1, nil)
}
