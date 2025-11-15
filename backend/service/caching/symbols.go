package caching

import (
	"encoding/json"
	"fmt"
	"log"
	"screener/backend/database"
	"screener/backend/model"

	"gorm.io/gorm"
)

const symbolsCacheKey = "cache:symbols:all"

// SymbolCache provides symbol list caching operations
type SymbolCache struct {
	cache *CacheService
	db    *gorm.DB
}

// NewSymbolCache creates a new symbol cache instance
func NewSymbolCache() *SymbolCache {
	return &SymbolCache{
		cache: NewCacheService(),
		db:    database.GetDB(),
	}
}

// LoadAllSymbols loads all symbols from the screener table into Redis
// This should be called once on server startup
func (s *SymbolCache) LoadAllSymbols() error {
	// Check if already cached
	exists, err := s.cache.Exists(symbolsCacheKey)
	if err == nil && exists {
		log.Println("Symbols already cached in Redis")
		return nil
	}

	// Fetch from database
	var symbols []string
	if err := s.db.Model(&model.Screener{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return fmt.Errorf("failed to load symbols from database: %w", err)
	}

	if len(symbols) == 0 {
		log.Println("Warning: No symbols found in screener table")
		return nil
	}

	// Store in Redis with no TTL (permanent until manually refreshed)
	data, err := json.Marshal(symbols)
	if err != nil {
		return fmt.Errorf("failed to marshal symbols: %w", err)
	}

	// Use Set with 0 TTL to make it permanent
	err = s.cache.Set(symbolsCacheKey, data, 0)
	if err != nil {
		return fmt.Errorf("failed to cache symbols: %w", err)
	}

	log.Printf("Successfully loaded %d symbols into Redis cache", len(symbols))
	return nil
}

// GetAllSymbols retrieves all symbols from Redis cache
// Falls back to database if Redis is unavailable or cache is empty
func (s *SymbolCache) GetAllSymbols() ([]string, error) {
	// Try Redis first
	var symbols []string
	found, err := s.cache.GetJSON(symbolsCacheKey, &symbols)
	if err == nil && found && len(symbols) > 0 {
		return symbols, nil
	}

	// Redis miss or error - fallback to database
	log.Printf("Redis cache miss for symbols, falling back to database")
	if err := s.db.Model(&model.Screener{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return nil, fmt.Errorf("failed to load symbols from database: %w", err)
	}

	// Try to repopulate cache (non-blocking)
	if len(symbols) > 0 {
		go func() {
			if err := s.LoadAllSymbols(); err != nil {
				log.Printf("Warning: Failed to repopulate symbol cache: %v", err)
			}
		}()
	}

	return symbols, nil
}

// RefreshSymbols manually refreshes the symbol cache from the database
func (s *SymbolCache) RefreshSymbols() error {
	// Delete existing cache
	if err := s.cache.Delete(symbolsCacheKey); err != nil {
		log.Printf("Warning: Failed to delete existing symbol cache: %v", err)
	}

	// Reload from database
	return s.LoadAllSymbols()
}
