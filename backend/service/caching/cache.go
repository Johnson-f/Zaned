package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService provides caching operations
type CacheService struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheService creates a new cache service instance
func NewCacheService() *CacheService {
	return &CacheService{
		client: GetRedisClient(),
		ctx:    GetRedisContext(),
	}
}

// Get retrieves a value from cache by key
func (c *CacheService) Get(key string) ([]byte, error) {
	if c.client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	val, err := c.client.Get(c.ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss, not an error
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get from cache: %w", err)
	}

	return val, nil
}

// GetJSON retrieves and unmarshals a JSON value from cache
func (c *CacheService) GetJSON(key string, dest interface{}) (bool, error) {
	data, err := c.Get(key)
	if err != nil {
		return false, err
	}
	if data == nil {
		log.Printf("[CACHE MISS] Key: %s", key)
		return false, nil // Cache miss
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	log.Printf("[CACHE HIT] Key: %s", key)
	return true, nil
}

// Set stores a value in cache with TTL
// If ttl is 0, the key is stored permanently (no expiration)
func (c *CacheService) Set(key string, value []byte, ttl time.Duration) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// If TTL is 0, store permanently (no expiration)
	// If TTL is negative, don't cache
	if ttl < 0 {
		return nil
	}

	err := c.client.Set(c.ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// SetJSON marshals and stores a JSON value in cache with TTL
func (c *CacheService) SetJSON(key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data for cache: %w", err)
	}

	err = c.Set(key, data, ttl)
	if err == nil {
		log.Printf("[CACHE SET] Key: %s, TTL: %v", key, ttl)
	}
	return err
}

// Delete removes a key from cache
func (c *CacheService) Delete(key string) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	err := c.client.Del(c.ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete from cache: %w", err)
	}

	return nil
}

// DeletePattern deletes all keys matching a pattern
func (c *CacheService) DeletePattern(pattern string) error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	// Use SCAN to find all keys matching the pattern
	iter := c.client.Scan(c.ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(c.ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		err := c.client.Del(c.ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return nil
}

// Exists checks if a key exists in cache
func (c *CacheService) Exists(key string) (bool, error) {
	if c.client == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	count, err := c.client.Exists(c.ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check cache key existence: %w", err)
	}

	return count > 0, nil
}

// ClearAll clears all cache keys (use with caution)
func (c *CacheService) ClearAll() error {
	if c.client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	err := c.client.FlushDB(c.ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	return nil
}