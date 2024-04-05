package rate_limiter

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (r *RateLimiter) BuildServerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// lua 脚本返回 bool 值，判断是否限流
		allow, err := r.allow(c.Request.Context(), c.Request.RequestURI)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		if !allow {
			c.AbortWithError(http.StatusTooManyRequests, LimitedErr)
		}
		c.Next()
	}
}
