// Package resp 提供 API 统一响应结构（见架构文档 §8）。
package resp

import (
	"net/http"

	"github.com/dxcloud/cloud-api/pkg/errcode"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data"`
	RequestID string `json:"request_id"`
}

// RequestID 从 gin 上下文读取中间件写入的请求 ID。
func RequestID(c *gin.Context) string {
	if v, ok := c.Get("request_id"); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// OK 成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{Code: errcode.CodeSuccess, Message: "操作成功", Data: data, RequestID: RequestID(c)})
}

// Fail 错误响应：HTTP 状态码与业务码对齐，响应体保持统一结构。
func Fail(c *gin.Context, code int, message string) {
	status := http.StatusOK
	switch {
	case code == errcode.CodeBadRequest:
		status = http.StatusBadRequest
	case code == errcode.CodeConflict:
		status = http.StatusConflict
	case code == errcode.CodeUnauthorized:
		status = http.StatusUnauthorized
	case code == errcode.CodeForbidden:
		status = http.StatusForbidden
	case code == errcode.CodeTooManyRequests:
		status = http.StatusTooManyRequests
	case code == errcode.CodeNotFound || code == errcode.CodeNotImplemented:
		status = http.StatusNotFound
	case code >= 50000:
		status = http.StatusInternalServerError
	}
	c.JSON(status, Response{Code: code, Message: message, Data: nil, RequestID: RequestID(c)})
}
