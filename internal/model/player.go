package model

// ConnectionStatus 连接状态枚举
type ConnectionStatus int

const (
	ConnectionStatusConnected    ConnectionStatus = iota // 已连接
	ConnectionStatusDisconnected                         // 断线中
	ConnectionStatusTimedOut                             // 已超时
)

// Player 玩家实体
type Player struct {
	ID               string           `json:"id"`
	Nickname         string           `json:"nickname"`
	Hand             Hand             `json:"hand"`
	Alive            bool             `json:"alive"`
	DefuseCount      int              `json:"defuse_count"`
	ConnectionStatus ConnectionStatus `json:"connection_status"`
	IsReady          bool             `json:"is_ready"`
	JoinedAt         int64            `json:"joined_at"`
}

// NewPlayer 创建新玩家
func NewPlayer(id, nickname string) *Player {
	return &Player{
		ID:               id,
		Nickname:         nickname,
		Hand:             *NewHand(id),
		Alive:            true,
		DefuseCount:      0,
		ConnectionStatus: ConnectionStatusConnected,
		IsReady:          false,
	}
}

// PlayerInfo 玩家公开信息（对局中其他玩家可见）
type PlayerInfo struct {
	ID          string `json:"id"`
	Nickname    string `json:"nickname"`
	HandSize    int    `json:"hand_size"`
	Alive       bool   `json:"alive"`
	DefuseCount int    `json:"defuse_count"`
}

// ToPlayerInfo 从 Player 转换为 PlayerInfo（过滤敏感信息）
func (p *Player) ToPlayerInfo() PlayerInfo {
	return PlayerInfo{
		ID:          p.ID,
		Nickname:    p.Nickname,
		HandSize:    p.Hand.Count,
		Alive:       p.Alive,
		DefuseCount: p.DefuseCount,
	}
}