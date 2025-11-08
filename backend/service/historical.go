package service

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/filtering"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HistoricalService contains business logic for historical price operations
type HistoricalService struct {
	db *gorm.DB
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

// GetHistoricalBySymbolRangeInterval fetches historical records filtered by symbol, range, and interval
// Returns records ordered by epoch ascending
func (s *HistoricalService) GetHistoricalBySymbolRangeInterval(symbol, rangeParam, interval string) ([]model.Historical, error) {
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}
	if rangeParam == "" {
		return nil, errors.New("range is required")
	}
	if interval == "" {
		return nil, errors.New("interval is required")
	}

	var historical []model.Historical
	result := s.db.Where("symbol = ? AND range = ? AND interval = ?", symbol, rangeParam, interval).
		Order("epoch ASC").
		Find(&historical)
	if result.Error != nil {
		return nil, result.Error
	}

	return historical, nil
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

// SaveInsideDayResults saves current inside day symbols to database
// This method delegates to the filtering service for inside-day logic
func (s *HistoricalService) SaveInsideDayResults() error {
	insideDayService := filtering.NewInsideDayService()
	return insideDayService.SaveInsideDayResults()
}

// SaveHighVolumeQuarterResults saves high volume quarter symbols
// This method delegates to the filtering service for high-volume-quarter logic
func (s *HistoricalService) SaveHighVolumeQuarterResults() error {
	highVolumeQuarterService := filtering.NewHighVolumeQuarterService()
	return highVolumeQuarterService.SaveHighVolumeQuarterResults()
}

// SaveHighVolumeYearResults saves high volume year symbols
// This method delegates to the filtering service for high-volume-year logic
func (s *HistoricalService) SaveHighVolumeYearResults() error {
	highVolumeYearService := filtering.NewHighVolumeYearService()
	return highVolumeYearService.SaveHighVolumeYearResults()
}

// SaveHighVolumeEverResults saves high volume ever symbols
// This method delegates to the filtering service for high-volume-ever logic
func (s *HistoricalService) SaveHighVolumeEverResults() error {
	highVolumeEverService := filtering.NewHighVolumeEverService()
	return highVolumeEverService.SaveHighVolumeEverResults()
}

// GetScreenerResults fetches screener results with time period filtering
func (s *HistoricalService) GetScreenerResults(resultType string, period string) ([]string, error) {
	query := s.db.Model(&model.ScreenerResult{}).
		Where("type = ?", resultType)

	now := time.Now()
	var startDate time.Time

	switch period {
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	case "ytd":
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	case "all":
		// No date filter for "all time"
		startDate = time.Time{}
	default:
		return nil, fmt.Errorf("invalid period: %s. Valid periods: 7d, 30d, 90d, ytd, all", period)
	}

	if !startDate.IsZero() {
		query = query.Where("date >= ?", startDate)
	}

	// Get distinct symbols
	var symbols []string
	if err := query.Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch screener results: %w", err)
	}

	return symbols, nil
}
