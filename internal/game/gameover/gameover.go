package gameover

import (
	"boom-cat-be/internal/model"
)

// GameOverResult 游戏结束结果
type GameOverResult struct {
	IsOver     bool     `json:"is_over"`
	WinnerID   string   `json:"winner_id,omitempty"`
	AliveCount int      `json:"alive_count"`
	Eliminated []string `json:"eliminated,omitempty"`
}

// CheckGameOver 检查游戏是否结束
// 如果只剩 0 或 1 个存活玩家，游戏结束
func CheckGameOver(gs *model.GameState) GameOverResult {
	aliveIDs := make([]string, 0)
	eliminatedIDs := make([]string, 0)

	for _, p := range gs.Players {
		if p.Alive {
			aliveIDs = append(aliveIDs, p.ID)
		} else {
			eliminatedIDs = append(eliminatedIDs, p.ID)
		}
	}

	isOver := len(aliveIDs) <= 1
	winnerID := ""
	if isOver && len(aliveIDs) == 1 {
		winnerID = aliveIDs[0]
	}

	return GameOverResult{
		IsOver:     isOver,
		WinnerID:   winnerID,
		AliveCount: len(aliveIDs),
		Eliminated: eliminatedIDs,
	}
}

// HandleElimination 处理玩家出局
// 更新游戏状态，返回出局后的游戏结果
func HandleElimination(gs *model.GameState, playerID string) GameOverResult {
	for i, p := range gs.Players {
		if p.ID == playerID {
			gs.Players[i].Alive = false
			break
		}
	}
	return CheckGameOver(gs)
}

// GetWinner 获取获胜者信息
// 如果游戏未结束返回空字符串
func GetWinner(gs *model.GameState) string {
	if !gs.IsGameOver() {
		return ""
	}
	return gs.Winner()
}

// SetGameOver 设置游戏状态为结束
func SetGameOver(gs *model.GameState) {
	gs.Phase = model.GamePhaseGameOver
}

// ResetGameState 重置游戏状态（用于重新开始）
func ResetGameState(gs *model.GameState) {
	gs.Phase = model.GamePhaseWaiting
	gs.Deck = *model.NewDeck(make([]model.Card, 0))
	gs.DiscardPile = *model.NewDiscardPile()
	gs.ActionStack = *model.NewActionStack()
	gs.RoundCount = 0
}
