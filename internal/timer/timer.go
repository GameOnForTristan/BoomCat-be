package timer

import "time"

// TimerType 定时器类型
type TimerType int

const (
	TimerTypeTurn     TimerType = iota // 回合超时
	TimerTypeNopeWindow                // Nope 等待窗口超时
	TimerTypeReconnect                 // 重连超时
)

// TimerState 定时器状态
type TimerState int

const (
	TimerStateStopped TimerState = iota
	TimerStateRunning
	TimerStateExpired
)

// GameTimer 游戏定时器
type GameTimer struct {
	Type      TimerType  `json:"type"`
	Duration  int        `json:"duration"`   // 超时时间（秒）
	Remaining int        `json:"remaining"`  // 剩余时间（秒）
	State     TimerState `json:"state"`
	StartedAt int64      `json:"started_at"` // 启动时间戳（毫秒）
	PlayerID  string     `json:"player_id,omitempty"` // 关联的玩家 ID
}

// NewTurnTimer 创建回合定时器
func NewTurnTimer(timeoutSec int, playerID string) *GameTimer {
	return &GameTimer{
		Type:      TimerTypeTurn,
		Duration:  timeoutSec,
		Remaining: timeoutSec,
		State:     TimerStateStopped,
		PlayerID:  playerID,
	}
}

// NewNopeTimer 创建 Nope 窗口定时器
func NewNopeTimer(windowSec int) *GameTimer {
	return &GameTimer{
		Type:      TimerTypeNopeWindow,
		Duration:  windowSec,
		Remaining: windowSec,
		State:     TimerStateStopped,
	}
}

// NewReconnectTimer 创建重连定时器
func NewReconnectTimer(timeoutSec int, playerID string) *GameTimer {
	return &GameTimer{
		Type:      TimerTypeReconnect,
		Duration:  timeoutSec,
		Remaining: timeoutSec,
		State:     TimerStateStopped,
		PlayerID:  playerID,
	}
}

// Start 启动定时器
func (gt *GameTimer) Start() {
	gt.State = TimerStateRunning
	gt.StartedAt = time.Now().UnixMilli()
	gt.Remaining = gt.Duration
}

// Stop 停止定时器
func (gt *GameTimer) Stop() {
	gt.State = TimerStateStopped
	gt.Remaining = gt.Duration
}

// Tick 定时器滴答，减少剩余时间
// 返回 true 表示定时器已到期
func (gt *GameTimer) Tick(seconds int) bool {
	if gt.State != TimerStateRunning {
		return false
	}
	gt.Remaining -= seconds
	if gt.Remaining <= 0 {
		gt.Remaining = 0
		gt.State = TimerStateExpired
		return true
	}
	return false
}

// IsExpired 检查定时器是否已到期
func (gt *GameTimer) IsExpired() bool {
	return gt.State == TimerStateExpired
}

// IsRunning 检查定时器是否正在运行
func (gt *GameTimer) IsRunning() bool {
	return gt.State == TimerStateRunning
}

// Reset 重置定时器
func (gt *GameTimer) Reset() {
	gt.State = TimerStateStopped
	gt.Remaining = gt.Duration
	gt.StartedAt = 0
}

// Elapsed 返回已过时间（秒）
func (gt *GameTimer) Elapsed() int {
	if gt.StartedAt == 0 {
		return 0
	}
	return int((time.Now().UnixMilli() - gt.StartedAt) / 1000)
}

// TimerManager 定时器管理器，管理一个房间内的所有定时器
type TimerManager struct {
	TurnTimer     *GameTimer `json:"turn_timer"`
	NopeTimer     *GameTimer `json:"nope_timer"`
	ReconnectTimers map[string]*GameTimer `json:"reconnect_timers"`
}

// NewTimerManager 创建定时器管理器
func NewTimerManager(turnTimeout, nopeWindow, reconnectTimeout int) *TimerManager {
	return &TimerManager{
		ReconnectTimers: make(map[string]*GameTimer),
		TurnTimer:       NewTurnTimer(turnTimeout, ""),
		NopeTimer:       NewNopeTimer(nopeWindow),
	}
}

// StartTurnTimer 启动回合定时器
func (tm *TimerManager) StartTurnTimer(playerID string) {
	tm.TurnTimer.PlayerID = playerID
	tm.TurnTimer.Start()
}

// StopTurnTimer 停止回合定时器
func (tm *TimerManager) StopTurnTimer() {
	tm.TurnTimer.Stop()
}

// StartNopeTimer 启动 Nope 窗口定时器
func (tm *TimerManager) StartNopeTimer() {
	tm.NopeTimer.Start()
}

// StopNopeTimer 停止 Nope 窗口定时器
func (tm *TimerManager) StopNopeTimer() {
	tm.NopeTimer.Stop()
}

// StartReconnectTimer 启动重连定时器
func (tm *TimerManager) StartReconnectTimer(playerID string) {
	if _, ok := tm.ReconnectTimers[playerID]; !ok {
		tm.ReconnectTimers[playerID] = NewReconnectTimer(tm.TurnTimer.Duration, playerID)
	}
	tm.ReconnectTimers[playerID].Start()
}

// StopReconnectTimer 停止重连定时器
func (tm *TimerManager) StopReconnectTimer(playerID string) {
	if timer, ok := tm.ReconnectTimers[playerID]; ok {
		timer.Stop()
	}
}

// TickAll 所有定时器同步推进
func (tm *TimerManager) TickAll(seconds int) *TimerType {
	// 检查回合定时器
	if tm.TurnTimer.Tick(seconds) {
		t := TimerTypeTurn
		return &t
	}

	// 检查 Nope 窗口定时器
	if tm.NopeTimer.Tick(seconds) {
		t := TimerTypeNopeWindow
		return &t
	}

	// 检查重连定时器
	for pid, timer := range tm.ReconnectTimers {
		if timer.Tick(seconds) {
			t := TimerTypeReconnect
			timer.PlayerID = pid
			return &t
		}
	}

	return nil
}

// Reset 重置所有定时器
func (tm *TimerManager) Reset() {
	tm.TurnTimer.Reset()
	tm.NopeTimer.Reset()
	tm.ReconnectTimers = make(map[string]*GameTimer)
}