package view

import (
	"boom-cat-be/internal/model"
)

// SerializeGameState 按玩家视角序列化游戏状态
// - 自己：完整手牌
// - 其他玩家：仅手牌数量
func SerializeGameState(gs *model.GameState, room *model.Room, playerID string) *GameStateDTO {
	if gs == nil {
		return nil
	}

	dto := &GameStateDTO{
		GamePhase:   int(gs.Phase),
		RoundCount:  gs.RoundCount,
		YourHand:    []CardDTO{},
		YourDefuses: 0,
		Players:     make([]PlayerInfoDTO, 0, len(room.Players)),
		Turn:        ToTurnDTO(gs.Turn),
		ActionStack: serializeStack(gs.ActionStack.Stack),
		AliveCount:  0,
		DeckCount:   gs.Deck.Count,
	}

	// 玩家自己的手牌与拆解数
	if player, ok := room.FindPlayer(playerID); ok {
		dto.YourHand = ToCardsDTO(player.Hand.Cards)
		dto.YourDefuses = player.DefuseCount
	}

	// 玩家列表（公开信息）
	for _, p := range room.Players {
		dto.Players = append(dto.Players, ToPlayerInfoDTO(p.ToPlayerInfo()))
		if p.Alive {
			dto.AliveCount++
		}
	}

	// 弃牌堆顶
	if top, ok := gs.DiscardPile.TopCard(); ok {
		topDTO := ToCardDTO(top)
		dto.DiscardTop = &topDTO
	}

	return dto
}

// SerializeRoom 序列化房间为 RoomDTO
// 需要传入 inviteCode（RoomService 持有，不暴露在 Room 实体中）
func SerializeRoom(room *model.Room, inviteCode string) *RoomDTO {
	if room == nil {
		return nil
	}
	players := make([]PlayerDTO, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, PlayerDTO{
			ID:          p.ID,
			Nickname:    p.Nickname,
			IsReady:     p.IsReady,
			IsOwner:     room.IsOwner(p.ID),
			Connected:   p.ConnectionStatus == model.ConnectionStatusConnected,
			Alive:       p.Alive,
			HandSize:    p.Hand.Count,
			DefuseCount: p.DefuseCount,
		})
	}
	return &RoomDTO{
		ID:          room.ID,
		Name:        room.Name,
		OwnerID:     room.OwnerID,
		InviteCode:  inviteCode,
		MaxPlayers:  room.Settings.MaxPlayers,
		PlayerCount: room.PlayerCount(),
		Status:      string(room.Status),
		Players:     players,
		CreatedAt:   room.CreatedAt,
	}
}

// serializeStack 序列化动作栈
func serializeStack(stack []model.StackEntry) []StackEntryDTO {
	out := make([]StackEntryDTO, 0, len(stack))
	for _, e := range stack {
		out = append(out, StackEntryDTO{
			Action: map[string]interface{}{
				"type":              int(e.Action.Type),
				"card_ids":          e.Action.CardIDs,
				"target_player_id":  e.Action.TargetPlayerID,
				"target_card_type":  string(e.Action.TargetCardType),
			},
			ActorID:   e.ActorID,
			Timestamp: e.Timestamp,
		})
	}
	return out
}
