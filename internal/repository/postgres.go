package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// =====================================================================
// PostgreSQL 仓储实现
// 对应 sql/001_init.sql 中的表结构：
//   - users                    用户账号
//   - boomcat_game_records     游戏记录
//   - player_stats             玩家战绩
//   - user_tokens              刷新令牌
// =====================================================================

// pgErrUniqueViolation PostgreSQL 唯一约束冲突错误码
const pgErrUniqueViolation = "23505"

// =====================================================================
// PostgresUserRepository 用户信息仓储
// =====================================================================

// PostgresUserRepository 基于 pgx 的 UserRepository 实现
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresUserRepository 创建 PostgreSQL 用户仓储
func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// CreateUser 插入新用户，昵称冲突时返回 ErrDuplicateNickname
func (r *PostgresUserRepository) CreateUser(ctx context.Context, user *User) error {
	if user == nil {
		return errors.New("user is nil")
	}
	const sql = `
		INSERT INTO users (id, nickname, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}
	_, err := r.pool.Exec(ctx, sql,
		user.ID, user.Nickname, user.Password,
		user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateNickname
		}
		return fmt.Errorf("pg INSERT users: %w", err)
	}
	return nil
}

// GetUserByID 按 ID 查询用户
func (r *PostgresUserRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	const sql = `
		SELECT id, nickname, password, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u User
	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&u.ID, &u.Nickname, &u.Password, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pg SELECT users by id: %w", err)
	}
	return &u, nil
}

// GetUserByNickname 按昵称查询用户（登录用）
func (r *PostgresUserRepository) GetUserByNickname(ctx context.Context, nickname string) (*User, error) {
	const sql = `
		SELECT id, nickname, password, created_at, updated_at
		FROM users WHERE nickname = $1
	`
	var u User
	err := r.pool.QueryRow(ctx, sql, nickname).Scan(
		&u.ID, &u.Nickname, &u.Password, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pg SELECT users by nickname: %w", err)
	}
	return &u, nil
}

// UpdateUser 更新用户信息（昵称/密码哈希）
func (r *PostgresUserRepository) UpdateUser(ctx context.Context, user *User) error {
	const sql = `
		UPDATE users
		SET nickname = $2, password = $3, updated_at = NOW()
		WHERE id = $1
	`
	if user == nil {
		return errors.New("user is nil")
	}
	_, err := r.pool.Exec(ctx, sql, user.ID, user.Nickname, user.Password)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrDuplicateNickname
		}
		return fmt.Errorf("pg UPDATE users: %w", err)
	}
	return nil
}

// =====================================================================
// PostgresGameRecordRepository 游戏记录仓储
// =====================================================================

// PostgresGameRecordRepository 基于 pgx 的 GameRecordRepository 实现
type PostgresGameRecordRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresGameRecordRepository 创建 PostgreSQL 游戏记录仓储
func NewPostgresGameRecordRepository(pool *pgxpool.Pool) *PostgresGameRecordRepository {
	return &PostgresGameRecordRepository{pool: pool}
}

