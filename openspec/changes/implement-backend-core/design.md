## Context

爆炸猫后端基于 Go 1.25，使用 Gin + WebSocket + Redis + PostgreSQL 技术栈。当前项目只有空文件骨架，需从零实现完整后端逻辑。项目同时存在两种组织方式：按业务领域划分的 Module 3 结构（`internal/game/*`、`internal/room/` 等）和 MVC 分层结构（`internal/model/`、`internal/service/`、`internal/controller/` 等）。两者是互补关系——MVC 按架构层次组织，Module 3 按业务领域组织，中间通过依赖注入串联。

## Goals / Non-Goals

**Goals:**
- 实现完整的爆炸猫游戏后端，支持 2-5 人实时对战
- 实现三层存储架构：内存（Primary）→ Redis（热备）→ PostgreSQL（归档）
- 实现每房间一个 goroutine + channel 的并发模型，无锁设计
- 实现 Nope 链、攻击叠加、回合调度等核心游戏机制
- 实现用户认证、房间管理、WebSocket 实时通信

**Non-Goals:**
- 不涉及前端代码
- 不涉及分布式部署（单实例即可，不引入服务发现/负载均衡）
- 不涉及游戏回放 UI 界面
- 不涉及第三方登录（仅账号密码）

## Decisions

### 1. 编码顺序：纯逻辑优先，I/O 靠后

采用自底向上的 9 阶段编码策略：
- Phase 1-2：纯数据模型与领域逻辑，无 I/O，可独立测试
- Phase 3：基础设施层（配置、Redis、PostgreSQL）
- Phase 4：Service 层（业务编排，依赖接口而非实现）
- Phase 5-6：网络层与 Controller 层
- Phase 7：View 层与序列化
- Phase 8：中间件、路由、入口组装
- Phase 9：工具包（贯穿始终）

**理由**：纯逻辑层在编写时不需要依赖外部服务（Redis/PG），写完即可 `go test` 验证。如果先写 I/O 层，调试时需要启动外部服务，反馈周期长。

### 2. 并发模型：每房间一个 goroutine + channel

严格遵循文档 4.3 节设计。每个房间有一个独立的 goroutine 运行事件循环，所有玩家消息通过 channel 串行化，游戏逻辑在同一 goroutine 内串行执行，无需任何锁。

**理由**：爆炸猫每局 2-5 人，所有操作通过 channel 串行化天然有序，没有竞态条件。Go 的 goroutine 足够轻量，每房间一个 goroutine 不会成为瓶颈。

### 3. 依赖注入：构造函数注入接口

Service 层只依赖 Repository 接口（定义在 `repository/interfaces.go`），不依赖具体实现。Controller 层只依赖 Service 接口。所有依赖在 `cmd/server/main.go` 中组装。

**理由**：方便单元测试——测试 Service 时传入 mock Repository，测试 Controller 时传入 mock Service。

### 4. 信息存储：三层架构

内存（Primary）→ Redis（热备）→ PostgreSQL（归档）。游戏进行中的热路径全部走内存，I/O 操作异步执行不阻塞主流程。游戏结束后才写入 PostgreSQL。

**理由**：文档 4.1 节明确的三层架构，与 Redis 设计文件 `redis/001_init.redis` 保持一致。

### 5. 两个组织结构的映射关系

| MVC 层次 | Module 3 领域 | 说明 |
|---|---|---|
| Model | `game/*`, `session/` | 纯数据模型与业务规则 |
| Repository | `persistence/` | 数据持久化抽象 |
| Service | `room/`, `service/` | 业务编排 |
| Controller | `network/`, `controller/` | 请求入口与消息分发 |
| View | `network/`（广播部分） | 按视角过滤序列化 |

开发时优先填充 Module 3 领域目录（`internal/game/`、`internal/room/` 等），然后在其上构建 MVC 层。

## Risks / Trade-offs

| 风险 | 缓解措施 |
|---|---|
| 单 goroutine 处理所有玩家消息，如果某条消息处理过慢会阻塞后续消息 | 所有 I/O 操作异步执行（Redis 写入 goroutine 外执行），游戏逻辑本身是纯内存操作，纳秒级完成 |
| 服务器崩溃时内存状态丢失 | Redis 异步备份，重启后从 Redis 恢复房间状态 |
| Nope 窗口关闭瞬间的竞态 | drain 机制：先清空 inbox 积压消息再结算，见文档 4.6 节 |
| 空 .go 文件导致编译失败 | 每个文件先写 package 声明和必要 import，确保能编译通过后再填充逻辑 |