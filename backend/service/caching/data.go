package caching

import (
	"fmt"
	"log"
	"screener/backend/model"
)

// DataCache provides data caching operations for fetched data
// All data is saved to Redis ONLY (no immediate database writes)
type DataCache struct {
	cache *CacheService
}

// NewDataCache creates a new data cache instance
func NewDataCache() *DataCache {
	return &DataCache{
		cache: NewCacheService(),
	}
}

// CacheHistorical caches historical data in Redis
// Key format: cache:data:historical:{symbol}:{range}:{interval}
func (d *DataCache) CacheHistorical(symbol, rangeParam, interval string, data []model.Historical) error {
	if d.cache == nil {
		return fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:historical:%s:%s:%s", symbol, rangeParam, interval)

	// No TTL - data persists until background worker persists to database
	if err := d.cache.SetJSON(key, data, 0); err != nil {
		return fmt.Errorf("failed to cache historical data: %w", err)
	}

	log.Printf("[CACHE] Cached historical data: %s (%d records)", key, len(data))
	return nil
}

// CacheCompanyInfo caches company info in Redis
// Key format: cache:data:company-info:{symbol}
func (d *DataCache) CacheCompanyInfo(symbol string, data *model.CompanyInfo) error {
	if d.cache == nil {
		return fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:company-info:%s", symbol)

	// No TTL - data persists until background worker persists to database
	if err := d.cache.SetJSON(key, data, 0); err != nil {
		return fmt.Errorf("failed to cache company info: %w", err)
	}

	log.Printf("[CACHE] Cached company info: %s", key)
	return nil
}

// CacheFundamentalData caches fundamental data in Redis
// Key format: cache:data:fundamental:{symbol}:{statementType}:{frequency}
func (d *DataCache) CacheFundamentalData(symbol, statementType, frequency string, data *model.FundamentalData) error {
	if d.cache == nil {
		return fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:fundamental:%s:%s:%s", symbol, statementType, frequency)

	// No TTL - data persists until background worker persists to database
	if err := d.cache.SetJSON(key, data, 0); err != nil {
		return fmt.Errorf("failed to cache fundamental data: %w", err)
	}

	log.Printf("[CACHE] Cached fundamental data: %s", key)
	return nil
}

// CacheMarketStatistics caches market statistics in Redis
// Key format: cache:data:market-statistics:{date}
func (d *DataCache) CacheMarketStatistics(date string, data *model.MarketStatistics) error {
	if d.cache == nil {
		return fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:market-statistics:%s", date)

	// No TTL - data persists until background worker persists to database
	if err := d.cache.SetJSON(key, data, 0); err != nil {
		return fmt.Errorf("failed to cache market statistics: %w", err)
	}

	log.Printf("[CACHE] Cached market statistics: %s", key)
	return nil
}

// GetHistorical retrieves historical data from Redis
func (d *DataCache) GetHistorical(symbol, rangeParam, interval string) ([]model.Historical, bool, error) {
	if d.cache == nil {
		return nil, false, fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:historical:%s:%s:%s", symbol, rangeParam, interval)
	var data []model.Historical

	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, false, err
	}

	return data, found, nil
}

// GetCompanyInfo retrieves company info from Redis
func (d *DataCache) GetCompanyInfo(symbol string) (*model.CompanyInfo, bool, error) {
	if d.cache == nil {
		return nil, false, fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:company-info:%s", symbol)
	var data model.CompanyInfo

	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	return &data, true, nil
}

// GetFundamentalData retrieves fundamental data from Redis
func (d *DataCache) GetFundamentalData(symbol, statementType, frequency string) (*model.FundamentalData, bool, error) {
	if d.cache == nil {
		return nil, false, fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:fundamental:%s:%s:%s", symbol, statementType, frequency)
	var data model.FundamentalData

	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	return &data, true, nil
}

// GetMarketStatistics retrieves market statistics from Redis
func (d *DataCache) GetMarketStatistics(date string) (*model.MarketStatistics, bool, error) {
	if d.cache == nil {
		return nil, false, fmt.Errorf("cache service not initialized")
	}

	key := fmt.Sprintf("cache:data:market-statistics:%s", date)
	var data model.MarketStatistics

	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, false, err
	}

	if !found {
		return nil, false, nil
	}

	return &data, true, nil
}

// GetAllHistoricalKeys returns all historical data keys matching a pattern
func (d *DataCache) GetAllHistoricalKeys() ([]string, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	pattern := "cache:data:historical:*"
	return d.scanKeys(pattern)
}

// GetAllCompanyInfoKeys returns all company info keys
func (d *DataCache) GetAllCompanyInfoKeys() ([]string, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	pattern := "cache:data:company-info:*"
	return d.scanKeys(pattern)
}

// GetAllFundamentalDataKeys returns all fundamental data keys
func (d *DataCache) GetAllFundamentalDataKeys() ([]string, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	pattern := "cache:data:fundamental:*"
	return d.scanKeys(pattern)
}

// GetAllMarketStatisticsKeys returns all market statistics keys
func (d *DataCache) GetAllMarketStatisticsKeys() ([]string, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	pattern := "cache:data:market-statistics:*"
	return d.scanKeys(pattern)
}

// scanKeys scans Redis for keys matching a pattern
func (d *DataCache) scanKeys(pattern string) ([]string, error) {
	client := GetRedisClient()
	ctx := GetRedisContext()

	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

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

// GetHistoricalByKey retrieves historical data by key
func (d *DataCache) GetHistoricalByKey(key string) ([]model.Historical, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	var data []model.Historical
	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return data, nil
}

// GetCompanyInfoByKey retrieves company info by key
func (d *DataCache) GetCompanyInfoByKey(key string) (*model.CompanyInfo, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	var data model.CompanyInfo
	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return &data, nil
}

// GetFundamentalDataByKey retrieves fundamental data by key
func (d *DataCache) GetFundamentalDataByKey(key string) (*model.FundamentalData, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	var data model.FundamentalData
	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return &data, nil
}

// GetMarketStatisticsByKey retrieves market statistics by key
func (d *DataCache) GetMarketStatisticsByKey(key string) (*model.MarketStatistics, error) {
	if d.cache == nil {
		return nil, fmt.Errorf("cache service not initialized")
	}

	var data model.MarketStatistics
	found, err := d.cache.GetJSON(key, &data)
	if err != nil {
		return nil, err
	}

	if !found {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	return &data, nil
}

// DeleteKey deletes a key from Redis
func (d *DataCache) DeleteKey(key string) error {
	if d.cache == nil {
		return fmt.Errorf("cache service not initialized")
	}

	return d.cache.Delete(key)
}

// DeleteKeys deletes multiple keys from Redis
func (d *DataCache) DeleteKeys(keys []string) error {
	if d.cache == nil {
		return fmt.Errorf("cache service not initialized")
	}

	client := GetRedisClient()
	ctx := GetRedisContext()

	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	if len(keys) == 0 {
		return nil
	}

	err := client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}
