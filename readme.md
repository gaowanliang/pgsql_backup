# pgsql_backup

一个基于 Go 的 PostgreSQL 备份工具，包括全量备份、增量备份（WAL 归档）和表级备份。

## 功能

- 全量数据库备份
- 使用 WAL 归档的增量备份
- 表级备份
- 使用 cron 作业的自动备份调度
- 备份保留管理

## 配置

配置通过 `config.yaml` 文件管理。以下是一个示例配置：

```yaml
postgres:
  host: "localhost"
  port: 5432
  user: "postgres"
  password: "test123456"
  data_dir: "C:/Users/MissMirai/Desktop/backup"
  database: "cmsdb"
backup:
  only_remains_days: 148
  check_removes_interval: 24
  full_backup_dir: "C:/Users/MissMirai/Desktop/backup/full_backup"
  full_backup_interval: 148
  wal_archive_dir: "C:/Users/MissMirai/Desktop/backup/archive_dir"
  wal_archive_backup_dir: "C:/Users/MissMirai/Desktop/backup/wal_archive"
  clean_wal_archive_dir: true
  wal_archive_interval: 24
  pg_basebackup: "C:/Program Files/PostgreSQL/16/bin/pg_basebackup.exe"
  pg_dump: "C:/Program Files/PostgreSQL/16/bin/pg_dump.exe"
  table_backup_dir: "C:/Users/MissMirai/Desktop/backup/table_backup"
  tables: [ ]
  table_backup_interval: 24
```

## 使用方法

### 命令行参数

- `-f`: 执行全量备份
- `-i`: 执行增量备份（WAL 归档）
- `-t`: 执行表级备份
- `-r`: 删除旧备份文件

### 运行应用程序

要运行应用程序，请使用以下命令：

```sh
go run main.go [options]
```

例如，执行全量备份：

```sh
go run main.go -f
```

### 构建应用程序

要为不同平台构建应用程序，请使用 `goreleaser`。确保在环境中设置了有效的 `GITHUB_TOKEN`。

```sh
goreleaser release --snapshot --skip-publish --rm-dist
```

## 日志记录

日志记录到 `setupLogging` 函数中指定的文件。确保日志文件路径对应用程序可写。

## 许可证

此项目根据 MIT 许可证授权。有关详细信息，请参阅 `LICENSE` 文件。