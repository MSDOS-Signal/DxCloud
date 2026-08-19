DxCloud 备份清单
时间：2026-08-19 03:07:43
mysql.sql      MySQL 逻辑导出（single-transaction + routines + triggers）
redis-data/    Redis AOF + RDB（BGSAVE 后拷贝）
registry-data.tgz  Registry 存储卷 tar 包

恢复：.\tools\restore.ps1 -BackupDir "D:\deepseek_harness\own_project\backups\backup-20260819-030742"
