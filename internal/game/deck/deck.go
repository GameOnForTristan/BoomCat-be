package deck

import (
	"boom-cat-be/internal/game/card"
	"boom-cat-be/internal/model"
	"fmt"
)

// BuildInitialDeck 根据玩家人数构建初始牌库
// 规格：4 张拆解 + 行动牌各 4 张 + 20 张猫牌 + (玩家数-1) 张爆炸猫
// 返回的牌库已洗牌，但爆炸猫尚未插入（需调用 InsertExplodingKittens）
func BuildInitialDeck(playerCount int, comp card.DeckComposition) *model.Deck {
	if playerCount < 2 {
		playerCount = 2
	}
	if playerCount > 5 {
		playerCount = 5
	}

	var cards []model.Card

	// 拆解牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeDefuse, "defuse", comp.DefuseCount)...)
	// 攻击牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeAttack, "attack", comp.AttackCount)...)
	// 跳过牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeSkip, "skip", comp.SkipCount)...)
	// 索要牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeFavor, "favor", comp.FavorCount)...)
	// 洗牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeShuffle, "shuffle", comp.ShuffleCount)...)
	// 预言牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeSeeTheFuture, "see_future", comp.SeeFutureCount)...)
	// 否决牌
	cards = append(cards, card.GenerateActionCards(model.CardTypeNope, "nope", comp.NopeCount)...)
	// 猫牌
	cards = append(cards, card.GenerateCatCards("cat", comp.CatCardCount)...)

	deck := model.NewDeck(cards)
	deck.Shuffle()
	return deck
}

// InsertExplodingKittens 将爆炸猫插入牌库
// 每位玩家发 4 张初始手牌后，将 (playerCount - 1) 张爆炸猫插入牌库
func InsertExplodingKittens(deck *model.Deck, playerCount int) {
	kittenCount := playerCount - 1
	for i := 0; i < kittenCount; i++ {
		kitten := model.NewCard(model.CardTypeExplodingKitten, fmt.Sprintf("ek_%d", i))
		// 插入到牌库顶部
		deck.InsertAt(kitten, 0)
	}
}

// BuildGameDeck 构建完整游戏牌库（含爆炸猫）
// 1. 构建初始牌库
// 2. 发初始手牌
// 3. 将爆炸猫插入牌库并洗牌
// 返回牌库和初始手牌
func BuildGameDeck(playerCount int, playerIDs []string) (*model.Deck, map[string][]model.Card) {
	comp := card.DefaultDeckComposition()
	deck := BuildInitialDeck(playerCount, comp)

	// 发初始手牌（每人 4 张 + 1 张拆解）
	hands := make(map[string][]model.Card)
	for _, pid := range playerIDs {
		hand := make([]model.Card, 0, 5)
		// 每人一张拆解
		defuse, ok := deck.Draw()
		if ok && defuse.Type == model.CardTypeDefuse {
			hand = append(hand, defuse)
		}
		// 每人再抽 4 张
		for i := 0; i < 4; i++ {
			card, ok := deck.Draw()
			if ok {
				hand = append(hand, card)
			}
		}
		hands[pid] = hand
	}

	// 插入爆炸猫
	InsertExplodingKittens(deck, playerCount)
	// 再次洗牌（让爆炸猫随机分布）
	deck.Shuffle()

	return deck, hands
}