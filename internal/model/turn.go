package model

// TurnPhase 回合阶段枚举
type TurnPhase int

const (
	PhaseAction    TurnPhase = iota // 出牌阶段
	PhaseDrawing                    // 抽牌阶段
	PhaseNopeWindow                 // Nope 等待窗口
	PhaseGameOver                   // 游戏结束
)

// TurnEngine 回合调度器
type TurnEngine struct {
	Players      []string  `json:"players"`        // 存活玩家，按座位顺序
	CurrentIdx   int       `json:"current_idx"`     // 当前玩家在 players 中的索引
	Direction    int       `json:"direction"`       // 1 = 顺时针, -1 = 逆时针
	PendingDraws int       `json:"pending_draws"`   // 攻击牌累积的待抽牌数
	Phase        TurnPhase `json:"phase"`
}

// NewTurnEngine 创建回合调度器
func NewTurnEngine(players []string) *TurnEngine {
	return &TurnEngine{
		Players:      players,
		CurrentIdx:   0,
		Direction:    1,
		PendingDraws: 0,
		Phase:        PhaseAction,
	}
}

// nextIndex 获取下一个存活玩家的索引（环形）
func (te *TurnEngine) nextIndex() int {
	next := te.CurrentIdx + te.Direction
	n := len(te.Players)
	if next >= n {
		return 0
	}
	if next < 0 {
		return n - 1
	}
	return next
}

// CurrentPlayer 返回当前玩家 ID
func (te *TurnEngine) CurrentPlayer() string {
	if te.CurrentIdx < 0 || te.CurrentIdx >= len(te.Players) {
		return ""
	}
	return te.Players[te.CurrentIdx]
}

// NextPlayer 返回下一个存活玩家的 ID
func (te *TurnEngine) NextPlayer() string {
	idx := te.nextIndex()
	if idx < 0 || idx >= len(te.Players) {
		return ""
	}
	return te.Players[idx]
}

// OnSkip 玩家打出 Skip
func (te *TurnEngine) OnSkip() {
	te.CurrentIdx = te.nextIndex()
	te.Phase = PhaseAction
}

// OnAttack 玩家打出 Attack
func (te *TurnEngine) OnAttack() {
	te.PendingDraws += 2
	te.CurrentIdx = te.nextIndex()
	te.Phase = PhaseDrawing
}

// OnCardDrawn 抽了一张牌（非爆炸猫）
func (te *TurnEngine) OnCardDrawn() {
	te.PendingDraws--
	if te.PendingDraws <= 0 {
		te.CurrentIdx = te.nextIndex()
		te.Phase = PhaseAction
	}
}

// OnExplodingKittenDrawn 抽到爆炸猫（有拆解，插回牌库）
func (te *TurnEngine) OnExplodingKittenDrawn() {
	te.PendingDraws--
	if te.PendingDraws <= 0 {
		te.CurrentIdx = te.nextIndex()
		te.Phase = PhaseAction
	}
}

// RemovePlayer 玩家出局，从回合轮转中移除
func (te *TurnEngine) RemovePlayer(playerID string) {
	idx := -1
	for i, p := range te.Players {
		if p == playerID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return
	}

	if idx < te.CurrentIdx {
		te.CurrentIdx--
	} else if idx == te.CurrentIdx {
		te.CurrentIdx = te.nextIndex()
	}

	te.Players = append(te.Players[:idx], te.Players[idx+1:]...)

	if len(te.Players) <= 1 {
		te.Phase = PhaseGameOver
	}
}

// ReverseDirection 反转回合方向
func (te *TurnEngine) ReverseDirection() {
	te.Direction = -te.Direction
}

// SetPhase 设置回合阶段
func (te *TurnEngine) SetPhase(phase TurnPhase) {
	te.Phase = phase
}

// PlayerCount 返回存活玩家数
func (te *TurnEngine) PlayerCount() int {
	return len(te.Players)
}