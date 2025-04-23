package main

import (
	"e-commerce/logger"
	"e-commerce/notification-service/notification"
	"e-commerce/protobuf/protobuf"
	"net"

	"github.com/gin-gonic/gin"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"google.golang.org/grpc"
)

func main() {
	// Initialize the logger
	logger.InitLogger("notification-service")

	notification.Init()
	// Start gRPC server listener
	lis, err := net.Listen("tcp", ":50020")
	if err != nil {
		logger.Logger.Fatalf("Unable to listen on port 50020: %v", err)
	}

	// Connect to Redis Service
	conn, err := grpc.Dial("redis-service:50010", grpc.WithInsecure())
	if err != nil {
		logger.Logger.Fatalf("Unable to connect to Redis service at port 50010: %v", err)
	}
	defer conn.Close()
	redCli := protobuf.NewRedisServiceClient(conn)

	// Create gRPC server with Prometheus interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)
	grpc_prometheus.Register(grpcServer)

	// Register NotificationService with gRPC server
	protobuf.RegisterNotificationServiceServer(grpcServer, notification.NewNotificationService(redCli))

	// Start the gRPC server in a goroutine
	go func() {
		logger.Logger.Info("Notification Service running on port 50020...")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()
	logger.Logger.Info("Connected to Redis successfully")

	// Expose the /metrics endpoint for Prometheus
	router := gin.Default()
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Start the HTTP server for Prometheus metrics
	logger.Logger.Info("Prometheus metrics available at /metrics on port 8011")
	if err := router.Run(":8011"); err != nil {
		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
	}
}
