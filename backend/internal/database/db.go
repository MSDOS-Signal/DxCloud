// Package database 负责 MySQL 连接池与 SQL 迁移执行。
package database

import (
	"fmt"
	"time"

	"github.com/dxcloud/cloud-api/internal/config"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// Connect 建立 MySQL 连接，带重试（等待 compose 依赖 healthy），最多 30 次 × 2s。
func Connect(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	for i := 1; i <= 30; i++ {
		if err := sqlDB.Ping(); err == nil {
			log.Info("mysql connected",
				zap.String("host", cfg.MySQL.Host),
				zap.String("database", cfg.MySQL.Database),
			)
			return db, nil
		}
		log.Warn("mysql not ready, retrying", zap.Int("attempt", i))
		time.Sleep(2 * time.Second)
	}
	// Ping 全部失败：关闭已建立的连接池，避免资源泄漏
	_ = sqlDB.Close()
	return nil, fmt.Errorf("mysql unreachable after retries")
}
