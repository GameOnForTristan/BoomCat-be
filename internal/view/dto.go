// Package view 定义对外暴露的 DTO、响应格式与序列化器
package view

import (
	"time"

	"boom-cat-be/internal/model"
)

// =====================================================================
// 房间相关 DTO
// =====================================================================

// RoomDTO 房间对外数据传输对象
type RoomDTO struct {
	ID          string       `json:"room_id"`
	Name        string       `json:"name"`
	OwnerID     string       `json:"owner_id"`
	InviteCode  string       `json:"invite_code"`
	MaxPlayers  int          `json:"max_players"`
	PlayerCount int          `json:"player_count"`
	Status      string       `json:"status"`
	Players     []PlayerDTO  `json:"players"`
	CreatedAt   int64        `json:"created_at"`
}

// PlayerDTO 玩家对外数据传输对象
type PlayerDTO struct {
	ID          string `json:"id"`
	Nickname    string `json:"nickname"`
	IsReady     bool   `json:"is_ready"`
	IsOwner     bool   `json:"is_owner"`
	Connected   bool   `json:"connected"`
	Alive       bool   `json:"alive"`
	HandSize    int    `json:"hand_size,omitempty"`
	DefuseCount int    `json:"defuse_count,omitempty"`
}

// =====================================================================
// 游戏状态 DTO
// =====================================================================

// CardDTO 卡牌 DTO
type CardDTO struct {
	ID          string `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// TurnDTO 回合状态 DTO
type TurnDTO struct {
	CurrentPlayerID string   `json:"current_player_id"`
	PlayerOrder     []string `json:"player_order"`
	Direction       int      `json:"direction"`
	PendingDraws    int      `json:"pending_draws"`
	Phase           int      `json:"phase"`
	ActionsTaken    int      `json:"actions_taken"`
}

// PlayerInfoDTO 玩家公开信息 DTO（不含手牌内容）
type PlayerInfoDTO struct {
	ID          string `json:"id"`
	Nickname    string `json:"nickname"`
	HandSize    int    `json:"hand_size"`
	Alive       bool   `json:"alive"`
	DefuseCount int    `json:"defuse_count"`
}

// StackEntryDTO 动作栈条目 DTO
type StackEntryDTO struct {
	Action    map[string]interface{} `json:"action"`
	ActorID   string                 `json:"actor_id"`
	Timestamp int64                  `json:"timestamp"`
}

// GameStateDTO 游戏状态 DTO（按玩家视角过滤后）
type GameStateDTO struct {
	GamePhase    int              `json:"game_phase"`
	RoundCount   int              `json:"round_count"`
	YourHand     []CardDTO        `json:"your_hand"`
	YourDefuses  int              `json:"your_defuses"`
	Players      []PlayerInfoDTO  `json:"players"`
	Turn         TurnDTO          `json:"turn"`
	DiscardTop   *CardDTO         `json:"discard_top,omitempty"`
	ActionStack  []StackEntryDTO  `json:"action_stack"`
	AliveCount   int              `json:"alive_count"`
	DeckCount    int              `json:"deck_count"`
}

// =====================================================================
// 认证相关 DTO
// =====================================================================

// AuthResultDTO 认证结果 DTO
type AuthResultDTO struct {
	UserID       string    `json:"user_id"`
	Nickname     string    `json:"nickname"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// =====================================================================
// 转换辅助函数
// =====================================================================

// ToCardDTO 将 model.Card 转换为 CardDTO
func ToCardDTO(c model.Card) CardDTO {
	return CardDTO{
		ID:          c.ID,
		Type:        string(c.Type),
		Name:        c.Name,
		Description: c.Description,
	}
}

// ToCardsDTO 批量转换
func ToCardsDTO(cards []model.Card) []CardDTO {
	out := make([]CardDTO, 0, len(cards))
	for _, c := range cards {
		out = append(out, ToCardDTO(c))
	}
	return out
}

// ToPlayerInfoDTO 将 model.PlayerInfo 转换为 PlayerInfoDTO
func ToPlayerInfoDTO(p model.PlayerInfo) PlayerInfoDTO {
	return PlayerInfoDTO{
		ID:          p.ID,
		Nickname:    p.Nickname,
		HandSize:    p.HandSize,
		Alive:       p.Alive,
		DefuseCount: p.DefuseCount,
	}
}

// ToTurnDTO 将 model.TurnState 转换为 TurnDTO
func ToTurnDTO(t model.TurnState) TurnDTO {
	return TurnDTO{
		CurrentPlayerID: t.CurrentPlayerID,
		PlayerOrder:     append([]string{}, t.PlayerOrder...),
		Direction:       t.Direction,
		PendingDraws:    t.PendingDraws,
		Phase:           int(t.Phase),
		ActionsTaken:    t.ActionsTaken,
	}
}
