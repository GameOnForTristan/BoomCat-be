package model

// GamePhase 游戏阶段枚举
type GamePhase int

const (
	GamePhaseWaiting  GamePhase = iota // 等待中
	GamePhaseDealing                   // 发牌中
	GamePhasePlaying                   // 游戏中
	GamePhaseGameOver                  // 已结束
)

// TurnState 回合状态实体
type TurnState struct {
	CurrentPlayerID string   `json:"current_player_id"`
	PlayerOrder     []string `json:"player_order"`
	Direction       int      `json:"direction"`
	PendingDraws    int      `json:"pending_draws"`
	Phase           TurnPhase `json:"phase"`
	ActionsTaken    int      `json:"actions_taken"`
}

// GameState 游戏状态实体（聚合根）
type GameState struct {
	Deck        Deck          `json:"deck"`
	DiscardPile DiscardPile   `json:"discard_pile"`
	Turn        TurnState     `json:"turn"`
	ActionStack ActionStack   `json:"action_stack"`
	Phase       GamePhase     `json:"phase"`
	Players     []PlayerInfo  `json:"players"`
	RoundCount  int           `json:"round_count"`
}

// DiscardPile 弃牌堆实体
type DiscardPile struct {
	Cards []Card `json:"cards"`
}

// NewDiscardPile 创建弃牌堆
func NewDiscardPile() *DiscardPile {
	return &DiscardPile{
		Cards: make([]Card, 0),
	}
}

// Add 向弃牌堆添加一张牌
func (dp *DiscardPile) Add(card Card) {
	dp.Cards = append(dp.Cards, card)
}

// TopCard 返回弃牌堆最上面一张牌（最近打出）
func (dp *DiscardPile) TopCard() (Card, bool) {
	if len(dp.Cards) == 0 {
		return Card{}, false
	}
	return dp.Cards[len(dp.Cards)-1], true
}

// Count 返回弃牌堆牌数
func (dp *DiscardPile) Count() int {
	return len(dp.Cards)
}

// NewGameState 创建初始游戏状态
func NewGameState() *GameState {
	return &GameState{
		Deck:        *NewDeck(make([]Card, 0)),
		DiscardPile: *NewDiscardPile(),
		ActionStack: *NewActionStack(),
		Phase:       GamePhaseWaiting,
		Players:     make([]PlayerInfo, 0),
		RoundCount:  0,
	}
}

// IsGameOver 判断游戏是否结束（只剩一人存活）
func (gs *GameState) IsGameOver() bool {
	alive := 0
	for _, p := range gs.Players {
		if p.Alive {
			alive++
		}
	}
	return alive <= 1
}

// Winner 返回获胜者 ID（游戏结束时调用）
func (gs *GameState) Winner() string {
	for _, p := range gs.Players {
		if p.Alive {
			return p.ID
		}
	}
	return ""
}