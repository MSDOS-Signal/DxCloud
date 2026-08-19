// Package redisx 提供 Redis 客户端构造（带重试）。
package redisx

import (
	"context"
	"fmt"
	"time"

	"github.com/dxcloud/cloud-api/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Connect 建立 Redis 连接，带重试（等待 compose 依赖 healthy），最多 30 次 × 2s。
func Connect(cfg *config.Config, log *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
	})

	for i := 1; i <= 30; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err := client.Ping(ctx).Err()
		cancel()
		if err == nil {
			log.Info("redis connected", zap.String("addr", cfg.Redis.Host+":"+cfg.Redis.Port))
			return client, nil
		}
		log.Warn("redis not ready, retrying", zap.Int("attempt", i))
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("redis unreachable after retries")
}
