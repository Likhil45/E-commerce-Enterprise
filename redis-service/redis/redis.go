package redhand

import (
	"context"
	"e-commerce/logger"
	"e-commerce/protobuf/protobuf"
	redisstore "e-commerce/redis-service/redis-store"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

var RedisOpsErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_operations_errors_total",
		Help: "Total number of Redis operation errors",
	},
	[]string{"operation"},
)

func init() {
	prometheus.MustRegister(RedisOpsErrors)
}

type RedisServer struct {
	protobuf.UnimplementedRedisServiceServer
}

func (s *RedisServer) SetData(ctx context.Context, req *protobuf.SetRequest) (*protobuf.SetResponse, error) {
	redisstore.RedisOpsTotal.WithLabelValues("set").Inc()

	logger.Logger.Infof("SetData called with Key=%s, Value=%s", req.Key, req.Value)

	err := redisstore.Rdb.Set(ctx, req.Key, req.Value, 2*time.Hour).Err()
	if err != nil {
		RedisOpsErrors.WithLabelValues("set").Inc() // Increment error counter for SetData

		logger.Logger.Errorf("Failed to set data in Redis: Key=%s, Error=%v", req.Key, err)
		return &protobuf.SetResponse{Status: "FAILED"}, err
	}

	logger.Logger.Infof("Data successfully set in Redis: Key=%s, Value=%s", req.Key, req.Value)
	return &protobuf.SetResponse{Status: "SUCCESS"}, nil
}

func (s *RedisServer) GetData(ctx context.Context, req *protobuf.GetRequest) (*protobuf.GetResponse, error) {
	redisstore.RedisOpsTotal.WithLabelValues("get").Inc()

	logger.Logger.Infof("GetData called with Key=%s", req.Key)

	value, err := redisstore.Rdb.Get(ctx, req.Key).Result()
	if err == redis.Nil {
		RedisOpsErrors.WithLabelValues("get").Inc() // Increment error counter for GetData

		logger.Logger.Warnf("Key not found in Redis: Key=%s", req.Key)
		return &protobuf.GetResponse{Value: ""}, nil
	} else if err != nil {
		RedisOpsErrors.WithLabelValues("get").Inc() // Increment error counter for GetData

		logger.Logger.Errorf("Failed to get data from Redis: Key=%s, Error=%v", req.Key, err)
		return nil, err
	}

	logger.Logger.Infof("Data successfully retrieved from Redis: Key=%s, Value=%s", req.Key, value)
	return &protobuf.GetResponse{Value: value}, nil
}
