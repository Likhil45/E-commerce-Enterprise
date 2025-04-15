package main

import (
	"e-commerce/payment-service/payment"
	"e-commerce/protobuf/protobuf"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50080")
	if err != nil {
		log.Println("Unable to connect to port 50080", err)
	}
	kafkaconn, err1 := grpc.Dial("producer-service:50052", grpc.WithInsecure())
	if err1 != nil {
		log.Println("unable to connect to Kafka producer at port 50052", err)
	}
	defer kafkaconn.Close()
	kafkaClient := protobuf.NewKafkaProducerServiceClient(kafkaconn)

	Userconn, err1 := grpc.Dial("write-db-service:50001", grpc.WithInsecure())
	if err1 != nil {
		log.Println("unable to connect to User Service at port 50001", err)
	}
	defer Userconn.Close()
	UserClient := protobuf.NewUserServiceClient(Userconn)

	grpcServer := grpc.NewServer()
	protobuf.RegisterPaymentServiceServer(grpcServer, payment.NewPaymentService(kafkaClient, UserClient))
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to start gRPC server: %v", err)
	}

}
