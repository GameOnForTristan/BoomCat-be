package service

import (
	"sync"
	"time"

	"boom-cat-be/internal/model"
)

// TimerService 定时器服务接口
// 管理三种定时器：回合超时、Nope 窗口超时、重连超时
type TimerService interface {
	// StartTurnTimer 启动回合定时器（玩家出牌阶段）
	StartTurnTimer(roomID, playerID string)
	// StopTurnTimer 停止回合定时器
	StopTurnTimer(roomID string)
	// ResetTurnTimer 重置回合定时器（玩家做出动作后重置）
	ResetTurnTimer(roomID, playerID string)

	// StartNopeTimer 启动 Nope 窗口定时器
	StartNopeTimer(roomID string)
	// StopNopeTimer 停止 Nope 窗口定时器
	StopNopeTimer(roomID string)
	// ResetNopeTimer 重置 Nope 窗口定时器（新的 Nope 被打出时）
	ResetNopeTimer(roomID string)

	// StartReconnectTimer 启动重连定时器
	StartReconnectTimer(roomID, playerID string)
	// StopReconnectTimer 停止重连定时器
	StopReconnectTimer(roomID, playerID string)

	// StopAll 停止房间所有定时器（游戏结束时调用）
	StopAll(roomID string)
}

// roomTimers 房间级别的定时器集合
type roomTimers struct {
	mu        sync.Mutex
	turn      *time.Timer
	nope      *time.Timer
	reconnect map[string]*time.Timer
}

type timerService struct {
	mu         sync.Mutex
	rooms      map[string]*roomTimers
	roomReg    RoomRegistry

	turnTimeout      time.Duration
	nopeWindow       time.Duration
	reconnectTimeout time.Duration
}

// NewTimerService 创建定时器服务
func NewTimerService(
	roomReg RoomRegistry,
	turnTimeoutSec, nopeWindowSec, reconnectTimeoutSec int,
) TimerService {
	return &timerService{
		rooms:            make(map[string]*roomTimers),
		roomReg:          roomReg,
		turnTimeout:      time.Duration(turnTimeoutSec) * time.Second,
		nopeWindow:       time.Duration(nopeWindowSec) * time.Second,
		reconnectTimeout: time.Duration(reconnectTimeoutSec) * time.Second,
	}
}

// getOrCreate 获取或创建房间的定时器集合
func (s *timerService) getOrCreate(roomID string) *roomTimers {
	s.mu.Lock()
	defer s.mu.Unlock()
	rt, ok := s.rooms[roomID]
	if !ok {
		rt = &roomTimers{
			reconnect: make(map[string]*time.Timer),
		}
		s.rooms[roomID] = rt
	}
	return rt
}

// StartTurnTimer 启动回合定时器，超时后向房间 Inbox 发送 auto_draw 消息
func (s *timerService) StartTurnTimer(roomID, playerID string) {
	rt := s.getOrCreate(roomID)
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.turn != nil {
		rt.turn.Stop()
	}
	rt.turn = time.AfterFunc(s.turnTimeout, func() {
		s.dispatchTimeout(roomID, model.Message{
			PlayerID: playerID,
			Type:     "auto_draw",
		})
	})
}

// StopTurnTimer 停止回合定时器
func (s *timerService) StopTurnTimer(roomID string) {
	rt := s.getOrCreate(roomID)
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.turn != nil {
		rt.turn.Stop()
		rt.turn = nil
	}
}

// ResetTurnTimer 重置回合定时器
func (s *timerService) ResetTurnTimer(roomID, playerID string) {
	s.StartTurnTimer(roomID, playerID)
}

// StartNopeTimer 启动 Nope 窗口定时器，超时后向房间 Inbox 发送 nope_timeout 消息
func (s *timerService) StartNopeTimer(roomID string) {
	rt := s.getOrCreate(roomID)
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.nope != nil {
		rt.nope.Stop()
	}
	rt.nope = time.AfterFunc(s.nopeWindow, func() {
		s.dispatchTimeout(roomID, model.Message{
			Type: "nope_timeout",
		})
	})
}

// StopNopeTimer 停止 Nope 窗口定时器
func (s *timerService) StopNopeTimer(roomID string) {
	rt := s.getOrCreate(roomID)
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.nope != nil {
		rt.nope.Stop()
		rt.nope = nil
	}
}

// ResetNopeTimer 重置 Nope 窗口定时器
func (s *timerService) ResetNopeTimer(roomID string) {
	s.StartNopeTimer(roomID)
}

// StartReconnectTimer 启动重连定时器，超时后发送 reconnect_timeout 消息
func (s *timerService) StartReconnectTimer(roomID, playerID string) {
	rt := s.getOrCreate(roomID)
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if t, ok := rt.reconnect[playerID]; ok {
		t.Stop()
	}
	rt.reconnect[playerID] = time.AfterFunc(s.reconnectTimeout, func() {
		s.dispatchTimeout(roomID, model.Message{
			PlayerID: playerID,
			Type:     "reconnect_timeout",
		})
	})
}

// StopReconnectTimer 停止重连定时器
func (s *timerService) StopReconnectTimer(roomID, playerID string) {
	rt := s.getOrCreate(roomID)
	rt.mu.Lock()
	defer rt.mu.Unlock()
	if t, ok := rt.reconnect[playerID]; ok {
		t.Stop()
		delete(rt.reconnect, playerID)
	}
}

// StopAll 停止房间所有定时器
func (s *timerService) StopAll(roomID string) {
	s.mu.Lock()
	rt, ok := s.rooms[roomID]
	if !ok {
		s.mu.Unlock()
		return
	}
	delete(s.rooms, roomID)
	s.mu.Unlock()

	rt.mu.Lock()
	defer rt.mu.Unlock()
	if rt.turn != nil {
		rt.turn.Stop()
	}
	if rt.nope != nil {
		rt.nope.Stop()
	}
	for _, t := range rt.reconnect {
		t.Stop()
	}
	rt.reconnect = make(map[string]*time.Timer)
}

// dispatchTimeout 将超时消息推送到房间 Inbox
func (s *timerService) dispatchTimeout(roomID string, msg model.Message) {
	room, ok := s.roomReg.GetRoomEntity(roomID)
	if !ok {
		return
	}
	select {
	case room.Inbox <- msg:
	default:
		// Inbox 满，丢弃超时消息（理论上不应发生）
	}
}
