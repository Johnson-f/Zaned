package caching

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisCtx    context.Context
)

// InitRedis initializes the Redis client connection
func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	redisPassword := os.Getenv("REDIS_PASSWORD")

	// If REDIS_URL is provided, parse it
	var opts *redis.Options
	if redisURL != "" {
		parsedOpts, err := redis.ParseURL(redisURL)
		if err != nil {
			return fmt.Errorf("failed to parse REDIS_URL: %w", err)
		}
		opts = parsedOpts
	} else {
		// Fallback to individual environment variables
		host := os.Getenv("REDIS_HOST")
		port := os.Getenv("REDIS_PORT")
		if host == "" {
			host = "localhost"
		}
		if port == "" {
			port = "6379"
		}

		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: redisPassword,
			DB:       0, // Default DB
		}
	}

	// Add connection pooling and timeout settings
	opts.PoolSize = 10
	opts.MinIdleConns = 5
	opts.DialTimeout = 5 * time.Second
	opts.ReadTimeout = 3 * time.Second
	opts.WriteTimeout = 3 * time.Second
	opts.PoolTimeout = 4 * time.Second

	redisClient = redis.NewClient(opts)
	redisCtx = context.Background()

	// Test connection
	ctx, cancel := context.WithTimeout(redisCtx, 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Println("Redis connection established successfully")
	return nil
}

// GetRedisClient returns the Redis client instance
func GetRedisClient() *redis.Client {
	return redisClient
}

// GetRedisContext returns the Redis context
func GetRedisContext() context.Context {
	return redisCtx
}

// CloseRedis closes the Redis connection gracefully
func CloseRedis() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}
