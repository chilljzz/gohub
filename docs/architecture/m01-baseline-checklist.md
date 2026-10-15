# M01 基线与兼容性检查表

## 环境和 Git（实际填写，不要预填通过）

- 远程 main SHA：`ff9d7107994df6a02d24739b8254e67c79d913e8`（2026-10-15 已远程核对）
- 本地 `git branch --show-current`：待填写
- 本地 `git status --short`：待填写
- `git log -1 --format='%H %s'`：待填写
- Go / Docker / MySQL 版本：待填写

## 基线命令（在 WSL2 项目根目录）

```bash
go version
docker compose config --services
go vet ./...
go test ./... -count=1
go test -race ./... -count=1 -timeout=5m
```

不要在未隔离测试 DB 前运行 integration 测试。`internal/repository/outbox_integration_test.go` 需要仓库根目录 `.env.test.local`，其中 `GOHUB_TEST_MYSQL_DSN` 指向独立 MySQL，且 `outbox_events` 要先执行 Goose Migration 创建。

## 兼容性矩阵

| 场景 | 预期语义 | 实际结果 / 测试证据 |
|---|---|---|
| 首次 SendMessage | Message + Outbox 同事务，`Created=true` | 待执行 |
| 同一 `client_message_id` 重试 | 原 Message ID，`Created=false` | 待执行 |
| 相同幂等键、不同内容 | 返回冲突，不修改原消息 | 待执行 |
| Outbox 插入故障 | 事务回滚，不留下孤立 Message | 待执行 |
| 消息提交成功后 WS ACK | `message_ack` 包含消息/会话/客户端幂等 ID | 待执行 |
| 普通 Channel/Direct 历史读取 | 兼容已有列表/游标与 `/sync` | 待执行 |
| Redis Pub/Sub 短时不可用 | **预期设计**：历史可补拉；当前故障表现需实验 | 待执行 |
| Kafka 不可用 / Audit Consumer 停止 | **目标**：不阻断聊天业务；当前 `App.New()` Ping 是已知风险 | 待执行 |
| 旧客户端 Read/Delivered | 个人 Direct 回执；大群不做逐人广播 | 待执行 |

## Day4 Outbox Worker 本地核验

```bash
git branch -a
git log --all --oneline --grep='outbox' -i
rg -n 'SKIP LOCKED|locked_until|lease|OutboxWorker|ClaimBatch|PublishOutbox' --glob '*.go' .
```

这些命令只能定位候选代码，还需确认它是否被 `cmd/` 或 `App.Run()` 实际启动，并运行并发抢占、进程崩溃、Kafka 停止/恢复等集成测试。

## M01 结束门槛

- [ ] Git 工作区已核对，必要改动安全保留；基于最新 main 建了 M01 分支
- [ ] 文档中的“现状”和“目标”无混用
- [ ] 已记录 `go test`、`go test -race` 真实输出及失败原因
- [ ] API/WS 路径没有改动，未引入数据库破坏性迁移
- [ ] 明确标出 Day4 Worker 本地状态：已验证 / local-only / 未完成 / 暂无法核实
- [ ] 提交并记录结果 SHA，合并结果与 PR 链接（若有）

## 复盘

- 最后一次真实通过的测试：待填写
- M01 提交：待填写
- 未解决问题：待填写
- 下一单元：M02（若 Worker 已完整验证则按进度进入 M03）
