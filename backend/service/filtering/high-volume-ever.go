package filtering

import (
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HighVolumeEverService handles highest volume ever filtering logic
type HighVolumeEverService struct {
	db *gorm.DB
}

// NewHighVolumeEverService creates a new instance of HighVolumeEverService
func NewHighVolumeEverService() *HighVolumeEverService {
	return &HighVolumeEverService{
		db: database.GetDB(),
	}
}

// GetSymbolsWithHighestVolumeEver scans all symbols' daily bars (interval='1d')
// across all available history in the database and returns those whose most recent
// daily bar has the highest volume ever observed for that symbol.
func (s *HighVolumeEverService) GetSymbolsWithHighestVolumeEver() ([]string, error) {
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

	matches := make([]string, 0)
	for _, sym := range symbols {
		// Fetch ALL daily bars for this symbol
		var rows []model.Historical
		if err := s.db.Where("symbol = ? AND interval = ?", sym, "1d").
			Order("epoch ASC").
			Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

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

// SaveHighVolumeEverResults saves high volume ever symbols to database
func (s *HighVolumeEverService) SaveHighVolumeEverResults() error {
	symbols, err := s.GetSymbolsWithHighestVolumeEver()
	if err != nil {
		return fmt.Errorf("failed to get high volume ever symbols: %w", err)
	}

	today := time.Now().Truncate(24 * time.Hour)
	results := make([]model.ScreenerResult, 0, len(symbols))

	for _, symbol := range symbols {
		results = append(results, model.ScreenerResult{
			Type:   "high_volume_ever",
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

