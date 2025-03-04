package redhand

import (
	"context"
	"e-commerce/protobuf/protobuf"
	redisstore "e-commerce/redis-service/redis-store"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisServer struct {
	protobuf.UnimplementedRedisServiceServer
	// note   protobuf.NotificationServiceClient
}

// func NewRedisServer(note protobuf.NotificationServiceClient) *RedisServer {
// 	return &RedisServer{note: note}
// }

func (s *RedisServer) SetData(ctx context.Context, req *protobuf.SetRequest) (*protobuf.SetResponse, error) {
	err := redisstore.Rdb.Set(ctx, req.Key, req.Value, 2*time.Hour).Err()
	if err != nil {
		return &protobuf.SetResponse{Status: "FAILED"}, err
	}

	return &protobuf.SetResponse{Status: "SUCCESS"}, nil
}

func (s *RedisServer) GetData(ctx context.Context, req *protobuf.GetRequest) (*protobuf.GetResponse, error) {
	value, err := redisstore.Rdb.Get(ctx, req.Key).Result()
	if err == redis.Nil {
		return &protobuf.GetResponse{Value: ""}, nil
	} else if err != nil {
		return nil, err
	}
	return &protobuf.GetResponse{Value: value}, nil
}
