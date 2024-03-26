package d_ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

type GinResp struct {
}

func TestRateLimiter_BuildServerMiddleware(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	middleware := NewRateLimiter(rdb, "rate_limiter", time.Second*5, 2).BuildServerMiddleware()

	var count uint64
	router := gin.New()

	router.Use(middleware)
	router.GET("/", func(context *gin.Context) {
		count++
	})

	// 请求第一次
	w := PerformRequest(router, "GET", "/")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint64(1), count)

	// 间隔请求第二次
	time.Sleep(time.Second * 3)
	w = PerformRequest(router, "GET", "/")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint64(2), count)

	// 请求第三次，被限流
	w = PerformRequest(router, "GET", "/")
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Equal(t, uint64(2), count)

	// 间隔 5s 避免一个周期
	time.Sleep(time.Second * 2)
	w = PerformRequest(router, "GET", "/")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint64(3), count)
}

// PerformRequest for testing gin router.
func PerformRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
