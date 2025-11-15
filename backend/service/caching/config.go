package caching

import (
	"os"
	"time"
)

// CacheTTLConfig holds TTL configuration for different cache types
type CacheTTLConfig struct {
	CompanyInfo       time.Duration
	FundamentalData   time.Duration
	MarketStatistics  time.Duration
	ScreenerResults   time.Duration
	Historical        time.Duration
	Screener          time.Duration
	Symbols           time.Duration // TTL for symbols list used by cron jobs
	PersistenceSchedule time.Duration // Schedule for background persistence worker (e.g., 1h, 24h)
	EnableRedisFirst  bool           // Enable Redis-first mode (default: true)
}

var ttlConfig *CacheTTLConfig

// GetTTLConfig returns the TTL configuration, initializing it if needed
func GetTTLConfig() *CacheTTLConfig {
	if ttlConfig == nil {
		// Parse persistence schedule (default: 1 hour)
		persistenceScheduleStr := os.Getenv("CACHE_PERSISTENCE_SCHEDULE")
		if persistenceScheduleStr == "" {
			persistenceScheduleStr = "1h"
		}
		persistenceSchedule := parseDuration(persistenceScheduleStr, 1*time.Hour)
		
		// Parse enable Redis-first flag (default: true)
		enableRedisFirst := true
		if enableStr := os.Getenv("CACHE_ENABLE_REDIS_FIRST"); enableStr != "" {
			if enableStr == "false" || enableStr == "0" {
				enableRedisFirst = false
			}
		}

		ttlConfig = &CacheTTLConfig{
			CompanyInfo:        parseDuration(os.Getenv("CACHE_TTL_COMPANY_INFO"), 1*time.Hour),
			FundamentalData:    parseDuration(os.Getenv("CACHE_TTL_FUNDAMENTAL_DATA"), 1*time.Hour),
			MarketStatistics:   parseDuration(os.Getenv("CACHE_TTL_MARKET_STATISTICS"), 5*time.Minute),
			ScreenerResults:    parseDuration(os.Getenv("CACHE_TTL_SCREENER_RESULTS"), 15*time.Minute),
			Historical:         parseDuration(os.Getenv("CACHE_TTL_HISTORICAL"), 30*time.Minute),
			Screener:           parseDuration(os.Getenv("CACHE_TTL_SCREENER"), 10*time.Minute),
			Symbols:            parseDuration(os.Getenv("CACHE_TTL_SYMBOLS"), 1*time.Hour), // Cache symbols list for 1 hour
			PersistenceSchedule: persistenceSchedule,
			EnableRedisFirst:   enableRedisFirst,
		}
	}
	return ttlConfig
}

// parseDuration parses a duration string (e.g., "1h", "5m", "30s") with a default fallback
func parseDuration(envValue string, defaultValue time.Duration) time.Duration {
	if envValue == "" {
		return defaultValue
	}

	duration, err := time.ParseDuration(envValue)
	if err != nil {
		// If parsing fails, return default
		return defaultValue
	}

	return duration
}
