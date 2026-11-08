// Package errors 定义错误码与错误包装工具
package errors

import (
	"errors"
	"fmt"
)

// =====================================================================
// 错误码定义
// =====================================================================

// Code 错误码
type Code int

const (
	CodeOK              Code = 0
	CodeBadRequest      Code = 400
	CodeUnauthorized    Code = 401
	CodeForbidden       Code = 403
	CodeNotFound        Code = 404
	CodeConflict        Code = 409
	CodeInternalError   Code = 500
	CodeServiceUnavail  Code = 503
)

// String 返回错误码描述
func (c Code) String() string {
	switch c {
	case CodeOK:
		return "ok"
	case CodeBadRequest:
		return "bad_request"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeForbidden:
		return "forbidden"
	case CodeNotFound:
		return "not_found"
	case CodeConflict:
		return "conflict"
	case CodeInternalError:
		return "internal_error"
	case CodeServiceUnavail:
		return "service_unavailable"
	default:
		return "unknown"
	}
}

// =====================================================================
// AppError 错误类型
// =====================================================================

// AppError 应用错误，包含错误码与可读消息
type AppError struct {
	Code    Code
	Message string
	Cause   error
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap 支持 errors.Is/As
func (e *AppError) Unwrap() error { return e.Cause }

// New 创建一个 AppError
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap 包装底层错误，附加错误码与消息
func Wrap(code Code, message string, cause error) *AppError {
	return &AppError{Code: code, Message: message, Cause: cause}
}

// =====================================================================
// 便捷构造函数
// =====================================================================

func BadRequest(format string, args ...interface{}) *AppError {
	return New(CodeBadRequest, fmt.Sprintf(format, args...))
}

func Unauthorized(format string, args ...interface{}) *AppError {
	return New(CodeUnauthorized, fmt.Sprintf(format, args...))
}

func Forbidden(format string, args ...interface{}) *AppError {
	return New(CodeForbidden, fmt.Sprintf(format, args...))
}

func NotFound(format string, args ...interface{}) *AppError {
	return New(CodeNotFound, fmt.Sprintf(format, args...))
}

func Conflict(format string, args ...interface{}) *AppError {
	return New(CodeConflict, fmt.Sprintf(format, args...))
}

func Internal(format string, args ...interface{}) *AppError {
	return New(CodeInternalError, fmt.Sprintf(format, args...))
}

// =====================================================================
// 辅助函数
// =====================================================================

// Is wraps errors.Is
func Is(err, target error) bool { return errors.Is(err, target) }

// As wraps errors.As
func As(err error, target interface{}) bool { return errors.As(err, target) }

// CodeOf 从 error 中提取错误码；未找到则返回 CodeInternalError
func CodeOf(err error) Code {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Code
	}
	return CodeInternalError
}

// MessageOf 从 error 中提取消息
func MessageOf(err error) string {
	if err == nil {
		return ""
	}
	var ae *AppError
	if errors.As(err, &ae) {
		return ae.Message
	}
	return err.Error()
}
