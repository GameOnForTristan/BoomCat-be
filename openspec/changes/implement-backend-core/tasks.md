## 1. Model Layer — 纯数据模型

- [x] 1.1 实现 `model/card.go`：CardType 枚举（9种类型）、Card 值对象、NewCard 构造函数
- [x] 1.2 实现 `model/deck.go`：Deck 结构体、NewDeck、Shuffle、Draw、InsertAt、Peek 方法
- [x] 1.3 实现 `model/hand.go`：Hand 结构体、AddCard、RemoveCard、FindCard、ValidateCombo 方法
- [x] 1.4 实现 `model/player.go`：Player 实体、ConnectionStatus 枚举
- [x] 1.5 实现 `model/action.go`：ActionType 枚举、Action/StackEntry 值对象、ActionStack 结构体及 Resolve 方法
- [x] 1.6 实现 `model/turn.go`：TurnPhase 枚举、TurnEngine 结构体及回合流转方法
- [x] 1.7 实现 `model/game.go`：GamePhase 枚举、GameState 聚合根、PlayerInfo 结构体
- [x] 1.8 实现 `model/room.go`：Room 实体、RoomSettings 结构体
- [x] 1.9 实现 `model/errors.go`：领域错误定义（ErrInvalidMove、ErrNotYourTurn 等）

## 2. Domain Modules — 领域逻辑细化

- [x] 2.1 实现 `game/card/card.go`：卡牌创建工具函数、DeckComposition 配置、GenerateCatCards 和 GenerateActionCards
- [x] 2.2 实现 `game/deck/deck.go`：按人数初始化牌堆（BuildInitialDeck、BuildGameDeck、InsertExplodingKittens）
- [x] 2.3 实现 `game/hand/hand.go`：ComboType 枚举、FindValidCombos、GetAvailableSingles、SortHandByType
- [x] 2.4 实现 `game/turn/turn.go`：AdvanceTurn、HandleSkip/Attack/Draw、HandleElimination、阶段查询
- [x] 2.5 实现 `game/action/action.go`：PushAction/PushNope、Resolve、CanNope、PeekBaseAction
- [x] 2.6 实现 `game/validator/validator.go`：出牌/抽牌/Nope/目标/组合/拆解合法性校验
- [x] 2.7 实现 `game/gameover/gameover.go`：GameOverResult、CheckGameOver、HandleElimination
- [x] 2.8 实现 `session/session.go`：PlayerStateView、BuildPlayerView（按视角过滤）
- [x] 2.9 实现 `timer/timer.go`：GameTimer 结构体、TimerManager 管理三种定时器

## 3. Infrastructure Layer — 配置与持久化

- [x] 3.1 实现 `config/config.go`：Config 结构体、LoadConfig 函数（YAML + 环境变量）
- [x] 3.2 实现 `repository/interfaces.go`：Repository 接口定义（GameStateRepository、UserRepository、GameRecordRepository）
- [x] 3.3 实现 `repository/redis.go`：Redis 游戏状态快照读写（SaveGameState、LoadGameState）
- [x] 3.4 实现 `repository/redis.go`：Redis 房间元数据与邀请码映射操作
- [x] 3.5 实现 `repository/postgres.go`：PostgreSQL 用户信息 CRUD
- [x] 3.6 实现 `repository/postgres.go`：PostgreSQL 游戏记录写入与查询
- [x] 3.7 实现 `repository/postgres.go`：PostgreSQL 玩家战绩更新与排行榜查询
- [x] 3.8 实现 `persistence/persistence.go`：RedisAdapter 与 Database 上层封装

## 4. Service Layer — 业务编排

- [x] 4.1 实现 `service/auth_service.go`：注册（bcrypt 哈希）、登录、JWT 签发与验证
- [x] 4.2 实现 `service/room_service.go`：创建房间、加入房间（邀请码）、退出房间
- [x] 4.3 实现 `service/room_service.go`：准备/取消准备、开始游戏（全员准备校验）
- [x] 4.4 实现 `service/game_service.go`：出牌流程编排（校验→执行→入栈→开 Nope 窗口→广播）
- [x] 4.5 实现 `service/game_service.go`：抽牌流程（含爆炸猫判定→拆解/出局）
- [x] 4.6 实现 `service/game_service.go`：Nope 结算流程
- [x] 4.7 实现 `service/turn_service.go`：回合流转调度（含攻击累加、方向变更）
- [x] 4.8 实现 `service/timer_service.go`：三种超时管理（回合、Nope 窗口、重连）

## 5. Network Layer — 网络通信

- [x] 5.1 实现 `network/network.go`：GameSocket 封装（WebSocket 读写、消息解析）
- [x] 5.2 实现 `network/network.go`：EventHandler 消息路由与分发
- [x] 5.3 实现 `network/network.go`：Broadcaster 广播（全员/单人/部分玩家）

## 6. Controller Layer — 请求入口

- [x] 6.1 实现 `controller/auth_controller.go`：注册/登录/刷新 Token HTTP 接口
- [x] 6.2 实现 `controller/room_controller.go`：房间 CRUD HTTP 接口
- [x] 6.3 实现 `controller/ws_controller.go`：WebSocket 连接升级、读写管理、断线检测
- [x] 6.4 实现 `controller/game_controller.go`：WebSocket 消息分发（play_card、draw_card、play_nope 等）

## 7. View Layer — 序列化呈现

- [x] 7.1 实现 `view/serializer.go`：按玩家视角过滤游戏状态序列化
- [x] 7.2 实现 `view/event.go`：服务端推送事件定义（state_update、nope_prompt、game_over 等）
- [x] 7.3 实现 `view/response.go`：统一 HTTP 响应格式（code、message、data）
- [x] 7.4 实现 `view/dto.go`：RoomDTO、PlayerDTO、GameStateDTO 定义

## 8. Middleware + Router + Entry Point

- [ ] 8.1 实现 `middleware/auth.go`：JWT 认证中间件
- [x] 8.2 实现 `middleware/cors.go`：CORS 跨域中间件
- [x] 8.3 实现 `middleware/recovery.go`：Panic 恢复中间件
- [x] 8.4 实现 `middleware/logging.go`：请求日志中间件
- [x] 8.5 实现 `router/router.go`：主路由注册（HTTP + WebSocket）
- [x] 8.6 实现 `router/routes.go`：路由表定义
- [x] 8.7 实现 `cmd/server/main.go`：启动入口，依赖组装与初始化
- [x] 8.8 编写 `config/config.yaml`：默认配置文件

## 9. Tooling Package — 工具包

- [x] 9.1 实现 `pkg/logger/logger.go`：结构化日志（分级、格式、输出目标）
- [x] 9.2 实现 `pkg/errors/errors.go`：错误码定义与错误包装
- [x] 9.3 实现 `pkg/utils/utils.go`：ID 生成器、邀请码生成、时间工具
- [x] 9.4 实现 `pkg/types/types.go`：共享类型与常量定义