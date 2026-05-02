package errors

import (
	"fmt"
	"strings"
)

// CLIError CLI 错误类型
type CLIError struct {
	Code       string
	Message    string
	Detail     string
	Suggestion string
}

func (e *CLIError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Detail)
	}
	return e.Message
}

// FullError 返回完整错误信息（包含 Code），用于调试
func (e *CLIError) FullError() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Detail)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Is 判断是否为特定错误
func (e *CLIError) Is(target error) bool {
	t, ok := target.(*CLIError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// 鉴权错误
var (
	ErrAuthRequired = &CLIError{
		Code:       "AUTH_REQUIRED",
		Message:    "未登录",
		Suggestion: "请先执行: cbi auth login",
	}
	ErrTokenExpired = &CLIError{
		Code:       "TOKEN_EXPIRED",
		Message:    "认证已过期",
		Suggestion: "请重新登录: cbi auth login",
	}
	ErrInvalidAPIKey = &CLIError{
		Code:       "INVALID_API_KEY",
		Message:    "无效的 API Key",
		Suggestion: "请重新登录: cbi auth login",
	}
)

// 权限错误
var (
	ErrPermissionDenied = &CLIError{
		Code:       "PERMISSION_DENIED",
		Message:    "无权限执行此操作",
		Suggestion: "请确认账户权限或联系管理员",
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
		Code:       "MISSING_PROJECT_ID",
		Message:    "缺少专案 ID",
		Suggestion: "请通过 --repository-id 指定",
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
		Code:       "UPSTREAM_TIMEOUT",
		Message:    "服务器响应超时",
		Suggestion: "请稍后重试",
	}
	ErrUpstreamError = &CLIError{
		Code:       "UPSTREAM_ERROR",
		Message:    "服务器内部错误",
		Suggestion: "请稍后重试，如持续出现请联系管理员",
	}
	ErrNetworkError = &CLIError{
		Code:       "NETWORK_ERROR",
		Message:    "网络连接失败",
		Suggestion: "请检查网络连接",
	}
	ErrConfigNotFound = &CLIError{
		Code:       "CONFIG_NOT_FOUND",
		Message:    "配置文件不存在",
		Suggestion: "请先初始化: cbi config init",
	}
	ErrConfigInvalid = &CLIError{
		Code:       "CONFIG_INVALID",
		Message:    "配置文件格式错误",
		Suggestion: "请重新初始化: cbi config init --new",
	}
	ErrConfigWriteFailed = &CLIError{
		Code:    "CONFIG_WRITE_FAILED",
		Message: "配置文件写入失败",
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
		Code:       cliErr.Code,
		Message:    cliErr.Message,
		Detail:     err.Error(),
		Suggestion: cliErr.Suggestion,
	}
}

// FormatError 格式化错误为用户友好的消息
// verbose 控制是否显示技术细节
func FormatError(err error, verbose bool) string {
	if err == nil {
		return ""
	}

	// CLIError 类型
	if cliErr, ok := err.(*CLIError); ok {
		return formatCLIError(cliErr, verbose)
	}

	// 尝试识别常见底层错误模式
	msg := err.Error()
	return formatRawError(msg, verbose)
}

// formatCLIError 格式化 CLIError
func formatCLIError(e *CLIError, verbose bool) string {
	var sb strings.Builder

	sb.WriteString(e.Message)

	if e.Detail != "" && verbose {
		sb.WriteString("\n  详情: ")
		sb.WriteString(e.Detail)
	}

	if e.Suggestion != "" {
		sb.WriteString("\n  ")
		sb.WriteString(e.Suggestion)
	}

	return sb.String()
}

// formatRawError 格式化原始错误
func formatRawError(msg string, verbose bool) string {
	lower := strings.ToLower(msg)

	// 网络错误
	if strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "no such host") ||
		strings.Contains(lower, "dns") ||
		strings.Contains(lower, "i/o timeout") ||
		strings.Contains(lower, "network is unreachable") {
		result := "无法连接到服务器"
		if verbose {
			result += "\n  详情: " + msg
		}
		result += "\n  请检查网络连接"
		return result
	}

	// 超时错误
	if strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline exceeded") {
		result := "请求超时，请稍后重试"
		if verbose {
			result += "\n  详情: " + msg
		}
		return result
	}

	// 配置文件错误
	if strings.Contains(lower, "config.json") && strings.Contains(lower, "no such file") {
		result := "配置文件不存在"
		if verbose {
			result += "\n  详情: " + msg
		}
		result += "\n  请先初始化: cbi config init"
		return result
	}

	if strings.Contains(lower, "config.json") && (strings.Contains(lower, "invalid") || strings.Contains(lower, "unexpected")) {
		result := "配置文件格式错误"
		if verbose {
			result += "\n  详情: " + msg
		}
		result += "\n  请重新初始化: cbi config init --new"
		return result
	}

	// context 取消
	if strings.Contains(lower, "context canceled") {
		return "操作已取消"
	}

	// 通用：去掉文件路径，保留简短描述
	if !verbose {
		msg = sanitizeErrorMsg(msg)
	}
	return fmt.Sprintf("操作失败: %s", msg)
}

// sanitizeErrorMsg 清理错误消息中的开发信息
func sanitizeErrorMsg(msg string) string {
	// 去掉完整文件路径
	msg = strings.ReplaceAll(msg, "/Users/", "~/")
	msg = strings.ReplaceAll(msg, "/home/", "~/")
	// 如果消息仍然很长，截断
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}
	return msg
}
