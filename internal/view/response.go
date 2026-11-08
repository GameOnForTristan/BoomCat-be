package view

import (
	"encoding/json"
	"net/http"

	apperr "boom-cat-be/pkg/errors"
)

// Response 统一响应格式
// {"code": 0, "message": "success", "data": {...}}
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// WriteOK 写入成功响应（HTTP 200）
func WriteOK(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusOK, Response{
		Code:    int(apperr.CodeOK),
		Message: "success",
		Data:    data,
	})
}

// WriteCreated 写入创建成功响应（HTTP 201）
func WriteCreated(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusCreated, Response{
		Code:    int(apperr.CodeOK),
		Message: "created",
		Data:    data,
	})
}

// WriteError 写入错误响应，根据 AppError 自动映射 HTTP 状态码
func WriteError(w http.ResponseWriter, err error) {
	code := apperr.CodeOf(err)
	httpStatus := codeToHTTPStatus(code)
	writeResponse(w, httpStatus, Response{
		Code:    int(code),
		Message: apperr.MessageOf(err),
		Data:    nil,
	})
}

// WriteBadRequest 写入参数错误响应
func WriteBadRequest(w http.ResponseWriter, message string) {
	writeResponse(w, http.StatusBadRequest, Response{
		Code:    int(apperr.CodeBadRequest),
		Message: message,
		Data:    nil,
	})
}

// WriteUnauthorized 写入未授权响应
func WriteUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "unauthorized"
	}
	writeResponse(w, http.StatusUnauthorized, Response{
		Code:    int(apperr.CodeUnauthorized),
		Message: message,
		Data:    nil,
	})
}

// WriteForbidden 写入禁止访问响应
func WriteForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "forbidden"
	}
	writeResponse(w, http.StatusForbidden, Response{
		Code:    int(apperr.CodeForbidden),
		Message: message,
		Data:    nil,
	})
}

// WriteNotFound 写入未找到响应
func WriteNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "not found"
	}
	writeResponse(w, http.StatusNotFound, Response{
		Code:    int(apperr.CodeNotFound),
		Message: message,
		Data:    nil,
	})
}

// WriteInternal 写入服务器内部错误响应
func WriteInternal(w http.ResponseWriter, message string) {
	if message == "" {
		message = "internal server error"
	}
	writeResponse(w, http.StatusInternalServerError, Response{
		Code:    int(apperr.CodeInternalError),
		Message: message,
		Data:    nil,
	})
}

// writeResponse 写入 HTTP 响应
func writeResponse(w http.ResponseWriter, status int, body Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// codeToHTTPStatus 将应用错误码映射为 HTTP 状态码
func codeToHTTPStatus(code apperr.Code) int {
	switch code {
	case apperr.CodeOK:
		return http.StatusOK
	case apperr.CodeBadRequest:
		return http.StatusBadRequest
	case apperr.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperr.CodeForbidden:
		return http.StatusForbidden
	case apperr.CodeNotFound:
		return http.StatusNotFound
	case apperr.CodeConflict:
		return http.StatusConflict
	case apperr.CodeServiceUnavail:
		return http.StatusServiceUnavailable
	case apperr.CodeInternalError:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}
