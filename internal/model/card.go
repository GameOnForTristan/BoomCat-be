package model

// CardType 卡牌类型枚举
type CardType string

const (
	CardTypeExplodingKitten CardType = "exploding_kitten"
	CardTypeDefuse          CardType = "defuse"
	CardTypeSkip            CardType = "skip"
	CardTypeAttack          CardType = "attack"
	CardTypeFavor           CardType = "favor"
	CardTypeShuffle         CardType = "shuffle"
	CardTypeSeeTheFuture    CardType = "see_the_future"
	CardTypeNope            CardType = "nope"
	CardTypeCatCard         CardType = "cat_card"
)

// Card 卡牌值对象，不可变
type Card struct {
	ID          string
	Type        CardType
	Name        string
	Description string
}

// cardMeta 卡牌类型元数据
type cardMeta struct {
	Name        string
	Description string
}

var cardTypeMeta = map[CardType]cardMeta{
	CardTypeExplodingKitten: {"爆炸猫", "抽到即出局（除非有拆解）"},
	CardTypeDefuse:          {"拆解", "抵消爆炸猫，可插回牌库任意位置"},
	CardTypeSkip:            {"跳过", "跳过本回合抽牌阶段"},
	CardTypeAttack:          {"攻击", "下家抽 2 张牌，且下家不能出牌"},
	CardTypeFavor:           {"索要", "指定玩家必须给一张牌"},
	CardTypeShuffle:         {"洗牌", "随机洗乱牌库"},
	CardTypeSeeTheFuture:    {"预言", "偷看牌库顶 3 张牌"},
	CardTypeNope:            {"否决", "取消任意一张牌的效果（非拆解/爆炸猫）"},
	CardTypeCatCard:         {"猫牌", "三张相同换一张指定牌；五张不同换一张弃牌堆牌"},
}

// NewCard 创建一张新卡牌，根据类型自动填充 Name 和 Description
func NewCard(cardType CardType, id string) Card {
	meta := cardTypeMeta[cardType]
	return Card{
		ID:          id,
		Type:        cardType,
		Name:        meta.Name,
		Description: meta.Description,
	}
}

// GetCardTypeName 获取卡牌类型的中文展示名称
func GetCardTypeName(cardType CardType) string {
	if meta, ok := cardTypeMeta[cardType]; ok {
		return meta.Name
	}
	return string(cardType)
}

// IsCatCard 判断是否为猫牌类型
func IsCatCard(cardType CardType) bool {
	return cardType == CardTypeCatCard
}

// IsStackable 判断该类型牌是否能进入动作栈
func IsStackable(cardType CardType) bool {
	switch cardType {
	case CardTypeNope, CardTypeCatCard, CardTypeFavor, CardTypeSkip, CardTypeAttack, CardTypeShuffle, CardTypeSeeTheFuture:
		return true
	default:
		return false
	}
}