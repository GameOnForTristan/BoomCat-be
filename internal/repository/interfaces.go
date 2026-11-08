package repository

import (
	"context"
	"errors"
	"time"

	"boom-cat-be/internal/model"
)

// 仓储层共享错误
var (
	// ErrDuplicateNickname 用户昵称冲突（PostgreSQL 唯一约束）
	ErrDuplicateNickname = errors.New("duplicate nickname")
	// ErrRecordNotFound 记录不存在
	ErrRecordNotFound = errors.New("record not found")
)

// =====================================================================
// 仓储接口定义
// Service 层依赖这些接口，不依赖具体 Redis/PostgreSQL 实现。
// =====================================================================

// GameStateRepository 游戏状态快照仓储（Redis）
type GameStateRepository interface {
	// SaveGameState 异步保存游戏状态快照（24h TTL）
	SaveGameState(ctx context.Context, roomID string, state *model.GameState) error
	// LoadGameState 读取游戏状态快照，未命中返回 nil, nil
	LoadGameState(ctx context.Context, roomID string) (*model.GameState, error)
	// DeleteGameState 删除游戏状态快照（房间解散时调用）
	DeleteGameState(ctx context.Context, roomID string) error
}

// RoomMetaRepository 房间元数据仓储（Redis Hash）
type RoomMetaRepository interface {
	// SaveRoomMeta 写入房间元数据（24h TTL）
	SaveRoomMeta(ctx context.Context, roomID string, meta RoomMeta) error
	// LoadRoomMeta 读取房间元数据
	LoadRoomMeta(ctx context.Context, roomID string) (RoomMeta, error)
	// DeleteRoomMeta 删除房间元数据
	DeleteRoomMeta(ctx context.Context, roomID string) error
	// UpdateRoomStatus 单独更新房间状态字段
	UpdateRoomStatus(ctx context.Context, roomID string, status string) error
	// UpdateRoomPlayerCount 单独更新当前在线玩家数
	UpdateRoomPlayerCount(ctx context.Context, roomID string, count int) error
}

// InviteCodeRepository 邀请码 -> 房间 ID 映射仓储
type InviteCodeRepository interface {
	// SaveInviteCode 写入邀请码映射（24h TTL）
	SaveInviteCode(ctx context.Context, code, roomID string) error
	// LookupRoomID 通过邀请码查询房间 ID，未命中返回 ""
	LookupRoomID(ctx context.Context, code string) (string, error)
	// DeleteInviteCode 删除邀请码映射
	DeleteInviteCode(ctx context.Context, code string) error
}

// RoomPlayersRepository 房间在线玩家集合仓储（Redis Set）
type RoomPlayersRepository interface {
	// AddPlayer 玩家加入房间
	AddPlayer(ctx context.Context, roomID, playerID string) error
	// RemovePlayer 玩家退出房间
	RemovePlayer(ctx context.Context, roomID, playerID string) error
	// ListPlayers 列出房间所有在线玩家 ID
	ListPlayers(ctx context.Context, roomID string) ([]string, error)
	// IsMember 判断玩家是否在房间中
	IsMember(ctx context.Context, roomID, playerID string) (bool, error)
}

// SessionRepository 玩家会话缓存仓储（Redis String，配合 JWT）
type SessionRepository interface {
	// SaveSession 写入会话缓存（与 JWT 过期时间对齐的 TTL）
	SaveSession(ctx context.Context, token string, data SessionData, ttl time.Duration) error
	// LoadSession 读取会话缓存
	LoadSession(ctx context.Context, token string) (*SessionData, error)
	// DeleteSession 删除会话缓存（登出或刷新时）
	DeleteSession(ctx context.Context, token string) error
}

