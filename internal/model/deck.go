package model

import "math/rand"

// Deck 牌库实体，表示有序的卡牌队列
type Deck struct {
	Cards []Card `json:"cards"`
	Count int    `json:"count"`
}

// NewDeck 创建牌库实例
func NewDeck(cards []Card) *Deck {
	return &Deck{
		Cards: cards,
		Count: len(cards),
	}
}

// Shuffle 随机洗乱牌库
func (d *Deck) Shuffle() {
	rand.Shuffle(len(d.Cards), func(i, j int) {
		d.Cards[i], d.Cards[j] = d.Cards[j], d.Cards[i]
	})
}

// Draw 从牌库顶（第一张）抽一张牌
func (d *Deck) Draw() (Card, bool) {
	if len(d.Cards) == 0 {
		return Card{}, false
	}
	card := d.Cards[0]
	d.Cards = d.Cards[1:]
	d.Count = len(d.Cards)
	return card, true
}

// InsertAt 将卡牌插入到牌库指定位置（0 = 牌顶）
func (d *Deck) InsertAt(card Card, index int) bool {
	if index < 0 || index > len(d.Cards) {
		return false
	}
	d.Cards = append(d.Cards[:index], append([]Card{card}, d.Cards[index:]...)...)
	d.Count = len(d.Cards)
	return true
}

// Peek 查看牌库顶 N 张牌（不移除）
func (d *Deck) Peek(n int) []Card {
	if n <= 0 {
		return nil
	}
	if n > len(d.Cards) {
		n = len(d.Cards)
	}
	result := make([]Card, n)
	copy(result, d.Cards[:n])
	return result
}