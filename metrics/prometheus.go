package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	AIRequestCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_requests_total",
			Help: "AI请求总数",
		},
		[]string{"task", "model", "provider", "status"},
	)

	AIRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_request_duration_seconds",
			Help:    "AI内容耗时(秒)",
			Buckets: []float64{0.5, 1.5, 2, 4, 6, 8, 10, 12, 14, 18, 22, 30, 35, 40, 50, 60},
		},
		[]string{"task", "model", "provider"},
	)

	// // 使用 Summary 替代 Histogram (如果只需要知道百分位，不需要完整的分布)
	// httpRequestLatency = promauto.NewSummaryVec(
	// 	prometheus.SummaryOpts{
	// 		Name:       "http_request_latency_seconds",
	// 		Help:       "Latency of HTTP requests in seconds",
	// 		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}, // 50th, 90th, and 99th percentiles
	// 	},
	// 	[]string{"path", "method"},
	// )

	// HTTP 相关指标
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求延迟分布(秒)",
			Buckets: []float64{0.1, 0.3, 0.5, 0.7, 1, 1.5, 2, 3, 5, 10},
		},
		[]string{"method", "path"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP请求大小分布(字节)",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8), // 从100字节开始，×10倍数增长，8个桶
		},
		[]string{"method", "path"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP响应大小分布(字节)",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path"},
	)

	HTTPActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "当前活跃的HTTP请求数",
		},
	)
)

type Labels prometheus.Labels

// 用于 http.Handle("/metrics", promhttp.Handler())
func GetHttpHandler() http.Handler {
	return promhttp.Handler()
}
