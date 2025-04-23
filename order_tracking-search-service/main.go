package main

// import (
// 	"e-commerce/logger"
// 	"e-commerce/protobuf/protobuf"

// 	"net"
// 	"net/http"

// 	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
// 	"github.com/prometheus/client_golang/prometheus/promhttp"
// 	"google.golang.org/grpc"
// )

// func main() {
// 	// Initialize the logger
// 	logger.InitLogger("order_tracking-service")

// 	// Start gRPC server listener
// 	listener, err := net.Listen("tcp", ":50070")
// 	if err != nil {
// 		logger.Logger.Fatalf("Failed to listen on port 50070: %v", err)
// 	}

// 	// Create a gRPC server with Prometheus interceptors
// 	grpcServer := grpc.NewServer(
// 		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
// 		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
// 	)

// 	// Register Prometheus metrics for gRPC
// 	grpc_prometheus.Register(grpcServer)

// 	// Register the OrderTrackingService
// 	protobuf.RegisterOrderTrackingServiceServer(grpcServer, &.OrderTrackingServiceServer{})

// 	// Start the gRPC server in a goroutine
// 	go func() {
// 		logger.Logger.Info("Order Tracking Service running on port 50070...")
// 		if err := grpcServer.Serve(listener); err != nil {
// 			logger.Logger.Fatalf("Failed to start gRPC server: %v", err)
// 		}
// 	}()

// 	// Expose the /metrics endpoint for Prometheus
// 	http.Handle("/metrics", promhttp.Handler())
// 	logger.Logger.Info("Prometheus metrics available at /metrics on port 8018")
// 	if err := http.ListenAndServe(":8018", nil); err != nil {
// 		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
// 	}
// }
