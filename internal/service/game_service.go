package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"boom-cat-be/internal/game/action"
	"boom-cat-be/internal/game/deck"
	"boom-cat-be/internal/game/gameover"
	"boom-cat-be/internal/game/validator"
	"boom-cat-be/internal/model"
)

// Broadcaster 事件广播接口（由 network 层实现）
type Broadcaster interface {
	Broadcast(roomID string, event string, payload interface{})
	SendToPlayer(roomID, playerID string, event string, payload interface{})
}

// RoomRegistry 房间查询接口（由 RoomService 实现）
type RoomRegistry interface {
	GetRoomEntity(roomID string) (*model.Room, bool)
}

// PlayCardPayload play_card 消息负载
type PlayCardPayload struct {
	CardIDs        []string `json:"card_ids"`
	TargetPlayerID string   `json:"target_player_id,omitempty"`
	TargetCardType string   `json:"target_card_type,omitempty"`
	DeckPosition   int      `json:"deck_position,omitempty"` // 拆解牌插回位置
}

// DrawCardPayload draw_card 消息负载
type DrawCardPayload struct{}

// NopePayload play_nope 消息负载
type NopePayload struct{}

// DefusePlacePayload 拆解牌插回位置
type DefusePlacePayload struct {
	DeckPosition int `json:"deck_position"`
}

// GameService 游戏服务接口
type GameService interface {
	InitializeGame(ctx context.Context, room *model.Room) error
	StartRoomLoop(roomID string)
	StopRoomLoop(roomID string)
}

type gameService struct {
	roomRegistry RoomRegistry
	broadcaster  Broadcaster
	timerSvc     TimerService
	turnSvc      TurnService

	turnTimeout      time.Duration
	nopeWindow       time.Duration
	reconnectTimeout time.Duration

	stopped map[string]chan struct{}
}

// NewGameService 创建游戏服务
func NewGameService(
	roomRegistry RoomRegistry,
	broadcaster Broadcaster,
	timerSvc TimerService,
	turnSvc TurnService,
	turnTimeoutSec, nopeWindowSec, reconnectTimeoutSec int,
) GameService {
	return &gameService{
		roomRegistry:     roomRegistry,
		broadcaster:      broadcaster,
		timerSvc:         timerSvc,
		turnSvc:          turnSvc,
		turnTimeout:      time.Duration(turnTimeoutSec) * time.Second,
		nopeWindow:       time.Duration(nopeWindowSec) * time.Second,
		reconnectTimeout: time.Duration(reconnectTimeoutSec) * time.Second,
		stopped:          make(map[string]chan struct{}),
	}
}

// InitializeGame 初始化游戏状态：构建牌库、发牌、设置回合
func (s *gameService) InitializeGame(ctx context.Context, room *model.Room) error {
	if room == nil {
		return errors.New("room is nil")
	}
	if len(room.Players) < 2 {
		return model.ErrNotAllReady
	}

	playerIDs := make([]string, 0, len(room.Players))
	for _, p := range room.Players {
		playerIDs = append(playerIDs, p.ID)
	}

	gameDeck, hands := deck.BuildGameDeck(len(playerIDs), playerIDs)

	gs := model.NewGameState()
	gs.Deck = *gameDeck
	gs.Phase = model.GamePhasePlaying
	gs.Turn = model.TurnState{
		CurrentPlayerID: playerIDs[0],
		PlayerOrder:     append([]string{}, playerIDs...),
		Direction:       1,
		PendingDraws:    1,
		Phase:           model.PhaseAction,
	}

	for i := range room.Players {
		p := &room.Players[i]
		p.Hand = *model.NewHand(p.ID)
		for _, c := range hands[p.ID] {
			p.Hand.AddCard(c)
			if c.Type == model.CardTypeDefuse {
				p.DefuseCount++
			}
		}
		p.Alive = true
		p.IsReady = false
		gs.Players = append(gs.Players, p.ToPlayerInfo())
	}

	room.GameState = gs
	room.Status = model.RoomStatusPlaying

	// 广播初始状态
	s.broadcastState(room)
	return nil
}

