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
	// Read from environment variables (Docker / Render / Local)
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Fallback for local development
	if redisHost == "" {
		redisHost = "localhost"
	}
	if redisPort == "" {
		redisPort = "6379"
	}

	addr := redisHost + ":" + redisPort

	RedisClient = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})

	// Test the connection with a ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pong, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		log.Println("⚠️ Redis connection failed:", err)
		log.Println("⚠️ Recently viewed feature will be disabled")
		RedisClient = nil
		return
	}

	log.Println("✅ Redis connected successfully:", pong, "at", addr)
}

// CloseRedis closes the Redis connection
func CloseRedis() error {
	if RedisClient != nil {
		return RedisClient.Close()
	}
	return nil
}
