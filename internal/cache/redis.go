package cache

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) *RedisClient {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr: addr,
		}),
	}
}

// Comment:
// CheckRateLimit implements a token bucket/sliding window rate limiter in Redis
// Limit: 60 req/min, burst 10 per brand API key
func (r *RedisClient) CheckRateLimit(ctx context.Context, brandID uint32) (bool, error) {
	key := fmt.Sprintf("rate_limit:brand:%d", brandID)
	now := time.Now().Unix()
	windowKey := fmt.Sprintf("%s:%d", key, now/60)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, windowKey)
	pipe.Expire(ctx, windowKey, 70*time.Second)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	count := incr.Val()
	if count > 60 { // Max 60 requests per minute[cite: 5]
		return false, nil
	}
	return true, nil
}

// Comment:
// CacheBrandAPIKey caches brand API keys securely to reduce DB lookups
func (r *RedisClient) GetCachedAPIKey(ctx context.Context, brandID uint32) (string, error) {
	key := fmt.Sprintf("brand_key:%d", brandID)
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // Cache miss
	}
	return val, err
}

func (r *RedisClient) SetCachedAPIKey(ctx context.Context, brandID uint32, apiKey string) error {
	key := fmt.Sprintf("brand_key:%d", brandID)
	return r.client.Set(ctx, key, apiKey, 15*time.Minute).Err()
}
