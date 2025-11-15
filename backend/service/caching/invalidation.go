package caching

import (
	"fmt"
)

// InvalidationService provides cache invalidation operations
type InvalidationService struct {
	cache *CacheService
}

// NewInvalidationService creates a new invalidation service
func NewInvalidationService() *InvalidationService {
	return &InvalidationService{
		cache: NewCacheService(),
	}
}

// InvalidateCompanyInfo invalidates cache for a specific company by symbol
func (i *InvalidationService) InvalidateCompanyInfo(symbol string) error {
	key := GenerateKeyFromPath(fmt.Sprintf("company-info/%s", symbol))
	return i.cache.Delete(key)
}

// InvalidateAllCompanyInfo invalidates all company info cache entries
func (i *InvalidationService) InvalidateAllCompanyInfo() error {
	pattern := GeneratePattern("company-info")
	return i.cache.DeletePattern(pattern)
}

// InvalidateFundamentalData invalidates cache for a specific symbol's fundamental data
func (i *InvalidationService) InvalidateFundamentalData(symbol string) error {
	pattern := GeneratePattern(fmt.Sprintf("fundamental-data/symbol/%s", symbol))
	return i.cache.DeletePattern(pattern)
}

// InvalidateAllFundamentalData invalidates all fundamental data cache entries
func (i *InvalidationService) InvalidateAllFundamentalData() error {
	pattern := GeneratePattern("fundamental-data")
	return i.cache.DeletePattern(pattern)
}

// InvalidateScreenerResults invalidates cache for a specific screener result type
func (i *InvalidationService) InvalidateScreenerResults(resultType string) error {
	key := GenerateKeyFromQuery("screener-results", fmt.Sprintf("type=%s", resultType))
	return i.cache.DeletePattern(fmt.Sprintf("%s*", key))
}

// InvalidateAllScreenerResults invalidates all screener results cache entries
func (i *InvalidationService) InvalidateAllScreenerResults() error {
	pattern := GeneratePattern("screener-results")
	return i.cache.DeletePattern(pattern)
}

// InvalidateMarketStatistics invalidates market statistics cache
func (i *InvalidationService) InvalidateMarketStatistics() error {
	pattern := GeneratePattern("market-statistics")
	return i.cache.DeletePattern(pattern)
}

// InvalidateScreener invalidates screener cache (filtered queries)
func (i *InvalidationService) InvalidateScreener() error {
	pattern := GeneratePattern("screener")
	return i.cache.DeletePattern(pattern)
}

// InvalidateHistorical invalidates historical data cache for a specific symbol
func (i *InvalidationService) InvalidateHistorical(symbol string) error {
	pattern := GeneratePattern(fmt.Sprintf("historical/%s", symbol))
	return i.cache.DeletePattern(pattern)
}

// InvalidateAllHistorical invalidates all historical data cache entries
func (i *InvalidationService) InvalidateAllHistorical() error {
	pattern := GeneratePattern("historical")
	return i.cache.DeletePattern(pattern)
}

// InvalidateSymbols invalidates the cached symbols list
// This should be called when screener table is updated (symbols added/removed)
func (i *InvalidationService) InvalidateSymbols() error {
	key := GenerateKeyFromPath("screener/symbols")
	return i.cache.Delete(key)
}

// InvalidateByPattern invalidates cache entries matching a custom pattern
func (i *InvalidationService) InvalidateByPattern(pattern string) error {
	return i.cache.DeletePattern(pattern)
}

