package calculations

import (
	"errors"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/filtering/indicators"

	"gorm.io/gorm"
)

// IndicatorCalculationService handles indicator calculations
type IndicatorCalculationService struct {
	db *gorm.DB
}

// NewIndicatorCalculationService creates a new instance of IndicatorCalculationService
func NewIndicatorCalculationService() *IndicatorCalculationService {
	return &IndicatorCalculationService{
		db: database.GetDB(),
	}
}

// ComputeIndicators computes a snapshot of indicators for the most recent bar
// using historical table data constrained by symbol/range/interval and lookbacks.
// It expects data ordered by epoch ascending.
func (s *IndicatorCalculationService) ComputeIndicators(symbol, rangeParam, interval string, lookbacks indicators.IndicatorLookbacks) (*indicators.IndicatorSnapshot, error) {
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
		atr := AverageTrueRange(rows, lookbacks.ATR)
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
		adr := SimpleMovingAverage(rngSeries, lookbacks.ADR)
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
		v := SimpleMovingAverage(volDollarSeries, lookbacks.VolumeSMA)
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
		ma := SimpleMovingAverage(closes, lookbacks.MA)
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

	return &indicators.IndicatorSnapshot{
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
