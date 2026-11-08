package controller

import (
	"encoding/json"
	"net/http"
	"strings"

	"boom-cat-be/internal/service"
)

// AuthController 认证相关 HTTP 接口
type AuthController struct {
	authSvc service.AuthService
}

// NewAuthController 创建认证控制器
func NewAuthController(authSvc service.AuthService) *AuthController {
	return &AuthController{authSvc: authSvc}
}

// Register POST /api/v1/auth/register
// Body: {"nickname": "...", "password": "..."}
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Nickname string `json:"nickname"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.authSvc.Register(r.Context(), req.Nickname, req.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       result.UserID,
		"nickname":      result.Nickname,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt,
	})
}

// Login POST /api/v1/auth/login
// Body: {"user_id": "...", "password": "..."}
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.authSvc.Login(r.Context(), req.UserID, req.Password)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       result.UserID,
		"nickname":      result.Nickname,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt,
	})
}

// Refresh POST /api/v1/auth/refresh
// Body: {"refresh_token": "..."}
func (c *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := c.authSvc.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		writeAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       result.UserID,
		"nickname":      result.Nickname,
		"access_token":  result.AccessToken,
		"refresh_token": result.RefreshToken,
		"expires_at":    result.ExpiresAt,
	})
}

// Logout POST /api/v1/auth/logout
// Header: Authorization: Bearer <access_token>
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing token")
		return
	}
	if err := c.authSvc.Logout(r.Context(), token); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// ValidateToken 公开方法，供中间件使用
func (c *AuthController) ValidateToken(token string) (*service.Claims, error) {
	return c.authSvc.ValidateToken(token)
}

// writeAuthError 将认证错误映射到 HTTP 状态码
func writeAuthError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"),
		strings.Contains(msg, "unauthorized"),
		strings.Contains(msg, "expired"),
		strings.Contains(msg, "invalid"):
		writeError(w, http.StatusUnauthorized, msg)
	case strings.Contains(msg, "already taken"):
		writeError(w, http.StatusConflict, msg)
	case strings.Contains(msg, "must be"):
		writeError(w, http.StatusBadRequest, msg)
	default:
		writeError(w, http.StatusInternalServerError, msg)
	}
}

// writeJSON 写入统一 JSON 响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}

// writeError 写入错误 JSON 响应
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    status,
		"message": message,
		"data":    nil,
	})
}

// extractBearerToken 从 Authorization 头提取 Bearer 令牌
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(auth, prefix) {
		return ""
	}
	return strings.TrimPrefix(auth, prefix)
}
