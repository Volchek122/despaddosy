package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter interface {
	Allow(ctx context.Context, key string, rps, burst int) (bool, int, error)
}

type MemoryLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

func NewMemoryLimiter() *MemoryLimiter {
	return &MemoryLimiter{buckets: make(map[string]*bucket)}
}

func (l *MemoryLimiter) Allow(_ context.Context, key string, rps, burst int) (bool, int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(burst), lastCheck: time.Now()}
		l.buckets[key] = b
	}
	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * float64(rps)
	if b.tokens > float64(burst) {
		b.tokens = float64(burst)
	}
	b.lastCheck = now
	if b.tokens < 1 {
		return false, int(b.tokens), nil
	}
	b.tokens -= 1
	return true, int(b.tokens), nil
}

type RedisLimiter struct {
	client *redis.Client
}

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{client: client}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string, rps, burst int) (bool, int, error) {
	if rps <= 0 {
		return true, burst, nil
	}
	window := time.Second
	if rps > 0 {
		window = time.Second
	}
	pipe := l.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}
	count := int(incr.Val())
	if count > burst {
		return false, burst - count, nil
	}
	return true, burst - count, nil
}
