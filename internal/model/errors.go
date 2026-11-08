package model

import "errors"

// 游戏领域错误定义
var (
	ErrInvalidMove       = errors.New("invalid move")
	ErrNotYourTurn       = errors.New("not your turn")
	ErrCardNotFound      = errors.New("card not found")
	ErrInvalidCombo      = errors.New("invalid card combination")
	ErrInvalidPhase      = errors.New("invalid game phase for this operation")
	ErrInvalidTarget     = errors.New("invalid target player")
	ErrPlayerNotFound    = errors.New("player not found")
	ErrPlayerEliminated  = errors.New("player is already eliminated")
	ErrHandEmpty         = errors.New("hand is empty")
	ErrDeckEmpty         = errors.New("deck is empty")
	ErrRoomFull          = errors.New("room is full")
	ErrRoomAlreadyStarted = errors.New("game already started")
	ErrNotAllReady       = errors.New("not all players are ready")
	ErrNotOwner          = errors.New("only the room owner can perform this action")
	ErrDuplicateCard     = errors.New("duplicate card ID")
	ErrNopeNotAllowed    = errors.New("nope not allowed at this time")
	ErrActionTimeout     = errors.New("action timed out")
	ErrInvalidAction     = errors.New("invalid action type")
)

// DomainError 带错误码的领域错误
type DomainError struct {
	Code    int
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// NewDomainError 创建领域错误
func NewDomainError(code int, message string, err error) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}