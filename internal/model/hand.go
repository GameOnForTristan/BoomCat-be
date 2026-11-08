package model

// Hand 手牌实体
type Hand struct {
	PlayerID string `json:"player_id"`
	Cards    []Card `json:"cards"`
	Count    int    `json:"count"`
}

// NewHand 创建新手牌
func NewHand(playerID string) *Hand {
	return &Hand{
		PlayerID: playerID,
		Cards:    make([]Card, 0),
		Count:    0,
	}
}

// AddCard 向手牌中添加一张牌
func (h *Hand) AddCard(card Card) {
	h.Cards = append(h.Cards, card)
	h.Count = len(h.Cards)
}

// RemoveCard 从手牌中移除指定 ID 的卡牌，返回是否成功
func (h *Hand) RemoveCard(cardID string) (Card, bool) {
	for i, c := range h.Cards {
		if c.ID == cardID {
			card := h.Cards[i]
			h.Cards = append(h.Cards[:i], h.Cards[i+1:]...)
			h.Count = len(h.Cards)
			return card, true
		}
	}
	return Card{}, false
}

// FindCard 查找手牌中指定 ID 的卡牌
func (h *Hand) FindCard(cardID string) (Card, bool) {
	for _, c := range h.Cards {
		if c.ID == cardID {
			return c, true
		}
	}
	return Card{}, false
}

// FindCardsByType 查找手牌中所有指定类型的卡牌
func (h *Hand) FindCardsByType(cardType CardType) []Card {
	var result []Card
	for _, c := range h.Cards {
		if c.Type == cardType {
			result = append(result, c)
		}
	}
	return result
}

// ValidateThreeSameCombo 校验三张相同猫牌组合
// 返回三张牌的 ID 和是否合法
func (h *Hand) ValidateThreeSameCombo(cardIDs []string) bool {
	if len(cardIDs) != 3 {
		return false
	}

	var cards []Card
	for _, id := range cardIDs {
		card, found := h.FindCard(id)
		if !found {
			return false
		}
		cards = append(cards, card)
	}

	// 必须都是猫牌
	for _, c := range cards {
		if !IsCatCard(c.Type) {
			return false
		}
	}

	// 三张名称必须相同
	return cards[0].Name == cards[1].Name && cards[1].Name == cards[2].Name
}

// ValidateFiveDifferentCombo 校验五张不同猫牌组合
// 返回是否合法
func (h *Hand) ValidateFiveDifferentCombo(cardIDs []string) bool {
	if len(cardIDs) != 5 {
		return false
	}

	var cards []Card
	seen := make(map[string]bool)
	for _, id := range cardIDs {
		card, found := h.FindCard(id)
		if !found {
			return false
		}
		if !IsCatCard(card.Type) {
			return false
		}
		// 五张必须互不相同
		if seen[card.Name] {
			return false
		}
		seen[card.Name] = true
		cards = append(cards, card)
	}

	return len(cards) == 5
}

// RemoveCards 批量移除手牌中的指定卡牌
func (h *Hand) RemoveCards(cardIDs []string) bool {
	idSet := make(map[string]bool)
	for _, id := range cardIDs {
		idSet[id] = true
	}

	var newCards []Card
	for _, c := range h.Cards {
		if !idSet[c.ID] {
			newCards = append(newCards, c)
		}
	}

	if len(newCards) != len(h.Cards)-len(cardIDs) {
		return false
	}

	h.Cards = newCards
	h.Count = len(h.Cards)
	return true
}