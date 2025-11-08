package filtering

import (
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HighVolumeYearService handles highest volume in year filtering logic
type HighVolumeYearService struct {
	db *gorm.DB
}

// NewHighVolumeYearService creates a new instance of HighVolumeYearService
func NewHighVolumeYearService() *HighVolumeYearService {
	return &HighVolumeYearService{
		db: database.GetDB(),
	}
}

// GetSymbolsWithHighestVolumeInYear scans all symbols' daily bars (interval='1d')
// within the last 365 days (inclusive) and returns those whose most recent daily bar
// has the highest volume in that 365-day window.
func (s *HighVolumeYearService) GetSymbolsWithHighestVolumeInYear() ([]string, error) {
	// Collect distinct symbols that have daily bars
	var symbols []string
	if err := s.db.Model(&model.Historical{}).
		Where("interval = ?", "1d").
		Distinct("symbol").
		Pluck("symbol", &symbols).Error; err != nil {
		return nil, fmt.Errorf("failed to load symbols: %w", err)
	}
	if len(symbols) == 0 {
		return []string{}, nil
	}

	// Epoch window: now and 365 days ago
	now := time.Now()
	oneYearAgo := now.AddDate(0, 0, -365)
	maxEpoch := now.Unix()
	minEpoch := oneYearAgo.Unix()

	matches := make([]string, 0)
	for _, sym := range symbols {
		// Fetch last 365 days of daily bars for this symbol
		var rows []model.Historical
		if err := s.db.Where("symbol = ? AND interval = ? AND epoch BETWEEN ? AND ?", sym, "1d", minEpoch, maxEpoch).
			Order("epoch ASC").
			Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

		// Find max volume in window and compare with last bar
		var maxVol int64 = 0
		for _, r := range rows {
			if r.Volume > maxVol {
				maxVol = r.Volume
			}
		}
		last := rows[len(rows)-1]
		if last.Volume >= maxVol { // include ties as "highest"
			matches = append(matches, sym)
		}
	}

	return matches, nil
}

// SaveHighVolumeYearResults saves high volume year symbols to database
func (s *HighVolumeYearService) SaveHighVolumeYearResults() error {
	symbols, err := s.GetSymbolsWithHighestVolumeInYear()
	if err != nil {
		return fmt.Errorf("failed to get high volume year symbols: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	results := make([]model.ScreenerResult, 0, len(symbols))

	for _, symbol := range symbols {
		results = append(results, model.ScreenerResult{
			Type:   "high_volume_year",
			Symbol: symbol,
			Date:   today,
		})
	}

	if len(results) == 0 {
		return nil
	}

	return s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "type"},
			{Name: "symbol"},
			{Name: "date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{"updated_at"}),
	}).CreateInBatches(results, 100).Error
}
