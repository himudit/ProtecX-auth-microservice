package config

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func ConnectRedis() {
	// Load environment variables from .env
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		panic("REDIS_URL is not set")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		panic("Failed to parse REDIS_URL: " + err.Error())
	}

	RDB = redis.NewClient(opt)

	// Test connection
	_, err = RDB.Ping(context.Background()).Result()
	if err != nil {
		panic("Redis connection failed: " + err.Error())
	}

	fmt.Println("✅ Connected to Upstash Redis")
}
