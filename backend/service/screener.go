package service

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/caching"

	"gorm.io/gorm"
)

// ScreenerService contains business logic for screener operations
type ScreenerService struct {
	db    *gorm.DB
	cache *caching.CacheService
	ttl   *caching.CacheTTLConfig
}

// FilterOptions represents filtering options for screener queries
type FilterOptions struct {
	MinPrice  *float64
	MaxPrice  *float64
	MinVolume *int64
	MaxVolume *int64
	MinOpen   *float64
	MaxOpen   *float64
	MinHigh   *float64
	MaxHigh   *float64
	MinLow    *float64
	MaxLow    *float64
	MinClose  *float64
	MaxClose  *float64
}

// SortOptions represents sorting options for screener queries
type SortOptions struct {
	Field     string // "symbol", "open", "high", "low", "close", "volume", "created_at"
	Direction string // "asc" or "desc"
}

// PaginationOptions represents pagination options
type PaginationOptions struct {
	Page  int // 1-indexed page number
	Limit int // Number of records per page
}

// QueryResult represents a paginated query result
type QueryResult struct {
	Data       []model.Screener `json:"data"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int64            `json:"total"`
	TotalPages int              `json:"total_pages"`
}

// NewScreenerService creates a new instance of ScreenerService
func NewScreenerService() *ScreenerService {
	return &ScreenerService{
		db:    database.GetDB(),
		cache: caching.NewCacheService(),
		ttl:   caching.GetTTLConfig(),
	}
}

// GetAllScreeners fetches all screener records (read-only)
func (s *ScreenerService) GetAllScreeners() ([]model.Screener, error) {
	cacheKey := caching.GenerateKeyFromPath("screener")
	var screeners []model.Screener
	
	found, err := s.cache.GetJSON(cacheKey, &screeners)
	if err == nil && found {
		return screeners, nil
	}

	result := s.db.Find(&screeners)
	if result.Error != nil {
		return nil, result.Error
	}

	_ = s.cache.SetJSON(cacheKey, screeners, s.ttl.Screener)
	return screeners, nil
}

// GetScreenerByID fetches a screener record by ID
func (s *ScreenerService) GetScreenerByID(id string) (*model.Screener, error) {
	var screener model.Screener
	result := s.db.Where("id = ?", id).First(&screener)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &screener, nil
}

// GetScreenerBySymbol fetches a screener record by symbol
func (s *ScreenerService) GetScreenerBySymbol(symbol string) (*model.Screener, error) {
	cacheKey := caching.GenerateKeyFromPath(fmt.Sprintf("screener/symbol/%s", symbol))
	var screener model.Screener
	
	found, err := s.cache.GetJSON(cacheKey, &screener)
	if err == nil && found {
		return &screener, nil
	}

	result := s.db.Where("symbol = ?", symbol).First(&screener)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	_ = s.cache.SetJSON(cacheKey, screener, s.ttl.Screener)
	return &screener, nil
}

// GetScreenersBySymbols fetches multiple screener records by symbols
func (s *ScreenerService) GetScreenersBySymbols(symbols []string) ([]model.Screener, error) {
	var screeners []model.Screener
	result := s.db.Where("symbol IN ?", symbols).Find(&screeners)
	if result.Error != nil {
		return nil, result.Error
	}

	return screeners, nil
}

// GetScreenersWithFilters fetches screener records with filtering, sorting, and pagination
func (s *ScreenerService) GetScreenersWithFilters(
	filters *FilterOptions,
	sort *SortOptions,
	pagination *PaginationOptions,
) (*QueryResult, error) {
	query := s.db.Model(&model.Screener{})

	// Apply filters
	if filters != nil {
		if filters.MinPrice != nil {
			query = query.Where("close >= ?", *filters.MinPrice)
		}
		if filters.MaxPrice != nil {
			query = query.Where("close <= ?", *filters.MaxPrice)
		}
		if filters.MinVolume != nil {
			query = query.Where("volume >= ?", *filters.MinVolume)
		}
		if filters.MaxVolume != nil {
			query = query.Where("volume <= ?", *filters.MaxVolume)
		}
		if filters.MinOpen != nil {
			query = query.Where("open >= ?", *filters.MinOpen)
		}
		if filters.MaxOpen != nil {
			query = query.Where("open <= ?", *filters.MaxOpen)
		}
		if filters.MinHigh != nil {
			query = query.Where("high >= ?", *filters.MinHigh)
		}
		if filters.MaxHigh != nil {
			query = query.Where("high <= ?", *filters.MaxHigh)
		}
		if filters.MinLow != nil {
			query = query.Where("low >= ?", *filters.MinLow)
		}
		if filters.MaxLow != nil {
			query = query.Where("low <= ?", *filters.MaxLow)
		}
		if filters.MinClose != nil {
			query = query.Where("close >= ?", *filters.MinClose)
		}
		if filters.MaxClose != nil {
			query = query.Where("close <= ?", *filters.MaxClose)
		}
	}

	// Get total count before pagination
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Apply sorting
	if sort != nil && sort.Field != "" {
		direction := "ASC"
		if sort.Direction == "desc" {
			direction = "DESC"
		}
		// Validate field name to prevent SQL injection
		validFields := map[string]bool{
			"symbol":     true,
			"open":       true,
			"high":       true,
			"low":        true,
			"close":      true,
			"volume":     true,
			"created_at": true,
			"updated_at": true,
		}
		if validFields[sort.Field] {
			query = query.Order(fmt.Sprintf("%s %s", sort.Field, direction))
		}
	} else {
		// Default sorting by symbol
		query = query.Order("symbol ASC")
	}

	// Apply pagination
	if pagination != nil {
		if pagination.Page < 1 {
			pagination.Page = 1
		}
		if pagination.Limit < 1 {
			pagination.Limit = 10 // Default limit
		}
		if pagination.Limit > 100 {
			pagination.Limit = 100 // Max limit to prevent performance issues
		}
		offset := (pagination.Page - 1) * pagination.Limit
		query = query.Offset(offset).Limit(pagination.Limit)
	} else {
		// Default pagination if none provided - return all records (up to reasonable limit)
		pagination = &PaginationOptions{Page: 1, Limit: 100}
		query = query.Limit(100)
	}

	// Execute query
	var screeners []model.Screener
	result := query.Find(&screeners)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch screeners: %w", result.Error)
	}

	// Calculate total pages
	totalPages := 1
	if pagination.Limit > 0 {
		totalPages = int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))
	}

	return &QueryResult{
		Data:       screeners,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

// SearchScreenersBySymbol searches for screeners by symbol (case-insensitive partial match)
func (s *ScreenerService) SearchScreenersBySymbol(searchTerm string, limit int) ([]model.Screener, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var screeners []model.Screener
	result := s.db.Where("LOWER(symbol) LIKE LOWER(?)", "%"+searchTerm+"%").
		Limit(limit).
		Order("symbol ASC").
		Find(&screeners)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to search screeners: %w", result.Error)
	}

	return screeners, nil
}

// GetScreenersByPriceRange fetches screeners within a specific price range
func (s *ScreenerService) GetScreenersByPriceRange(minPrice, maxPrice float64) ([]model.Screener, error) {
	var screeners []model.Screener
	result := s.db.Where("close >= ? AND close <= ?", minPrice, maxPrice).
		Order("close ASC").
		Find(&screeners)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch screeners by price range: %w", result.Error)
	}

	return screeners, nil
}

// GetScreenersByVolumeRange fetches screeners within a specific volume range
func (s *ScreenerService) GetScreenersByVolumeRange(minVolume, maxVolume int64) ([]model.Screener, error) {
	var screeners []model.Screener
	result := s.db.Where("volume >= ? AND volume <= ?", minVolume, maxVolume).
		Order("volume DESC").
		Find(&screeners)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch screeners by volume range: %w", result.Error)
	}

	return screeners, nil
}

// GetTopGainers fetches top gainers based on price change (high - low)
func (s *ScreenerService) GetTopGainers(limit int) ([]model.Screener, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var screeners []model.Screener
	result := s.db.Order("(high - low) DESC").
		Limit(limit).
		Find(&screeners)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch top gainers: %w", result.Error)
	}

	return screeners, nil
}

// GetMostActive fetches most active stocks by volume
func (s *ScreenerService) GetMostActive(limit int) ([]model.Screener, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var screeners []model.Screener
	result := s.db.Order("volume DESC").
		Limit(limit).
		Find(&screeners)

	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch most active stocks: %w", result.Error)
	}

	return screeners, nil
}

// GetCount returns the total count of screener records
func (s *ScreenerService) GetCount() (int64, error) {
	var count int64
	result := s.db.Model(&model.Screener{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to get count: %w", result.Error)
	}

	return count, nil
}
