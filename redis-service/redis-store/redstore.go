package redisstore

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client
var ctx = context.Background()

func RedInit() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password by default
		DB:       0,  // Default DB
	})

	// Test connection
	_, err := Rdb.Ping(ctx).Result()
	if err != nil {
		log.Printf("\nFailed to connect to Redis: %v", err)
	}
	fmt.Println("Connected to Redis")
}
