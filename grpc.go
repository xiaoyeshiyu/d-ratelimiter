package d_ratelimiter

import (
	"context"

	"google.golang.org/grpc"
)

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
