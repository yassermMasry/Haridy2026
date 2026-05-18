package middleware

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var requestCount uint64
var errorCount uint64
var slowRequestCount uint64
var totalLatencyMicros uint64

func Observability() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		atomic.AddUint64(&requestCount, 1)
		traceID := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Header("X-Trace-ID", traceID)
		c.Set("trace_id", traceID)
		c.Next()
		elapsed := time.Since(start)
		if c.Writer.Status() >= 500 {
			atomic.AddUint64(&errorCount, 1)
		}
		if elapsed > 750*time.Millisecond {
			atomic.AddUint64(&slowRequestCount, 1)
		}
		atomic.AddUint64(&totalLatencyMicros, uint64(elapsed.Microseconds()))
		c.Header("X-Response-Time", elapsed.String())
	}
}

func MetricsText() string {
	requests := atomic.LoadUint64(&requestCount)
	return fmt.Sprintf("# HELP haridy_http_requests_total Total HTTP requests\n# TYPE haridy_http_requests_total counter\nharidy_http_requests_total %d\n# HELP haridy_http_errors_total Total HTTP 5xx responses\n# TYPE haridy_http_errors_total counter\nharidy_http_errors_total %d\n# HELP haridy_http_slow_requests_total Requests slower than 750ms\n# TYPE haridy_http_slow_requests_total counter\nharidy_http_slow_requests_total %d\n# HELP haridy_http_latency_average_microseconds Average HTTP latency\n# TYPE haridy_http_latency_average_microseconds gauge\nharidy_http_latency_average_microseconds %.2f\n", requests, atomic.LoadUint64(&errorCount), atomic.LoadUint64(&slowRequestCount), averageLatency(requests))
}

func averageLatency(requests uint64) float64 {
	if requests == 0 {
		return 0
	}
	return float64(atomic.LoadUint64(&totalLatencyMicros)) / float64(requests)
}
