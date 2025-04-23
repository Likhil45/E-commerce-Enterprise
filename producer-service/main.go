package main

import (
	"e-commerce/logger"
	"e-commerce/producer-service/handlers"
	"e-commerce/protobuf/protobuf"
	"net"

	"github.com/gin-gonic/gin"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	// Initialize the logger
	logger.InitLogger("producer-service")

	// Initialize Prometheus metrics
	handlers.Init()

	// Start gRPC server
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		logger.Logger.Fatalf("Failed to listen on port 50052: %v", err)
	}

	// Initialize Kafka producer
	kafkaProducer, err := handlers.NewKafkaProducer([]string{"kafka:9092"})
	if err != nil {
		logger.Logger.Fatalf("Failed to initialize Kafka producer: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	// Register Prometheus metrics for gRPC
	grpc_prometheus.Register(grpcServer)

	// Register KafkaProducerService with gRPC server
	protobuf.RegisterKafkaProducerServiceServer(grpcServer, kafkaProducer)

	// Start the gRPC server in a goroutine
	go func() {
		logger.Logger.Info("Kafka Producer Service running on port 50052...")
		if err := grpcServer.Serve(listener); err != nil {
			logger.Logger.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Start HTTP server for Prometheus metrics
	r := gin.Default()
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	logger.Logger.Info("Prometheus metrics available at /metrics on port 8007")
	if err := r.Run(":8007"); err != nil {
		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
	}
}
