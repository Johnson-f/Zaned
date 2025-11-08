package screening

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/filtering/indicators/calculations"

	"gorm.io/gorm"
)

// ATRScreeningService handles ATR (Average True Range) screening logic
type ATRScreeningService struct {
	db *gorm.DB
}

// NewATRScreeningService creates a new instance of ATRScreeningService
func NewATRScreeningService() *ATRScreeningService {
	return &ATRScreeningService{
		db: database.GetDB(),
	}
}

// GetSymbolsByATR scans all symbols with the given range/interval and returns those
// whose ATR% (Average True Range as percentage) falls within the specified thresholds.
// ATR% = ATR(lookback) / close * 100
// ATR is calculated as SMA of True Range (max of: high-low, |high-prevClose|, |low-prevClose|)
func (s *ATRScreeningService) GetSymbolsByATR(rangeParam, interval string, lookback int, minATR, maxATR *float64) ([]string, error) {
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

		// Calculate ATR%: ATR over lookback period, then divide by last close
		atr := calculations.AverageTrueRange(rows, lookback)
		last := rows[len(rows)-1]
		if last.Close == 0 {
			continue // skip if no valid close price
		}
		atrPercent := (atr / last.Close) * 100.0

		// Apply filters if provided
		matchesThreshold := true
		if minATR != nil && atrPercent < *minATR {
			matchesThreshold = false
		}
		if maxATR != nil && atrPercent > *maxATR {
			matchesThreshold = false
		}
		if matchesThreshold {
			matches = append(matches, sym)
		}
	}

	return matches, nil
}

// GetATRForSymbol calculates and returns ATR% for a specific symbol.
// ATR% = ATR(lookback) / close * 100
// ATR is calculated as SMA of True Range (max of: high-low, |high-prevClose|, |low-prevClose|)
func (s *ATRScreeningService) GetATRForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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

	// Calculate ATR%: ATR over lookback period, then divide by last close
	atr := calculations.AverageTrueRange(rows, lookback)
	last := rows[len(rows)-1]
	if last.Close == 0 {
		return 0, errors.New("invalid close price (zero)")
	}
	atrPercent := (atr / last.Close) * 100.0

	return atrPercent, nil
}

