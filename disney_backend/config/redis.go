package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis initializes the Redis client connection
func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")

	if redisURL == "" {
		log.Println("⚠️ REDIS_URL not set – Redis disabled")
		return
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Println("❌ Invalid REDIS_URL:", err)
		return
	}

	RedisClient = redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := RedisClient.Ping(ctx).Err(); err != nil {
		log.Println("❌ Redis connection failed:", err)
		log.Println("⚠️ Recently viewed feature will be disabled")
		RedisClient = nil
		return
	}

	log.Println("✅ Redis connected successfully")
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
