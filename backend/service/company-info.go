package service

import (
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"

	"gorm.io/gorm"
)

// CompanyInfoService contains business logic for company info operations
type CompanyInfoService struct {
	db *gorm.DB
}

// NewCompanyInfoService creates a new instance of CompanyInfoService
func NewCompanyInfoService() *CompanyInfoService {
	return &CompanyInfoService{
		db: database.GetDB(),
	}
}

// GetAllCompanyInfo fetches all company info records (read-only)
func (s *CompanyInfoService) GetAllCompanyInfo() ([]model.CompanyInfo, error) {
	var companyInfo []model.CompanyInfo
	result := s.db.Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info: %w", result.Error)
	}

	return companyInfo, nil
}

// GetCompanyInfoBySymbol fetches company info by symbol
func (s *CompanyInfoService) GetCompanyInfoBySymbol(symbol string) (*model.CompanyInfo, error) {
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}

	var companyInfo model.CompanyInfo
	result := s.db.Where("symbol = ?", symbol).First(&companyInfo)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, fmt.Errorf("failed to fetch company info: %w", result.Error)
	}

	return &companyInfo, nil
}

// GetCompanyInfoBySymbols fetches multiple company info records by symbols
func (s *CompanyInfoService) GetCompanyInfoBySymbols(symbols []string) ([]model.CompanyInfo, error) {
	if len(symbols) == 0 {
		return []model.CompanyInfo{}, nil
	}

	var companyInfo []model.CompanyInfo
	result := s.db.Where("symbol IN ?", symbols).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info: %w", result.Error)
	}

	return companyInfo, nil
}

// SearchCompanyInfo searches company info by name, sector, or industry
func (s *CompanyInfoService) SearchCompanyInfo(searchTerm string) ([]model.CompanyInfo, error) {
	if searchTerm == "" {
		return []model.CompanyInfo{}, nil
	}

	var companyInfo []model.CompanyInfo
	searchPattern := "%" + searchTerm + "%"
	result := s.db.Where(
		"name ILIKE ? OR sector ILIKE ? OR industry ILIKE ? OR symbol ILIKE ?",
		searchPattern, searchPattern, searchPattern, searchPattern,
	).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search company info: %w", result.Error)
	}

	return companyInfo, nil
}

// GetCompanyInfoBySector fetches all company info records for a specific sector
func (s *CompanyInfoService) GetCompanyInfoBySector(sector string) ([]model.CompanyInfo, error) {
	if sector == "" {
		return []model.CompanyInfo{}, nil
	}

	var companyInfo []model.CompanyInfo
	result := s.db.Where("sector = ?", sector).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info by sector: %w", result.Error)
	}

	return companyInfo, nil
}

// GetCompanyInfoByIndustry fetches all company info records for a specific industry
func (s *CompanyInfoService) GetCompanyInfoByIndustry(industry string) ([]model.CompanyInfo, error) {
	if industry == "" {
		return []model.CompanyInfo{}, nil
	}

	var companyInfo []model.CompanyInfo
	result := s.db.Where("industry = ?", industry).Find(&companyInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch company info by industry: %w", result.Error)
	}

	return companyInfo, nil
}
