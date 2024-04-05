package rate_limiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/sliding_window.lua
var luaRateLimiterWindows string

type RateLimiter struct {
	client   redis.Cmdable
	key      string
	duration time.Duration
	rate     uint64
}

func NewRateLimiter(client redis.Cmdable, key string, duration time.Duration, rate uint64) *RateLimiter {
	return &RateLimiter{
		client:   client,
		key:      key,
		duration: duration,
		rate:     rate,
	}
}

func (r *RateLimiter) allow(ctx context.Context, key string) (bool, error) {
	if r.key != "" {
		key = r.key
	}

	return r.client.Eval(ctx, luaRateLimiterWindows, []string{key},
		time.Now().Add(-r.duration).UnixMicro(), r.rate, time.Now().UnixMicro(), r.duration.Milliseconds()).
		Bool()
}
