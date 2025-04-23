package redisstore

import (
	"context"
	"e-commerce/logger"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var ctx = context.Background()

var RedisOpsTotal = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "redis_operations_total",
		Help: "Total number of Redis operations",
	},
	[]string{"operation"},
)

func RedInit() {
	logger.Logger.Info("Initializing Redis client...")

	Rdb = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "", // No password by default
		DB:       0,  // Default DB
	})

	// Test connection
	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		logger.Logger.Errorf("Failed to connect to Redis: %v", err)
		return
	}

	logger.Logger.Info("Connected to Redis successfully")
}
