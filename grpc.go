package d_ratelimiter

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
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

func (r *RateLimiter) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// lua 脚本返回 bool 值，判断是否限流
		allow, err := r.allow(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}
		if !allow {
			err = LimitedErr
			return
		}

		return handler(ctx, req)
	}
}

func (r *RateLimiter) allow(ctx context.Context, key string) (bool, error) {
	if r.key != "" {
		key = r.key
	}

	return r.client.Eval(ctx, luaRateLimiterWindows, []string{key},
		time.Now().Add(-r.duration).UnixMicro(), r.rate, time.Now().UnixMicro(), r.duration.Seconds()).
		Bool()
}
