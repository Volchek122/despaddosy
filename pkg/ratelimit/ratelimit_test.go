package ratelimit

import (
	"context"
	"testing"
)

func TestMemoryLimiter(t *testing.T) {
	limiter := NewMemoryLimiter()
	ctx := context.Background()
	allowed, _, err := limiter.Allow(ctx, "key", 1, 1)
	if err != nil || !allowed {
		t.Fatal("expected first request allowed")
	}
	allowed, _, _ = limiter.Allow(ctx, "key", 1, 1)
	if allowed {
		t.Fatal("expected second request blocked")
	}
}
