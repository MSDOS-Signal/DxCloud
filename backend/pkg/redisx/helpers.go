// Package redisx：Redis 客户端构造与常用原语封装。
// 仅使用 Redis 3.0 兼容原语（SCAN/INCR/EXPIRE/SET/HASH），
// 保证开发用 Redis 3.0.504 与生产 redis:7 均可运行。
package redisx

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func Set(ctx context.Context, r *redis.Client, key, value string, ttl time.Duration) error {
	return r.Set(ctx, key, value, ttl).Err()
}

func Get(ctx context.Context, r *redis.Client, key string) (string, error) {
	return r.Get(ctx, key).Result()
}

func Del(ctx context.Context, r *redis.Client, keys ...string) error {
	return r.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, r *redis.Client, key string) (bool, error) {
	n, err := r.Exists(ctx, key).Result()
	return n > 0, err
}

func Incr(ctx context.Context, r *redis.Client, key string) (int64, error) {
	return r.Incr(ctx, key).Result()
}

func Expire(ctx context.Context, r *redis.Client, key string, ttl time.Duration) error {
	return r.Expire(ctx, key, ttl).Err()
}

func TTL(ctx context.Context, r *redis.Client, key string) (time.Duration, error) {
	return r.TTL(ctx, key).Result()
}

func HSet(ctx context.Context, r *redis.Client, key string, fields map[string]any) error {
	return r.HSet(ctx, key, fields).Err()
}

func HGetAll(ctx context.Context, r *redis.Client, key string) (map[string]string, error) {
	return r.HGetAll(ctx, key).Result()
}

func HGet(ctx context.Context, r *redis.Client, key, field string) (string, error) {
	return r.HGet(ctx, key, field).Result()
}

// ScanKeys 按 pattern 扫描全部 key（Redis 3.0 兼容，避免 KEYS 阻塞）。
func ScanKeys(ctx context.Context, r *redis.Client, pattern string) ([]string, error) {
	var out []string
	var cursor uint64
	for {
		keys, next, err := r.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return nil, err
		}
		out = append(out, keys...)
		cursor = next
		if cursor == 0 {
			return out, nil
		}
	}
}

// HGetAllAndDel 原子读取并删除 hash key（Lua 脚本保证 HGETALL+DEL 原子性，兼容 Redis 3.0+）。
// 用于一次性令牌消费：避免 HGetAll 与 Del 之间的并发窗口导致 token 被复用。
func HGetAllAndDel(ctx context.Context, r *redis.Client, key string) (map[string]string, error) {
	script := `local d = redis.call('HGETALL', KEYS[1])
redis.call('DEL', KEYS[1])
return d`
	res, err := r.Eval(ctx, script, []string{key}).Result()
	if err != nil {
		return nil, err
	}
	arr, ok := res.([]interface{})
	if !ok {
		return nil, errors.New("unexpected script result type")
	}
	m := make(map[string]string, len(arr)/2)
	for i := 0; i+1 < len(arr); i += 2 {
		field, _ := arr[i].(string)
		value, _ := arr[i+1].(string)
		m[field] = value
	}
	return m, nil
}
