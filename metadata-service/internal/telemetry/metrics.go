package telemetry

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "metadata-service"

// InitMetrics sets up the OTel meter provider with a Prometheus exporter.
// Must be called before any metrics are recorded.
func InitMetrics() (*sdkmetric.MeterProvider, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	return provider, nil
}

// GetMeter returns a named meter from the global provider.
func GetMeter() metric.Meter {
	return otel.Meter(meterName)
}

// MetricsHandler returns a Gin handler that serves Prometheus metrics.
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RawMetricsHandler returns the raw http.Handler for Prometheus metrics.
func RawMetricsHandler() http.Handler {
	return promhttp.Handler()
}
