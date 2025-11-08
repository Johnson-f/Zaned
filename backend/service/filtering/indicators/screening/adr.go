package screening

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/filtering/indicators/calculations"

	"gorm.io/gorm"
)

// ADRScreeningService handles ADR (Average Daily Range) screening logic
type ADRScreeningService struct {
	db *gorm.DB
}

// NewADRScreeningService creates a new instance of ADRScreeningService
func NewADRScreeningService() *ADRScreeningService {
	return &ADRScreeningService{
		db: database.GetDB(),
	}
}

// GetSymbolsByADR scans all symbols with the given range/interval and returns those
// whose ADR% (Average Daily Range as percentage) falls within the specified thresholds.
// ADR% = SMA(high-low, lookback) / close * 100
func (s *ADRScreeningService) GetSymbolsByADR(rangeParam, interval string, lookback int, minADR, maxADR *float64) ([]string, error) {
	if rangeParam == "" || interval == "" || lookback <= 0 {
		return nil, errors.New("range, interval, and lookback (positive) are required")
	}

	// Collect distinct symbols that have records for this range/interval
	var symbols []string
	if err := s.db.Model(&model.Historical{}).
		Where("range = ? AND interval = ?", rangeParam, interval).
		Distinct("symbol").
		Pluck("symbol", &symbols).Error; err != nil {
		return nil, fmt.Errorf("failed to load symbols: %w", err)
	}
	if len(symbols) == 0 {
		return []string{}, nil
	}

	matches := make([]string, 0)
	for _, sym := range symbols {
		// Fetch historical data for this symbol
		var rows []model.Historical
		if err := s.db.Where("symbol = ? AND range = ? AND interval = ?", sym, rangeParam, interval).
			Order("epoch ASC").
			Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

		// Calculate ADR%: SMA of (high-low) over lookback period, then divide by last close
		rngSeries := make([]float64, 0, len(rows))
		for _, r := range rows {
			rngSeries = append(rngSeries, r.High-r.Low)
		}
		adr := calculations.SimpleMovingAverage(rngSeries, lookback)
		last := rows[len(rows)-1]
		if last.Close == 0 {
			continue // skip if no valid close price
		}
		adrPercent := (adr / last.Close) * 100.0

		// Apply filters if provided
		matchesThreshold := true
		if minADR != nil && adrPercent < *minADR {
			matchesThreshold = false
		}
		if maxADR != nil && adrPercent > *maxADR {
			matchesThreshold = false
		}
		if matchesThreshold {
			matches = append(matches, sym)
		}
	}

	return matches, nil
}

// GetADRForSymbol calculates and returns ADR% for a specific symbol.
// ADR% = SMA(high-low, lookback) / close * 100
func (s *ADRScreeningService) GetADRForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
	if symbol == "" || rangeParam == "" || interval == "" || lookback <= 0 {
		return 0, errors.New("symbol, range, interval, and lookback (positive) are required")
	}

	var rows []model.Historical
	if err := s.db.Where("symbol = ? AND range = ? AND interval = ?", symbol, rangeParam, interval).
		Order("epoch ASC").
		Find(&rows).Error; err != nil {
		return 0, fmt.Errorf("failed to fetch historical data: %w", err)
	}
	if len(rows) == 0 {
		return 0, errors.New("no historical data found for symbol")
	}

	// Calculate ADR%: SMA of (high-low) over lookback period, then divide by last close
	rngSeries := make([]float64, 0, len(rows))
	for _, r := range rows {
		rngSeries = append(rngSeries, r.High-r.Low)
	}
	adr := calculations.SimpleMovingAverage(rngSeries, lookback)
	last := rows[len(rows)-1]
	if last.Close == 0 {
		return 0, errors.New("invalid close price (zero)")
	}
	adrPercent := (adr / last.Close) * 100.0

	return adrPercent, nil
}

