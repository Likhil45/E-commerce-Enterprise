package main

import (
	"e-commerce/protobuf/protobuf"
	redhand "e-commerce/redis-service/redis"
	redisstore "e-commerce/redis-service/redis-store"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	redisstore.RedInit()
	defer redisstore.Rdb.Close()

	lsn, err := net.Listen("tcp", ":50010")
	if err != nil {
		log.Println("Unable to listen to 50010 port", err)
	}

	grpcServer := grpc.NewServer()
	protobuf.RegisterRedisServiceServer(grpcServer, &redhand.RedisServer{})

	err1 := grpcServer.Serve(lsn)
	if err1 != nil {
		log.Println("Unable to Serve ", err1)
	}

}
