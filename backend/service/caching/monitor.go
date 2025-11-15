package caching

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// CacheStats holds cache statistics
type CacheStats struct {
	TotalKeys    int64
	MemoryUsage  string
	HitRate      float64
	MissRate     float64
}

// GetCacheStats returns current cache statistics
func GetCacheStats() (*CacheStats, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := GetRedisContext()

	// Get total number of keys
	dbSize, err := client.DBSize(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get database size: %w", err)
	}

	// Get memory info
	info, err := client.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory info: %w", err)
	}

	// Parse used_memory_human from info
	memoryUsage := "unknown"
	lines := info
	for _, line := range strings.Split(lines, "\n") {
		if strings.HasPrefix(line, "used_memory_human:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				memoryUsage = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	return &CacheStats{
		TotalKeys:   dbSize,
		MemoryUsage: memoryUsage,
	}, nil
}

// LogCacheStats logs current cache statistics
func LogCacheStats() {
	stats, err := GetCacheStats()
	if err != nil {
		log.Printf("[CACHE] Failed to get stats: %v", err)
		return
	}

	log.Printf("[CACHE STATS] Total Keys: %d, Memory Usage: %s", stats.TotalKeys, stats.MemoryUsage)
}

// ListCacheKeys lists all cache keys matching a pattern (for debugging)
func ListCacheKeys(pattern string) ([]string, error) {
	client := GetRedisClient()
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	ctx := GetRedisContext()
	var keys []string

	iter := client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	return keys, nil
}

// GetKeyTTL returns the remaining TTL for a key
func GetKeyTTL(key string) (time.Duration, error) {
	client := GetRedisClient()
	if client == nil {
		return 0, fmt.Errorf("redis client not initialized")
	}

	ctx := GetRedisContext()
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL: %w", err)
	}

	return ttl, nil
}

