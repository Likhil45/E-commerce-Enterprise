package main

import (
	"e-commerce/logger"
	"e-commerce/protobuf/protobuf"
	redhand "e-commerce/redis-service/redis"
	redisstore "e-commerce/redis-service/redis-store"
	"net"
	"net/http"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Initialize the logger
	logger.InitLogger("redis-service")

	// Initialize Redis
	redisstore.RedInit()
	defer redisstore.Rdb.Close()

	// Start gRPC server
	lsn, err := net.Listen("tcp", ":50010")
	if err != nil {
		logger.Logger.Fatalf("Unable to listen on port 50010: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	)

	// Register Prometheus metrics for gRPC
	grpc_prometheus.Register(grpcServer)
	protobuf.RegisterRedisServiceServer(grpcServer, &redhand.RedisServer{})
	reflection.Register(grpcServer)

	// Start the gRPC server in a goroutine
	go func() {
		logger.Logger.Info("Redis Service running on port 50010...")
		if err := grpcServer.Serve(lsn); err != nil {
			logger.Logger.Fatalf("Failed to serve gRPC server: %v", err)
		}
	}()

	// Expose the /metrics endpoint for Prometheus
	http.Handle("/metrics", promhttp.Handler())
	logger.Logger.Info("Prometheus metrics available at /metrics on port 8015")
	if err := http.ListenAndServe(":8015", nil); err != nil {
		logger.Logger.Fatalf("Failed to start metrics server: %v", err)
	}
}
