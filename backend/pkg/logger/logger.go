// Package logger 提供 zap 结构化 JSON 日志（日志规范见 docs/phase0-architecture.md §42）。
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New 按级别创建生产级 JSON logger。
func New(level string) *zap.Logger {
	lvl := zapcore.InfoLevel
	_ = lvl.UnmarshalText([]byte(level))

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
