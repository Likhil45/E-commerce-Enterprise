package main

import (
	invhandler "e-commerce/inventory-service/handlers"
	"e-commerce/logger"
	"e-commerce/protobuf/protobuf"
	"net"
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Initialize the logger
	logger.InitLogger("inventory-service")

	// Connect to Database Service
	invhandler.Init()
	dbConn, err := grpc.Dial("write-db-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to Database Service: %v", err)
	}
	defer dbConn.Close()
	dbClient := protobuf.NewDatabaseServiceClient(dbConn)
	logger.Logger.Info("Connected to Database Service successfully")

	// Connect to Kafka Producer Service
	kafkaConn, err := grpc.Dial("producer-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Logger.Fatalf("Failed to connect to Kafka Producer Service: %v", err)
	}
	defer kafkaConn.Close()
	kafkaClient := protobuf.NewKafkaProducerServiceClient(kafkaConn)
	logger.Logger.Info("Connected to Kafka Producer Service successfully")

	// Start gRPC Server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Logger.Fatalf("Failed to start listener on port 50051: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	// Register Prometheus metrics for gRPC
	grpc_prometheus.Register(grpcServer)
	protobuf.RegisterInventoryServiceServer(grpcServer, invhandler.NewInventoryService(dbClient, kafkaClient))

	go func() {
		logger.Logger.Info("Inventory Service running on port 50051...")
		if err := grpcServer.Serve(listener); err != nil {
			logger.Logger.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Expose the /metrics endpoint for Prometheus
	http.Handle("/metrics", promhttp.Handler())
	logger.Logger.Info("Prometheus metrics available at /metrics on port 8013")
	if err := http.ListenAndServe(":8013", nil); err != nil {
		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
	}
}
