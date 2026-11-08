## Why

爆炸猫是一款完全信息不对称的实时在线桌游，后端需要处理 WebSocket 长连接、游戏状态机、Nope 链、回合调度等复杂逻辑。当前项目仅有空文件骨架，需要从零实现完整的后端核心逻辑，使游戏可运行。

## What Changes

- 实现完整的游戏核心模型层（Card、Deck、Hand、ActionStack、TurnEngine、GameState）
- 实现房间管理与玩家生命周期
- 实现用户认证系统（注册、登录、JWT）
- 实现 Redis 游戏状态快照与 PostgreSQL 持久化存储
- 实现 WebSocket 实时通信与状态广播
- 实现 Service 层业务编排与超时控制
- 实现 HTTP RESTful 接口与路由注册
- 实现中间件（认证、CORS、异常恢复、日志）

## Capabilities

### New Capabilities
- `game-core`: 核心游戏模型，包括 Card/Deck/Hand/ActionStack/TurnEngine/GameState 纯逻辑
- `room-lifecycle`: 房间创建、加入、退出、准备、开始，玩家连接状态管理
- `user-auth`: 用户注册、登录、JWT Token 签发与验证
- `game-persistence`: Redis 状态热备与 PostgreSQL 归档存储（用户信息、游戏记录、排行榜）
- `real-time-communication`: WebSocket 连接管理、消息分发、状态广播、断线重连
- `game-services`: Service 层业务编排，包括出牌流程、回合调度、超时控制
- `http-api`: HTTP RESTful 接口（房间操作、用户认证）与路由注册
- `middleware-layer`: HTTP 中间件（认证、CORS、异常恢复、请求日志）
- `config-and-tooling`: 配置管理、结构化日志、错误码定义、通用工具函数

### Modified Capabilities

无（本项目从零开始，尚无现有 spec）

## Impact

- 所有 `internal/` 下的包将写入具体实现代码
- 依赖 `gin-gonic/gin`、`gorilla/websocket`、`go-redis/redis/v9`、`jackc/pgx/v5`、`golang-jwt/jwt/v5`、`google/uuid`
- 数据库迁移脚本 `sql/001_init.sql` 已就绪，Redis 结构文档 `redis/001_init.redis` 已就绪
- 不影响现有前端代码