// SaveGameRecord 写入一局游戏记录
func (r *PostgresGameRecordRepository) SaveGameRecord(ctx context.Context, record *GameRecord) error {
	if record == nil {
		return errors.New("record is nil")
	}
	const sql = `
		INSERT INTO boomcat_game_records
			(room_id, players, action_log, result, created_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Now()
	}
	if record.FinishedAt.IsZero() {
		record.FinishedAt = time.Now()
	}
	err := r.pool.QueryRow(ctx, sql,
		record.RoomID,
		[]byte(record.Players),
		[]byte(record.ActionLog),
		[]byte(record.Result),
		record.CreatedAt,
		record.FinishedAt,
	).Scan(&record.ID)
	if err != nil {
		return fmt.Errorf("pg INSERT boomcat_game_records: %w", err)
	}
	return nil
}

// GetGameRecord 查询单条游戏记录
func (r *PostgresGameRecordRepository) GetGameRecord(ctx context.Context, id string) (*GameRecord, error) {
	const sql = `
		SELECT id, room_id, players, action_log, result, created_at, finished_at
		FROM boomcat_game_records WHERE id = $1
	`
	var rec GameRecord
	var players, actionLog, result []byte
	err := r.pool.QueryRow(ctx, sql, id).Scan(
		&rec.ID, &rec.RoomID, &players, &actionLog, &result,
		&rec.CreatedAt, &rec.FinishedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pg SELECT boomcat_game_records by id: %w", err)
	}
	rec.Players = JSONB(players)
	rec.ActionLog = JSONB(actionLog)
	rec.Result = JSONB(result)
	return &rec, nil
}

// ListGameRecordsByPlayer 查询某玩家参与的历史对局
// 利用 JSONB 玩家列表的 GIN 索引按 @> 操作符过滤
func (r *PostgresGameRecordRepository) ListGameRecordsByPlayer(ctx context.Context, playerID string, limit, offset int) ([]*GameRecord, error) {
	if limit <= 0 {
		limit = 20
	}
	const sql = `
		SELECT id, room_id, players, action_log, result, created_at, finished_at
		FROM boomcat_game_records
		WHERE players @> $1::jsonb
		ORDER BY finished_at DESC
		LIMIT $2 OFFSET $3
	`
	filter := fmt.Sprintf(`[{"id":"%s"}]`, playerID)
	rows, err := r.pool.Query(ctx, sql, filter, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("pg SELECT boomcat_game_records by player: %w", err)
	}
	defer rows.Close()

	var records []*GameRecord
	for rows.Next() {
		var rec GameRecord
		var players, actionLog, result []byte
		if err := rows.Scan(
			&rec.ID, &rec.RoomID, &players, &actionLog, &result,
			&rec.CreatedAt, &rec.FinishedAt,
		); err != nil {
			return nil, fmt.Errorf("pg scan boomcat_game_records: %w", err)
		}
		rec.Players = JSONB(players)
		rec.ActionLog = JSONB(actionLog)
		rec.Result = JSONB(result)
		records = append(records, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pg rows boomcat_game_records: %w", err)
	}
	return records, nil
}

// =====================================================================
// PostgresPlayerStatsRepository 玩家战绩仓储
// =====================================================================

// PostgresPlayerStatsRepository 基于 pgx 的 PlayerStatsRepository 实现
type PostgresPlayerStatsRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresPlayerStatsRepository 创建 PostgreSQL 玩家战绩仓储
func NewPostgresPlayerStatsRepository(pool *pgxpool.Pool) *PostgresPlayerStatsRepository {
	return &PostgresPlayerStatsRepository{pool: pool}
}

// GetPlayerStats 查询玩家战绩，不存在时返回 nil, nil
func (r *PostgresPlayerStatsRepository) GetPlayerStats(ctx context.Context, playerID string) (*PlayerStats, error) {
	const sql = `
		SELECT player_id, total_games, wins, losses, win_rate,
		       total_kills, total_draws, total_played, created_at, updated_at
		FROM player_stats WHERE player_id = $1
	`
	var s PlayerStats
	err := r.pool.QueryRow(ctx, sql, playerID).Scan(
		&s.PlayerID, &s.TotalGames, &s.Wins, &s.Losses, &s.WinRate,
		&s.TotalKills, &s.TotalDraws, &s.TotalPlayed,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pg SELECT player_stats: %w", err)
	}
	return &s, nil
}

// UpsertPlayerStats 按增量更新或初始化玩家战绩
// 使用 ON CONFLICT 做 upsert，win_rate 由 GENERATED 列自动维护
func (r *PostgresPlayerStatsRepository) UpsertPlayerStats(ctx context.Context, playerID string, delta PlayerStatsDelta) error {
	const sql = `
		INSERT INTO player_stats
			(player_id, total_games, wins, losses, total_kills, total_draws, total_played)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (player_id) DO UPDATE SET
			total_games  = player_stats.total_games + EXCLUDED.total_games,
			wins         = player_stats.wins + EXCLUDED.wins,
			losses       = player_stats.losses + EXCLUDED.losses,
			total_kills  = player_stats.total_kills + EXCLUDED.total_kills,
			total_draws  = player_stats.total_draws + EXCLUDED.total_draws,
			total_played = player_stats.total_played + EXCLUDED.total_played,
			updated_at   = NOW()
	`
	_, err := r.pool.Exec(ctx, sql,
		playerID,
		delta.TotalGames, delta.Wins, delta.Losses,
		delta.TotalKills, delta.TotalDraws, delta.TotalPlayed,
	)
	if err != nil {
		return fmt.Errorf("pg UPSERT player_stats: %w", err)
	}
	return nil
}

// ListTopPlayers 按胜率倒序获取排行榜
func (r *PostgresPlayerStatsRepository) ListTopPlayers(ctx context.Context, limit int) ([]*PlayerStats, error) {
	if limit <= 0 {
		limit = 50
	}
	const sql = `
		SELECT player_id, total_games, wins, losses, win_rate,
		       total_kills, total_draws, total_played, created_at, updated_at
		FROM player_stats
		WHERE total_games > 0
		ORDER BY win_rate DESC, total_games DESC
		LIMIT $1
	`
	rows, err := r.pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("pg SELECT player_stats top: %w", err)
	}
	defer rows.Close()

	var stats []*PlayerStats
	for rows.Next() {
		var s PlayerStats
		if err := rows.Scan(
			&s.PlayerID, &s.TotalGames, &s.Wins, &s.Losses, &s.WinRate,
			&s.TotalKills, &s.TotalDraws, &s.TotalPlayed,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("pg scan player_stats: %w", err)
		}
		stats = append(stats, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("pg rows player_stats: %w", err)
	}
	return stats, nil
}

// =====================================================================
// PostgresTokenRepository 用户刷新令牌仓储
// =====================================================================

// PostgresTokenRepository 基于 pgx 的 TokenRepository 实现
type PostgresTokenRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTokenRepository 创建 PostgreSQL 令牌仓储
func NewPostgresTokenRepository(pool *pgxpool.Pool) *PostgresTokenRepository {
	return &PostgresTokenRepository{pool: pool}
}

// SaveRefreshToken 写入刷新令牌
func (r *PostgresTokenRepository) SaveRefreshToken(ctx context.Context, token *RefreshToken) error {
	if token == nil {
		return errors.New("token is nil")
	}
	const sql = `
		INSERT INTO user_tokens (user_id, refresh_token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}
	return r.pool.QueryRow(ctx, sql,
		token.UserID, token.Token, token.ExpiresAt, token.CreatedAt,
	).Scan(&token.ID)
}

