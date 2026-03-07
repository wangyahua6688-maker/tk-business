# seed_draw_history

用于给 `tk_special_lottery`、`tk_draw_record` 生成一批可联调的“开奖历史”测试数据。

## 用法

```bash
cd /Users/dollar/go_space/tk-business
go run ./tools/seed_draw_history
```

可选：通过环境变量覆盖默认数据库连接。

```bash
TK_DB_DSN='user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local' go run ./tools/seed_draw_history
```

## 说明

- 脚本会确保存在两个彩种：`macau`（澳彩）、`hk`（港彩）；
- 每个彩种会 upsert 12 期开奖记录；
- 最新一期会自动标记为 `is_current=1`；
- 会同步更新 `tk_special_lottery.current_issue` 与 `next_draw_at`。
