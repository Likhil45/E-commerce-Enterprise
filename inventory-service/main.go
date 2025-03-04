package main

import (
	invhandler "e-commerce/inventory-service/handlers"
	"fmt"
	"log"
	"net"

	"e-commerce/protobuf/protobuf"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to Database Service
	dbConn, err := grpc.Dial("localhost:50001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Database Service: %v", err)
	}
	defer dbConn.Close()
	dbClient := protobuf.NewDatabaseServiceClient(dbConn)

	// Connect to Kafka Producer Service
	kafkaConn, err := grpc.Dial("localhost:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Kafka Producer Service: %v", err)
	}
	defer kafkaConn.Close()
	kafkaClient := protobuf.NewKafkaProducerServiceClient(kafkaConn)

	// Start gRPC Server
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to start listener: %v", err)
	}

	grpcServer := grpc.NewServer()
	protobuf.RegisterInventoryServiceServer(grpcServer, invhandler.NewInventoryService(dbClient, kafkaClient))

	fmt.Println("Inventory Service running on port 50051...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}
}
