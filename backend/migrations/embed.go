// Package migrations 内嵌 SQL 迁移文件。
// 规则：文件按 000001_xxx.sql 递增编号；只增不改；由 internal/database.Migrate 按序执行一次。
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
