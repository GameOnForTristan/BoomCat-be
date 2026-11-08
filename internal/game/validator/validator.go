package validator

import (
	"boom-cat-be/internal/model"
)

// PlayContext 出牌上下文
type PlayContext struct {
	PlayerID      string
	GamePhase     model.GamePhase
	TurnPhase     model.TurnPhase
	CurrentPlayer string
	IsAlive       bool
	HasDefuse     bool
	HandSize      int
	CardCount     int
}

// ValidatePlay 校验出牌合法性
// 返回 (是否合法, 错误)
func ValidatePlay(ctx PlayContext) (bool, error) {
	// 游戏必须在进行中
	if ctx.GamePhase != model.GamePhasePlaying {
		return false, model.ErrInvalidPhase
	}

	// 玩家必须存活
	if !ctx.IsAlive {
		return false, model.ErrPlayerEliminated
	}

	// 必须是当前玩家的回合
	if ctx.PlayerID != ctx.CurrentPlayer {
		return false, model.ErrNotYourTurn
	}

	// 必须在出牌阶段
	if ctx.TurnPhase != model.PhaseAction {
		return false, model.ErrInvalidPhase
	}

	// 手牌不能为空
	if ctx.HandSize <= 0 {
		return false, model.ErrHandEmpty
	}

	return true, nil
}

// ValidateDraw 校验抽牌合法性
func ValidateDraw(ctx PlayContext) (bool, error) {
	if ctx.GamePhase != model.GamePhasePlaying {
		return false, model.ErrInvalidPhase
	}
	if !ctx.IsAlive {
		return false, model.ErrPlayerEliminated
	}
	if ctx.PlayerID != ctx.CurrentPlayer {
		return false, model.ErrNotYourTurn
	}
	if ctx.TurnPhase != model.PhaseDrawing {
		return false, model.ErrInvalidPhase
	}
	return true, nil
}

// ValidateNope 校验出 Nope 合法性
func ValidateNope(ctx PlayContext) (bool, error) {
	if ctx.GamePhase != model.GamePhasePlaying {
		return false, model.ErrInvalidPhase
	}
	if !ctx.IsAlive {
		return false, model.ErrPlayerEliminated
	}
	// Nope 可以在任何阶段打出（只要有动作栈），但不在出牌阶段时限制更宽松
	return true, nil
}

// ValidateTarget 校验目标玩家合法性
func ValidateTarget(targetID string, playerIDs []string, excludeID string) (bool, error) {
	if targetID == "" {
		return false, model.ErrInvalidTarget
	}
	if targetID == excludeID {
		return false, model.ErrInvalidTarget
	}
	found := false
	for _, pid := range playerIDs {
		if pid == targetID {
			found = true
			break
		}
	}
	if !found {
		return false, model.ErrPlayerNotFound
	}
	return true, nil
}

// ValidateCombo 校验猫牌组合合法性
// 支持三张相同和五张不同
func ValidateCombo(hand *model.Hand, cardIDs []string) (bool, error) {
	if len(cardIDs) == 3 {
		if hand.ValidateThreeSameCombo(cardIDs) {
			return true, nil
		}
		return false, model.ErrInvalidCombo
	}
	if len(cardIDs) == 5 {
		if hand.ValidateFiveDifferentCombo(cardIDs) {
			return true, nil
		}
		return false, model.ErrInvalidCombo
	}
	if len(cardIDs) == 1 {
		card, found := hand.FindCard(cardIDs[0])
		if !found {
			return false, model.ErrCardNotFound
		}
		// 单张只能出行动牌
		switch card.Type {
		case model.CardTypeSkip, model.CardTypeAttack, model.CardTypeFavor,
			model.CardTypeShuffle, model.CardTypeSeeTheFuture, model.CardTypeNope:
			return true, nil
		default:
			return false, model.ErrInvalidCombo
		}
	}
	return false, model.ErrInvalidCombo
}

// ValidateDefuse 校验使用拆解牌的合法性
func ValidateDefuse(hand *model.Hand) (bool, error) {
	if hand.Count <= 0 {
		return false, model.ErrHandEmpty
	}
	cards := hand.FindCardsByType(model.CardTypeDefuse)
	if len(cards) == 0 {
		return false, model.ErrCardNotFound
	}
	return true, nil
}