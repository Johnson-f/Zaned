package service

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HistoricalService contains business logic for historical price operations
type HistoricalService struct {
	db *gorm.DB
}

// IndicatorLookbacks controls window sizes for indicator calculations
type IndicatorLookbacks struct {
	ATR       int // e.g., 14
	ADR       int // e.g., 14
	VolumeSMA int // e.g., 50
	MA        int // e.g., 50
}

// IndicatorSnapshot represents a single-bar snapshot of computed indicators
type IndicatorSnapshot struct {
	Symbol               string  `json:"symbol"`
	Range                string  `json:"range"`
	Interval             string  `json:"interval"`
	Epoch                int64   `json:"epoch"`
	ATRPercent           float64 `json:"atr_percent"`
	ADRPercent           float64 `json:"adr_percent"`
	DailyClosingRangePct float64 `json:"daily_closing_range_percent"`
	VolumeDollarsSMA_M   float64 `json:"volume_dollars_sma_m"`
	DailyVolumeDollarsM  float64 `json:"daily_volume_dollars_m"`
	PercentGainFromMA    float64 `json:"percent_gain_from_ma"`
	InsideDay            bool    `json:"inside_day"`
}

// ComputeIndicators computes a snapshot of indicators for the most recent bar
// using historical table data constrained by symbol/range/interval and lookbacks.
// It expects data ordered by epoch ascending.
func (s *HistoricalService) ComputeIndicators(symbol, rangeParam, interval string, lookbacks IndicatorLookbacks) (*IndicatorSnapshot, error) {
	if symbol == "" || rangeParam == "" || interval == "" {
		return nil, errors.New("symbol, range and interval are required")
	}

	var rows []model.Historical
	if err := s.db.Where("symbol = ? AND range = ? AND interval = ?", symbol, rangeParam, interval).
		Order("epoch ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("no historical data")
	}

	// Most recent bar (data is ordered ASC)
	last := rows[len(rows)-1]

	// ATR% using TR over consecutive bars, then SMA(ATR, n) / close * 100
	var atrPct float64 = 0
	if lookbacks.ATR > 0 {
		atr := averageTrueRange(rows, lookbacks.ATR)
		if last.Close != 0 {
			atrPct = (atr / last.Close) * 100.0
		}
	}

	// ADR% = SMA(high-low, n) / close * 100
	var adrPct float64 = 0
	if lookbacks.ADR > 0 {
		rngSeries := make([]float64, 0, len(rows))
		for _, r := range rows {
			rngSeries = append(rngSeries, r.High-r.Low)
		}
		adr := simpleMovingAverage(rngSeries, lookbacks.ADR)
		if last.Close != 0 {
			adrPct = (adr / last.Close) * 100.0
		}
	}

	// Daily Closing Range % = (close - low) / (high - low) * 100 for the last bar
	var dcr float64 = 0
	denom := last.High - last.Low
	if denom > 0 {
		dcr = (last.Close - last.Low) / denom * 100.0
	}

	// Volume dollars SMA (in millions) over lookbacks.VolumeSMA
	var volDollarsSMAM float64 = 0
	if lookbacks.VolumeSMA > 0 {
		volDollarSeries := make([]float64, 0, len(rows))
		for _, r := range rows {
			volDollarSeries = append(volDollarSeries, float64(r.Volume)*r.Close)
		}
		v := simpleMovingAverage(volDollarSeries, lookbacks.VolumeSMA)
		volDollarsSMAM = v / 1_000_000.0
	}

	// Daily volume dollars (last bar), in millions
	dailyVolDollarsM := (float64(last.Volume) * last.Close) / 1_000_000.0

	// Percent gain from MA = (close - SMA(close, n)) / SMA(close, n) * 100
	var pctFromMA float64 = 0
	if lookbacks.MA > 0 {
		closes := make([]float64, 0, len(rows))
		for _, r := range rows {
			closes = append(closes, r.Close)
		}
		ma := simpleMovingAverage(closes, lookbacks.MA)
		if ma != 0 {
			pctFromMA = ((last.Close - ma) / ma) * 100.0
		}
	}

	// Inside day: last high < previous high AND last low > previous low
	insideDay := false
	if len(rows) >= 2 {
		prev := rows[len(rows)-2]
		insideDay = last.High < prev.High && last.Low > prev.Low
	}

	return &IndicatorSnapshot{
		Symbol:               symbol,
		Range:                rangeParam,
		Interval:             interval,
		Epoch:                last.Epoch,
		ATRPercent:           atrPct,
		ADRPercent:           adrPct,
		DailyClosingRangePct: dcr,
		VolumeDollarsSMA_M:   volDollarsSMAM,
		DailyVolumeDollarsM:  dailyVolDollarsM,
		PercentGainFromMA:    pctFromMA,
		InsideDay:            insideDay,
	}, nil
}

