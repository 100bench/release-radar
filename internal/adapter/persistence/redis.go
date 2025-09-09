package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/mackb/releaseradar/pkg/idempotency"
	"github.com/redis/go-redis/v9"
)

type RedisIdempotencyStorage struct {
	client *redis.Client
}

func NewRedisIdempotencyStorage(redisClient *redis.Client) idempotency.Storage {
	return &RedisIdempotencyStorage{client: redisClient}
}

func (r *RedisIdempotencyStorage) CheckAndSet(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	// SETNX (Set if Not Exists) returns 1 if the key was set, 0 if it already existed.
	set, err := r.client.SetNX(ctx, key, true, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx key in redis: %w", err)
	}
	return set, nil
}
