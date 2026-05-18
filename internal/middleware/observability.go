package middleware

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var requestCount uint64
var errorCount uint64

func Observability() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		atomic.AddUint64(&requestCount, 1)
		traceID := fmt.Sprintf("%d", time.Now().UnixNano())
		c.Header("X-Trace-ID", traceID)
		c.Set("trace_id", traceID)
		c.Next()
		if c.Writer.Status() >= 500 {
			atomic.AddUint64(&errorCount, 1)
		}
		c.Header("X-Response-Time", time.Since(start).String())
	}
}

func MetricsText() string {
	return fmt.Sprintf("# HELP haridy_http_requests_total Total HTTP requests\n# TYPE haridy_http_requests_total counter\nharidy_http_requests_total %d\n# HELP haridy_http_errors_total Total HTTP 5xx responses\n# TYPE haridy_http_errors_total counter\nharidy_http_errors_total %d\n", atomic.LoadUint64(&requestCount), atomic.LoadUint64(&errorCount))
}
