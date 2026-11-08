package action

import (
	"boom-cat-be/internal/model"
	"time"
)

// PushAction 向动作栈推送一个动作
func PushAction(stack *model.ActionStack, action model.Action, actorID string) {
	stack.Push(model.StackEntry{
		Action:    action,
		ActorID:   actorID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// PushNope 向动作栈推送一个 Nope 动作
func PushNope(stack *model.ActionStack, actorID string) {
	stack.Push(model.StackEntry{
		Action: model.Action{
			Type: model.ActionNope,
		},
		ActorID:   actorID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// Resolve 结算动作栈，返回结算结果
// 奇数次 Nope = 动作被阻止，偶数次 Nope = 动作生效
func Resolve(stack *model.ActionStack) model.ResolveResult {
	return stack.Resolve()
}

// CanNope 检查当前是否可以出 Nope
// 条件：栈非空，且栈顶不是 Nope（不能连续 Nope 同一动作）
func CanNope(stack *model.ActionStack) bool {
	if stack.Len() == 0 {
		return false
	}
	return !stack.TopIsNope()
}

// IsNopeChainActive 检查 Nope 链是否激活（栈中有待结算动作）
func IsNopeChainActive(stack *model.ActionStack) bool {
	return stack.Len() > 0
}

// PeekBaseAction 查看栈底的基础动作（不移除）
func PeekBaseAction(stack *model.ActionStack) (model.StackEntry, bool) {
	if stack.Len() == 0 {
		return model.StackEntry{}, false
	}
	return stack.Stack[0], true
}

// ClearStack 清空动作栈
func ClearStack(stack *model.ActionStack) {
	stack.Clear()
}

// StackSize 返回栈中条目数
func StackSize(stack *model.ActionStack) int {
	return stack.Len()
}
