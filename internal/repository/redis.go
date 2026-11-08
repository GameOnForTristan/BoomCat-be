package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"boom-cat-be/internal/model"

	"github.com/redis/go-redis/v9"
)

// =====================================================================
// Redis 仓储实现
// 对应 redis/001_init.redis 中的 Key 设计：
//   - game_state:room:{room_id}    String, 86400s, 游戏状态快照
//   - room:{room_id}               Hash,   86400s, 房间元数据
//   - invite_code:{code}           String, 86400s, 邀请码映射
//   - room_players:{room_id}       Set,    86400s, 房间在线玩家
//   - session:{jwt_token}          String, 259200s, 会话缓存
// =====================================================================

const (
	redisTTLGameState  = 24 * time.Hour
	redisTTLRoom       = 24 * time.Hour
	redisTTLInviteCode = 24 * time.Hour
	redisTTLRoomPlayer = 24 * time.Hour
)

// =====================================================================
// RedisGameStateRepository 游戏状态快照仓储
// =====================================================================

// RedisGameStateRepository 基于 go-redis 的 GameStateRepository 实现
type RedisGameStateRepository struct {
	client *redis.Client
}

// NewRedisGameStateRepository 创建 Redis 游戏状态仓储
func NewRedisGameStateRepository(client *redis.Client) *RedisGameStateRepository {
	return &RedisGameStateRepository{client: client}
}

// SaveGameState 序列化整个 GameState 为 JSON 写入 Redis
func (r *RedisGameStateRepository) SaveGameState(ctx context.Context, roomID string, state *model.GameState) error {
	if state == nil {
		return errors.New("game state is nil")
	}
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal game state: %w", err)
	}
	key := fmt.Sprintf("game_state:room:%s", roomID)
	if err := r.client.Set(ctx, key, data, redisTTLGameState).Err(); err != nil {
		return fmt.Errorf("redis SET %s: %w", key, err)
	}
	return nil
}

// LoadGameState 从 Redis 读取并反序列化游戏状态
func (r *RedisGameStateRepository) LoadGameState(ctx context.Context, roomID string) (*model.GameState, error) {
	key := fmt.Sprintf("game_state:room:%s", roomID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis GET %s: %w", key, err)
	}
	var state model.GameState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("unmarshal game state: %w", err)
	}
	return &state, nil
}

// DeleteGameState 删除游戏状态快照
func (r *RedisGameStateRepository) DeleteGameState(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("game_state:room:%s", roomID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL %s: %w", key, err)
	}
	return nil
}

// =====================================================================
// RedisRoomMetaRepository 房间元数据仓储
// =====================================================================

// RedisRoomMetaRepository 基于 go-redis 的 RoomMetaRepository 实现
type RedisRoomMetaRepository struct {
	client *redis.Client
}

// NewRedisRoomMetaRepository 创建 Redis 房间元数据仓储
func NewRedisRoomMetaRepository(client *redis.Client) *RedisRoomMetaRepository {
	return &RedisRoomMetaRepository{client: client}
}

// SaveRoomMeta 使用 HSET 写入房间元数据 Hash
func (r *RedisRoomMetaRepository) SaveRoomMeta(ctx context.Context, roomID string, meta RoomMeta) error {
	key := fmt.Sprintf("room:%s", roomID)
	fields := map[string]interface{}{
		"name":          meta.Name,
		"owner_id":      meta.OwnerID,
		"status":        meta.Status,
		"max_players":   meta.MaxPlayers,
		"player_count":  meta.PlayerCount,
		"created_at":    meta.CreatedAt,
	}
	if err := r.client.HSet(ctx, key, fields).Err(); err != nil {
		return fmt.Errorf("redis HSET %s: %w", key, err)
	}
	if err := r.client.Expire(ctx, key, redisTTLRoom).Err(); err != nil {
		return fmt.Errorf("redis EXPIRE %s: %w", key, err)
	}
	return nil
}

// LoadRoomMeta 读取房间元数据 Hash
func (r *RedisRoomMetaRepository) LoadRoomMeta(ctx context.Context, roomID string) (RoomMeta, error) {
	var meta RoomMeta
	key := fmt.Sprintf("room:%s", roomID)
	res, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return meta, fmt.Errorf("redis HGETALL %s: %w", key, err)
	}
	if len(res) == 0 {
		return meta, nil
	}
	meta.Name = res["name"]
	meta.OwnerID = res["owner_id"]
	meta.Status = res["status"]
	meta.CreatedAt = parseInt64(res["created_at"])
	meta.MaxPlayers = parseInt(res["max_players"])
	meta.PlayerCount = parseInt(res["player_count"])
	return meta, nil
}

// DeleteRoomMeta 删除房间元数据
func (r *RedisRoomMetaRepository) DeleteRoomMeta(ctx context.Context, roomID string) error {
	key := fmt.Sprintf("room:%s", roomID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL %s: %w", key, err)
	}
	return nil
}

// UpdateRoomStatus 单独更新房间状态字段
func (r *RedisRoomMetaRepository) UpdateRoomStatus(ctx context.Context, roomID string, status string) error {
	key := fmt.Sprintf("room:%s", roomID)
	if err := r.client.HSet(ctx, key, "status", status).Err(); err != nil {
		return fmt.Errorf("redis HSET %s status: %w", key, err)
	}
	return nil
}

// UpdateRoomPlayerCount 单独更新当前在线玩家数
func (r *RedisRoomMetaRepository) UpdateRoomPlayerCount(ctx context.Context, roomID string, count int) error {
	key := fmt.Sprintf("room:%s", roomID)
	if err := r.client.HSet(ctx, key, "player_count", count).Err(); err != nil {
		return fmt.Errorf("redis HSET %s player_count: %w", key, err)
	}
	return nil
}

