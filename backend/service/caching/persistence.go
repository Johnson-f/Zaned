package caching

import (
	"fmt"
	"log"
	"screener/backend/database"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PersistenceService handles batch persistence of Redis data to database
type PersistenceService struct {
	dataCache *DataCache
	db        *gorm.DB
}

// NewPersistenceService creates a new persistence service instance
func NewPersistenceService() *PersistenceService {
	return &PersistenceService{
		dataCache: NewDataCache(),
		db:        database.GetDB(),
	}
}

// PersistHistoricalData scans Redis for historical data keys and batch upserts to database
func (p *PersistenceService) PersistHistoricalData() error {
	keys, err := p.dataCache.GetAllHistoricalKeys()
	if err != nil {
		return fmt.Errorf("failed to get historical keys: %w", err)
	}

	if len(keys) == 0 {
		log.Println("[PERSIST] No historical data keys found in Redis")
		return nil
	}

	log.Printf("[PERSIST] Found %d historical data keys to persist", len(keys))

	totalPersisted := 0
	batchSize := 1000

	for _, key := range keys {
		// Parse key: cache:data:historical:{symbol}:{range}:{interval}
		parts := strings.Split(key, ":")
		if len(parts) != 6 {
			log.Printf("[PERSIST] Warning: Invalid historical key format: %s", key)
			continue
		}

		// Get data from Redis
		historical, err := p.dataCache.GetHistoricalByKey(key)
		if err != nil {
			log.Printf("[PERSIST] Warning: Failed to get historical data for key %s: %v", key, err)
			continue
		}

		if len(historical) == 0 {
			continue
		}

		// Batch upsert to database
		err = p.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "symbol"}, {Name: "epoch"}, {Name: "range"}, {Name: "interval"},
			},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"open":       gorm.Expr("excluded.open"),
				"high":       gorm.Expr("excluded.high"),
				"low":        gorm.Expr("excluded.low"),
				"close":      gorm.Expr("excluded.close"),
				"adj_close":  gorm.Expr("excluded.adj_close"),
				"volume":     gorm.Expr("excluded.volume"),
				"updated_at": gorm.Expr("NOW()"),
			}),
		}).CreateInBatches(historical, batchSize).Error

		if err != nil {
			log.Printf("[PERSIST] Error persisting historical data for key %s: %v", key, err)
			continue
		}

		// Delete key from Redis after successful persistence
		if err := p.dataCache.DeleteKey(key); err != nil {
			log.Printf("[PERSIST] Warning: Failed to delete key %s after persistence: %v", key, err)
		}

		totalPersisted += len(historical)
		log.Printf("[PERSIST] Persisted %d historical records from key %s", len(historical), key)
	}

	log.Printf("[PERSIST] Historical data persistence completed: %d total records persisted", totalPersisted)
	return nil
}

// PersistCompanyInfo scans Redis for company info keys and batch upserts to database
func (p *PersistenceService) PersistCompanyInfo() error {
	keys, err := p.dataCache.GetAllCompanyInfoKeys()
	if err != nil {
		return fmt.Errorf("failed to get company info keys: %w", err)
	}

	if len(keys) == 0 {
		log.Println("[PERSIST] No company info keys found in Redis")
		return nil
	}

	log.Printf("[PERSIST] Found %d company info keys to persist", len(keys))

	totalPersisted := 0

	for _, key := range keys {
		// Get data from Redis
		companyInfo, err := p.dataCache.GetCompanyInfoByKey(key)
		if err != nil {
			log.Printf("[PERSIST] Warning: Failed to get company info for key %s: %v", key, err)
			continue
		}

		// Upsert to database
		err = p.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "symbol"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "price", "after_hours_price", "change", "percent_change",
				"open", "high", "low", "year_high", "year_low",
				"volume", "avg_volume", "market_cap", "beta", "pe",
				"earnings_date", "sector", "industry", "about", "employees",
				"five_days_return", "one_month_return", "three_month_return",
				"six_month_return", "ytd_return", "year_return",
				"three_year_return", "five_year_return", "ten_year_return",
				"max_return", "logo", "updated_at",
			}),
		}).Create(companyInfo).Error

		if err != nil {
			log.Printf("[PERSIST] Error persisting company info for key %s: %v", key, err)
			continue
		}

		// Delete key from Redis after successful persistence
		if err := p.dataCache.DeleteKey(key); err != nil {
			log.Printf("[PERSIST] Warning: Failed to delete key %s after persistence: %v", key, err)
		}

		totalPersisted++
		log.Printf("[PERSIST] Persisted company info for key %s", key)
	}

	log.Printf("[PERSIST] Company info persistence completed: %d records persisted", totalPersisted)
	return nil
}

