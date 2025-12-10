package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
	)
)

// Metrics returns a middleware that collects Prometheus metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip metrics endpoint itself
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		httpRequestsInFlight.Inc()
		start := time.Now()

		c.Next()

		httpRequestsInFlight.Dec()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())

		// Normalize path for metrics (avoid high cardinality)
		path := normalizePath(c.FullPath())
		if path == "" {
			path = c.Request.URL.Path
		}

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
	}
}

// normalizePath normalizes the path to avoid high cardinality in metrics
func normalizePath(path string) string {
	if path == "" {
		return "unknown"
	}
	return path
}

// MetricsWithPrefix returns metrics middleware with custom metric prefix
func MetricsWithPrefix(prefix string) gin.HandlerFunc {
	requestsTotal := promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: prefix + "_http_requests_total",
			Help: "Total number of HTTP requests for " + prefix,
		},
		[]string{"method", "path", "status"},
	)

	requestDuration := promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prefix + "_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds for " + prefix,
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		path := normalizePath(c.FullPath())
		if path == "" {
			path = c.Request.URL.Path
		}

		requestsTotal.WithLabelValues(c.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path, status).Observe(duration)
	}
}
