package controller

import (
	"encoding/json"
	"net/http"

	"boom-cat-be/internal/model"
	"boom-cat-be/internal/network"
	"boom-cat-be/internal/service"
)

// GameController 游戏相关 HTTP 与 WebSocket 消息分发控制器
// WebSocket 消息通过 EventHandler 推送到 room.Inbox，由 game_service 的房间 goroutine 处理
// 本控制器提供 HTTP 端点供客户端查询游戏状态
type GameController struct {
	gameSvc    service.GameService
	roomSvc    service.RoomService
	authSvc    service.AuthService
	broadcaster *network.Broadcaster
}

// NewGameController 创建游戏控制器
func NewGameController(
	gameSvc service.GameService,
	roomSvc service.RoomService,
	authSvc service.AuthService,
	broadcaster *network.Broadcaster,
) *GameController {
	return &GameController{
		gameSvc:    gameSvc,
		roomSvc:    roomSvc,
		authSvc:    authSvc,
		broadcaster: broadcaster,
	}
}

// NewEventHandler 创建一个绑定到房间消息分发的 EventHandler
// 供 WSController 在每个连接上使用
func (c *GameController) NewEventHandler() *network.EventHandler {
	return network.NewEventHandler(roomDispatcherAdapter{roomSvc: c.roomSvc})
}

// GameStateResponse 游戏状态查询响应
type GameStateResponse struct {
	RoomID    string      `json:"room_id"`
	Status    string      `json:"status"`
	Phase     int         `json:"phase"`
	Players   interface{} `json:"players"`
	Turn      interface{} `json:"turn"`
	DeckCount int         `json:"deck_count"`
}

// GetState GET /api/v1/rooms/{room_id}/state
// 返回当前游戏状态（按玩家视角过滤）
func (c *GameController) GetState(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return
	}
	claims, err := c.authSvc.ValidateToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roomID := pathParam(r, "room_id")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}

	// 通过 RoomService 内部接口获取房间实体
	rs, ok := c.roomSvc.(interface {
		GetRoomEntity(roomID string) (*model.Room, bool)
	})
	if !ok {
		writeError(w, http.StatusInternalServerError, "room service incompatible")
		return
	}
	room, ok := rs.GetRoomEntity(roomID)
	if !ok {
		writeError(w, http.StatusNotFound, "room not found")
		return
	}

	// 玩家必须在房间内
	inRoom := false
	for _, p := range room.Players {
		if p.ID == claims.UserID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		writeError(w, http.StatusForbidden, "player not in room")
		return
	}

	if room.GameState == nil {
		writeJSON(w, http.StatusOK, GameStateResponse{
			RoomID: room.ID,
			Status: string(room.Status),
		})
		return
	}

	// 构建玩家视角
	player, _ := room.FindPlayer(claims.UserID)
	hand := []model.Card{}
	if player != nil {
		hand = player.Hand.Cards
	}
	players := make([]model.PlayerInfo, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, p.ToPlayerInfo())
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"room_id":     room.ID,
		"status":      string(room.Status),
		"game_phase":  int(room.GameState.Phase),
		"round_count": room.GameState.RoundCount,
		"your_hand":   hand,
		"players":     players,
		"turn":        room.GameState.Turn,
		"deck_count":  room.GameState.Deck.Count,
		"alive_count": countAlive(room),
	})
}

// SendEvent POST /api/v1/rooms/{room_id}/event
// 通用事件发送接口（调试用，将事件转发到房间 Inbox）
func (c *GameController) SendEvent(w http.ResponseWriter, r *http.Request) {
	token := extractBearerToken(r)
	if token == "" {
		writeError(w, http.StatusUnauthorized, "missing authorization token")
		return
	}
	claims, err := c.authSvc.ValidateToken(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	roomID := pathParam(r, "room_id")
	if roomID == "" {
		writeError(w, http.StatusBadRequest, "room_id is required")
		return
	}

	var req struct {
		Type    string          `json:"type"`
		Payload json.RawMessage `json:"payload,omitempty"`
		ReqID   string          `json:"req_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}

	msg := model.Message{
		PlayerID: claims.UserID,
		Type:     req.Type,
		Payload:  req.Payload,
		ReqID:    req.ReqID,
	}
	if err := c.roomSvc.DispatchMessage(roomID, msg); err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "dispatched"})
}

// countAlive 统计存活玩家数
func countAlive(room *model.Room) int {
	n := 0
	for _, p := range room.Players {
		if p.Alive {
			n++
		}
	}
	return n
}