// simpleMovingAverage returns SMA over the last N values of the series.
// If there are fewer than N points, it averages available points; if series is empty returns 0.
func simpleMovingAverage(series []float64, n int) float64 {
	if n <= 0 || len(series) == 0 {
		return 0
	}
	if len(series) < n {
		// average over available
		var sum float64
		for _, v := range series {
			sum += v
		}
		return sum / float64(len(series))
	}
	var sum float64
	for i := len(series) - n; i < len(series); i++ {
		sum += series[i]
	}
	return sum / float64(n)
}

// averageTrueRange computes ATR over the last N bars using Wilder's SMA of True Range.
// If fewer than N bars, it averages available TRs.
func averageTrueRange(rows []model.Historical, n int) float64 {
	if n <= 0 || len(rows) == 0 {
		return 0
	}
	// Build TR series
	trs := make([]float64, 0, len(rows))
	for i := range rows {
		cur := rows[i]
		var prevClose float64
		if i == 0 {
			prevClose = cur.Close
		} else {
			prevClose = rows[i-1].Close
		}
		hl := cur.High - cur.Low
		hc := abs(cur.High - prevClose)
		lc := abs(cur.Low - prevClose)
		tr := hl
		if hc > tr {
			tr = hc
		}
		if lc > tr {
			tr = lc
		}
		trs = append(trs, tr)
	}
	return simpleMovingAverage(trs, n)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

// NewHistoricalService creates a new instance of HistoricalService
func NewHistoricalService() *HistoricalService {
	return &HistoricalService{
		db: database.GetDB(),
	}
}

// GetAllHistorical fetches all historical records (read-only)
func (s *HistoricalService) GetAllHistorical() ([]model.Historical, error) {
	var historical []model.Historical
	result := s.db.Find(&historical)
	if result.Error != nil {
		return nil, result.Error
	}

	return historical, nil
}

// GetHistoricalByID fetches a historical record by ID
func (s *HistoricalService) GetHistoricalByID(id string) (*model.Historical, error) {
	var historical model.Historical
	result := s.db.Where("id = ?", id).First(&historical)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &historical, nil
}

// GetSymbolsWithDailyInsideDay scans all symbols in the historical table
// and returns those whose latest daily bar is an inside day compared to the
// immediately previous daily bar.
// An inside day means: current high < previous high AND current low > previous low
func (s *HistoricalService) GetSymbolsWithDailyInsideDay() ([]string, error) {
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

// GetSymbolsWithHighestVolumeInQuarter scans all symbols' daily bars (interval='1d')
// within the last 90 days (inclusive) and returns those whose most recent daily bar
// has the highest volume in that 90-day window.
func (s *HistoricalService) GetSymbolsWithHighestVolumeInQuarter() ([]string, error) {
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

	// Epoch window: now and 90 days ago
	now := time.Now()
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	maxEpoch := now.Unix()
	minEpoch := ninetyDaysAgo.Unix()

	matches := make([]string, 0)
	for _, sym := range symbols {
		// Fetch last 90 days of daily bars for this symbol
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

// GetSymbolsWithHighestVolumeInYear scans all symbols' daily bars (interval='1d')
// within the last 365 days (inclusive) and returns those whose most recent daily bar
// has the highest volume in that 365-day window.
func (s *HistoricalService) GetSymbolsWithHighestVolumeInYear() ([]string, error) {
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

// GetSymbolsWithHighestVolumeEver scans all symbols' daily bars (interval='1d')
// across all available history in the database and returns those whose most recent
// daily bar has the highest volume ever observed for that symbol.
func (s *HistoricalService) GetSymbolsWithHighestVolumeEver() ([]string, error) {
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

// GetSymbolsByADR scans all symbols with the given range/interval and returns those
// whose ADR% (Average Daily Range as percentage) falls within the specified thresholds.
// ADR% = SMA(high-low, lookback) / close * 100
func (s *HistoricalService) GetSymbolsByADR(rangeParam, interval string, lookback int, minADR, maxADR *float64) ([]string, error) {
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
		adr := simpleMovingAverage(rngSeries, lookback)
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

// GetSymbolsByATR scans all symbols with the given range/interval and returns those
// whose ATR% (Average True Range as percentage) falls within the specified thresholds.
// ATR% = ATR(lookback) / close * 100
// ATR is calculated as SMA of True Range (max of: high-low, |high-prevClose|, |low-prevClose|)
func (s *HistoricalService) GetSymbolsByATR(rangeParam, interval string, lookback int, minATR, maxATR *float64) ([]string, error) {
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
		atr := averageTrueRange(rows, lookback)
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

// GetADRForSymbol calculates and returns ADR% for a specific symbol.
// ADR% = SMA(high-low, lookback) / close * 100
func (s *HistoricalService) GetADRForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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
	adr := simpleMovingAverage(rngSeries, lookback)
	last := rows[len(rows)-1]
	if last.Close == 0 {
		return 0, errors.New("invalid close price (zero)")
	}
	adrPercent := (adr / last.Close) * 100.0

	return adrPercent, nil
}

// GetATRForSymbol calculates and returns ATR% for a specific symbol.
// ATR% = ATR(lookback) / close * 100
// ATR is calculated as SMA of True Range (max of: high-low, |high-prevClose|, |low-prevClose|)
func (s *HistoricalService) GetATRForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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
	atr := averageTrueRange(rows, lookback)
	last := rows[len(rows)-1]
	if last.Close == 0 {
		return 0, errors.New("invalid close price (zero)")
	}
	atrPercent := (atr / last.Close) * 100.0

	return atrPercent, nil
}

// GetSymbolsByAvgVolumeDollars scans all symbols and returns those whose average daily
// volume in dollars (SMA of volume*close over lookback) falls within the thresholds.
// Volume in dollars = volume * close, then SMA over lookback, then convert to millions ($M)
func (s *HistoricalService) GetSymbolsByAvgVolumeDollars(rangeParam, interval string, lookback int, minVolDollarsM, maxVolDollarsM *float64) ([]string, error) {
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
		avgVolDollarsM := simpleMovingAverage(volDollarSeries, lookback) / 1_000_000.0

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
func (s *HistoricalService) GetSymbolsByAvgVolumePercent(rangeParam, interval string, lookback int, minVolPercent, maxVolPercent *float64) ([]string, error) {
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
		avgVolume := simpleMovingAverage(volumes, lookback)
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
func (s *HistoricalService) GetAvgVolumeDollarsForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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
	avgVolDollarsM := simpleMovingAverage(volDollarSeries, lookback) / 1_000_000.0

	return avgVolDollarsM, nil
}

// GetAvgVolumePercentForSymbol calculates and returns current volume as a percentage
// of average volume for a specific symbol. Volume % = (current volume / SMA(volume, lookback)) * 100
func (s *HistoricalService) GetAvgVolumePercentForSymbol(symbol, rangeParam, interval string, lookback int) (float64, error) {
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
	avgVolume := simpleMovingAverage(volumes, lookback)
	if avgVolume == 0 {
		return 0, errors.New("average volume is zero")
	}

	last := rows[len(rows)-1]
	volPercent := (float64(last.Volume) / avgVolume) * 100.0

	return volPercent, nil
}

// CreateHistorical creates a new historical record
func (s *HistoricalService) CreateHistorical(historical *model.Historical) error {
	if historical == nil {
		return errors.New("historical record cannot be nil")
	}

	result := s.db.Create(historical)
	if result.Error != nil {
		return fmt.Errorf("failed to create historical record: %w", result.Error)
	}

	return nil
}

// CreateHistoricalBatch creates multiple historical records in a single transaction
func (s *HistoricalService) CreateHistoricalBatch(historical []model.Historical) error {
	if len(historical) == 0 {
		return errors.New("historical records cannot be empty")
	}

	result := s.db.CreateInBatches(historical, 100)
	if result.Error != nil {
		return fmt.Errorf("failed to create historical records: %w", result.Error)
	}

	return nil
}

// UpsertHistorical upserts a historical record based on symbol, epoch, range, and interval
func (s *HistoricalService) UpsertHistorical(historical *model.Historical) error {
	if historical == nil {
		return errors.New("historical record cannot be nil")
	}

	result := s.db.Where("symbol = ? AND epoch = ? AND range = ? AND interval = ?",
		historical.Symbol, historical.Epoch, historical.Range, historical.Interval).
		Assign(map[string]interface{}{
			"open":      historical.Open,
			"high":      historical.High,
			"low":       historical.Low,
			"close":     historical.Close,
			"adj_close": historical.AdjClose,
			"volume":    historical.Volume,
		}).
		FirstOrCreate(historical)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert historical record: %w", result.Error)
	}

	return nil
}

// UpsertHistoricalBatch upserts multiple historical records
func (s *HistoricalService) UpsertHistoricalBatch(historical []model.Historical) error {
	if len(historical) == 0 {
		return errors.New("historical records cannot be empty")
	}

	// Prefer bulk upsert using ON CONFLICT for performance
	// Requires unique index on (symbol, epoch, range, interval)
	return s.db.Clauses(clause.OnConflict{
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
	}).CreateInBatches(historical, 100).Error
}

// UpdateHistorical updates an existing historical record
func (s *HistoricalService) UpdateHistorical(id string, historical *model.Historical) error {
	if historical == nil {
		return errors.New("historical record cannot be nil")
	}

	result := s.db.Model(&model.Historical{}).Where("id = ?", id).Updates(historical)
	if result.Error != nil {
		return fmt.Errorf("failed to update historical record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}
