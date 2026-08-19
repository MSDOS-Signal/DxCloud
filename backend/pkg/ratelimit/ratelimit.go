// Package ratelimit 提供基于 Redis 固定窗口计数器的限流器。
package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) *Limiter {
	return &Limiter{rdb: rdb}
}

// Allow 固定窗口限流：key 在 window 内最多允许 limit 次。
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	n, err := l.rdb.Incr(ctx, key).Result()
	if err != nil {
		return false, err
	}
	if n == 1 {
		_ = l.rdb.Expire(ctx, key, window).Err()
	}
	return n <= int64(limit), nil
}