// GetRefreshToken 查询刷新令牌
func (r *PostgresTokenRepository) GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error) {
	const sql = `
		SELECT id, user_id, refresh_token, expires_at, created_at
		FROM user_tokens WHERE refresh_token = $1
	`
	var t RefreshToken
	err := r.pool.QueryRow(ctx, sql, token).Scan(
		&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("pg SELECT user_tokens: %w", err)
	}
	return &t, nil
}

// DeleteRefreshToken 删除单条刷新令牌
func (r *PostgresTokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	const sql = `DELETE FROM user_tokens WHERE refresh_token = $1`
	_, err := r.pool.Exec(ctx, sql, token)
	if err != nil {
		return fmt.Errorf("pg DELETE user_tokens: %w", err)
	}
	return nil
}

// DeleteTokensByUser 删除某用户所有刷新令牌
func (r *PostgresTokenRepository) DeleteTokensByUser(ctx context.Context, userID string) error {
	const sql = `DELETE FROM user_tokens WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, sql, userID)
	if err != nil {
		return fmt.Errorf("pg DELETE user_tokens by user: %w", err)
	}
	return nil
}

// =====================================================================
// 辅助函数
// =====================================================================

// isUniqueViolation 判断错误是否为 PostgreSQL 唯一约束冲突
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgErrUniqueViolation
	}
	return false
}
