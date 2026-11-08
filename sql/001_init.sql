-- ============================================================
-- 爆炸猫（Exploding Kittens）在线桌游 - 数据库建表脚本
-- 数据库: PostgreSQL
-- 说明: 用户信息、游戏记录、排行榜
-- ============================================================

-- ------------------------------------------------------------
-- 1. 用户表（users）
-- 存储玩家账号信息，注册时写入，登录时验证。
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS users (
    id          VARCHAR(32)     PRIMARY KEY,              -- 用户唯一标识，如 "u_abc123"
    nickname    VARCHAR(16)     NOT NULL,                 -- 玩家昵称（2-16 字符）
    password    VARCHAR(256)    NOT NULL,                 -- 密码哈希值（bcrypt）
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),   -- 注册时间
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()    -- 更新时间
);

CREATE INDEX idx_users_nickname ON users (nickname);

COMMENT ON TABLE  users     IS '玩家账号信息';
COMMENT ON COLUMN users.id  IS '用户唯一标识，格式 u_<随机>';


-- ------------------------------------------------------------
-- 2. 游戏记录表（boomcat_game_records）
-- 对局结束后写入，存储完整对局信息，用于历史查询和回放。
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS boomcat_game_records (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),  -- 记录唯一标识
    room_id         VARCHAR(32)     NOT NULL,                              -- 房间 ID
    players         JSONB           NOT NULL,                              -- 玩家列表及排名
    action_log      JSONB           NOT NULL,                              -- 全量操作日志（可用于回放）
    result          JSONB           NOT NULL,                              -- 结果信息（胜者、出局顺序等）
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),                -- 游戏开始时间
    finished_at     TIMESTAMPTZ     NOT NULL,                              -- 游戏结束时间
    duration_sec    INTEGER         GENERATED ALWAYS AS (
        EXTRACT(EPOCH FROM finished_at - created_at)
    ) STORED                                                                 -- 对局时长（秒）
);

CREATE INDEX idx_boomcat_game_records_room_id    ON boomcat_game_records (room_id);
CREATE INDEX idx_boomcat_game_records_finished   ON boomcat_game_records (finished_at DESC);

COMMENT ON TABLE  boomcat_game_records               IS '对局历史记录';
COMMENT ON COLUMN boomcat_game_records.players       IS 'JSON 格式：[{"id":"u_001","nickname":"玩家A","rank":1,"alive":true}, ...]';
COMMENT ON COLUMN boomcat_game_records.action_log    IS 'JSON 格式：[{"seq":1,"type":"play_card","player":"u_001","card_ids":["cat_01"],"ts":1712345678}, ...]';
COMMENT ON COLUMN boomcat_game_records.result        IS 'JSON 格式：{"winner_id":"u_001","winner_nickname":"玩家A","elimination_order":["u_003","u_002"]}';


-- ------------------------------------------------------------
-- 3. 玩家战绩表（player_stats）
-- 记录每位玩家的累计战绩，用于排行榜展示。
-- 对局结束后更新，由应用层维护聚合数据。
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS player_stats (
    player_id       VARCHAR(32)     PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,  -- 玩家 ID
    total_games     INTEGER         NOT NULL DEFAULT 0,      -- 总对局数
    wins            INTEGER         NOT NULL DEFAULT 0,      -- 获胜次数
    losses          INTEGER         NOT NULL DEFAULT 0,      -- 失败次数
    win_rate        NUMERIC(5,4)    GENERATED ALWAYS AS (
        CASE WHEN total_games > 0
             THEN ROUND(wins::NUMERIC / total_games, 4)
             ELSE 0
        END
    ) STORED,                                                 -- 胜率
    total_kills     INTEGER         NOT NULL DEFAULT 0,      -- 总淘汰数（让对手爆炸出局）
    total_draws     INTEGER         NOT NULL DEFAULT 0,      -- 总抽牌数
    total_played    INTEGER         NOT NULL DEFAULT 0,      -- 总出牌数
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW(),  -- 首次记录时间
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()   -- 最后更新时间
);

CREATE INDEX idx_player_stats_win_rate ON player_stats (win_rate DESC);

COMMENT ON TABLE  player_stats                IS '玩家累计战绩与统计数据';
COMMENT ON COLUMN player_stats.win_rate       IS '胜率，保留 4 位小数，如 0.7500 表示 75%';
COMMENT ON COLUMN player_stats.total_kills    IS '累计让对手爆炸出局的次数';


-- ------------------------------------------------------------
-- 4. 用户认证 Token 表（user_tokens）
-- 存储 JWT 刷新令牌，用于 Token 续期。
-- ------------------------------------------------------------
CREATE TABLE IF NOT EXISTS user_tokens (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),  -- 记录唯一标识
    user_id         VARCHAR(32)     NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- 用户 ID
    refresh_token   VARCHAR(512)    NOT NULL,                            -- 刷新令牌
    expires_at      TIMESTAMPTZ     NOT NULL,                            -- 过期时间
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT NOW()               -- 创建时间
);

CREATE INDEX idx_user_tokens_user_id    ON user_tokens (user_id);
CREATE INDEX idx_user_tokens_expires    ON user_tokens (expires_at);

COMMENT ON TABLE  user_tokens             IS '用户刷新令牌记录';
COMMENT ON COLUMN user_tokens.refresh_token IS 'JWT 刷新令牌，哈希后存储';