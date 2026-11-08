package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"boom-cat-be/internal/service"
)

// RoomController 房间相关 HTTP 接口
type RoomController struct {
	roomSvc service.RoomService
	authSvc service.AuthService
}

// NewRoomController 创建房间控制器
func NewRoomController(roomSvc service.RoomService, authSvc service.AuthService) *RoomController {
	return &RoomController{roomSvc: roomSvc, authSvc: authSvc}
}

// extractUser 从请求中提取认证用户
func (c *RoomController) extractUser(r *http.Request) (*service.Claims, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, errors.New("missing authorization token")
	}
	return c.authSvc.ValidateToken(token)
}

// Create POST /api/v1/rooms
// Body: {"name": "...", "max_players": 4}
func (c *RoomController) Create(w http.ResponseWriter, r *http.Request) {
	user, err := c.extractUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	var req struct {
		Name       string `json:"name"`
		MaxPlayers int    `json:"max_players"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	room, err := c.roomSvc.CreateRoom(r.Context(), user.UserID, user.Nickname, req.Name, req.MaxPlayers)
	if err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, room)
}

// Get GET /api/v1/rooms/{room_id}
func (c *RoomController) Get(w http.ResponseWriter, r *http.Request) {
	if _, err := c.extractUser(r); err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	roomID := pathParam(r, "room_id")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}
	room, err := c.roomSvc.GetRoom(r.Context(), roomID)
	if err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, room)
}

// Join POST /api/v1/rooms/join
// Body: {"invite_code": "..."}
func (c *RoomController) Join(w http.ResponseWriter, r *http.Request) {
	user, err := c.extractUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var req struct {
		InviteCode string `json:"invite_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.InviteCode == "" {
		writeError(w, http.StatusBadRequest, "invite_code is required")
		return
	}
	room, err := c.roomSvc.JoinRoom(r.Context(), req.InviteCode, user.UserID, user.Nickname)
	if err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, room)
}

// Leave POST /api/v1/rooms/{room_id}/leave
func (c *RoomController) Leave(w http.ResponseWriter, r *http.Request) {
	user, err := c.extractUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	roomID := pathParam(r, "room_id")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}
	if err := c.roomSvc.LeaveRoom(r.Context(), roomID, user.UserID); err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "left"})
}

// ToggleReady POST /api/v1/rooms/{room_id}/ready
// 切换准备状态，返回新的准备状态
func (c *RoomController) ToggleReady(w http.ResponseWriter, r *http.Request) {
	user, err := c.extractUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	roomID := pathParam(r, "room_id")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}
	ready, err := c.roomSvc.ToggleReady(r.Context(), roomID, user.UserID)
	if err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"is_ready": ready,
	})
}

// Start POST /api/v1/rooms/{room_id}/start
// 房主开始游戏
func (c *RoomController) Start(w http.ResponseWriter, r *http.Request) {
	user, err := c.extractUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}
	roomID := pathParam(r, "room_id")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}
	if err := c.roomSvc.StartGame(r.Context(), roomID, user.UserID); err != nil {
		writeRoomError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "game started"})
}

// writeRoomError 将房间服务错误映射到 HTTP 状态码
func writeRoomError(w http.ResponseWriter, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not found"):
		writeError(w, http.StatusNotFound, msg)
	case strings.Contains(msg, "full"):
		writeError(w, http.StatusConflict, msg)
	case strings.Contains(msg, "already started"):
		writeError(w, http.StatusConflict, msg)
	case strings.Contains(msg, "owner"):
		writeError(w, http.StatusForbidden, msg)
	case strings.Contains(msg, "not all ready"):
		writeError(w, http.StatusBadRequest, msg)
	case strings.Contains(msg, "invalid"):
		writeError(w, http.StatusBadRequest, msg)
	default:
		writeError(w, http.StatusInternalServerError, msg)
	}
}

// pathParam 从 URL path 中提取参数
// 形如 /api/v1/rooms/{room_id}/leave -> 提取 room_id
// 这里基于 path 末尾或倒数第二段解析，因为 net/http 不带路径参数
func pathParam(r *http.Request, name string) string {
	// 优先从 context 取（middleware 注入），其次从 query 取
	if v := r.URL.Query().Get(name); v != "" {
		return v
	}
	// 从 path 解析：/api/v1/rooms/<room_id>/...
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// 找到 "rooms" 后的位置
	for i, p := range parts {
		if p == "rooms" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
