# w_* -> tk_* 数据迁移工具

## 用途

将历史 `w_*` 业务表数据迁移到新命名的 `tk_*` 表（本工具当前覆盖）：

- `w_external_link` -> `tk_external_link`
- `w_lottery_category` -> `tk_lottery_category`
- `w_lottery_info` -> `tk_lottery_info`

## 特性

- 幂等：可重复执行；
- 先做结构对齐，再做数据导入；
- 采用“先更新再插入”的方式，避免重复数据；
- 自动输出迁移后的计数统计。

## 执行

默认连接：`nb_sys_001`（本地开发）。

```bash
cd /Users/dollar/go_space/tk-business
go run ./tools/migrate_w_to_tk
```

指定数据库连接：

```bash
cd /Users/dollar/go_space/tk-business
TK_DB_DSN='user:pass@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local' go run ./tools/migrate_w_to_tk
```