// PersistFundamentalData scans Redis for fundamental data keys and batch upserts to database
func (p *PersistenceService) PersistFundamentalData() error {
	keys, err := p.dataCache.GetAllFundamentalDataKeys()
	if err != nil {
		return fmt.Errorf("failed to get fundamental data keys: %w", err)
	}

	if len(keys) == 0 {
		log.Println("[PERSIST] No fundamental data keys found in Redis")
		return nil
	}

	log.Printf("[PERSIST] Found %d fundamental data keys to persist", len(keys))

	totalPersisted := 0

	for _, key := range keys {
		// Get data from Redis
		fundamentalData, err := p.dataCache.GetFundamentalDataByKey(key)
		if err != nil {
			log.Printf("[PERSIST] Warning: Failed to get fundamental data for key %s: %v", key, err)
			continue
		}

		// Upsert to database
		err = p.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "symbol"},
				{Name: "statement_type"},
				{Name: "frequency"},
			},
			DoUpdates: clause.AssignmentColumns([]string{
				"statement", "updated_at",
			}),
		}).Create(fundamentalData).Error

		if err != nil {
			log.Printf("[PERSIST] Error persisting fundamental data for key %s: %v", key, err)
			continue
		}

		// Delete key from Redis after successful persistence
		if err := p.dataCache.DeleteKey(key); err != nil {
			log.Printf("[PERSIST] Warning: Failed to delete key %s after persistence: %v", key, err)
		}

		totalPersisted++
		log.Printf("[PERSIST] Persisted fundamental data for key %s", key)
	}

	log.Printf("[PERSIST] Fundamental data persistence completed: %d records persisted", totalPersisted)
	return nil
}

// PersistMarketStatistics scans Redis for market statistics keys and batch upserts to database
func (p *PersistenceService) PersistMarketStatistics() error {
	keys, err := p.dataCache.GetAllMarketStatisticsKeys()
	if err != nil {
		return fmt.Errorf("failed to get market statistics keys: %w", err)
	}

	if len(keys) == 0 {
		log.Println("[PERSIST] No market statistics keys found in Redis")
		return nil
	}

	log.Printf("[PERSIST] Found %d market statistics keys to persist", len(keys))

	totalPersisted := 0

	for _, key := range keys {
		// Get data from Redis
		marketStats, err := p.dataCache.GetMarketStatisticsByKey(key)
		if err != nil {
			log.Printf("[PERSIST] Warning: Failed to get market statistics for key %s: %v", key, err)
			continue
		}

		// Upsert to database
		err = p.db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "date"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"stocks_up", "stocks_down", "stocks_unchanged", "total_stocks", "updated_at",
			}),
		}).Create(marketStats).Error

		if err != nil {
			log.Printf("[PERSIST] Error persisting market statistics for key %s: %v", key, err)
			continue
		}

		// Delete key from Redis after successful persistence
		if err := p.dataCache.DeleteKey(key); err != nil {
			log.Printf("[PERSIST] Warning: Failed to delete key %s after persistence: %v", key, err)
		}

		totalPersisted++
		log.Printf("[PERSIST] Persisted market statistics for key %s", key)
	}

	log.Printf("[PERSIST] Market statistics persistence completed: %d records persisted", totalPersisted)
	return nil
}

// PersistAll persists all cached data types to database
func (p *PersistenceService) PersistAll() error {
	log.Println("[PERSIST] Starting persistence of all cached data types...")

	// Persist in order: historical, company-info, fundamental, market-statistics
	if err := p.PersistHistoricalData(); err != nil {
		log.Printf("[PERSIST] Error persisting historical data: %v", err)
		// Continue with other types
	}

	if err := p.PersistCompanyInfo(); err != nil {
		log.Printf("[PERSIST] Error persisting company info: %v", err)
		// Continue with other types
	}

	if err := p.PersistFundamentalData(); err != nil {
		log.Printf("[PERSIST] Error persisting fundamental data: %v", err)
		// Continue with other types
	}

	if err := p.PersistMarketStatistics(); err != nil {
		log.Printf("[PERSIST] Error persisting market statistics: %v", err)
		// Continue
	}

	log.Println("[PERSIST] Persistence of all data types completed")
	return nil
}

