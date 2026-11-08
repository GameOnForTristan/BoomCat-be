// Package types 定义跨模块共享的类型与常量
package types

import "time"

// =====================================================================
// 通用响应码
// =====================================================================

// ResponseCode 统一响应码
type ResponseCode int

const (
	CodeOK              ResponseCode = 0
	CodeBadRequest      ResponseCode = 400
	CodeUnauthorized    ResponseCode = 401
	CodeForbidden       ResponseCode = 403
	CodeNotFound        ResponseCode = 404
	CodeConflict        ResponseCode = 409
	CodeInternalError   ResponseCode = 500
	CodeServiceUnavail  ResponseCode = 503
)

// String 返回响应码的字符串描述
func (c ResponseCode) String() string {
	switch c {
	case CodeOK:
		return "success"
	case CodeBadRequest:
		return "bad request"
	case CodeUnauthorized:
		return "unauthorized"
	case CodeForbidden:
		return "forbidden"
	case CodeNotFound:
		return "not found"
	case CodeConflict:
		return "conflict"
	case CodeInternalError:
		return "internal server error"
	case CodeServiceUnavail:
		return "service unavailable"
	default:
		return "unknown"
	}
}

// =====================================================================
// 事件类型常量（服务端推送）
// =====================================================================

// EventType 服务端推送事件类型
type EventType string

const (
	EventStateUpdate      EventType = "state_update"
	EventNopePrompt       EventType = "nope_prompt"
	EventNopeResolved     EventType = "nope_resolved"
	EventSeeTheFuture     EventType = "see_the_future"
	EventDefuseRequired   EventType = "defuse_required"
	EventExplodingDefused EventType = "exploding_defused"
	EventPlayerEliminated EventType = "player_eliminated"
	EventPlayerConnect    EventType = "player_connect"
	EventPlayerDisconnect EventType = "player_disconnect"
	EventPlayerTimeout    EventType = "player_timeout"
	EventFavorCompleted   EventType = "favor_completed"
	EventGameOver         EventType = "game_over"
	EventError            EventType = "error"
)

// =====================================================================
// 客户端消息类型常量
// =====================================================================

// ClientMessageType 客户端消息类型
type ClientMessageType string

const (
	MsgPlayCard     ClientMessageType = "play_card"
	MsgDrawCard     ClientMessageType = "draw_card"
	MsgPlayNope     ClientMessageType = "play_nope"
	MsgPassNope     ClientMessageType = "pass_nope"
	MsgSelectTarget ClientMessageType = "select_target"
	MsgDefusePlace  ClientMessageType = "defuse_place"
)

// =====================================================================
// 房间与游戏状态常量
// =====================================================================

// RoomStatus 房间状态
type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

// ConnectionStatus 玩家连接状态
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusTimedOut     ConnectionStatus = "timed_out"
)

// =====================================================================
// 通用时间戳类型
// =====================================================================

// Timestamp 毫秒级 Unix 时间戳
type Timestamp int64

// Now 返回当前时间戳
func Now() Timestamp { return Timestamp(time.Now().UnixMilli()) }

// Time 将时间戳转为 time.Time
func (t Timestamp) Time() time.Time { return time.UnixMilli(int64(t)) }

// ISOTime 返回 ISO 8601 格式字符串
func (t Timestamp) ISOTime() string { return t.Time().Format(time.RFC3339) }
