package errors

import (
	"fmt"
)

// CLIError CLI 错误类型
type CLIError struct {
	Code    string
	Message string
	Detail  string
}

func (e *CLIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// 鉴权错误
var (
	ErrAuthRequired = &CLIError{
		Code:    "AUTH_REQUIRED",
		Message: "未登录，请先执行 cbi auth login",
	}
	ErrTokenExpired = &CLIError{
		Code:    "TOKEN_EXPIRED",
		Message: "认证已过期，请重新登录",
	}
	ErrInvalidAPIKey = &CLIError{
		Code:    "INVALID_API_KEY",
		Message: "无效的 API Key",
	}
)

// 权限错误
var (
	ErrPermissionDenied = &CLIError{
		Code:    "PERMISSION_DENIED",
		Message: "无权限执行此操作",
	}
)

// 参数错误
var (
	ErrInvalidArgument = &CLIError{
		Code:    "INVALID_ARGUMENT",
		Message: "参数错误",
	}
	ErrInvalidScope = &CLIError{
		Code:    "INVALID_SCOPE",
		Message: "无效的范围参数",
	}
	ErrMissingProjectID = &CLIError{
		Code:    "MISSING_PROJECT_ID",
		Message: "缺少专案 ID",
	}
	ErrMissingCollectionID = &CLIError{
		Code:    "MISSING_COLLECTION_ID",
		Message: "缺少专案集 ID",
	}
	ErrMissingAssetID = &CLIError{
		Code:    "MISSING_ASSET_ID",
		Message: "缺少素材 ID",
	}
	ErrMissingTaskID = &CLIError{
		Code:    "MISSING_TASK_ID",
		Message: "缺少任务 ID",
	}
)

// 资源错误
var (
	ErrNotFound = &CLIError{
		Code:    "NOT_FOUND",
		Message: "资源不存在",
	}
	ErrProjectNotFound = &CLIError{
		Code:    "PROJECT_NOT_FOUND",
		Message: "专案不存在",
	}
	ErrCollectionNotFound = &CLIError{
		Code:    "COLLECTION_NOT_FOUND",
		Message: "专案集不存在",
	}
	ErrAssetNotFound = &CLIError{
		Code:    "ASSET_NOT_FOUND",
		Message: "素材不存在",
	}
	ErrTaskNotFound = &CLIError{
		Code:    "TASK_NOT_FOUND",
		Message: "任务不存在",
	}
	ErrWorkspaceNotFound = &CLIError{
		Code:    "WORKSPACE_NOT_FOUND",
		Message: "workspace 不存在",
	}
)

// 网络/上游错误
var (
	ErrUpstreamTimeout = &CLIError{
		Code:    "UPSTREAM_TIMEOUT",
		Message: "上游服务超时",
	}
	ErrUpstreamError = &CLIError{
		Code:    "UPSTREAM_ERROR",
		Message: "上游服务错误",
	}
	ErrNetworkError = &CLIError{
		Code:    "NETWORK_ERROR",
		Message: "网络连接错误",
	}
)

// NewCLIError 创建新的 CLI 错误
func NewCLIError(code, message string) *CLIError {
	return &CLIError{
		Code:    code,
		Message: message,
	}
}

// NewCLIErrorWithDetail 创建带详情的 CLI 错误
func NewCLIErrorWithDetail(code, message, detail string) *CLIError {
	return &CLIError{
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

// WrapError 包装错误为 CLI 错误
func WrapError(err error, cliErr *CLIError) *CLIError {
	return &CLIError{
		Code:    cliErr.Code,
		Message: cliErr.Message,
		Detail:  err.Error(),
	}
}