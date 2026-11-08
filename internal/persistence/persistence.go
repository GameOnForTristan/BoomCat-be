package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"boom-cat-be/internal/config"
	"boom-cat-be/internal/repository"
)

// =====================================================================
// 持久化层封装
// 提供 RedisAdapter（Redis 客户端 + 全部 Redis 仓储聚合）和
// Database（PostgreSQL 连接池 + 全部 PG 仓储聚合）两个上层对象，
// 供 main.go 一次性组装、注入到 Service 层。
// =====================================================================

// RedisAdapter 聚合 Redis 客户端与全部 Redis 仓储
type RedisAdapter struct {
	client *redis.Client

	GameState  repository.GameStateRepository
	RoomMeta   repository.RoomMetaRepository
	InviteCode repository.InviteCodeRepository
	RoomPlayers repository.RoomPlayersRepository
	Session    repository.SessionRepository
}

// NewRedisAdapter 创建 Redis 客户端并组装所有 Redis 仓储
func NewRedisAdapter(cfg config.RedisConfig) (*RedisAdapter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &RedisAdapter{
		client:      client,
		GameState:   repository.NewRedisGameStateRepository(client),
		RoomMeta:    repository.NewRedisRoomMetaRepository(client),
		InviteCode:  repository.NewRedisInviteCodeRepository(client),
		RoomPlayers: repository.NewRedisRoomPlayersRepository(client),
		Session:     repository.NewRedisSessionRepository(client),
	}, nil
}

// Client 暴露底层 Redis 客户端（用于 Health check / 高级用法）
func (a *RedisAdapter) Client() *redis.Client {
	return a.client
}

// Ping 检查 Redis 连接
func (a *RedisAdapter) Ping(ctx context.Context) error {
	if a == nil || a.client == nil {
		return errors.New("redis adapter not initialized")
	}
	return a.client.Ping(ctx).Err()
}

// Close 关闭 Redis 连接
func (a *RedisAdapter) Close() error {
	if a == nil || a.client == nil {
		return nil
	}
	return a.client.Close()
}

// Database 聚合 PostgreSQL 连接池与全部 PG 仓储
type Database struct {
	pool *pgxpool.Pool

	User        repository.UserRepository
	GameRecord  repository.GameRecordRepository
	PlayerStats repository.PlayerStatsRepository
	Token       repository.TokenRepository
}

// NewDatabase 创建 PostgreSQL 连接池并组装所有 PG 仓储
func NewDatabase(ctx context.Context, cfg config.DatabaseConfig) (*Database, error) {
	poolCfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password,
		cfg.Name, cfg.SSLMode, cfg.MaxConns,
	))
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}

	return &Database{
		pool:        pool,
		User:        repository.NewPostgresUserRepository(pool),
		GameRecord:  repository.NewPostgresGameRecordRepository(pool),
		PlayerStats: repository.NewPostgresPlayerStatsRepository(pool),
		Token:       repository.NewPostgresTokenRepository(pool),
	}, nil
}

// Pool 暴露底层 pgxpool（用于事务 / 高级查询）
func (d *Database) Pool() *pgxpool.Pool {
	return d.pool
}

// Ping 检查 PostgreSQL 连接
func (d *Database) Ping(ctx context.Context) error {
	if d == nil || d.pool == nil {
		return errors.New("database not initialized")
	}
	return d.pool.Ping(ctx)
}

// Close 关闭连接池
func (d *Database) Close() {
	if d == nil || d.pool == nil {
		return
	}
	d.pool.Close()
}
