package middleware

import (
	"time"

	"github.com/chilljzz/gohub/internal/metrics"
	"github.com/gin-gonic/gin"
)

func PrometheusMetrics(
	appMetrics *metrics.Metrics,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		start := time.Now()

		ctx.Next()

		route := ctx.FullPath()

		if route == "" {
			route = "unmatched"
		}
		if route == "/ping" {
			return
		}
		if route == "/metrics" {
			return
		}
		appMetrics.ObserveHTTPRequest(
			ctx.Request.Method,
			route,
			ctx.Writer.Status(),
			time.Since(start),
		)
	}
}
