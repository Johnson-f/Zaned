package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"screener/backend/database"
	"screener/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MarketStatisticsService contains business logic for market statistics aggregation
type MarketStatisticsService struct {
	db         *gorm.DB
	aggregator *DailyAggregator
}

// DailyAggregator holds in-memory aggregation state for the current day
type DailyAggregator struct {
	mu          sync.RWMutex
	today       time.Time
	counts      map[string]int // "up", "down", "unchanged"
	lastUpdated time.Time
}

// NewMarketStatisticsService constructs a new MarketStatisticsService
func NewMarketStatisticsService() *MarketStatisticsService {
	return &MarketStatisticsService{
		db: database.GetDB(),
		aggregator: &DailyAggregator{
			today:  time.Now().Truncate(24 * time.Hour),
			counts: make(map[string]int),
		},
	}
}

// parsePercentChange converts "-5.06%" or "+0.01%" string to float64
func parsePercentChange(percentStr string) (float64, error) {
	percentStr = strings.TrimSpace(percentStr)
	percentStr = strings.TrimSuffix(percentStr, "%")
	return strconv.ParseFloat(percentStr, 64)
}

// categorizeStock determines if stock is up, down, or unchanged based on +0.01% threshold
func categorizeStock(percentChange string) string {
	percent, err := parsePercentChange(percentChange)
	if err != nil {
		return "unchanged" // Default if parsing fails
	}

	if percent >= 0.01 {
		return "up"
	} else if percent <= -0.01 {
		return "down"
	}
	return "unchanged"
}

// AggregateQuotes processes quotes and updates daily counts
// Accepts both simpleQuote and detailedQuote types (both have PercentChange field)
func (s *MarketStatisticsService) AggregateQuotes(ctx context.Context, quotes interface{}) error {
	s.aggregator.mu.Lock()
	defer s.aggregator.mu.Unlock()

	// Reset if it's a new day
	today := time.Now().Truncate(24 * time.Hour)
	if !s.aggregator.today.Equal(today) {
		s.aggregator.today = today
		s.aggregator.counts = make(map[string]int)
	}

	// Handle different quote types using type assertion
	switch q := quotes.(type) {
	case []simpleQuote:
		for _, quote := range q {
			category := categorizeStock(quote.PercentChange)
			s.aggregator.counts[category]++
		}
	case []detailedQuote:
		for _, quote := range q {
			category := categorizeStock(quote.PercentChange)
			s.aggregator.counts[category]++
		}
	default:
		return fmt.Errorf("unsupported quote type: %T", quotes)
	}

	s.aggregator.lastUpdated = time.Now()
	return nil
}

// GetCurrentDayStats returns current day's aggregated stats
func (s *MarketStatisticsService) GetCurrentDayStats() (map[string]int, error) {
	s.aggregator.mu.RLock()
	defer s.aggregator.mu.RUnlock()

	stats := make(map[string]int)
	stats["up"] = s.aggregator.counts["up"]
	stats["down"] = s.aggregator.counts["down"]
	stats["unchanged"] = s.aggregator.counts["unchanged"]
	stats["total"] = stats["up"] + stats["down"] + stats["unchanged"]

	return stats, nil
}

// StoreEndOfDayStats saves today's aggregated stats to database
func (s *MarketStatisticsService) StoreEndOfDayStats(ctx context.Context) error {
	stats, err := s.GetCurrentDayStats()
	if err != nil {
		return err
	}

	today := time.Now().Truncate(24 * time.Hour)

	marketStats := model.MarketStatistics{
		Date:            today,
		StocksUp:        stats["up"],
		StocksDown:      stats["down"],
		StocksUnchanged: stats["unchanged"],
		TotalStocks:     stats["total"],
	}

	// Upsert using ON CONFLICT
	result := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "date"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"stocks_up", "stocks_down", "stocks_unchanged", "total_stocks", "updated_at",
		}),
	}).Create(&marketStats)

	return result.Error
}

// GetHistoricalStats fetches market statistics for charting
func (s *MarketStatisticsService) GetHistoricalStats(ctx context.Context, startDate, endDate time.Time) ([]model.MarketStatistics, error) {
	var stats []model.MarketStatistics
	err := s.db.Where("date >= ? AND date <= ?", startDate, endDate).
		Order("date ASC").
		Find(&stats).Error
	return stats, err
}
