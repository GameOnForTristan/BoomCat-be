package controller

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"boom-cat-be/internal/model"
	"boom-cat-be/internal/network"
	"boom-cat-be/internal/service"
)

// WSController WebSocket 连接控制器
type WSController struct {
	authSvc     service.AuthService
	roomSvc     service.RoomService
	broadcaster *network.Broadcaster
	upgrader    websocket.Upgrader

	mu        sync.Mutex
	reconnect map[string]*network.GameSocket // playerID -> socket (for reconnect tracking)
}

// NewWSController 创建 WebSocket 控制器
func NewWSController(authSvc service.AuthService, roomSvc service.RoomService, broadcaster *network.Broadcaster) *WSController {
	return &WSController{
		authSvc:     authSvc,
		roomSvc:     roomSvc,
		broadcaster: broadcaster,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // 生产环境应校验 Origin
			},
		},
		reconnect: make(map[string]*network.GameSocket),
	}
}

// HandleWS GET /ws?room_id={room_id}&token={token}
// 升级 HTTP 到 WebSocket，校验 JWT，注册连接并启动读写循环
func (c *WSController) HandleWS(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room_id")
	token := r.URL.Query().Get("token")
	if roomID == "" || token == "" {
		http.Error(w, "room_id and token are required", http.StatusBadRequest)
		return
	}

	claims, err := c.authSvc.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusUnauthorized)
		return
	}

	// 校验玩家在房间内
	room, err := c.roomSvc.GetRoom(r.Context(), roomID)
	if err != nil {
		http.Error(w, "room not found", http.StatusNotFound)
		return
	}
	inRoom := false
	for _, p := range room.Players {
		if p.ID == claims.UserID {
			inRoom = true
			break
		}
	}
	if !inRoom {
		http.Error(w, "player not in room", http.StatusForbidden)
		return
	}

	// 升级到 WebSocket
	conn, err := c.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade 已写入错误响应
	}

	socket := network.NewGameSocket(conn, claims.UserID, roomID)
	c.broadcaster.Register(roomID, claims.UserID, socket)
	c.handleReconnect(claims.UserID, socket)

	// 通知房间玩家上线
	_ = c.roomSvc.DispatchMessage(roomID, model.Message{
		PlayerID: claims.UserID,
		Type:     "player_connect",
	})

	// 启动读写循环（EventHandler 由 GameController 提供）
	eventHandler := network.NewEventHandler(roomDispatcherAdapter{roomSvc: c.roomSvc})
	go socket.WritePump()
	go socket.ReadPump(eventHandler, c)
}

// handleReconnect 处理重连：保留新连接，关闭旧连接
func (c *WSController) handleReconnect(playerID string, newSocket *network.GameSocket) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.reconnect[playerID]; ok && old != newSocket {
		go old.Close()
	}
	c.reconnect[playerID] = newSocket
}

// OnDisconnect 实现 network DisconnectListener 接口
// 由 GameSocket.ReadPump 在退出时调用
func (c *WSController) OnDisconnect(socket *network.GameSocket) {
	c.broadcaster.Unregister(socket.RoomID(), socket.PlayerID())
	c.mu.Lock()
	if cur, ok := c.reconnect[socket.PlayerID()]; ok && cur == socket {
		delete(c.reconnect, socket.PlayerID())
	}
	c.mu.Unlock()

	// 通知房间玩家断开
	_ = c.roomSvc.DispatchMessage(socket.RoomID(), model.Message{
		PlayerID: socket.PlayerID(),
		Type:     "player_disconnect",
	})
}

// roomDispatcherAdapter 将 RoomService 适配为 network.MessageDispatcher
type roomDispatcherAdapter struct {
	roomSvc service.RoomService
}

func (a roomDispatcherAdapter) DispatchMessage(roomID string, msg model.Message) error {
	return a.roomSvc.DispatchMessage(roomID, msg)
}
