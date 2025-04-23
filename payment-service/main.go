package main

import (
	"e-commerce/logger"
	metricsprom "e-commerce/payment-service/metrics"
	"e-commerce/payment-service/payment"
	"e-commerce/protobuf/protobuf"
	"net"

	"github.com/gin-gonic/gin"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	// Initialize the logger
	logger.InitLogger("payment-service")

	// Start gRPC server listener
	lis, err := net.Listen("tcp", ":50080")
	if err != nil {
		logger.Logger.Fatalf("Unable to connect to port 50080: %v", err)
	}

	// Connect to Kafka Producer Service
	kafkaconn, err := grpc.Dial("producer-service:50052", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Unable to connect to Kafka producer at port 50052: %v", err)
	}
	defer kafkaconn.Close()
	kafkaClient := protobuf.NewKafkaProducerServiceClient(kafkaconn)

	// Connect to User Service
	Userconn, err := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Unable to connect to User Service at port 50001: %v", err)
	}
	defer Userconn.Close()
	UserClient := protobuf.NewUserServiceClient(Userconn)

	// Create a gRPC server with Prometheus interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	// Register Prometheus metrics for gRPC
	grpc_prometheus.Register(grpcServer)

	// Register PaymentService with gRPC server
	protobuf.RegisterPaymentServiceServer(grpcServer, payment.NewPaymentService(kafkaClient, UserClient))

	// Start the gRPC server in a goroutine
	go func() {
		logger.Logger.Info("Payment Service running on port 50080...")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Initialize Prometheus metrics
	metricsprom.Init()

	// Create a Gin router for Prometheus metrics
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start the HTTP server for Prometheus metrics
	logger.Logger.Info("Prometheus metrics available at /metrics on port 8005")
	if err := router.Run(":8005"); err != nil {
		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
	}
}
