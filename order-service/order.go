package main

import (
	"e-commerce/logger"
	"e-commerce/order-service/orderhand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "endpoint", "status"},
)

func main() {
	// Initialize the logger
	logger.InitLogger("order-service")

	// Register Prometheus metrics
	prometheus.MustRegister(httpRequestsTotal)

	// Create a Gin router
	r := gin.Default()

	// Middleware to track HTTP requests and log them
	r.Use(func(c *gin.Context) {
		logger.Logger.Infof("Incoming request: Method=%s, Path=%s", c.Request.Method, c.FullPath())
		c.Next()
		status := http.StatusText(c.Writer.Status())
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		logger.Logger.Infof("Request completed: Method=%s, Path=%s, Status=%s", c.Request.Method, c.FullPath(), status)
	})

	// Expose the /metrics endpoint for Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	logger.Logger.Info("Prometheus metrics available at /metrics")

	// Define the /order/create endpoint
	r.POST("/order/create", func(c *gin.Context) {
		logger.Logger.Info("Handling /order/create request")
		orderhand.CreateOrder(c)
	})

	// Start the HTTP server
	logger.Logger.Info("Order Service running on port 8083...")
	if err := r.Run(":8083"); err != nil {
		logger.Logger.Fatalf("Failed to start Order Service: %v", err)
	}
}
