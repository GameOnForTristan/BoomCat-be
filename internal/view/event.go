package view

import "time"

// ServerEvent 服务端推送事件
type ServerEvent struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	ReqID     string      `json:"req_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// NewEvent 创建服务端事件，自动填充时间戳
func NewEvent(eventType string, payload interface{}) ServerEvent {
	return ServerEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
}

// =====================================================================
// 常用事件构造函数
// =====================================================================

// StateUpdateEvent 状态更新事件
func StateUpdateEvent(state *GameStateDTO) ServerEvent {
	return NewEvent("state_update", state)
}

// NopePromptEvent 否决提示事件
type NopePromptPayload struct {
	ActorID   string `json:"actor_id"`
	CardType  string `json:"card_type"`
	WindowSec int    `json:"window_sec"`
}

func NopePromptEvent(actorID, cardType string, windowSec int) ServerEvent {
	return NewEvent("nope_prompt", NopePromptPayload{
		ActorID:   actorID,
		CardType:  cardType,
		WindowSec: windowSec,
	})
}

// NopeResolvedEvent 否决链结算事件
type NopeResolvedPayload struct {
	Blocked     bool                   `json:"blocked"`
	NopeCount   int                    `json:"nope_count"`
	BaseAction  map[string]interface{} `json:"base_action"`
}

func NopeResolvedEvent(blocked bool, nopeCount int, baseAction map[string]interface{}) ServerEvent {
	return NewEvent("nope_resolved", NopeResolvedPayload{
		Blocked:    blocked,
		NopeCount:  nopeCount,
		BaseAction: baseAction,
	})
}

// SeeTheFutureEvent 预言事件
func SeeTheFutureEvent(cards interface{}) ServerEvent {
	return NewEvent("see_the_future", cards)
}

// DefuseRequiredEvent 拆解要求事件
type DefuseRequiredPayload struct {
	ExplodingKittenID string `json:"exploding_kitten_id"`
}

func DefuseRequiredEvent(kittenID string) ServerEvent {
	return NewEvent("defuse_required", DefuseRequiredPayload{
		ExplodingKittenID: kittenID,
	})
}

// ExplodingDefusedEvent 爆炸猫拆解事件
type ExplodingDefusedPayload struct {
	PlayerID string `json:"player_id"`
}

func ExplodingDefusedEvent(playerID string) ServerEvent {
	return NewEvent("exploding_defused", ExplodingDefusedPayload{PlayerID: playerID})
}

// PlayerEliminatedEvent 玩家出局事件
type PlayerEliminatedPayload struct {
	PlayerID string `json:"player_id"`
}

func PlayerEliminatedEvent(playerID string) ServerEvent {
	return NewEvent("player_eliminated", PlayerEliminatedPayload{PlayerID: playerID})
}

// PlayerConnectEvent 玩家上线事件
func PlayerConnectEvent(playerID string) ServerEvent {
	return NewEvent("player_connect", PlayerEliminatedPayload{PlayerID: playerID})
}

// PlayerDisconnectEvent 玩家断线事件
func PlayerDisconnectEvent(playerID string) ServerEvent {
	return NewEvent("player_disconnect", PlayerEliminatedPayload{PlayerID: playerID})
}

// PlayerTimeoutEvent 玩家超时事件
func PlayerTimeoutEvent(playerID string) ServerEvent {
	return NewEvent("player_timeout", PlayerEliminatedPayload{PlayerID: playerID})
}

// FavorCompletedEvent 索要完成事件
type FavorCompletedPayload struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func FavorCompletedEvent(fromID, toID string) ServerEvent {
	return NewEvent("favor_completed", FavorCompletedPayload{From: fromID, To: toID})
}

// GameOverEvent 游戏结束事件
type GameOverPayload struct {
	WinnerID    string   `json:"winner_id,omitempty"`
	AliveCount  int      `json:"alive_count"`
	Eliminated  []string `json:"eliminated,omitempty"`
}

func GameOverEvent(winnerID string, aliveCount int, eliminated []string) ServerEvent {
	return NewEvent("game_over", GameOverPayload{
		WinnerID:   winnerID,
		AliveCount: aliveCount,
		Eliminated: eliminated,
	})
}

// ErrorEvent 错误事件
type ErrorPayload struct {
	ReqID   string `json:"req_id,omitempty"`
	Message string `json:"message"`
}

func ErrorEvent(reqID, message string) ServerEvent {
	return NewEvent("error", ErrorPayload{ReqID: reqID, Message: message})
}
