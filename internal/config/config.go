package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 全局配置根结构
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Redis    RedisConfig    `yaml:"redis"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Game     GameConfig      `yaml:"game"`
	Log      LogConfig       `yaml:"log"`
}

// ServerConfig HTTP/WebSocket 服务配置
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
	Mode         string `yaml:"mode"`
}

// RedisConfig Redis 连接配置
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

// DatabaseConfig PostgreSQL 连接配置
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`
	MaxConns int    `yaml:"max_conns"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret        string `yaml:"jwt_secret"`
	AccessTokenTTL   int    `yaml:"access_token_ttl"`
	RefreshTokenTTL  int    `yaml:"refresh_token_ttl"`
	BcryptCost       int    `yaml:"bcrypt_cost"`
}

// GameConfig 游戏参数配置
type GameConfig struct {
	DefaultMaxPlayers      int `yaml:"default_max_players"`
	MinPlayers             int `yaml:"min_players"`
	MaxPlayers             int `yaml:"max_players"`
	TurnTimeoutSec         int `yaml:"turn_timeout_sec"`
	NopeWindowSec          int `yaml:"nope_window_sec"`
	ReconnectTimeoutSec    int `yaml:"reconnect_timeout_sec"`
	StateSnapshotIntervalMs int `yaml:"state_snapshot_interval_ms"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
}

// defaultConfig 返回带合理默认值的配置
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         8080,
			ReadTimeout:  10,
			WriteTimeout: 10,
			Mode:         "debug",
		},
		Redis: RedisConfig{
			Host:     "127.0.0.1",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     5432,
			User:     "boomcat",
			Password: "",
			Name:     "boomcat",
			SSLMode:  "disable",
			MaxConns: 20,
		},
		Auth: AuthConfig{
			JWTSecret:       "change-me-in-production",
			AccessTokenTTL:  3600,
			RefreshTokenTTL: 259200,
			BcryptCost:      10,
		},
		Game: GameConfig{
			DefaultMaxPlayers:       5,
			MinPlayers:             2,
			MaxPlayers:             5,
			TurnTimeoutSec:          60,
			NopeWindowSec:           3,
			ReconnectTimeoutSec:     90,
			StateSnapshotIntervalMs: 200,
		},
		Log: LogConfig{
			Level:    "info",
			Format:   "text",
			Output:   "stdout",
			FilePath: "./logs/server.log",
		},
	}
}

// LoadConfig 加载 YAML 配置文件，并应用环境变量覆盖
// 环境变量约定：BOOMCAT_<SECTION>_<FIELD>，例如 BOOMCAT_SERVER_PORT
func LoadConfig(path string) (*Config, error) {
	cfg := defaultConfig()

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("read config file: %w", err)
			}
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("parse config yaml: %w", err)
			}
		}
	}

	applyEnvOverrides(cfg)

	return cfg, nil
}

// applyEnvOverrides 用环境变量覆盖配置项
func applyEnvOverrides(cfg *Config) {
	// Server
	envStr("BOOMCAT_SERVER_HOST", &cfg.Server.Host)
	envInt("BOOMCAT_SERVER_PORT", &cfg.Server.Port)
	envInt("BOOMCAT_SERVER_READ_TIMEOUT", &cfg.Server.ReadTimeout)
	envInt("BOOMCAT_SERVER_WRITE_TIMEOUT", &cfg.Server.WriteTimeout)
	envStr("BOOMCAT_SERVER_MODE", &cfg.Server.Mode)

	// Redis
	envStr("BOOMCAT_REDIS_HOST", &cfg.Redis.Host)
	envInt("BOOMCAT_REDIS_PORT", &cfg.Redis.Port)
	envStr("BOOMCAT_REDIS_PASSWORD", &cfg.Redis.Password)
	envInt("BOOMCAT_REDIS_DB", &cfg.Redis.DB)
	envInt("BOOMCAT_REDIS_POOL_SIZE", &cfg.Redis.PoolSize)

	// Database
	envStr("BOOMCAT_DB_HOST", &cfg.Database.Host)
	envInt("BOOMCAT_DB_PORT", &cfg.Database.Port)
	envStr("BOOMCAT_DB_USER", &cfg.Database.User)
	envStr("BOOMCAT_DB_PASSWORD", &cfg.Database.Password)
	envStr("BOOMCAT_DB_NAME", &cfg.Database.Name)
	envStr("BOOMCAT_DB_SSL_MODE", &cfg.Database.SSLMode)
	envInt("BOOMCAT_DB_MAX_CONNS", &cfg.Database.MaxConns)

	// Auth
	envStr("BOOMCAT_AUTH_JWT_SECRET", &cfg.Auth.JWTSecret)
	envInt("BOOMCAT_AUTH_ACCESS_TTL", &cfg.Auth.AccessTokenTTL)
	envInt("BOOMCAT_AUTH_REFRESH_TTL", &cfg.Auth.RefreshTokenTTL)
	envInt("BOOMCAT_AUTH_BCRYPT_COST", &cfg.Auth.BcryptCost)

	// Game
	envInt("BOOMCAT_GAME_DEFAULT_MAX_PLAYERS", &cfg.Game.DefaultMaxPlayers)
	envInt("BOOMCAT_GAME_MIN_PLAYERS", &cfg.Game.MinPlayers)
	envInt("BOOMCAT_GAME_MAX_PLAYERS", &cfg.Game.MaxPlayers)
	envInt("BOOMCAT_GAME_TURN_TIMEOUT", &cfg.Game.TurnTimeoutSec)
	envInt("BOOMCAT_GAME_NOPE_WINDOW", &cfg.Game.NopeWindowSec)
	envInt("BOOMCAT_GAME_RECONNECT_TIMEOUT", &cfg.Game.ReconnectTimeoutSec)

	// Log
	envStr("BOOMCAT_LOG_LEVEL", &cfg.Log.Level)
	envStr("BOOMCAT_LOG_FORMAT", &cfg.Log.Format)
	envStr("BOOMCAT_LOG_OUTPUT", &cfg.Log.Output)
	envStr("BOOMCAT_LOG_FILE_PATH", &cfg.Log.FilePath)
}

// envStr 若环境变量存在且非空，则覆盖 target 字符串
func envStr(key string, target *string) {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		*target = v
	}
}

// envInt 若环境变量存在且可解析为整数，则覆盖 target
func envInt(key string, target *int) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil {
		return
	}
	*target = n
}

// ServerAddr 返回 HTTP 服务监听地址 host:port
func (c *Config) ServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// RedisAddr 返回 Redis 地址 host:port
func (c *Config) RedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Redis.Host, c.Redis.Port)
}

// PostgresDSN 返回 PostgreSQL 连接 DSN
func (c *Config) PostgresDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port,
		c.Database.User, c.Database.Password,
		c.Database.Name, c.Database.SSLMode)
}
