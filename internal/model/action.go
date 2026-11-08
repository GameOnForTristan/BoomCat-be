package model

// ActionType 动作类型枚举
type ActionType int

const (
	ActionPlayCard     ActionType = iota // 出牌
	ActionNope                           // 出否决
	ActionDrawCard                       // 抽牌
	ActionEndTurn                        // 结束回合
	ActionSelectTarget                   // 选择目标
)

// Action 动作值对象
type Action struct {
	Type            ActionType `json:"type"`
	CardIDs         []string   `json:"card_ids,omitempty"`
	TargetPlayerID  string     `json:"target_player_id,omitempty"`
	TargetCardType  CardType   `json:"target_card_type,omitempty"`
}

// StackEntry 动作栈条目
type StackEntry struct {
	Action    Action `json:"action"`
	ActorID   string `json:"actor_id"`
	Timestamp int64  `json:"timestamp"`
}

// ResolveResult 动作栈结算结果
type ResolveResult struct {
	Action        Action `json:"action"`
	ActionBlocked bool   `json:"action_blocked"`
	NopeCount     int    `json:"nope_count"`
}

// ActionStack 动作栈（Nope 链核心）
type ActionStack struct {
	Stack []StackEntry `json:"stack"`
}

// NewActionStack 创建新动作栈
func NewActionStack() *ActionStack {
	return &ActionStack{
		Stack: make([]StackEntry, 0),
	}
}

// Push 向栈顶压入一个条目
func (s *ActionStack) Push(entry StackEntry) {
	s.Stack = append(s.Stack, entry)
}

// Pop 从栈顶弹出一个条目
func (s *ActionStack) Pop() (StackEntry, bool) {
	if len(s.Stack) == 0 {
		return StackEntry{}, false
	}
	entry := s.Stack[len(s.Stack)-1]
	s.Stack = s.Stack[:len(s.Stack)-1]
	return entry, true
}

// Top 查看栈顶条目（不移除）
func (s *ActionStack) Top() (StackEntry, bool) {
	if len(s.Stack) == 0 {
		return StackEntry{}, false
	}
	return s.Stack[len(s.Stack)-1], true
}

// TopIsNope 判断栈顶是否为 Nope 动作
func (s *ActionStack) TopIsNope() bool {
	entry, ok := s.Top()
	if !ok {
		return false
	}
	return entry.Action.Type == ActionNope
}

// Resolve 结算动作栈（奇偶抵消）
// 从栈底向上遍历，Nope 每出现一次翻转一次生效状态
func (s *ActionStack) Resolve() ResolveResult {
	if len(s.Stack) == 0 {
		return ResolveResult{}
	}

	baseAction := s.Stack[0].Action
	nopeCount := 0
	for i := 1; i < len(s.Stack); i++ {
		if s.Stack[i].Action.Type == ActionNope {
			nopeCount++
		}
	}

	// 清空栈
	s.Stack = make([]StackEntry, 0)

	return ResolveResult{
		Action:        baseAction,
		ActionBlocked: nopeCount%2 == 1,
		NopeCount:     nopeCount,
	}
}

// Len 返回栈中条目数
func (s *ActionStack) Len() int {
	return len(s.Stack)
}

// Clear 清空动作栈
func (s *ActionStack) Clear() {
	s.Stack = make([]StackEntry, 0)
}