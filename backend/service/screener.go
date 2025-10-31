package service

import (
	"errors"
	"screener/backend/database"
	"screener/backend/model"

	"gorm.io/gorm"
)

// ScreenerService contains business logic for screener operations
type ScreenerService struct {
	db *gorm.DB
}

// NewScreenerService creates a new instance of ScreenerService
func NewScreenerService() *ScreenerService {
	return &ScreenerService{
		db: database.GetDB(),
	}
}

// GetAllScreeners fetches all screener records (read-only)
func (s *ScreenerService) GetAllScreeners() ([]model.Screener, error) {
	var screeners []model.Screener
	result := s.db.Find(&screeners)
	if result.Error != nil {
		return nil, result.Error
	}

	return screeners, nil
}

// GetScreenerByID fetches a screener record by ID
func (s *ScreenerService) GetScreenerByID(id string) (*model.Screener, error) {
	var screener model.Screener
	result := s.db.Where("id = ?", id).First(&screener)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &screener, nil
}

// GetScreenerBySymbol fetches a screener record by symbol
func (s *ScreenerService) GetScreenerBySymbol(symbol string) (*model.Screener, error) {
	var screener model.Screener
	result := s.db.Where("symbol = ?", symbol).First(&screener)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, result.Error
	}

	return &screener, nil
}

// GetScreenersBySymbols fetches multiple screener records by symbols
func (s *ScreenerService) GetScreenersBySymbols(symbols []string) ([]model.Screener, error) {
	var screeners []model.Screener
	result := s.db.Where("symbol IN ?", symbols).Find(&screeners)
	if result.Error != nil {
		return nil, result.Error
	}

	return screeners, nil
}
