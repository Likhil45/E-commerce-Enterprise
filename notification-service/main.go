package main

import (
	"e-commerce/notification-service/notification"
	"e-commerce/protobuf/protobuf"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":50020")
	if err != nil {
		log.Println("Unable to Serve at port 5020", err)
	}
	conn, err := grpc.Dial("redis-service:50010", grpc.WithInsecure())
	if err != nil {
		log.Println("Unable to dial to 50010 -redis", err)
	}
	defer conn.Close()
	redCli := protobuf.NewRedisServiceClient(conn)

	grpcServer := grpc.NewServer()
	protobuf.RegisterNotificationServiceServer(grpcServer, notification.NewNotificationService(redCli))

	err1 := grpcServer.Serve(lis)
	if err1 != nil {
		log.Println("Unable to Serve ", err1)
	}

	fmt.Println("Connected to Redis")

}
