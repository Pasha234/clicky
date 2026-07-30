package ratelimit

import (
	"clicky-go-collector/internal/config"
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const window = time.Minute

type Limiter interface {
	Allow(ctx context.Context, token, ip string) (Decision, error)
	Ready(ctx context.Context) error
	Close() error
}

type Decision struct {
	Allowed           bool
	ExceededDimension string
}

type RedisLimiter struct {
	client *redis.Client
	limit  int
}

func NewRedisLimiter(ctx context.Context, cfg config.Redis, limit int) (*RedisLimiter, error) {
	options, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}

	return &RedisLimiter{client: client, limit: limit}, nil
}

func (l *RedisLimiter) Allow(ctx context.Context, token, ip string) (Decision, error) {
	bucket := time.Now().UTC().Format("200601021504")
	keys := []string{
		limiterKey("token", token, bucket),
		limiterKey("ip", ip, bucket),
	}

	pipe := l.client.TxPipeline()
	counts := make([]*redis.IntCmd, len(keys))
	for i, key := range keys {
		counts[i] = pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return Decision{}, fmt.Errorf("increment rate-limit counters: %w", err)
	}

	if counts[0].Val() > int64(l.limit) {
		return Decision{ExceededDimension: "token"}, nil
	}
	if counts[1].Val() > int64(l.limit) {
		return Decision{ExceededDimension: "ip"}, nil
	}

	return Decision{Allowed: true}, nil
}

func (l *RedisLimiter) Ready(ctx context.Context) error {
	if err := l.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis is not ready: %w", err)
	}

	return nil
}

func (l *RedisLimiter) Close() error {
	return l.client.Close()
}

func limiterKey(kind, value, bucket string) string {
	digest := sha256.Sum256([]byte(value))
	return fmt.Sprintf("collector:rate-limit:%s:%x:%s", kind, digest, bucket)
}

type AllowAll struct{}

func (AllowAll) Allow(context.Context, string, string) (Decision, error) {
	return Decision{Allowed: true}, nil
}
func (AllowAll) Ready(context.Context) error { return nil }
func (AllowAll) Close() error                { return nil }