// UserRepository PostgreSQL 用户信息仓储
type UserRepository interface {
	// CreateUser 创建用户
	CreateUser(ctx context.Context, user *User) error
	// GetUserByID 按 ID 查询用户
	GetUserByID(ctx context.Context, id string) (*User, error)
	// GetUserByNickname 按昵称查询用户（登录用）
	GetUserByNickname(ctx context.Context, nickname string) (*User, error)
	// UpdateUser 更新用户信息（昵称/密码哈希）
	UpdateUser(ctx context.Context, user *User) error
}

// GameRecordRepository PostgreSQL 游戏记录仓储
type GameRecordRepository interface {
	// SaveGameRecord 写入一局游戏记录
	SaveGameRecord(ctx context.Context, record *GameRecord) error
	// GetGameRecord 查询单条游戏记录
	GetGameRecord(ctx context.Context, id string) (*GameRecord, error)
	// ListGameRecordsByPlayer 查询某玩家的历史对局（分页）
	ListGameRecordsByPlayer(ctx context.Context, playerID string, limit, offset int) ([]*GameRecord, error)
}

// PlayerStatsRepository PostgreSQL 玩家战绩仓储
type PlayerStatsRepository interface {
	// GetPlayerStats 查询玩家战绩
	GetPlayerStats(ctx context.Context, playerID string) (*PlayerStats, error)
	// UpsertPlayerStats 初始化或更新玩家战绩
	UpsertPlayerStats(ctx context.Context, playerID string, delta PlayerStatsDelta) error
	// ListTopPlayers 按胜率倒序获取排行榜
	ListTopPlayers(ctx context.Context, limit int) ([]*PlayerStats, error)
}

// TokenRepository PostgreSQL 用户刷新令牌仓储
type TokenRepository interface {
	// SaveRefreshToken 写入刷新令牌
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	// GetRefreshToken 查询刷新令牌
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	// DeleteRefreshToken 删除刷新令牌
	DeleteRefreshToken(ctx context.Context, token string) error
	// DeleteTokensByUser 删除某用户所有刷新令牌（登出全部设备）
	DeleteTokensByUser(ctx context.Context, userID string) error
}

// =====================================================================
// 仓储层共享的数据传输对象
// =====================================================================

// RoomMeta 房间元数据（对应 Redis Hash 字段）
type RoomMeta struct {
	Name        string `redis:"name"`
	OwnerID     string `redis:"owner_id"`
	Status      string `redis:"status"`
	MaxPlayers  int    `redis:"max_players"`
	PlayerCount int    `redis:"player_count"`
	CreatedAt   int64  `redis:"created_at"`
}

// SessionData 会话缓存内容
type SessionData struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

// User PostgreSQL users 表对应实体
type User struct {
	ID        string    `db:"id"`
	Nickname  string    `db:"nickname"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GameRecord PostgreSQL boomcat_game_records 表对应实体
type GameRecord struct {
	ID         string    `db:"id"`
	RoomID     string    `db:"room_id"`
	Players    JSONB     `db:"players"`
	ActionLog  JSONB     `db:"action_log"`
	Result     JSONB     `db:"result"`
	CreatedAt  time.Time `db:"created_at"`
	FinishedAt time.Time `db:"finished_at"`
}

// JSONB 包装 []byte，便于 pgx 写入 JSONB 字段
type JSONB []byte

// PlayerStats PostgreSQL player_stats 表对应实体
type PlayerStats struct {
	PlayerID    string    `db:"player_id"`
	TotalGames  int       `db:"total_games"`
	Wins        int       `db:"wins"`
	Losses      int       `db:"losses"`
	WinRate     float64   `db:"win_rate"`
	TotalKills  int       `db:"total_kills"`
	TotalDraws  int       `db:"total_draws"`
	TotalPlayed int       `db:"total_played"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// PlayerStatsDelta 战绩增量
type PlayerStatsDelta struct {
	TotalGames  int
	Wins        int
	Losses      int
	TotalKills  int
	TotalDraws  int
	TotalPlayed int
}

// RefreshToken PostgreSQL user_tokens 表对应实体
type RefreshToken struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Token     string    `db:"refresh_token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
