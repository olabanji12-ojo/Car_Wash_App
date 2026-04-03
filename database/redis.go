package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// ConnectRedis initializes the Redis connection
func ConnectRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URI")
	if redisURL == "" {
		redisURL = "localhost:6379"
		log.Println("⚠️ REDIS_URI not found in .env, defaulting to localhost:6379")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Ping the Redis server to verify connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Printf("❌ Redis connection failed: %v", err)
		return nil
	}

	RedisClient = client
	fmt.Println("✅ Redis connected successfully")
	return RedisClient
}

// GetRedisKey generates a consistent key for caching
func GetRedisKey(prefix, id string) string {
	return fmt.Sprintf("%s:%s", prefix, id)
}
