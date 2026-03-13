package telemetry

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// MetricsMiddleware returns a Gin middleware that records HTTP request metrics.
func MetricsMiddleware() gin.HandlerFunc {
	meter := GetMeter()

	requestDuration, _ := meter.Float64Histogram(
		"http_server_request_duration_seconds",
		metric.WithDescription("Duration of HTTP server requests in seconds"),
		metric.WithUnit("s"),
	)

	requestTotal, _ := meter.Int64Counter(
		"http_server_request_total",
		metric.WithDescription("Total number of HTTP server requests"),
	)

	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.status_code", fmt.Sprintf("%d", c.Writer.Status())),
		}

		requestDuration.Record(c.Request.Context(), duration, metric.WithAttributes(attrs...))
		requestTotal.Add(c.Request.Context(), 1, metric.WithAttributes(attrs...))
	}
}
