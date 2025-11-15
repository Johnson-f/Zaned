package service

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/caching"

	"gorm.io/gorm"
)

// CompanyInfoService contains business logic for company info operations
type CompanyInfoService struct {
	db    *gorm.DB
	cache *caching.CacheService
	ttl   *caching.CacheTTLConfig
}

// NewCompanyInfoService creates a new instance of CompanyInfoService
func NewCompanyInfoService() *CompanyInfoService {
	return &CompanyInfoService{
		db:    database.GetDB(),
		cache: caching.NewCacheService(),
		ttl:   caching.GetTTLConfig(),
	}
}

// GetAllCompanyInfo fetches all company info records (read-only)
func (s *CompanyInfoService) GetAllCompanyInfo() ([]model.CompanyInfo, error) {
	// Try to get from cache
	cacheKey := caching.GenerateKeyFromPath("company-info")
	var companyInfo []model.CompanyInfo
	
	found, err := s.cache.GetJSON(cacheKey, &companyInfo)
	if err == nil && found {
		return companyInfo, nil
	}

	// Cache miss - query database
	result := s.db.Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info: %w", result.Error)
	}

	// Store in cache
	_ = s.cache.SetJSON(cacheKey, companyInfo, s.ttl.CompanyInfo)

	return companyInfo, nil
}

// GetCompanyInfoBySymbol fetches company info by symbol
func (s *CompanyInfoService) GetCompanyInfoBySymbol(symbol string) (*model.CompanyInfo, error) {
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}

	// Try to get from cache
	cacheKey := caching.GenerateKeyFromPath(fmt.Sprintf("company-info/%s", symbol))
	var companyInfo model.CompanyInfo
	
	found, err := s.cache.GetJSON(cacheKey, &companyInfo)
	if err == nil && found {
		return &companyInfo, nil
	}

	// Cache miss - query database
	result := s.db.Where("symbol = ?", symbol).First(&companyInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, fmt.Errorf("failed to fetch company info: %w", result.Error)
	}

	// Store in cache
	_ = s.cache.SetJSON(cacheKey, companyInfo, s.ttl.CompanyInfo)

	return &companyInfo, nil
}

// GetCompanyInfoBySymbols fetches multiple company info records by symbols
func (s *CompanyInfoService) GetCompanyInfoBySymbols(symbols []string) ([]model.CompanyInfo, error) {
	if len(symbols) == 0 {
		return []model.CompanyInfo{}, nil
	}

	// Create cache key from symbols (sorted for consistency)
	cacheKey := caching.GenerateKey("company-info/symbols", map[string]string{
		"symbols": fmt.Sprintf("%v", symbols),
	})
	var companyInfo []model.CompanyInfo
	
	found, err := s.cache.GetJSON(cacheKey, &companyInfo)
	if err == nil && found {
		return companyInfo, nil
	}

	// Cache miss - query database
	result := s.db.Where("symbol IN ?", symbols).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info: %w", result.Error)
	}

	// Store in cache
	_ = s.cache.SetJSON(cacheKey, companyInfo, s.ttl.CompanyInfo)

	return companyInfo, nil
}

// SearchCompanyInfo searches company info by name, sector, or industry
func (s *CompanyInfoService) SearchCompanyInfo(searchTerm string) ([]model.CompanyInfo, error) {
	if searchTerm == "" {
		return []model.CompanyInfo{}, nil
	}

	// Try to get from cache
	cacheKey := caching.GenerateKey("company-info/search", map[string]string{
		"q": searchTerm,
	})
	var companyInfo []model.CompanyInfo
	
	found, err := s.cache.GetJSON(cacheKey, &companyInfo)
	if err == nil && found {
		return companyInfo, nil
	}

	// Cache miss - query database
	searchPattern := "%" + searchTerm + "%"
	result := s.db.Where(
		"name ILIKE ? OR sector ILIKE ? OR industry ILIKE ? OR symbol ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search company info: %w", result.Error)
	}

	// Store in cache
	_ = s.cache.SetJSON(cacheKey, companyInfo, s.ttl.CompanyInfo)

	return companyInfo, nil
}

// GetCompanyInfoBySector fetches all company info records for a specific sector
func (s *CompanyInfoService) GetCompanyInfoBySector(sector string) ([]model.CompanyInfo, error) {
	if sector == "" {
		return []model.CompanyInfo{}, nil
	}

	// Try to get from cache
	cacheKey := caching.GenerateKeyFromPath(fmt.Sprintf("company-info/sector/%s", sector))
	var companyInfo []model.CompanyInfo
	
	found, err := s.cache.GetJSON(cacheKey, &companyInfo)
	if err == nil && found {
		return companyInfo, nil
	}

	// Cache miss - query database
	result := s.db.Where("sector = ?", sector).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info by sector: %w", result.Error)
	}

	// Store in cache
	_ = s.cache.SetJSON(cacheKey, companyInfo, s.ttl.CompanyInfo)

	return companyInfo, nil
}

// GetCompanyInfoByIndustry fetches all company info records for a specific industry
func (s *CompanyInfoService) GetCompanyInfoByIndustry(industry string) ([]model.CompanyInfo, error) {
	if industry == "" {
		return []model.CompanyInfo{}, nil
	}

	// Try to get from cache
	cacheKey := caching.GenerateKeyFromPath(fmt.Sprintf("company-info/industry/%s", industry))
	var companyInfo []model.CompanyInfo
	
	found, err := s.cache.GetJSON(cacheKey, &companyInfo)
	if err == nil && found {
		return companyInfo, nil
	}

	// Cache miss - query database
	result := s.db.Where("industry = ?", industry).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info by industry: %w", result.Error)
	}

	// Store in cache
	_ = s.cache.SetJSON(cacheKey, companyInfo, s.ttl.CompanyInfo)

	return companyInfo, nil
}
