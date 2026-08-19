// Package errcode 定义全局业务错误码。
// 规范：0 成功；40xxx 客户端错误；50xxx 服务端错误。
package errcode

const (
	CodeSuccess = 0

	CodeBadRequest     = 40000 // 参数错误
	CodeForbidden      = 40001 // 无权限（RBAC / 资源归属校验失败）
	CodeConflict       = 40009 // 资源冲突（用户名已存在等）
	CodeTooManyRequests = 42900 // 触发限流
	CodeUnauthorized   = 40100 // 未登录 / token 失效
	CodeNotFound       = 40400 // 资源不存在
	CodeNotImplemented = 40401 // 端点未实现（按阶段交付，后续 Phase 替换为真实 handler）

	CodeInternal = 50000 // 服务器内部错误
	CodeDBDown   = 50300 // 依赖未就绪（MySQL / Redis）
)
