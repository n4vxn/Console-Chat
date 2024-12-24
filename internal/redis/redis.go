package redis

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client
var ctx = context.Background()

func InitRedis() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", 
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}
	log.Println("Connected to Redis")
}

func PubUserEvent(channel string, message string) {
	if rdb == nil {
		log.Fatalf("Redis client is not initialized")
	}
	err := rdb.Publish(ctx, channel, message).Err()
	if err != nil {
		log.Printf("Failed to publish user event: %v", err)
	}
}
