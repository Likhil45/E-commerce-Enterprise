package main

import (
	"e-commerce/producer-service/handlers"
	"e-commerce/protobuf/protobuf"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50052")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	kafkaProducer, err := handlers.NewKafkaProducer([]string{"kafka:9092"})
	if err != nil {
		log.Fatalf("Failed to initialize Kafka producer: %v", err)
	}

	grpcServer := grpc.NewServer()
	protobuf.RegisterKafkaProducerServiceServer(grpcServer, kafkaProducer)

	log.Println("Kafka Producer Service running on port 50052...")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
