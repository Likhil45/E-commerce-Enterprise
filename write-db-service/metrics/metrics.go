package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Total number of HTTP requests
	RequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Duration of HTTP requests
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response times for HTTP requests",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Total number of errors
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of errors",
		},
		[]string{"method", "endpoint", "error"},
	)
	// Database query duration
	DBQueryDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Histogram of database query durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query"},
	)

	// gRPC call duration
	GRPCCallDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_call_duration_seconds",
			Help:    "Histogram of gRPC call durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

func InitMetrics() {
	// Register metrics with Prometheus
	prometheus.MustRegister(RequestsTotal)
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ErrorsTotal)
	prometheus.MustRegister(DBQueryDuration)
	prometheus.MustRegister(GRPCCallDuration)
}
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process the request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		method := c.Request.Method
		endpoint := c.FullPath()

		RequestsTotal.WithLabelValues(method, endpoint, http.StatusText(status)).Inc()
		RequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		if status >= 400 {
			ErrorsTotal.WithLabelValues(method, endpoint, http.StatusText(status)).Inc()
		}
	}
}
