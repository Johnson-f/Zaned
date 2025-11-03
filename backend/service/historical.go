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

// HistoricalFilterOptions represents filtering options for historical queries
type HistoricalFilterOptions struct {
	Symbol    *string
	MinEpoch  *int64
	MaxEpoch  *int64
	Range     *string
	Interval  *string
	MinOpen   *float64
	MaxOpen   *float64
	MinHigh   *float64
	MaxHigh   *float64
	MinLow    *float64
	MaxLow    *float64
	MinClose  *float64
	MaxClose  *float64
	MinVolume *int64
	MaxVolume *int64
}

// HistoricalSortOptions represents sorting options for historical queries
type HistoricalSortOptions struct {
	Field     string // "symbol", "epoch", "open", "high", "low", "close", "volume", "created_at"
	Direction string // "asc" or "desc"
}

// HistoricalPaginationOptions represents pagination options
type HistoricalPaginationOptions struct {
	Page  int // 1-indexed page number
	Limit int // Number of records per page
}

// HistoricalQueryResult represents a paginated query result
type HistoricalQueryResult struct {
	Data       []model.Historical `json:"data"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	Total      int64              `json:"total"`
	TotalPages int                `json:"total_pages"`
}

// VolumeMetricsResult represents volume metrics for a stock
type VolumeMetricsResult struct {
	Symbol                 string `json:"symbol"`
	HighestVolumeInYear    int64  `json:"highest_volume_in_year"`
	HighestVolumeInQuarter int64  `json:"highest_volume_in_quarter"`
	HighestVolumeEver      int64  `json:"highest_volume_ever"`
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

// GetHistoricalBySymbol fetches all historical records for a specific symbol
func (s *HistoricalService) GetHistoricalBySymbol(symbol string) ([]model.Historical, error) {
	var historical []model.Historical
	result := s.db.Where("symbol = ?", symbol).
		Order("epoch ASC").
		Find(&historical)
	if result.Error != nil {
		return nil, result.Error
	}

	return historical, nil
}

// GetHistoricalBySymbolAndParams fetches historical records by symbol, range, and interval
func (s *HistoricalService) GetHistoricalBySymbolAndParams(symbol, rangeParam, interval string) ([]model.Historical, error) {
	var historical []model.Historical
	result := s.db.Where("symbol = ? AND range = ? AND interval = ?", symbol, rangeParam, interval).
		Order("epoch ASC").
		Find(&historical)
	if result.Error != nil {
		return nil, result.Error
	}

	return historical, nil
}

// GetHistoricalByEpochRange fetches historical records within an epoch range
func (s *HistoricalService) GetHistoricalByEpochRange(symbol string, minEpoch, maxEpoch int64) ([]model.Historical, error) {
	var historical []model.Historical
	result := s.db.Where("symbol = ? AND epoch >= ? AND epoch <= ?", symbol, minEpoch, maxEpoch).
		Order("epoch ASC").
		Find(&historical)
	if result.Error != nil {
		return nil, result.Error
	}

	return historical, nil
}

// GetHistoricalWithFilters fetches historical records with filtering, sorting, and pagination
func (s *HistoricalService) GetHistoricalWithFilters(
	filters *HistoricalFilterOptions,
	sort *HistoricalSortOptions,
	pagination *HistoricalPaginationOptions,
) (*HistoricalQueryResult, error) {
	query := s.db.Model(&model.Historical{})

	// Apply filters
	if filters != nil {
		if filters.Symbol != nil && *filters.Symbol != "" {
			query = query.Where("symbol = ?", *filters.Symbol)
		}
		if filters.MinEpoch != nil {
			query = query.Where("epoch >= ?", *filters.MinEpoch)
		}
		if filters.MaxEpoch != nil {
			query = query.Where("epoch <= ?", *filters.MaxEpoch)
		}
		if filters.Range != nil && *filters.Range != "" {
			query = query.Where("range = ?", *filters.Range)
		}
		if filters.Interval != nil && *filters.Interval != "" {
			query = query.Where("interval = ?", *filters.Interval)
		}
		if filters.MinOpen != nil {
			query = query.Where("open >= ?", *filters.MinOpen)
		}
		if filters.MaxOpen != nil {
			query = query.Where("open <= ?", *filters.MaxOpen)
		}
		if filters.MinHigh != nil {
			query = query.Where("high >= ?", *filters.MinHigh)
		}
		if filters.MaxHigh != nil {
			query = query.Where("high <= ?", *filters.MaxHigh)
		}
		if filters.MinLow != nil {
			query = query.Where("low >= ?", *filters.MinLow)
		}
		if filters.MaxLow != nil {
			query = query.Where("low <= ?", *filters.MaxLow)
		}
		if filters.MinClose != nil {
			query = query.Where("close >= ?", *filters.MinClose)
		}
		if filters.MaxClose != nil {
			query = query.Where("close <= ?", *filters.MaxClose)
		}
		if filters.MinVolume != nil {
			query = query.Where("volume >= ?", *filters.MinVolume)
		}
		if filters.MaxVolume != nil {
			query = query.Where("volume <= ?", *filters.MaxVolume)
		}
	}

	// Get total count before pagination
	var total int64
	countQuery := query
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count records: %w", err)
	}

	// Apply sorting
	if sort != nil && sort.Field != "" {
		direction := "ASC"
		if sort.Direction == "desc" {
			direction = "DESC"
		}
		// Validate field name to prevent SQL injection
		validFields := map[string]bool{
			"symbol":     true,
			"epoch":      true,
			"open":       true,
			"high":       true,
			"low":        true,
			"close":      true,
			"volume":     true,
			"created_at": true,
			"updated_at": true,
		}
		if validFields[sort.Field] {
			query = query.Order(fmt.Sprintf("%s %s", sort.Field, direction))
		}
	} else {
		// Default sorting by epoch
		query = query.Order("epoch ASC")
	}

	// Apply pagination
	if pagination != nil {
		if pagination.Page < 1 {
			pagination.Page = 1
		}
		if pagination.Limit < 1 {
			pagination.Limit = 10 // Default limit
		}
		if pagination.Limit > 100 {
			pagination.Limit = 100 // Max limit to prevent performance issues
		}
		offset := (pagination.Page - 1) * pagination.Limit
		query = query.Offset(offset).Limit(pagination.Limit)
	} else {
		// Default pagination if none provided
		pagination = &HistoricalPaginationOptions{Page: 1, Limit: 100}
		query = query.Limit(100)
	}

	// Execute query
	var historical []model.Historical
	result := query.Find(&historical)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch historical data: %w", result.Error)
	}

	// Calculate total pages
	totalPages := 1
	if pagination.Limit > 0 {
		totalPages = int((total + int64(pagination.Limit) - 1) / int64(pagination.Limit))
	}

	return &HistoricalQueryResult{
		Data:       historical,
		Page:       pagination.Page,
		Limit:      pagination.Limit,
		Total:      total,
		TotalPages: totalPages,
	}, nil
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
            "open":      gorm.Expr("excluded.open"),
            "high":      gorm.Expr("excluded.high"),
            "low":       gorm.Expr("excluded.low"),
            "close":     gorm.Expr("excluded.close"),
            "adj_close": gorm.Expr("excluded.adj_close"),
            "volume":    gorm.Expr("excluded.volume"),
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

// DeleteHistorical deletes a historical record by ID
func (s *HistoricalService) DeleteHistorical(id string) error {
	result := s.db.Where("id = ?", id).Delete(&model.Historical{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete historical record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

// DeleteHistoricalBySymbol deletes all historical records for a specific symbol
func (s *HistoricalService) DeleteHistoricalBySymbol(symbol string) error {
	result := s.db.Where("symbol = ?", symbol).Delete(&model.Historical{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete historical records: %w", result.Error)
	}

	return nil
}

// DeleteHistoricalBySymbolAndParams deletes historical records by symbol, range, and interval
func (s *HistoricalService) DeleteHistoricalBySymbolAndParams(symbol, rangeParam, interval string) error {
	result := s.db.Where("symbol = ? AND range = ? AND interval = ?", symbol, rangeParam, interval).
		Delete(&model.Historical{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete historical records: %w", result.Error)
	}

	return nil
}

// GetCount returns the total count of historical records
func (s *HistoricalService) GetCount() (int64, error) {
	var count int64
	result := s.db.Model(&model.Historical{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to get count: %w", result.Error)
	}

	return count, nil
}

// GetCountBySymbol returns the count of historical records for a specific symbol
func (s *HistoricalService) GetCountBySymbol(symbol string) (int64, error) {
	var count int64
	result := s.db.Model(&model.Historical{}).Where("symbol = ?", symbol).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to get count: %w", result.Error)
	}

	return count, nil
}

// dailyVolume represents a day's aggregated volume
type dailyVolume struct {
	Date   time.Time
	Volume int64
}

// GetStocksVolumeMetrics calculates volume metrics for all stocks in the database
// It fetches range="1d" and interval="30m" data, aggregates into daily volumes,
// and calculates highest volume in year (365 days), quarter (90 days), and ever
func (s *HistoricalService) GetStocksVolumeMetrics() ([]VolumeMetricsResult, error) {
	// Get all unique symbols from the historical table
	var symbols []string
	result := s.db.Model(&model.Historical{}).
		Distinct("symbol").
		Pluck("symbol", &symbols)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get unique symbols: %w", result.Error)
	}

	if len(symbols) == 0 {
		return []VolumeMetricsResult{}, nil
	}

	results := make([]VolumeMetricsResult, 0, len(symbols))
	currentTime := time.Now()
	oneYearAgo := currentTime.AddDate(0, 0, -365)
	oneQuarterAgo := currentTime.AddDate(0, 0, -90)

	// Convert current time to epoch for comparison
	currentEpoch := currentTime.Unix()
	oneYearAgoEpoch := oneYearAgo.Unix()
	oneQuarterAgoEpoch := oneQuarterAgo.Unix()

	// Process each symbol
	for _, symbol := range symbols {
		// Fetch historical records with range="1d" and interval="30m"
		var historical []model.Historical
		result := s.db.Where("symbol = ? AND range = ? AND interval = ?", symbol, "1d", "30m").
			Order("epoch ASC").
			Find(&historical)
		if result.Error != nil {
			// Log error but continue with next symbol
			continue
		}

		if len(historical) == 0 {
			// No data for this symbol, add with zero metrics
			results = append(results, VolumeMetricsResult{
				Symbol:                 symbol,
				HighestVolumeInYear:    0,
				HighestVolumeInQuarter: 0,
				HighestVolumeEver:      0,
			})
			continue
		}

		// Group records by day and sum volumes
		dailyVolumes := s.groupByDayAndSumVolumes(historical)

		// Calculate metrics
		highestInYear := s.findHighestVolumeInPeriod(dailyVolumes, oneYearAgoEpoch, currentEpoch)
		highestInQuarter := s.findHighestVolumeInPeriod(dailyVolumes, oneQuarterAgoEpoch, currentEpoch)
		highestEver := s.findHighestVolumeEver(dailyVolumes)

		results = append(results, VolumeMetricsResult{
			Symbol:                 symbol,
			HighestVolumeInYear:    highestInYear,
			HighestVolumeInQuarter: highestInQuarter,
			HighestVolumeEver:      highestEver,
		})
	}

	return results, nil
}

// groupByDayAndSumVolumes groups historical records by day and sums their volumes
func (s *HistoricalService) groupByDayAndSumVolumes(historical []model.Historical) []dailyVolume {
	if len(historical) == 0 {
		return []dailyVolume{}
	}

	// Map to store daily volumes: date (as string) -> volume sum
	dailyMap := make(map[string]int64)

	for _, h := range historical {
		// Convert epoch to time
		t := time.Unix(h.Epoch, 0)
		// Get date at midnight (normalize to start of day)
		dateKey := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
		dateKeyStr := dateKey.Format("2006-01-02")

		// Sum volumes for the same day
		dailyMap[dateKeyStr] += h.Volume
	}

	// Convert map to slice of dailyVolume
	dailyVolumes := make([]dailyVolume, 0, len(dailyMap))
	for dateStr, volume := range dailyMap {
		// Parse the date string back to time.Time
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}
		dailyVolumes = append(dailyVolumes, dailyVolume{
			Date:   date,
			Volume: volume,
		})
	}

	return dailyVolumes
}

// findHighestVolumeInPeriod finds the highest daily volume within a time period (epoch range)
func (s *HistoricalService) findHighestVolumeInPeriod(dailyVolumes []dailyVolume, minEpoch, maxEpoch int64) int64 {
	if len(dailyVolumes) == 0 {
		return 0
	}

	// Convert epochs to time for date comparison
	minDate := time.Unix(minEpoch, 0)
	maxDate := time.Unix(maxEpoch, 0)

	// Normalize to start of day for proper date comparison
	minDateNormalized := time.Date(minDate.Year(), minDate.Month(), minDate.Day(), 0, 0, 0, 0, minDate.Location())
	maxDateNormalized := time.Date(maxDate.Year(), maxDate.Month(), maxDate.Day(), 0, 0, 0, 0, maxDate.Location())

	var highest int64 = 0
	for _, dv := range dailyVolumes {
		// Compare dates (normalized to start of day) instead of exact epochs
		// This ensures we include any day that overlaps with the period
		if !dv.Date.Before(minDateNormalized) && !dv.Date.After(maxDateNormalized) {
			if dv.Volume > highest {
				highest = dv.Volume
			}
		}
	}

	return highest
}

// findHighestVolumeEver finds the highest daily volume across all time
func (s *HistoricalService) findHighestVolumeEver(dailyVolumes []dailyVolume) int64 {
	if len(dailyVolumes) == 0 {
		return 0
	}

	var highest int64 = 0
	for _, dv := range dailyVolumes {
		if dv.Volume > highest {
			highest = dv.Volume
		}
	}

	return highest
}
