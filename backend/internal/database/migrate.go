package database

import (
	"fmt"
	"io/fs"
	"sort"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Migrate 按文件名顺序执行尚未应用的 SQL 迁移（只增不改）。
// 已执行版本记录在 schema_migrations 表；同名文件幂等跳过。
// 注意：依赖 DSN 中的 multiStatements=true 一次性执行整份文件。
func Migrate(db *gorm.DB, fsys fs.FS, log *zap.Logger) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version    VARCHAR(191) NOT NULL,
		applied_at DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
		PRIMARY KEY (version)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`).Error; err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", e.Name()).Scan(&count).Error; err != nil {
			return fmt.Errorf("check version %s: %w", e.Name(), err)
		}
		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(fsys, e.Name())
		if err != nil {
			return fmt.Errorf("read %s: %w", e.Name(), err)
		}
		if err := db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("apply %s: %w", e.Name(), err)
		}
		if err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", e.Name()).Error; err != nil {
			return fmt.Errorf("record %s: %w", e.Name(), err)
		}
		log.Info("migration applied", zap.String("file", e.Name()))
	}
	return nil
}
