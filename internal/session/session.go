package session

import (
	"boom-cat-be/internal/model"
)

// PlayerStateView 玩家视角的游戏状态
// 只包含该玩家应当看到的信息
type PlayerStateView struct {
	GamePhase    model.GamePhase            `json:"game_phase"`
	RoundCount   int                        `json:"round_count"`
	YourHand     []model.Card               `json:"your_hand"`
	YourDefuses  int                        `json:"your_defuses"`
	Players      []model.PlayerInfo         `json:"players"`
	Turn         model.TurnState            `json:"turn"`
	DiscardTop   *model.Card                `json:"discard_top,omitempty"`
	ActionStack  []model.StackEntry         `json:"action_stack"`
	AliveCount   int                        `json:"alive_count"`
}

// BuildPlayerView 为指定玩家构建游戏状态视图
// 过滤掉其他玩家的手牌信息，只暴露手牌数量
func BuildPlayerView(gs *model.GameState, playerID string) *PlayerStateView {
	view := &PlayerStateView{
		GamePhase:   gs.Phase,
		RoundCount:  gs.RoundCount,
		YourHand:    make([]model.Card, 0),
		YourDefuses: 0,
		Players:     make([]model.PlayerInfo, 0),
		Turn:        gs.Turn,
		ActionStack: gs.ActionStack.Stack,
		AliveCount:  0,
	}

	// 构建玩家信息，过滤掉自己手牌的具体内容
	for _, p := range gs.Players {
		if p.ID == playerID {
			// 玩家自己的信息：包含完整手牌
			// 从房间 Player 中获取，这里需要从 GameState 反查
			view.YourDefuses = p.DefuseCount
		}
		view.Players = append(view.Players, p)
		if p.Alive {
			view.AliveCount++
		}
	}

	// 弃牌堆顶牌
	if topCard, ok := gs.DiscardPile.TopCard(); ok {
		view.DiscardTop = &topCard
	}

	return view
}

// BuildPublicView 构建完全公开的游戏状态（仅用于已结束的游戏）
func BuildPublicView(gs *model.GameState) *PlayerStateView {
	view := &PlayerStateView{
		GamePhase:   gs.Phase,
		RoundCount:  gs.RoundCount,
		YourHand:    make([]model.Card, 0),
		Players:     make([]model.PlayerInfo, 0),
		Turn:        gs.Turn,
		ActionStack: gs.ActionStack.Stack,
		AliveCount:  0,
	}

	for _, p := range gs.Players {
		view.Players = append(view.Players, p)
		if p.Alive {
			view.AliveCount++
		}
	}

	if topCard, ok := gs.DiscardPile.TopCard(); ok {
		view.DiscardTop = &topCard
	}

	return view
}

// BuildHandForPlayer 构建玩家自己的手牌视图
// 从 Player 实体的 Hand 中获取完整手牌
func BuildHandForPlayer(player *model.Player) []model.Card {
	cards := make([]model.Card, len(player.Hand.Cards))
	copy(cards, player.Hand.Cards)
	return cards
}

// ToPlayerInfoList 将 Player 列表转换为 PlayerInfo 列表（过滤敏感信息）
func ToPlayerInfoList(players []model.Player) []model.PlayerInfo {
	info := make([]model.PlayerInfo, 0, len(players))
	for _, p := range players {
		info = append(info, p.ToPlayerInfo())
	}
	return info
}