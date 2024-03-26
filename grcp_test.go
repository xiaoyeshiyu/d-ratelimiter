//go:build e2e

package d_ratelimiter

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type Resp struct {
}

func TestRateLimiter_BuildServerInterceptor(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	interceptor := NewRateLimiter(rdb, "rate_limiter", time.Second*5, 2).BuildServerInterceptor()

	var count uint64
	handler := func(ctx context.Context, req any) (resp any, err error) {
		count++
		return &Resp{}, nil
	}

	// 请求第一次
	resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &Resp{}, resp)
	assert.Equal(t, uint64(1), count)

	// 请求第二次
	resp, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &Resp{}, resp)
	assert.Equal(t, uint64(2), count)

	// 请求第三次，被限流
	resp, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	require.Equal(t, err, LimitedErr)
	assert.Nil(t, resp)
	assert.Equal(t, uint64(2), count)

	// 休眠 5s 避免一个周期
	time.Sleep(time.Second * 5)
	resp, err = interceptor(context.Background(), nil, &grpc.UnaryServerInfo{}, handler)
	require.NoError(t, err)
	assert.Equal(t, &Resp{}, resp)
	assert.Equal(t, uint64(3), count)
}