// =====================================================================
// RedisInviteCodeRepository 邀请码 -> 房间 ID 映射仓储
// =====================================================================

// RedisInviteCodeRepository 基于 go-redis 的 InviteCodeRepository 实现
type RedisInviteCodeRepository struct {
	client *redis.Client
}

// NewRedisInviteCodeRepository 创建 Redis 邀请码仓储
func NewRedisInviteCodeRepository(client *redis.Client) *RedisInviteCodeRepository {
	return &RedisInviteCodeRepository{client: client}
}

// SaveInviteCode 写入邀请码映射
func (r *RedisInviteCodeRepository) SaveInviteCode(ctx context.Context, code, roomID string) error {
	key := fmt.Sprintf("invite_code:%s", code)
	if err := r.client.Set(ctx, key, roomID, redisTTLInviteCode).Err(); err != nil {
		return fmt.Errorf("redis SET %s: %w", key, err)
	}
	return nil
}

// LookupRoomID 通过邀请码查询房间 ID，未命中返回 ""
func (r *RedisInviteCodeRepository) LookupRoomID(ctx context.Context, code string) (string, error) {
	key := fmt.Sprintf("invite_code:%s", code)
	roomID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", nil
		}
		return "", fmt.Errorf("redis GET %s: %w", key, err)
	}
	return roomID, nil
}

// DeleteInviteCode 删除邀请码映射
func (r *RedisInviteCodeRepository) DeleteInviteCode(ctx context.Context, code string) error {
	key := fmt.Sprintf("invite_code:%s", code)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL %s: %w", key, err)
	}
	return nil
}

// =====================================================================
// RedisRoomPlayersRepository 房间在线玩家集合仓储
// =====================================================================

// RedisRoomPlayersRepository 基于 go-redis 的 RoomPlayersRepository 实现
type RedisRoomPlayersRepository struct {
	client *redis.Client
}

// NewRedisRoomPlayersRepository 创建 Redis 房间玩家集合仓储
func NewRedisRoomPlayersRepository(client *redis.Client) *RedisRoomPlayersRepository {
	return &RedisRoomPlayersRepository{client: client}
}

// AddPlayer 玩家加入房间
func (r *RedisRoomPlayersRepository) AddPlayer(ctx context.Context, roomID, playerID string) error {
	key := fmt.Sprintf("room_players:%s", roomID)
	if err := r.client.SAdd(ctx, key, playerID).Err(); err != nil {
		return fmt.Errorf("redis SADD %s: %w", key, err)
	}
	return r.client.Expire(ctx, key, redisTTLRoomPlayer).Err()
}

// RemovePlayer 玩家退出房间
func (r *RedisRoomPlayersRepository) RemovePlayer(ctx context.Context, roomID, playerID string) error {
	key := fmt.Sprintf("room_players:%s", roomID)
	if err := r.client.SRem(ctx, key, playerID).Err(); err != nil {
		return fmt.Errorf("redis SREM %s: %w", key, err)
	}
	return nil
}

// ListPlayers 列出房间所有在线玩家 ID
func (r *RedisRoomPlayersRepository) ListPlayers(ctx context.Context, roomID string) ([]string, error) {
	key := fmt.Sprintf("room_players:%s", roomID)
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("redis SMEMBERS %s: %w", key, err)
	}
	return members, nil
}

// IsMember 判断玩家是否在房间中
func (r *RedisRoomPlayersRepository) IsMember(ctx context.Context, roomID, playerID string) (bool, error) {
	key := fmt.Sprintf("room_players:%s", roomID)
	ok, err := r.client.SIsMember(ctx, key, playerID).Result()
	if err != nil {
		return false, fmt.Errorf("redis SISMEMBER %s: %w", key, err)
	}
	return ok, nil
}

// =====================================================================
// RedisSessionRepository 玩家会话缓存仓储
// =====================================================================

// RedisSessionRepository 基于 go-redis 的 SessionRepository 实现
type RedisSessionRepository struct {
	client *redis.Client
}

// NewRedisSessionRepository 创建 Redis 会话仓储
func NewRedisSessionRepository(client *redis.Client) *RedisSessionRepository {
	return &RedisSessionRepository{client: client}
}

// SaveSession 写入会话缓存
func (r *RedisSessionRepository) SaveSession(ctx context.Context, token string, data SessionData, ttl time.Duration) error {
	if token == "" {
		return errors.New("empty session token")
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	key := fmt.Sprintf("session:%s", token)
	if err := r.client.Set(ctx, key, payload, ttl).Err(); err != nil {
		return fmt.Errorf("redis SET %s: %w", key, err)
	}
	return nil
}

// LoadSession 读取会话缓存
func (r *RedisSessionRepository) LoadSession(ctx context.Context, token string) (*SessionData, error) {
	key := fmt.Sprintf("session:%s", token)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("redis GET %s: %w", key, err)
	}
	var s SessionData
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &s, nil
}

// DeleteSession 删除会话缓存
func (r *RedisSessionRepository) DeleteSession(ctx context.Context, token string) error {
	key := fmt.Sprintf("session:%s", token)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis DEL %s: %w", key, err)
	}
	return nil
}

// =====================================================================
// 辅助函数
// =====================================================================

func parseInt(s string) int {
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0
	}
	return n
}

func parseInt64(s string) int64 {
	var n int64
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil {
		return 0
	}
	return n
}
