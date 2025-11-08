package screening

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/filtering/indicators/calculations"

	"gorm.io/gorm"
)

// VolumeScreeningService handles volume-based screening logic
type VolumeScreeningService struct {
	db *gorm.DB
}

// NewVolumeScreeningService creates a new instance of VolumeScreeningService
func NewVolumeScreeningService() *VolumeScreeningService {
	return &VolumeScreeningService{
		db: database.GetDB(),
	}
}

// GetSymbolsByAvgVolumeDollars scans all symbols and returns those whose average daily
// volume in dollars (SMA of volume*close over lookback) falls within the thresholds.
// Volume in dollars = volume * close, then SMA over lookback, then convert to millions ($M)
func (s *VolumeScreeningService) GetSymbolsByAvgVolumeDollars(rangeParam, interval string, lookback int, minVolDollarsM, maxVolDollarsM *float64) ([]string, error) {
	if rangeParam == "" || interval == "" || lookback <= 0 {
		return nil, errors.New("range, interval, and lookback (positive) are required")
	}

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
		var rows []model.Historical
		if err := s.db.Where("symbol = ? AND range = ? AND interval = ?", sym, rangeParam, interval).
			Order("epoch ASC").
			Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

		// Calculate average volume in dollars: SMA(volume * close, lookback) / 1M
		volDollarSeries := make([]float64, 0, len(rows))
		for _, r := range rows {
			volDollarSeries = append(volDollarSeries, float64(r.Volume)*r.Close)
		}
		avgVolDollarsM := calculations.SimpleMovingAverage(volDollarSeries, lookback) / 1_000_000.0

		matchesThreshold := true
		if minVolDollarsM != nil && avgVolDollarsM < *minVolDollarsM {
			matchesThreshold = false
		}
		if maxVolDollarsM != nil && avgVolDollarsM > *maxVolDollarsM {
			matchesThreshold = false
		}
		if matchesThreshold {
			matches = append(matches, sym)
		}
	}

	return matches, nil
}

// GetSymbolsByAvgVolumePercent scans all symbols and returns those whose current volume
// as a percentage of average volume (SMA over lookback) falls within the thresholds.
// Volume % = (current volume / SMA(volume, lookback)) * 100
func (s *VolumeScreeningService) GetSymbolsByAvgVolumePercent(rangeParam, interval string, lookback int, minVolPercent, maxVolPercent *float64) ([]string, error) {
	if rangeParam == "" || interval == "" || lookback <= 0 {
		return nil, errors.New("range, interval, and lookback (positive) are required")
	}

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
		var rows []model.Historical
		if err := s.db.Where("symbol = ? AND range = ? AND interval = ?", sym, rangeParam, interval).
			Order("epoch ASC").
			Find(&rows).Error; err != nil {
			continue
		}
		if len(rows) == 0 {
			continue
		}

		// Calculate volume %: (current volume / SMA(volume, lookback)) * 100
		volumes := make([]float64, 0, len(rows))
		for _, r := range rows {
			volumes = append(volumes, float64(r.Volume))
		}
		avgVolume := calculations.SimpleMovingAverage(volumes, lookback)
		if avgVolume == 0 {
			continue // skip if average is zero
		}
		last := rows[len(rows)-1]
		volPercent := (float64(last.Volume) / avgVolume) * 100.0

		matchesThreshold := true
		if minVolPercent != nil && volPercent < *minVolPercent {
			matchesThreshold = false
		}
		if maxVolPercent != nil && volPercent > *maxVolPercent {
			matchesThreshold = false
		}
		if matchesThreshold {
			matches = append(matches, sym)
		}
	}

	return matches, nil
}

// GetAvgVolumeDollarsForSymbol calculates and returns average daily volume in dollars (millions)
// for a specific symbol. Volume $ = SMA(volume * close, lookback) / 1,000,000
func (s *VolumeScreeningService) GetAvgVolumeDollarsForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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

	volDollarSeries := make([]float64, 0, len(rows))
	for _, r := range rows {
		volDollarSeries = append(volDollarSeries, float64(r.Volume)*r.Close)
	}
	avgVolDollarsM := calculations.SimpleMovingAverage(volDollarSeries, lookback) / 1_000_000.0

	return avgVolDollarsM, nil
}

// GetAvgVolumePercentForSymbol calculates and returns current volume as a percentage
// of average volume for a specific symbol. Volume % = (current volume / SMA(volume, lookback)) * 100
func (s *VolumeScreeningService) GetAvgVolumePercentForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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

	volumes := make([]float64, 0, len(rows))
	for _, r := range rows {
		volumes = append(volumes, float64(r.Volume))
	}
	avgVolume := calculations.SimpleMovingAverage(volumes, lookback)
	if avgVolume == 0 {
		return 0, errors.New("average volume is zero")
	}

	last := rows[len(rows)-1]
	volPercent := (float64(last.Volume) / avgVolume) * 100.0

	return volPercent, nil
}

