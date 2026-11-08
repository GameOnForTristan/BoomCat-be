package network

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"boom-cat-be/internal/model"
)

// ClientMessage 客户端->服务端统一消息格式
type ClientMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
	ReqID   string          `json:"req_id,omitempty"`
}

// ServerMessage 服务端->客户端统一消息格式
type ServerMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	ReqID     string      `json:"req_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewServerMessage 创建服务端消息并填充时间戳
func NewServerMessage(msgType string, payload interface{}) ServerMessage {
	return ServerMessage{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

// =====================================================================
// GameSocket - 单个 WebSocket 连接的封装
// =====================================================================

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = 30 * time.Second
	bufSize    = 256
)

// GameSocket 封装单个玩家的 WebSocket 连接
type GameSocket struct {
	conn     *websocket.Conn
	playerID string
	roomID   string

	send     chan []byte
	closeMu  sync.Mutex
	closed   bool
	closeCh  chan struct{}
}

// NewGameSocket 创建 GameSocket
func NewGameSocket(conn *websocket.Conn, playerID, roomID string) *GameSocket {
	return &GameSocket{
		conn:     conn,
		playerID: playerID,
		roomID:   roomID,
		send:     make(chan []byte, bufSize),
		closeCh:  make(chan struct{}),
	}
}

// PlayerID 返回玩家 ID
func (s *GameSocket) PlayerID() string { return s.playerID }

// RoomID 返回房间 ID
func (s *GameSocket) RoomID() string { return s.roomID }

// Send 将消息加入发送队列（非阻塞，队列满则丢弃）
func (s *GameSocket) Send(message ServerMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		return
	}
	select {
	case s.send <- data:
	default:
		// 队列满，丢弃消息避免阻塞广播方
	}
}

// SendRaw 直接发送原始 JSON 字节
func (s *GameSocket) SendRaw(data []byte) {
	select {
	case s.send <- data:
	default:
	}
}

// Close 关闭连接（幂等）
func (s *GameSocket) Close() {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	close(s.closeCh)
	close(s.send)
	_ = s.conn.Close()
}

// IsClosed 检查是否已关闭
func (s *GameSocket) IsClosed() bool {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()
	return s.closed
}

// Closed 返回关闭信号 channel
func (s *GameSocket) Closed() <-chan struct{} { return s.closeCh }

// DisconnectListener 连接断开监听器（由 controller 层实现）
type DisconnectListener interface {
	OnDisconnect(socket *GameSocket)
}

// noopDisconnectListener 默认空实现
type noopDisconnectListener struct{}

func (noopDisconnectListener) OnDisconnect(*GameSocket) {}

// ReadPump 读取循环：从 WebSocket 读消息并交给 handler 处理
// 必须在单独 goroutine 中调用，每个连接一个
// listener 用于在连接断开时通知上层（如 WSController 注销连接）
func (s *GameSocket) ReadPump(handler *EventHandler, listener DisconnectListener) {
	if listener == nil {
		listener = noopDisconnectListener{}
	}
	defer func() {
		handler.OnDisconnect(s)
		listener.OnDisconnect(s)
		s.Close()
	}()

	s.conn.SetReadDeadline(time.Now().Add(pongWait))
	s.conn.SetPongHandler(func(string) error {
		s.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := s.conn.ReadMessage()
		if err != nil {
			return
		}
		var msg ClientMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			s.Send(NewServerMessage("error", map[string]string{
				"message": "invalid message format",
			}))
			continue
		}
		handler.OnMessage(s, &msg)
	}
}

// WritePump 写入循环：从 send 队列读消息并写入 WebSocket
// 必须在单独 goroutine 中调用，每个连接一个
func (s *GameSocket) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		s.Close()
	}()

	for {
		select {
		case message, ok := <-s.send:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				s.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := s.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			s.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-s.closeCh:
			return
		}
	}
}

// =====================================================================
// Broadcaster - 向房间内的玩家广播消息
// 实现 service.Broadcaster 接口
// =====================================================================

// Broadcaster 管理房间内所有玩家的 WebSocket 连接，支持广播与定向发送
type Broadcaster struct {
	mu      sync.RWMutex
	sockets map[string]map[string]*GameSocket // roomID -> playerID -> socket
}

// NewBroadcaster 创建广播器
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		sockets: make(map[string]map[string]*GameSocket),
	}
}

// Register 注册一个连接到房间
func (b *Broadcaster) Register(roomID, playerID string, socket *GameSocket) {
	b.mu.Lock()
	defer b.mu.Unlock()
	room, ok := b.sockets[roomID]
	if !ok {
		room = make(map[string]*GameSocket)
		b.sockets[roomID] = room
	}
	// 若已存在旧连接，关闭之（同一玩家新连接替代旧连接）
	if old, ok := room[playerID]; ok {
		go old.Close()
	}
	room[playerID] = socket
}

// Unregister 从房间中移除连接
func (b *Broadcaster) Unregister(roomID, playerID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	room, ok := b.sockets[roomID]
	if !ok {
		return
	}
	delete(room, playerID)
	if len(room) == 0 {
		delete(b.sockets, roomID)
	}
}

// Broadcast 向房间内所有玩家广播消息
func (b *Broadcaster) Broadcast(roomID string, event string, payload interface{}) {
	msg := NewServerMessage(event, payload)
	b.mu.RLock()
	room, ok := b.sockets[roomID]
	if !ok {
		b.mu.RUnlock()
		return
	}
	sockets := make([]*GameSocket, 0, len(room))
	for _, s := range room {
		sockets = append(sockets, s)
	}
	b.mu.RUnlock()
	for _, s := range sockets {
		s.Send(msg)
	}
}

// SendToPlayer 向房间内指定玩家发送消息
func (b *Broadcaster) SendToPlayer(roomID, playerID string, event string, payload interface{}) {
	msg := NewServerMessage(event, payload)
	b.mu.RLock()
	room, ok := b.sockets[roomID]
	if !ok {
		b.mu.RUnlock()
		return
	}
	socket, ok := room[playerID]
	b.mu.RUnlock()
	if ok {
		socket.Send(msg)
	}
}

// SendToPlayers 向部分玩家发送消息
func (b *Broadcaster) SendToPlayers(roomID string, playerIDs []string, event string, payload interface{}) {
	msg := NewServerMessage(event, payload)
	b.mu.RLock()
	room, ok := b.sockets[roomID]
	if !ok {
		b.mu.RUnlock()
		return
	}
	sockets := make([]*GameSocket, 0, len(playerIDs))
	for _, pid := range playerIDs {
		if s, ok := room[pid]; ok {
			sockets = append(sockets, s)
		}
	}
	b.mu.RUnlock()
	for _, s := range sockets {
		s.Send(msg)
	}
}

// GetConnectedPlayers 返回房间内已连接的玩家 ID 列表
func (b *Broadcaster) GetConnectedPlayers(roomID string) []string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	room, ok := b.sockets[roomID]
	if !ok {
		return nil
	}
	out := make([]string, 0, len(room))
	for pid := range room {
		out = append(out, pid)
	}
	return out
}

// =====================================================================
// EventHandler - 消息路由与分发
// =====================================================================

// MessageHandler 处理客户端消息的函数类型
type MessageHandler func(socket *GameSocket, msg *ClientMessage)

// MessageDispatcher 将消息推送到房间 Inbox 的接口（由 service.RoomService 实现）
type MessageDispatcher interface {
	DispatchMessage(roomID string, msg model.Message) error
}

// EventHandler 消息路由器
type EventHandler struct {
	dispatcher MessageDispatcher
}

// NewEventHandler 创建消息路由器
func NewEventHandler(dispatcher MessageDispatcher) *EventHandler {
	return &EventHandler{dispatcher: dispatcher}
}

// OnMessage 处理收到的客户端消息：解析为 model.Message 并派发到房间 Inbox
func (h *EventHandler) OnMessage(socket *GameSocket, msg *ClientMessage) {
	if h.dispatcher == nil {
		return
	}
	modelMsg := model.Message{
		PlayerID: socket.playerID,
		Type:     msg.Type,
		Payload:  msg.Payload,
		ReqID:    msg.ReqID,
	}
	if err := h.dispatcher.DispatchMessage(socket.roomID, modelMsg); err != nil {
		socket.Send(NewServerMessage("error", map[string]string{
			"req_id": msg.ReqID,
			"message": err.Error(),
		}))
	}
}

// OnDisconnect 处理连接断开事件（默认空实现，由 DisconnectListener 处理）
func (h *EventHandler) OnDisconnect(socket *GameSocket) {
	// 实际断开逻辑由 WSController.OnDisconnect 处理（注销广播、通知房间）
}
