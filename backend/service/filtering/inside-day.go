package filtering

import (
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// InsideDayService handles inside day filtering logic
type InsideDayService struct {
	db *gorm.DB
}

// NewInsideDayService creates a new instance of InsideDayService
func NewInsideDayService() *InsideDayService {
	return &InsideDayService{
		db: database.GetDB(),
	}
}

// GetSymbolsWithDailyInsideDay scans all symbols in the historical table
// and returns those whose latest daily bar is an inside day compared to the
// immediately previous daily bar.
// An inside day means: current high < previous high AND current low > previous low
func (s *InsideDayService) GetSymbolsWithDailyInsideDay() ([]string, error) {
	// Fetch all unique symbols that have daily records (10y range, 1d interval)
	var symbols []string
	if err := s.db.Model(&model.Historical{}).
		Where("range = ? AND interval = ?", "10y", "1d").
		Distinct("symbol").
		Pluck("symbol", &symbols).Error; err != nil {
		return nil, fmt.Errorf("failed to load symbols: %w", err)
	}

	if len(symbols) == 0 {
		return []string{}, nil
	}

	matches := make([]string, 0)
	for _, sym := range symbols {
		// Get last two DAILY bars (from 10y range) by epoch desc for this symbol
		var rows []model.Historical
		if err := s.db.Where("symbol = ? AND range = ? AND interval = ?", sym, "10y", "1d").
			Order("epoch DESC").
			Limit(2).
			Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) < 2 {
			continue
		}

		// rows[0] = most recent (current day)
		// rows[1] = previous day
		current := rows[0]
		previous := rows[1]

		// Inside day condition: current high < previous high AND current low > previous low
		if current.High < previous.High && current.Low > previous.Low {
			matches = append(matches, sym)
		}
	}

	return matches, nil
}

// SaveInsideDayResults saves current inside day symbols to database
func (s *InsideDayService) SaveInsideDayResults() error {
	symbols, err := s.GetSymbolsWithDailyInsideDay()
	if err != nil {
		return fmt.Errorf("failed to get inside day symbols: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	results := make([]model.ScreenerResult, 0, len(symbols))

	for _, symbol := range symbols {
		results = append(results, model.ScreenerResult{
			Type:   "inside_day",
			Symbol: symbol,
			Date:   today,
		})
	}

	if len(results) == 0 {
		return nil
	}

	// Upsert using ON CONFLICT
	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "type"},
			{Name: "symbol"},
			{Name: "date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).CreateInBatches(results, 100).Error
}
