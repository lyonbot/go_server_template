package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

// PrometheusMiddleware 返回一个 Gin 中间件，用于收集 HTTP 指标
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown" // 对于未匹配的路由
		}

		// 请求开始时，活跃请求数+1
		HTTPActiveRequests.Inc()

		// 记录请求大小
		HTTPRequestSize.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(float64(c.Request.ContentLength))

		// 处理请求
		c.Next()

		// 请求结束时，活跃请求数-1
		HTTPActiveRequests.Dec()

		// 记录状态码
		status := c.Writer.Status()

		// 记录请求总数
		HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			string(rune(status)),
		).Inc()

		// 记录请求延迟
		HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(time.Since(start).Seconds())

		// 记录响应大小
		HTTPResponseSize.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(float64(c.Writer.Size()))
	}
}
