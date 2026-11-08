package service

import (
	"boom-cat-be/internal/model"
)

// TurnService 回合服务接口
// 直接操作 model.TurnState（GameState 内嵌字段），避免 TurnEngine/TurnState 双重抽象
type TurnService interface {
	// AdvanceTurn 推进到下一个玩家
	AdvanceTurn(gs *model.GameState) string
	// HandleSkip 处理 Skip 牌效果
	HandleSkip(gs *model.GameState)
	// HandleAttack 处理 Attack 牌效果（累加 2 张待抽）
	HandleAttack(gs *model.GameState)
	// HandleDraw 处理一次抽牌，返回是否回合切换
	HandleDraw(gs *model.GameState) bool
	// HandleExplodingKitten 处理抽到爆炸猫（已使用拆解）
	HandleExplodingKitten(gs *model.GameState)
	// HandleElimination 玩家出局，返回是否游戏结束
	HandleElimination(gs *model.GameState, playerID string) bool
	// HandleReverse 反转方向
	HandleReverse(gs *model.GameState)
	// CurrentPlayer 当前玩家
	CurrentPlayer(gs *model.GameState) string
	// NextPlayer 下一个玩家
	NextPlayer(gs *model.GameState) string
	// PlayerOrder 存活玩家顺序
	PlayerOrder(gs *model.GameState) []string
	// IsTurnAction 是否处于出牌阶段
	IsTurnAction(gs *model.GameState) bool
	// IsTurnDrawing 是否处于抽牌阶段
	IsTurnDrawing(gs *model.GameState) bool
	// IsGameOver 是否游戏结束
	IsGameOver(gs *model.GameState) bool
}

type turnService struct{}

// NewTurnService 创建回合服务
func NewTurnService() TurnService {
	return &turnService{}
}

// nextIndex 计算下一个存活玩家索引（环形，考虑方向）
func nextIndex(order []string, currentIdx, direction int) int {
	n := len(order)
	if n == 0 {
		return 0
	}
	next := currentIdx + direction
	if next >= n {
		return 0
	}
	if next < 0 {
		return n - 1
	}
	return next
}

// findCurrentIndex 在 PlayerOrder 中找到 CurrentPlayerID 的索引
func findCurrentIndex(gs *model.GameState) int {
	for i, pid := range gs.Turn.PlayerOrder {
		if pid == gs.Turn.CurrentPlayerID {
			return i
		}
	}
	return 0
}

// AdvanceTurn 推进到下一个玩家
func (s *turnService) AdvanceTurn(gs *model.GameState) string {
	if gs == nil || len(gs.Turn.PlayerOrder) == 0 {
		return ""
	}
	idx := findCurrentIndex(gs)
	next := nextIndex(gs.Turn.PlayerOrder, idx, gs.Turn.Direction)
	gs.Turn.CurrentPlayerID = gs.Turn.PlayerOrder[next]
	gs.Turn.Phase = model.PhaseAction
	return gs.Turn.CurrentPlayerID
}

// HandleSkip 处理 Skip 牌效果
func (s *turnService) HandleSkip(gs *model.GameState) {
	if gs == nil {
		return
	}
	s.AdvanceTurn(gs)
}

// HandleAttack 处理 Attack 牌效果：累加 2 张待抽并轮到下家
func (s *turnService) HandleAttack(gs *model.GameState) {
	if gs == nil {
		return
	}
	gs.Turn.PendingDraws += 2
	s.AdvanceTurn(gs)
	gs.Turn.Phase = model.PhaseDrawing
}

// HandleDraw 处理一次抽牌
// 返回 true 表示回合已切换
func (s *turnService) HandleDraw(gs *model.GameState) bool {
	if gs == nil {
		return false
	}
	gs.Turn.PendingDraws--
	if gs.Turn.PendingDraws <= 0 {
		gs.Turn.PendingDraws = 0
		s.AdvanceTurn(gs)
		return true
	}
	return false
}

// HandleExplodingKitten 处理抽到爆炸猫（已使用拆解）
func (s *turnService) HandleExplodingKitten(gs *model.GameState) {
	if gs == nil {
		return
	}
	gs.Turn.PendingDraws--
	if gs.Turn.PendingDraws <= 0 {
		gs.Turn.PendingDraws = 0
		s.AdvanceTurn(gs)
	}
}

// HandleElimination 玩家出局，从轮转中移除
func (s *turnService) HandleElimination(gs *model.GameState, playerID string) bool {
	if gs == nil {
		return false
	}
	order := gs.Turn.PlayerOrder
	idx := -1
	for i, pid := range order {
		if pid == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return len(order) <= 1
	}

	currentIdx := findCurrentIndex(gs)
	if idx < currentIdx {
		currentIdx--
	} else if idx == currentIdx {
		// 当前玩家出局，下一个玩家接管
		next := nextIndex(order, idx, gs.Turn.Direction)
		if next >= len(order) {
			next = 0
		}
		if next == idx {
			// 只剩一人
		}
	}

	order = append(order[:idx], order[idx+1:]...)
	gs.Turn.PlayerOrder = order

	// 重新计算 CurrentPlayerID
	if len(order) > 0 {
		if currentIdx >= len(order) {
			currentIdx = 0
		}
		gs.Turn.CurrentPlayerID = order[currentIdx]
	}

	if len(order) <= 1 {
		gs.Turn.Phase = model.PhaseGameOver
		return true
	}
	return false
}

// HandleReverse 反转方向
func (s *turnService) HandleReverse(gs *model.GameState) {
	if gs == nil {
		return
	}
	gs.Turn.Direction = -gs.Turn.Direction
}

// CurrentPlayer 当前玩家
func (s *turnService) CurrentPlayer(gs *model.GameState) string {
	if gs == nil {
		return ""
	}
	return gs.Turn.CurrentPlayerID
}

// NextPlayer 下一个玩家
func (s *turnService) NextPlayer(gs *model.GameState) string {
	if gs == nil || len(gs.Turn.PlayerOrder) == 0 {
		return ""
	}
	idx := findCurrentIndex(gs)
	next := nextIndex(gs.Turn.PlayerOrder, idx, gs.Turn.Direction)
	return gs.Turn.PlayerOrder[next]
}

// PlayerOrder 存活玩家顺序
func (s *turnService) PlayerOrder(gs *model.GameState) []string {
	if gs == nil {
		return nil
	}
	out := make([]string, len(gs.Turn.PlayerOrder))
	copy(out, gs.Turn.PlayerOrder)
	return out
}

// IsTurnAction 是否处于出牌阶段
func (s *turnService) IsTurnAction(gs *model.GameState) bool {
	if gs == nil {
		return false
	}
	return gs.Turn.Phase == model.PhaseAction
}

// IsTurnDrawing 是否处于抽牌阶段
func (s *turnService) IsTurnDrawing(gs *model.GameState) bool {
	if gs == nil {
		return false
	}
	return gs.Turn.Phase == model.PhaseDrawing
}

// IsGameOver 是否游戏结束
func (s *turnService) IsGameOver(gs *model.GameState) bool {
	if gs == nil {
		return true
	}
	return gs.Turn.Phase == model.PhaseGameOver || gs.Phase == model.GamePhaseGameOver
}
