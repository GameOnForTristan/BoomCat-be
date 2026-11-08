package card

import (
	"boom-cat-be/internal/model"
	"fmt"
)

// 五张不同猫牌的名称
var CatCardNames = []string{
	"彩虹猫",     // Rainbow Cat
	"水煮猫",     // Cattermelon
	"玉米猫",     // Taco Cat
	"毛土豆猫",   // Hairy Potato Cat
	"胡子猫",     // Beard Cat
}

// NewCatCard 创建一张指定名称的猫牌
func NewCatCard(id, name string) model.Card {
	return model.Card{
		ID:   id,
		Type: model.CardTypeCatCard,
		Name: name,
	}
}

// GenerateCatCards 生成指定数量的猫牌（5种名称平均分配）
func GenerateCatCards(baseID string, count int) []model.Card {
	if count <= 0 {
		return nil
	}
	cards := make([]model.Card, 0, count)
	nameCount := len(CatCardNames)
	for i := 0; i < count; i++ {
		name := CatCardNames[i%nameCount]
		cardID := fmt.Sprintf("%s_cat_%d", baseID, i)
		cards = append(cards, NewCatCard(cardID, name))
	}
	return cards
}

// GenerateActionCards 生成指定类型和数量的行动牌
func GenerateActionCards(cardType model.CardType, baseID string, count int) []model.Card {
	cards := make([]model.Card, 0, count)
	for i := 0; i < count; i++ {
		cardID := fmt.Sprintf("%s_%d", baseID, i)
		cards = append(cards, model.NewCard(cardType, cardID))
	}
	return cards
}

// DeckComposition 牌库组成配置
type DeckComposition struct {
	DefuseCount    int
	AttackCount    int
	SkipCount      int
	FavorCount     int
	ShuffleCount   int
	SeeFutureCount int
	NopeCount      int
	CatCardCount   int
}

// DefaultDeckComposition 返回默认牌库组成（标准 Exploding Kittens 配置）
func DefaultDeckComposition() DeckComposition {
	return DeckComposition{
		DefuseCount:    4,
		AttackCount:    4,
		SkipCount:      4,
		FavorCount:     4,
		ShuffleCount:   4,
		SeeFutureCount: 4,
		NopeCount:      4,
		CatCardCount:   20,
	}
}