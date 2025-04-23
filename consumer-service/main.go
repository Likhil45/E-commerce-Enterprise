package main

import (
	"context"
	"e-commerce/consumer-service/consmetrics"
	"e-commerce/consumer-service/consumer"
	"e-commerce/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var ActiveConsumerSessions = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "kafka_consumer_active_sessions",
		Help: "Number of active Kafka consumer sessions",
	},
)

func main() {
	// Initialize the logger
	logger.InitLogger("consumer-service")

	brokers := []string{"kafka:9092"}
	groupID := "ecommerce-consumer-group"
	topics := []string{"OrderCreated", "OutOfStock", "InventoryReserved", "PaymentProcessed", "OrderConfirmed", "PaymentFailed"}

	// Initialize the Kafka consumer
	consumerService, err := consumer.NewKafkaConsumer(brokers, groupID, topics)
	if err != nil {
		logger.Logger.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Logger.Info("Shutting down consumer...")
		cancel()
	}()

	// Increment consumer restarts and active sessions
	consmetrics.Init()

	// Start the consumer
	go func() {
		logger.Logger.Info("Starting Kafka Consumer...")
		consumerService.ProcessMessage(ctx)
		logger.Logger.Info("Kafka Consumer stopped")
	}()

	// Expose the /metrics endpoint for Prometheus
	http.Handle("/metrics", promhttp.Handler())
	logger.Logger.Info("Prometheus metrics available at /metrics on port 8017")
	if err := http.ListenAndServe(":8017", nil); err != nil {
		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
	}
}