// StartRoomLoop 启动房间消息循环（每房间一个 goroutine）
func (s *gameService) StartRoomLoop(roomID string) {
	room, ok := s.roomRegistry.GetRoomEntity(roomID)
	if !ok {
		return
	}

	stopCh := make(chan struct{})
	// 复用 stopped map 作为停止信号注册表
	// 注意：此处不持有锁，假设 StartRoomLoop 由 RoomService.StartGame 单次调用
	prev, exists := s.stopped[roomID]
	if exists {
		close(prev)
	}
	s.stopped[roomID] = stopCh

	go s.roomLoop(room, stopCh)
}

// StopRoomLoop 停止房间消息循环
func (s *gameService) StopRoomLoop(roomID string) {
	if ch, ok := s.stopped[roomID]; ok {
		close(ch)
		delete(s.stopped, roomID)
	}
}

// roomLoop 房间消息循环：从 Inbox 读取消息并分发
// Nope 窗口定时器由 TimerService 管理，超时后通过 "nope_timeout" 消息回到此循环
func (s *gameService) roomLoop(room *model.Room, stopCh chan struct{}) {
	for {
		select {
		case <-stopCh:
			return
		case msg, ok := <-room.Inbox:
			if !ok {
				return
			}
			// 定时器触发的消息：nope_timeout / auto_draw / reconnect_timeout
			switch msg.Type {
			case "nope_timeout":
				s.resolveStack(room)
				continue
			case "auto_draw":
				s.handleDrawCard(room, msg)
				continue
			case "reconnect_timeout":
				s.handleReconnectTimeout(room, msg)
				continue
			}

			// Nope 在 Nope 窗口内：直接入栈并重置 Nope 定时器
			if msg.Type == "play_nope" {
				if s.isNopeWindowActive(room) {
					s.handleNope(room, msg)
					s.timerSvc.ResetNopeTimer(room.ID)
					continue
				}
			}

			// 其他消息进入正常分发
			s.handleMessage(room, msg)
		}
	}
}

// isNopeWindowActive 检查是否处于 Nope 窗口期（动作栈非空且栈顶非 Nope）
func (s *gameService) isNopeWindowActive(room *model.Room) bool {
	if room.GameState == nil {
		return false
	}
	return action.IsNopeChainActive(&room.GameState.ActionStack)
}

// handleReconnectTimeout 处理重连超时：将玩家标记为超时
func (s *gameService) handleReconnectTimeout(room *model.Room, msg model.Message) {
	player, ok := room.FindPlayer(msg.PlayerID)
	if !ok {
		return
	}
	player.ConnectionStatus = model.ConnectionStatusTimedOut
	s.broadcastEvent(room, "player_timeout", map[string]string{
		"player_id": msg.PlayerID,
	})
}

// opensNopeWindow 判断动作类型是否需要打开 Nope 窗口
func opensNopeWindow(actionType model.ActionType) bool {
	switch actionType {
	case model.ActionPlayCard:
		return true
	default:
		return false
	}
}

