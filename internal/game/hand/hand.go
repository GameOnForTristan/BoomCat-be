package hand

import (
	"boom-cat-be/internal/model"
	"sort"
)

// ComboType 组合类型
type ComboType int

const (
	ComboNone              ComboType = iota
	ComboSingleCard                  // 单张出牌（Skip, Attack, Favor, Shuffle, SeeTheFuture, Nope）
	ComboThreeSameCat               // 三张相同猫牌
	ComboFiveDifferentCat           // 五张不同猫牌
)

// ComboInfo 组合信息
type ComboInfo struct {
	Type    ComboType
	CardIDs []string
}

// CanPlayAnyCard 检查手牌中是否有可出的牌
func CanPlayAnyCard(hand *model.Hand) bool {
	return hand.Count > 0
}

// GetAvailableSingles 返回所有可单出的牌（非猫牌、非拆解、非爆炸猫）
func GetAvailableSingles(hand *model.Hand) []model.Card {
	var singles []model.Card
	for _, c := range hand.Cards {
		switch c.Type {
		case model.CardTypeSkip, model.CardTypeAttack, model.CardTypeFavor,
			model.CardTypeShuffle, model.CardTypeSeeTheFuture, model.CardTypeNope:
			singles = append(singles, c)
		}
	}
	return singles
}

// FindValidCombos 查找手牌中所有有效的组合
func FindValidCombos(hand *model.Hand) []ComboInfo {
	var combos []ComboInfo

	// 单牌
	for _, c := range hand.Cards {
		switch c.Type {
		case model.CardTypeSkip, model.CardTypeAttack, model.CardTypeFavor,
			model.CardTypeShuffle, model.CardTypeSeeTheFuture, model.CardTypeNope:
			combos = append(combos, ComboInfo{
				Type:    ComboSingleCard,
				CardIDs: []string{c.ID},
			})
		}
	}

	// 三张相同猫牌
	catCards := hand.FindCardsByType(model.CardTypeCatCard)
	nameGroups := make(map[string][]string)
	for _, c := range catCards {
		nameGroups[c.Name] = append(nameGroups[c.Name], c.ID)
	}
	for _, ids := range nameGroups {
		if len(ids) >= 3 {
			combos = append(combos, ComboInfo{
				Type:    ComboThreeSameCat,
				CardIDs: ids[:3],
			})
		}
	}

	// 五张不同猫牌
	if len(catCards) >= 5 {
		seen := make(map[string]bool)
		var fiveIDs []string
		for _, c := range catCards {
			if !seen[c.Name] {
				seen[c.Name] = true
				fiveIDs = append(fiveIDs, c.ID)
			}
			if len(fiveIDs) == 5 {
				break
			}
		}
		if len(fiveIDs) == 5 {
			combos = append(combos, ComboInfo{
				Type:    ComboFiveDifferentCat,
				CardIDs: fiveIDs,
			})
		}
	}

	return combos
}

// SortHandByType 按类型对手牌排序（便于展示）
func SortHandByType(hand *model.Hand) {
	sort.Slice(hand.Cards, func(i, j int) bool {
		if hand.Cards[i].Type != hand.Cards[j].Type {
			return hand.Cards[i].Type < hand.Cards[j].Type
		}
		return hand.Cards[i].Name < hand.Cards[j].Name
	})
}