# 服务边界与表所有权（M01 草案）

| 领域 | 当前 Go 代码入口（真实文件/目录） | 未来逻辑所有权 | 拆分顺序 |
|---|---|---|---|
| Core | `internal/service/conversation_service.go`、`internal/repository/conversation_repository.go`、User/Friend/Team/Channel 等服务 | 用户、好友、团队、会话、授权 | 后于 Message 服务分离，前期接口保留本地调用 |
| Message | `internal/service/conversation_message_service.go`、`channel_message_service.go`、`internal/repository/message_repository.go` | `messages`、`outbox_events`、幂等和同步规则 | M08–M10 gRPC 拆分 |
| Gateway | `internal/controller/ws_message_handle.go`、`internal/ws/` | WebSocket 连接、读写队列、外部协议适配 | M16 后 |
| Realtime | `internal/realtime/redis_broker.go`、`internal/service/presence_service.go` | 暂态在线目录、节点定向路由、通知 | M17–M20 |
| Audit | `internal/messaging/message_created_consumer.go` 当前只有日志 | 独立审计数据、事件去重与 Offset | M03–M04 首个独立进程 |
| Message Data | `internal/repository/message_repository.go` 中的持久化能力 | 存储接口、冷热读、分片路由（拟建） | M28 后评估是否另起进程 |

## 不变式

- HTTP `/api`、WS 事件名称、ACK、`client_message_id`、`/sync` 先保持兼容。
- 核心聊天主写入不依赖 Audit 成功。
- Message 存储和异步事件必须可在宕机后恢复。
- 以 `event_id` 业务唯一键处理至少一次投递；不能宣称端到端 Exactly-once。
- 不在 Gateway 绕过 Message API 直接跨服务访问 `messages`；过渡期的直接 SQL 要标记债务并由 M10 移除。
- 服务可以早期共用一套 MySQL **物理实例**，但逻辑表所有权须明确；未来不同服务不随意 JOIN。

## M01 暂不做

- 不创建 `cmd/message-service`、`cmd/gateway` 或 `.proto`。
- 不改现有业务表、不回改已执行的 Goose 迁移。
- 不替换 Redis/Kafka/MySQL，也不声称 ScyllaDB 已采用。
