package turn

import (
	"boom-cat-be/internal/model"
)

// AdvanceTurn 推进到下一个玩家
// 返回新的当前玩家 ID
func AdvanceTurn(te *model.TurnEngine) string {
	te.OnSkip()
	return te.CurrentPlayer()
}

// HandleSkip 处理 Skip 牌效果：跳过当前玩家，进入下一位
func HandleSkip(te *model.TurnEngine) {
	te.OnSkip()
}

// HandleAttack 处理 Attack 牌效果：累积 2 张待抽，轮到下家
func HandleAttack(te *model.TurnEngine) {
	te.OnAttack()
}

// HandleDraw 处理抽牌：减少待抽计数，如果归零则推进到下一玩家
// 返回 true 表示回合已切换
func HandleDraw(te *model.TurnEngine) bool {
	te.OnCardDrawn()
	return te.Phase == model.PhaseAction
}

// HandleExplodingKitten 处理抽到爆炸猫（已使用拆解）
// 玩家存活，继续抽牌流程
func HandleExplodingKitten(te *model.TurnEngine) {
	te.OnExplodingKittenDrawn()
}

// HandleElimination 处理玩家出局
// 从 TurnEngine 中移除该玩家，如果游戏结束则返回 true
func HandleElimination(te *model.TurnEngine, playerID string) bool {
	te.RemovePlayer(playerID)
	return te.Phase == model.PhaseGameOver
}

// HandleReverse 处理方向反转
func HandleReverse(te *model.TurnEngine) {
	te.ReverseDirection()
}

// GetNextPlayer 获取当前玩家的下一个玩家 ID
func GetNextPlayer(te *model.TurnEngine) string {
	return te.NextPlayer()
}

// IsTurnAction 检查当前阶段是否允许出牌
func IsTurnAction(te *model.TurnEngine) bool {
	return te.Phase == model.PhaseAction
}

// IsTurnDrawing 检查当前阶段是否允许抽牌
func IsTurnDrawing(te *model.TurnEngine) bool {
	return te.Phase == model.PhaseDrawing
}

// IsGameOver 检查是否游戏结束
func IsGameOver(te *model.TurnEngine) bool {
	return te.Phase == model.PhaseGameOver
}

// GetPlayerOrder 获取当前存活玩家列表（按座位顺序）
func GetPlayerOrder(te *model.TurnEngine) []string {
	result := make([]string, len(te.Players))
	copy(result, te.Players)
	return result
}