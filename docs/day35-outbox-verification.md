# GoHub Day 35 Transactional Outbox 验证

## 1. 代码版本
- Git Commit:
- Go 版本:
- MySQL 版本:
- goose 版本:

## 2. Migration
- SQL 文件:
- 迁移版本:
- 建表结果:
- 重复运行结果:

## 3. 事务验证
- 首次写入:
- 重复请求:
- 不同内容冲突:
- Outbox 唯一键冲突:
- 事务回滚结果:

## 4. 测试命令
- go test ./...
- go test -race ./...
- go test -tags=integration ./internal/repository -v

## 5. 已知限制
- 尚未实现 Outbox Worker
- 尚未验证 Kafka 恢复投递
- 尚未完成并发幂等压力测试