// handleMessage 分发消息到对应处理器
func (s *gameService) handleMessage(room *model.Room, msg model.Message) {
	switch msg.Type {
	case "play_card":
		s.handlePlayCard(room, msg)
	case "draw_card":
		s.handleDrawCard(room, msg)
	case "select_target":
		s.handleSelectTarget(room, msg)
	case "defuse_place":
		s.handleDefusePlace(room, msg)
	case "pass_nope":
		// 玩家主动跳过 Nope，由定时器统一处理
	default:
		s.sendError(room, msg.PlayerID, msg.ReqID, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

// handlePlayCard 处理出牌：校验 -> 入栈 -> 打开 Nope 窗口
func (s *gameService) handlePlayCard(room *model.Room, msg model.Message) {
	gs := room.GameState
	if gs == nil {
		s.sendError(room, msg.PlayerID, msg.ReqID, "game not initialized")
		return
	}

	payload, ok := msg.Payload.(PlayCardPayload)
	if !ok {
		// 尝试从 map 转换
		payload = s.coercePlayCardPayload(msg.Payload)
	}

	player, ok := room.FindPlayer(msg.PlayerID)
	if !ok {
		s.sendError(room, msg.PlayerID, msg.ReqID, "player not found")
		return
	}

	// 校验出牌
	if _, err := validator.ValidatePlay(validator.PlayContext{
		PlayerID:      msg.PlayerID,
		GamePhase:     gs.Phase,
		TurnPhase:     gs.Turn.Phase,
		CurrentPlayer: gs.Turn.CurrentPlayerID,
		IsAlive:       player.Alive,
		HandSize:      player.Hand.Count,
		CardCount:     len(payload.CardIDs),
	}); err != nil {
		s.sendError(room, msg.PlayerID, msg.ReqID, err.Error())
		return
	}

	if _, err := validator.ValidateCombo(&player.Hand, payload.CardIDs); err != nil {
		s.sendError(room, msg.PlayerID, msg.ReqID, err.Error())
		return
	}

	// 校验目标玩家（如有）
	if payload.TargetPlayerID != "" {
		playerIDs := make([]string, 0, len(room.Players))
		for _, p := range room.Players {
			playerIDs = append(playerIDs, p.ID)
		}
		if _, err := validator.ValidateTarget(payload.TargetPlayerID, playerIDs, msg.PlayerID); err != nil {
			s.sendError(room, msg.PlayerID, msg.ReqID, err.Error())
			return
		}
	}

	// 从手牌移除并加入弃牌堆
	var cards []model.Card
	for _, id := range payload.CardIDs {
		c, ok := player.Hand.RemoveCard(id)
		if !ok {
			s.sendError(room, msg.PlayerID, msg.ReqID, "card not found")
			return
		}
		cards = append(cards, c)
	}

	// 构造动作并入栈
	act := model.Action{
		Type:           model.ActionPlayCard,
		CardIDs:        payload.CardIDs,
		TargetPlayerID: payload.TargetPlayerID,
	}
	action.PushAction(&gs.ActionStack, act, msg.PlayerID)

	// 弃牌堆顶为最后打出的牌
	for _, c := range cards {
		gs.DiscardPile.Add(c)
	}

	// 启动 Nope 窗口定时器
	s.timerSvc.StartNopeTimer(room.ID)
	s.broadcastState(room)
	s.broadcastNopePrompt(room, msg.PlayerID, cards[0].Type)

	return
}

// handleNope 处理 Nope：入栈，重置由调用方处理
func (s *gameService) handleNope(room *model.Room, msg model.Message) {
	gs := room.GameState
	if gs == nil {
		return
	}

	if !action.CanNope(&gs.ActionStack) {
		s.sendError(room, msg.PlayerID, msg.ReqID, model.ErrNopeNotAllowed.Error())
		return
	}

	player, ok := room.FindPlayer(msg.PlayerID)
	if !ok {
		return
	}

	// 校验 Nope 出牌合法性
	if _, err := validator.ValidateNope(validator.PlayContext{
		PlayerID:      msg.PlayerID,
		GamePhase:     gs.Phase,
		TurnPhase:     gs.Turn.Phase,
		CurrentPlayer: gs.Turn.CurrentPlayerID,
		IsAlive:       player.Alive,
	}); err != nil {
		s.sendError(room, msg.PlayerID, msg.ReqID, err.Error())
		return
	}

	// 从手牌移除一张 Nope
	nopeCards := player.Hand.FindCardsByType(model.CardTypeNope)
	if len(nopeCards) == 0 {
		s.sendError(room, msg.PlayerID, msg.ReqID, "no nope card in hand")
		return
	}
	player.Hand.RemoveCard(nopeCards[0].ID)
	gs.DiscardPile.Add(nopeCards[0])

	action.PushNope(&gs.ActionStack, msg.PlayerID)
	s.broadcastState(room)
}

// resolveStack 结算 Nope 链，执行未被阻止的基础动作
func (s *gameService) resolveStack(room *model.Room) {
	gs := room.GameState
	if gs == nil || gs.ActionStack.Len() == 0 {
		return
	}

	s.timerSvc.StopNopeTimer(room.ID)
	result := action.Resolve(&gs.ActionStack)

	if result.ActionBlocked {
		s.broadcastEvent(room, "nope_resolved", map[string]interface{}{
			"blocked":     true,
			"nope_count":  result.NopeCount,
			"base_action": result.Action,
		})
		s.broadcastState(room)
		return
	}

	s.broadcastEvent(room, "nope_resolved", map[string]interface{}{
		"blocked":     false,
		"nope_count":  result.NopeCount,
		"base_action": result.Action,
	})

	// 执行基础动作（出牌效果）
	s.executeAction(room, result.Action)
	s.broadcastState(room)

	// 检查游戏结束
	if res := gameover.CheckGameOver(gs); res.IsOver {
		s.handleGameOver(room, res)
	}
}

// executeAction 执行已生效的出牌动作
func (s *gameService) executeAction(room *model.Room, act model.Action) {
	gs := room.GameState
	if len(act.CardIDs) == 0 {
		return
	}

	// 取第一张牌的类型作为动作类型（多张猫牌组合也取第一张）
	topCard, _ := gs.DiscardPile.TopCard()
	switch topCard.Type {
	case model.CardTypeSkip:
		s.turnSvc.HandleSkip(gs)
	case model.CardTypeAttack:
		s.turnSvc.HandleAttack(gs)
	case model.CardTypeShuffle:
		gs.Deck.Shuffle()
	case model.CardTypeSeeTheFuture:
		// 通知当前玩家偷看牌顶 3 张
		future := gs.Deck.Peek(3)
		s.broadcaster.SendToPlayer(room.ID, gs.Turn.CurrentPlayerID, "see_the_future", future)
	case model.CardTypeFavor:
		if act.TargetPlayerID != "" {
			s.handleFavor(room, act.TargetPlayerID, gs.Turn.CurrentPlayerID)
		}
	case model.CardTypeCatCard:
		s.handleCatCombo(room, act)
	default:
		// 其他类型（含 Nope）此处不处理
	}
}

// handleFavor 索要牌：从目标玩家手牌随机抽一张给当前玩家
func (s *gameService) handleFavor(room *model.Room, targetID, requesterID string) {
	target, ok := room.FindPlayer(targetID)
	if !ok || target.Hand.Count == 0 {
		return
	}
	requester, ok := room.FindPlayer(requesterID)
	if !ok {
		return
	}
	idx := len(target.Hand.Cards) - 1 // 取最后一张
	card := target.Hand.Cards[idx]
	target.Hand.RemoveCard(card.ID)
	requester.Hand.AddCard(card)
	s.broadcastEvent(room, "favor_completed", map[string]string{
		"from": targetID,
		"to":   requesterID,
	})
}

// handleCatCombo 猫牌组合效果（3 张相同或 5 张不同）
// 实现简化版：直接抽目标玩家手牌或弃牌堆顶
func (s *gameService) handleCatCombo(room *model.Room, act model.Action) {
	gs := room.GameState
	if len(act.CardIDs) == 5 {
		// 五张不同：拿弃牌堆顶
		if top, ok := gs.DiscardPile.TopCard(); ok {
			player, ok := room.FindPlayer(gs.Turn.CurrentPlayerID)
			if ok {
				player.Hand.AddCard(top)
				gs.DiscardPile.Cards = gs.DiscardPile.Cards[:len(gs.DiscardPile.Cards)-1]
			}
		}
		return
	}
	// 三张相同：从目标玩家手牌抽一张
	if act.TargetPlayerID != "" {
		s.handleFavor(room, act.TargetPlayerID, gs.Turn.CurrentPlayerID)
	}
}

// handleDrawCard 抽牌流程：含爆炸猫判定 -> 拆解/出局
func (s *gameService) handleDrawCard(room *model.Room, msg model.Message) {
	gs := room.GameState
	if gs == nil {
		s.sendError(room, msg.PlayerID, msg.ReqID, "game not initialized")
		return
	}

	player, ok := room.FindPlayer(msg.PlayerID)
	if !ok {
		s.sendError(room, msg.PlayerID, msg.ReqID, "player not found")
		return
	}

	if _, err := validator.ValidateDraw(validator.PlayContext{
		PlayerID:      msg.PlayerID,
		GamePhase:     gs.Phase,
		TurnPhase:     gs.Turn.Phase,
		CurrentPlayer: gs.Turn.CurrentPlayerID,
		IsAlive:       player.Alive,
	}); err != nil {
		s.sendError(room, msg.PlayerID, msg.ReqID, err.Error())
		return
	}

	card, ok := gs.Deck.Draw()
	if !ok {
		s.sendError(room, msg.PlayerID, msg.ReqID, "deck is empty")
		return
	}

	if card.Type == model.CardTypeExplodingKitten {
		// 抽到爆炸猫：检查是否有拆解
		if player.DefuseCount > 0 {
			player.DefuseCount--
			// 移除一张拆解牌
			defuses := player.Hand.FindCardsByType(model.CardTypeDefuse)
			if len(defuses) > 0 {
				player.Hand.RemoveCard(defuses[0].ID)
				gs.DiscardPile.Add(defuses[0])
			}
			// 通知玩家选择插回位置
			s.broadcaster.SendToPlayer(room.ID, msg.PlayerID, "defuse_required", map[string]interface{}{
				"exploding_kitten_id": card.ID,
			})
			// 简化处理：默认插回牌库顶
			gs.Deck.InsertAt(card, 0)
			s.turnSvc.HandleExplodingKitten(gs)
			s.broadcastEvent(room, "exploding_defused", map[string]string{
				"player_id": msg.PlayerID,
			})
		} else {
			// 出局
			player.Alive = false
			gs.DiscardPile.Add(card)
			res := gameover.HandleElimination(gs, msg.PlayerID)
			s.turnSvc.HandleElimination(gs, msg.PlayerID)
			s.broadcastEvent(room, "player_eliminated", map[string]string{
				"player_id": msg.PlayerID,
			})
			if res.IsOver {
				s.handleGameOver(room, res)
				return
			}
		}
	} else {
		// 普通牌：加入手牌
		player.Hand.AddCard(card)
		s.turnSvc.HandleDraw(gs)
	}

	s.timerSvc.ResetTurnTimer(room.ID, gs.Turn.CurrentPlayerID)
	s.broadcastState(room)

	// 检查游戏结束
	if res := gameover.CheckGameOver(gs); res.IsOver {
		s.handleGameOver(room, res)
	}
}

// handleSelectTarget 处理目标选择（如猫牌组合）
func (s *gameService) handleSelectTarget(room *model.Room, msg model.Message) {
	// 简化实现：直接更新待执行动作的目标
	// 完整实现需要状态机记录"等待目标选择"阶段
	s.broadcastState(room)
}

// handleDefusePlace 处理拆解牌插回位置
func (s *gameService) handleDefusePlace(room *model.Room, msg model.Message) {
	gs := room.GameState
	if gs == nil {
		return
	}
	// 简化：玩家选择位置已通过 DefusePlacePayload.DeckPosition 传递
	// 完整实现需要校验阶段并更新爆炸猫位置
	s.broadcastState(room)
}

// handleGameOver 游戏结束处理
func (s *gameService) handleGameOver(room *model.Room, res gameover.GameOverResult) {
	gs := room.GameState
	gameover.SetGameOver(gs)
	room.Status = model.RoomStatusFinished
	s.broadcastEvent(room, "game_over", map[string]interface{}{
		"winner_id":    res.WinnerID,
		"alive_count":  res.AliveCount,
		"eliminated":   res.Eliminated,
	})
	s.timerSvc.StopAll(room.ID)
}

// broadcastState 广播状态更新（按视角过滤）
func (s *gameService) broadcastState(room *model.Room) {
	if s.broadcaster == nil || room.GameState == nil {
		return
	}
	for _, p := range room.Players {
		if !p.Alive {
			continue
		}
		s.broadcaster.SendToPlayer(room.ID, p.ID, "state_update", buildPlayerView(room, p.ID))
	}
}

// broadcastNopePrompt 广播 Nope 提示（除出牌者外的玩家可否决）
func (s *gameService) broadcastNopePrompt(room *model.Room, actorID string, cardType model.CardType) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.Broadcast(room.ID, "nope_prompt", map[string]interface{}{
		"actor_id":     actorID,
		"card_type":    string(cardType),
		"window_sec":   int(s.nopeWindow / time.Second),
	})
}

// broadcastEvent 广播通用事件
func (s *gameService) broadcastEvent(room *model.Room, event string, payload interface{}) {
	if s.broadcaster != nil {
		s.broadcaster.Broadcast(room.ID, event, payload)
	}
}

// sendError 向单个玩家发送错误
func (s *gameService) sendError(room *model.Room, playerID, reqID, message string) {
	if s.broadcaster == nil {
		return
	}
	s.broadcaster.SendToPlayer(room.ID, playerID, "error", map[string]interface{}{
		"req_id": reqID,
		"message": message,
	})
}

// coercePlayCardPayload 将 map[string]interface{} 转为 PlayCardPayload
func (s *gameService) coercePlayCardPayload(raw interface{}) PlayCardPayload {
	m, ok := raw.(map[string]interface{})
	if !ok {
		return PlayCardPayload{}
	}
	p := PlayCardPayload{}
	if v, ok := m["card_ids"].([]interface{}); ok {
		for _, item := range v {
			if s, ok := item.(string); ok {
				p.CardIDs = append(p.CardIDs, s)
			}
		}
	}
	if v, ok := m["target_player_id"].(string); ok {
		p.TargetPlayerID = v
	}
	if v, ok := m["target_card_type"].(string); ok {
		p.TargetCardType = v
	}
	if v, ok := m["deck_position"].(float64); ok {
		p.DeckPosition = int(v)
	}
	return p
}

// buildPlayerView 构建玩家视角状态
func buildPlayerView(room *model.Room, playerID string) interface{} {
	gs := room.GameState
	player, ok := room.FindPlayer(playerID)
	if !ok {
		return nil
	}
	hand := make([]model.Card, len(player.Hand.Cards))
	copy(hand, player.Hand.Cards)

	players := make([]model.PlayerInfo, 0, len(room.Players))
	for _, p := range room.Players {
		players = append(players, p.ToPlayerInfo())
	}

	view := map[string]interface{}{
		"game_phase":   int(gs.Phase),
		"round_count":  gs.RoundCount,
		"your_hand":    hand,
		"your_defuses": player.DefuseCount,
		"players":      players,
		"turn":         gs.Turn,
		"action_stack": gs.ActionStack.Stack,
		"alive_count":  countAlive(room),
		"deck_count":   gs.Deck.Count,
	}
	if top, ok := gs.DiscardPile.TopCard(); ok {
		view["discard_top"] = top
	}
	return view
}

func countAlive(room *model.Room) int {
	n := 0
	for _, p := range room.Players {
		if p.Alive {
			n++
		}
	}
	return n
}
