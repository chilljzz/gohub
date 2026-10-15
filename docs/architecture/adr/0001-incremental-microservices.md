# ADR-0001：按功能边界渐进式拆分 GoHub

- **状态**：Proposed（M01 等待本地验证和审查）
- **记录日期**：2026-10-15
- **参考基线**：`main@ff9d7107994df6a02d24739b8254e67c79d913e8`

## 背景（Context）

当前 GoHub 是一个 Gin + WebSocket 的单进程应用。`internal/app/app.go` 同时初始化 Redis 订阅、Kafka 发布端、Kafka 消费端和 HTTP 服务。消息通过 `MessageRepository.CreateWithOutbox` 原子写入 `messages/outbox_events`，但尚未在远程 main 发现可靠投递 Worker。若立即切分全部服务，会扩大网络超时、鉴权、数据一致性和本地部署复杂度，也可能破坏旧 WS/HTTP 行为。

## 决策驱动因素

1. 旧客户端 ACK、历史同步和 Direct Receipt 必须继续工作。
2. 让 Kafka 真正承载审计/搜索/统计等可独立失败的下游任务。
3. 在面试中可展示真实服务进程、错误恢复和可复现实验。
4. 优先可控迁移，不依赖高并发数字假设。

## 备选方案

- **A：保持单体直至功能全部完成。** 优点：简单；缺点：故障隔离与可扩容能力难以验证。
- **B：立即拆所有微服务。** 优点：目标架构接近；缺点：工作量、部署成本、分布式故障风险显著增加。
- **C：先模块边界与可靠事件，再分阶段拆 Audit → Message → Gateway。** 优点：每一步可测试、可回滚；缺点：迁移期间有适配器和混合拓扑。

## 决策（Decision）

选择 **C**：继续使用 Monorepo，不重排全部 `internal/`；先以 M01 建立架构基线，M02 核对 Outbox Worker，M03–M04 完成独立 Audit，再以 gRPC 拆 Message Service，最后拆分多 Gateway / Dispatcher。Kafka 仍承担可恢复的异步事件总线，消息事实来源仍为数据库。

## 后果（Consequences）

- 积极：旧 API 可逐步迁移、可验证 Kafka Consumer 独立性、便于回滚。
- 消极：一段时期内旧单体和新服务并存，需要兼容测试、双路由切换和运维文档。
- 风险控制：每次拆分前先有契约测试，单点故障注入；出问题时回滚到原来的本地调用，而不是回退数据库历史迁移。

## 验收证据（待填）

- 本地分支：TBD
- 单元测试：未执行
- 集成测试：未执行
- 评审结论：待用户审查
