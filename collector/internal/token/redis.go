package token

import (
	"clicky-go-collector/internal/config"
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisValidator serves active-token mappings from Redis and falls back to
// PostgreSQL on a miss or a Redis error. The fallback makes Redis an
// acceleration layer rather than a single point of failure.
type RedisValidator struct {
	client   *redis.Client
	fallback Validator
	ttl      time.Duration
}

func NewRedisValidator(
	ctx context.Context,
	cfg config.Redis,
	fallback Validator,
) (*RedisValidator, error) {
	options, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse Redis URL: %w", err)
	}

	client := redis.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("connect to Redis: %w", err)
	}

	return &RedisValidator{client: client, fallback: fallback, ttl: cfg.TokenCacheTTL}, nil
}

func (v *RedisValidator) Validate(ctx context.Context, value string) (string, error) {
	siteID, err := v.client.Get(ctx, redisTokenKey(value)).Result()
	if err == nil {
		return siteID, nil
	}

	// A Redis error must not prevent the PostgreSQL source of truth from
	// validating a legitimate browser event.
	siteID, fallbackErr := v.fallback.Validate(ctx, value)
	if fallbackErr != nil {
		return "", fallbackErr
	}

	if err := v.client.Set(ctx, redisTokenKey(value), siteID, v.ttl).Err(); err != nil {
		return siteID, nil
	}

	return siteID, nil
}

func (v *RedisValidator) Ready(ctx context.Context) error {
	if ready, ok := v.fallback.(Readiness); ok {
		return ready.Ready(ctx)
	}

	return nil
}

func (v *RedisValidator) Close() error {
	return v.client.Close()
}

func redisTokenKey(value string) string {
	digest := sha256.Sum256([]byte(value))
	return fmt.Sprintf("collector:token:%x", digest)
}
