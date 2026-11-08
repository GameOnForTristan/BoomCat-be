package model

// RoomSettings 房间设置
type RoomSettings struct {
	MaxPlayers        int `json:"max_players"`
	TurnTimeout       int `json:"turn_timeout"`
	NopeWindow        int `json:"nope_window"`
	ReconnectTimeout  int `json:"reconnect_timeout"`
}

// DefaultRoomSettings 返回默认房间设置
func DefaultRoomSettings() RoomSettings {
	return RoomSettings{
		MaxPlayers:       5,
		TurnTimeout:      60,
		NopeWindow:       3,
		ReconnectTimeout: 90,
	}
}

// RoomStatus 房间状态
type RoomStatus string

const (
	RoomStatusWaiting  RoomStatus = "waiting"
	RoomStatusPlaying  RoomStatus = "playing"
	RoomStatusFinished RoomStatus = "finished"
)

// Room 房间实体
type Room struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	OwnerID   string       `json:"owner_id"`
	Players   []Player     `json:"players"`
	GameState *GameState   `json:"game_state,omitempty"`
	Status    RoomStatus   `json:"status"`
	Settings  RoomSettings `json:"settings"`
	CreatedAt int64        `json:"created_at"`
	Inbox     chan Message `json:"-"`
}

// Message 玩家消息通道消息
type Message struct {
	PlayerID string      `json:"player_id"`
	Type     string      `json:"type"`
	Payload  interface{} `json:"payload"`
	ReqID    string      `json:"req_id,omitempty"`
}

// NewRoom 创建新房间
func NewRoom(id, name, ownerID string, settings RoomSettings, createdAt int64) *Room {
	return &Room{
		ID:        id,
		Name:      name,
		OwnerID:   ownerID,
		Players:   make([]Player, 0),
		GameState: nil,
		Status:    RoomStatusWaiting,
		Settings:  settings,
		CreatedAt: createdAt,
		Inbox:     make(chan Message, 100),
	}
}

// PlayerCount 返回当前玩家数
func (r *Room) PlayerCount() int {
	return len(r.Players)
}

// FindPlayer 查找房间中的玩家
func (r *Room) FindPlayer(playerID string) (*Player, bool) {
	for i := range r.Players {
		if r.Players[i].ID == playerID {
			return &r.Players[i], true
		}
	}
	return nil, false
}

// IsOwner 判断玩家是否为房主
func (r *Room) IsOwner(playerID string) bool {
	return r.OwnerID == playerID
}

// AllReady 判断是否全员准备
func (r *Room) AllReady() bool {
	if len(r.Players) < 2 {
		return false
	}
	for _, p := range r.Players {
		if !p.IsReady {
			return false
		}
	}
	return true
}

// CanStart 判断是否可以开始游戏
func (r *Room) CanStart() bool {
	return r.Status == RoomStatusWaiting &&
		len(r.Players) >= 2 &&
		r.AllReady()
}