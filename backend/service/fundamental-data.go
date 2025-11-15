package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"screener/backend/database"
	"screener/backend/model"
	"screener/backend/service/caching"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// FundamentalDataService contains business logic for fundamental data operations
type FundamentalDataService struct {
	db    *gorm.DB
	cache *caching.CacheService
	ttl   *caching.CacheTTLConfig
}

// NewFundamentalDataService creates a new instance of FundamentalDataService
func NewFundamentalDataService() *FundamentalDataService {
	return &FundamentalDataService{
		db:    database.GetDB(),
		cache: caching.NewCacheService(),
		ttl:   caching.GetTTLConfig(),
	}
}

// StatementRow represents a single row in the financial statement
type StatementRow struct {
	Breakdown string            `json:"Breakdown"`
	Dates     map[string]string `json:"-"` // Dates will be parsed from JSON
}

// StatementData represents the parsed statement structure
type StatementData map[string]StatementRow

// ParsedStatement represents a parsed financial statement
type ParsedStatement struct {
	Rows []StatementRowWithDates `json:"rows"`
}

// StatementRowWithDates represents a statement row with parsed date values
type StatementRowWithDates struct {
	Breakdown string             `json:"breakdown"`
	Dates     map[string]float64 `json:"dates"`    // Date -> numeric value
	RawDates  map[string]string  `json:"rawDates"` // Date -> raw string value
}

// FundamentalMetrics represents calculated metrics from financial data
type FundamentalMetrics struct {
	Symbol            string             `json:"symbol"`
	StatementType     string             `json:"statementType"`
	Frequency         string             `json:"frequency"`
	RevenueGrowthQoQ  *float64           `json:"revenueGrowthQoQ,omitempty"`  // Quarter over Quarter %
	RevenueGrowthYoY  *float64           `json:"revenueGrowthYoY,omitempty"`  // Year over Year %
	EPS               map[string]float64 `json:"eps,omitempty"`               // Date -> EPS value
	GrossProfitMargin map[string]float64 `json:"grossProfitMargin,omitempty"` // Date -> %
	OperatingMargin   map[string]float64 `json:"operatingMargin,omitempty"`   // Date -> %
	NetMargin         map[string]float64 `json:"netMargin,omitempty"`         // Date -> %
	TotalRevenue      map[string]float64 `json:"totalRevenue,omitempty"`
	GrossProfit       map[string]float64 `json:"grossProfit,omitempty"`
	OperatingIncome   map[string]float64 `json:"operatingIncome,omitempty"`
	NetIncome         map[string]float64 `json:"netIncome,omitempty"`
	ParsedStatement   *ParsedStatement   `json:"parsedStatement,omitempty"`
}

// GetAllFundamentalData fetches all fundamental data records
func (s *FundamentalDataService) GetAllFundamentalData() ([]model.FundamentalData, error) {
	cacheKey := caching.GenerateKeyFromPath("fundamental-data")
	var fundamentalData []model.FundamentalData
	
	found, err := s.cache.GetJSON(cacheKey, &fundamentalData)
	if err == nil && found {
		return fundamentalData, nil
	}

	result := s.db.Find(&fundamentalData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch all fundamental data: %w", result.Error)
	}
	
	_ = s.cache.SetJSON(cacheKey, fundamentalData, s.ttl.FundamentalData)
	return fundamentalData, nil
}

// GetFundamentalDataBySymbol fetches fundamental data by symbol
func (s *FundamentalDataService) GetFundamentalDataBySymbol(symbol string) ([]model.FundamentalData, error) {
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}

	cacheKey := caching.GenerateKeyFromPath(fmt.Sprintf("fundamental-data/symbol/%s", symbol))
	var fundamentalData []model.FundamentalData
	
	found, err := s.cache.GetJSON(cacheKey, &fundamentalData)
	if err == nil && found {
		return fundamentalData, nil
	}

	result := s.db.Where("symbol = ?", symbol).Find(&fundamentalData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch fundamental data by symbol: %w", result.Error)
	}
	
	_ = s.cache.SetJSON(cacheKey, fundamentalData, s.ttl.FundamentalData)
	return fundamentalData, nil
}

// GetFundamentalDataBySymbolAndType fetches fundamental data by symbol and statement type
func (s *FundamentalDataService) GetFundamentalDataBySymbolAndType(symbol, statementType string) (*model.FundamentalData, error) {
	if symbol == "" || statementType == "" {
		return nil, errors.New("symbol and statement type are required")
	}

	var fundamentalData model.FundamentalData
	result := s.db.Where("symbol = ? AND statement_type = ?", symbol, statementType).First(&fundamentalData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, fmt.Errorf("failed to fetch fundamental data: %w", result.Error)
	}
	return &fundamentalData, nil
}

// GetFundamentalDataBySymbolTypeAndFrequency fetches fundamental data by symbol, statement type, and frequency
// Checks Redis first, then database
func (s *FundamentalDataService) GetFundamentalDataBySymbolTypeAndFrequency(symbol, statementType, frequency string) (*model.FundamentalData, error) {
	if symbol == "" || statementType == "" || frequency == "" {
		return nil, errors.New("symbol, statement type, and frequency are required")
	}

	// Check Redis first (using DataCache)
	dataCache := caching.NewDataCache()
	fundamentalData, found, err := dataCache.GetFundamentalData(symbol, statementType, frequency)
	if err == nil && found {
		return fundamentalData, nil
	}

	// Redis miss - check database
	var dbFundamentalData model.FundamentalData
	result := s.db.Where("symbol = ? AND statement_type = ? AND frequency = ?", symbol, statementType, frequency).First(&dbFundamentalData)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("record not found")
		}
		return nil, fmt.Errorf("failed to fetch fundamental data: %w", result.Error)
	}
	
	// If found in database, cache it in Redis for next time
	_ = dataCache.CacheFundamentalData(symbol, statementType, frequency, &dbFundamentalData)
	return &dbFundamentalData, nil
}

// GetFundamentalDataByStatementType fetches all fundamental data for a specific statement type
func (s *FundamentalDataService) GetFundamentalDataByStatementType(statementType string) ([]model.FundamentalData, error) {
	if statementType == "" {
		return []model.FundamentalData{}, nil
	}

	var fundamentalData []model.FundamentalData
	result := s.db.Where("statement_type = ?", statementType).Find(&fundamentalData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch fundamental data by statement type: %w", result.Error)
	}
	return fundamentalData, nil
}

// GetFundamentalDataByFrequency fetches all fundamental data for a specific frequency
func (s *FundamentalDataService) GetFundamentalDataByFrequency(frequency string) ([]model.FundamentalData, error) {
	if frequency == "" {
		return []model.FundamentalData{}, nil
	}

	var fundamentalData []model.FundamentalData
	result := s.db.Where("frequency = ?", frequency).Find(&fundamentalData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to fetch fundamental data by frequency: %w", result.Error)
	}
	return fundamentalData, nil
}

// GetFundamentalMetrics calculates metrics from fundamental data
func (s *FundamentalDataService) GetFundamentalMetrics(symbol, statementType, frequency string) (*FundamentalMetrics, error) {
	fundamentalData, err := s.GetFundamentalDataBySymbolTypeAndFrequency(symbol, statementType, frequency)
	if err != nil {
		return nil, err
	}

	return s.calculateMetrics(fundamentalData)
}

// calculateMetrics calculates various financial metrics from the statement data
func (s *FundamentalDataService) calculateMetrics(fundamentalData *model.FundamentalData) (*FundamentalMetrics, error) {
	metrics := &FundamentalMetrics{
		Symbol:        fundamentalData.Symbol,
		StatementType: fundamentalData.StatementType,
		Frequency:     fundamentalData.Frequency,
	}

	// Parse the statement JSON
	parsedStatement, err := s.parseStatement(fundamentalData.Statement)
	if err != nil {
		return nil, fmt.Errorf("failed to parse statement: %w", err)
	}
	metrics.ParsedStatement = parsedStatement

	// Extract key metrics
	metrics.TotalRevenue = s.extractMetric(parsedStatement, "Total Revenue")
	metrics.GrossProfit = s.extractMetric(parsedStatement, "Gross Profit")
	metrics.OperatingIncome = s.extractMetric(parsedStatement, "Operating Income")
	metrics.NetIncome = s.extractMetric(parsedStatement, "Net Income Common Stockholders")
	if len(metrics.NetIncome) == 0 {
		// Try alternative name
		metrics.NetIncome = s.extractMetric(parsedStatement, "Net Income(Attributable to Parent Company Shareholders)")
	}

	// Extract EPS
	metrics.EPS = s.extractMetric(parsedStatement, "Diluted EPS")
	if len(metrics.EPS) == 0 {
		metrics.EPS = s.extractMetric(parsedStatement, "Basic EPS")
	}

	// Calculate margins
	metrics.GrossProfitMargin = s.calculateMargin(metrics.GrossProfit, metrics.TotalRevenue)
	metrics.OperatingMargin = s.calculateMargin(metrics.OperatingIncome, metrics.TotalRevenue)
	metrics.NetMargin = s.calculateMargin(metrics.NetIncome, metrics.TotalRevenue)

	// Calculate growth rates
	if fundamentalData.Frequency == "quarterly" {
		metrics.RevenueGrowthQoQ = s.calculateQoQGrowth(metrics.TotalRevenue)
	}
	metrics.RevenueGrowthYoY = s.calculateYoYGrowth(metrics.TotalRevenue)

	return metrics, nil
}

// parseStatement parses the JSONB statement string into a structured format
func (s *FundamentalDataService) parseStatement(statementJSON string) (*ParsedStatement, error) {
	var rawStatement map[string]interface{}
	if err := json.Unmarshal([]byte(statementJSON), &rawStatement); err != nil {
		return nil, fmt.Errorf("failed to unmarshal statement JSON: %w", err)
	}

	parsed := &ParsedStatement{
		Rows: make([]StatementRowWithDates, 0),
	}

	for _, value := range rawStatement {
		rowMap, ok := value.(map[string]interface{})
		if !ok {
			continue
		}

		breakdown, _ := rowMap["Breakdown"].(string)
		row := StatementRowWithDates{
			Breakdown: breakdown,
			Dates:     make(map[string]float64),
			RawDates:  make(map[string]string),
		}

		// Extract date fields (everything except "Breakdown")
		for k, v := range rowMap {
			if k == "Breakdown" {
				continue
			}
			// Date keys are in format "YYYY-MM-DD"
			if strVal, ok := v.(string); ok {
				row.RawDates[k] = strVal
				// Try to parse as float
				if numVal, err := s.parseNumericValue(strVal); err == nil {
					row.Dates[k] = numVal
				}
			}
		}

		parsed.Rows = append(parsed.Rows, row)
	}

	// Keep original order for now

	return parsed, nil
}

// parseNumericValue parses a string value to float64, handling "*" and other non-numeric values
func (s *FundamentalDataService) parseNumericValue(value string) (float64, error) {
	if value == "*" || value == "" || value == "N/A" {
		return 0, fmt.Errorf("non-numeric value")
	}
	return strconv.ParseFloat(value, 64)
}

// extractMetric extracts a specific metric from the parsed statement
func (s *FundamentalDataService) extractMetric(statement *ParsedStatement, breakdown string) map[string]float64 {
	result := make(map[string]float64)
	for _, row := range statement.Rows {
		if row.Breakdown == breakdown {
			return row.Dates
		}
	}
	return result
}

// calculateMargin calculates margin percentage (metric / revenue * 100)
func (s *FundamentalDataService) calculateMargin(metric, revenue map[string]float64) map[string]float64 {
	margin := make(map[string]float64)
	for date, rev := range revenue {
		if metricVal, ok := metric[date]; ok && rev != 0 {
			margin[date] = (metricVal / rev) * 100
		}
	}
	return margin
}

// calculateQoQGrowth calculates Quarter over Quarter growth percentage
func (s *FundamentalDataService) calculateQoQGrowth(values map[string]float64) *float64 {
	if len(values) < 2 {
		return nil
	}

	// Sort dates to get chronological order
	dates := make([]string, 0, len(values))
	for date := range values {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Get the two most recent quarters
	if len(dates) < 2 {
		return nil
	}

	current := values[dates[0]]
	previous := values[dates[1]]

	if previous == 0 {
		return nil
	}

	growth := ((current - previous) / previous) * 100
	return &growth
}

// calculateYoYGrowth calculates Year over Year growth percentage
func (s *FundamentalDataService) calculateYoYGrowth(values map[string]float64) *float64 {
	if len(values) < 2 {
		return nil
	}

	// Sort dates to get chronological order
	dates := make([]string, 0, len(values))
	for date := range values {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	// Get the two most recent periods
	if len(dates) < 2 {
		return nil
	}

	current := values[dates[0]]
	previous := values[dates[1]]

	if previous == 0 {
		return nil
	}

	growth := ((current - previous) / previous) * 100
	return &growth
}

// FilterFundamentalData filters fundamental data based on various criteria
type FundamentalDataFilter struct {
	Symbols        []string `json:"symbols,omitempty"`
	StatementTypes []string `json:"statementTypes,omitempty"`
	Frequencies    []string `json:"frequencies,omitempty"`
}

// FilterFundamentalData applies filters to fundamental data
func (s *FundamentalDataService) FilterFundamentalData(filter FundamentalDataFilter) ([]model.FundamentalData, error) {
	query := s.db.Model(&model.FundamentalData{})

	if len(filter.Symbols) > 0 {
		query = query.Where("symbol IN ?", filter.Symbols)
	}

	if len(filter.StatementTypes) > 0 {
		query = query.Where("statement_type IN ?", filter.StatementTypes)
	}

	if len(filter.Frequencies) > 0 {
		query = query.Where("frequency IN ?", filter.Frequencies)
	}

	var fundamentalData []model.FundamentalData
	result := query.Find(&fundamentalData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to filter fundamental data: %w", result.Error)
	}

	return fundamentalData, nil
}

// GetStocksWithRevenueGrowth filters stocks based on revenue growth criteria
type RevenueGrowthFilter struct {
	MinQoQGrowth  *float64 `json:"minQoQGrowth,omitempty"` // Minimum QoQ growth %
	MaxQoQGrowth  *float64 `json:"maxQoQGrowth,omitempty"` // Maximum QoQ growth %
	MinYoYGrowth  *float64 `json:"minYoYGrowth,omitempty"` // Minimum YoY growth %
	MaxYoYGrowth  *float64 `json:"maxYoYGrowth,omitempty"` // Maximum YoY growth %
	StatementType string   `json:"statementType"`          // "income" or other
	Frequency     string   `json:"frequency"`              // "annual" or "quarterly"
}

// GetStocksWithRevenueGrowth returns stocks that match revenue growth criteria
func (s *FundamentalDataService) GetStocksWithRevenueGrowth(filter RevenueGrowthFilter) ([]FundamentalMetrics, error) {
	// Get all income statements with the specified frequency
	dbFilter := FundamentalDataFilter{
		StatementTypes: []string{filter.StatementType},
		Frequencies:    []string{filter.Frequency},
	}

	fundamentalData, err := s.FilterFundamentalData(dbFilter)
	if err != nil {
		return nil, err
	}

	var results []FundamentalMetrics

	for _, fd := range fundamentalData {
		metrics, err := s.calculateMetrics(&fd)
		if err != nil {
			continue // Skip if metrics calculation fails
		}

		// Check QoQ growth if filter specified
		if filter.MinQoQGrowth != nil || filter.MaxQoQGrowth != nil {
			if metrics.RevenueGrowthQoQ == nil {
				continue // Skip if QoQ not available
			}
			if filter.MinQoQGrowth != nil && *metrics.RevenueGrowthQoQ < *filter.MinQoQGrowth {
				continue
			}
			if filter.MaxQoQGrowth != nil && *metrics.RevenueGrowthQoQ > *filter.MaxQoQGrowth {
				continue
			}
		}

		// Check YoY growth if filter specified
		if filter.MinYoYGrowth != nil || filter.MaxYoYGrowth != nil {
			if metrics.RevenueGrowthYoY == nil {
				continue // Skip if YoY not available
			}
			if filter.MinYoYGrowth != nil && *metrics.RevenueGrowthYoY < *filter.MinYoYGrowth {
				continue
			}
			if filter.MaxYoYGrowth != nil && *metrics.RevenueGrowthYoY > *filter.MaxYoYGrowth {
				continue
			}
		}

		results = append(results, *metrics)
	}

	return results, nil
}

// GetStocksWithEPSRange filters stocks based on EPS criteria
type EPSFilter struct {
	MinEPS        *float64 `json:"minEPS,omitempty"`
	MaxEPS        *float64 `json:"maxEPS,omitempty"`
	Date          string   `json:"date,omitempty"` // Specific date to check (e.g., "2024-09-30"), or latest if empty
	StatementType string   `json:"statementType"`
	Frequency     string   `json:"frequency"`
}

// GetStocksWithEPSRange returns stocks that match EPS criteria
func (s *FundamentalDataService) GetStocksWithEPSRange(filter EPSFilter) ([]FundamentalMetrics, error) {
	dbFilter := FundamentalDataFilter{
		StatementTypes: []string{filter.StatementType},
		Frequencies:    []string{filter.Frequency},
	}

	fundamentalData, err := s.FilterFundamentalData(dbFilter)
	if err != nil {
		return nil, err
	}

	var results []FundamentalMetrics

	for _, fd := range fundamentalData {
		metrics, err := s.calculateMetrics(&fd)
		if err != nil {
			continue
		}

		var epsValue float64
		var found bool

		if filter.Date != "" {
			epsValue, found = metrics.EPS[filter.Date]
		} else {
			// Get latest EPS
			if len(metrics.EPS) > 0 {
				dates := make([]string, 0, len(metrics.EPS))
				for date := range metrics.EPS {
					dates = append(dates, date)
				}
				sort.Strings(dates)
				if len(dates) > 0 {
					epsValue, found = metrics.EPS[dates[0]]
				}
			}
		}

		if !found {
			continue
		}

		if filter.MinEPS != nil && epsValue < *filter.MinEPS {
			continue
		}
		if filter.MaxEPS != nil && epsValue > *filter.MaxEPS {
			continue
		}

		results = append(results, *metrics)
	}

	return results, nil
}

// GetStocksWithMarginRange filters stocks based on margin criteria
type MarginFilter struct {
	MarginType    string   `json:"marginType"` // "gross", "operating", "net"
	MinMargin     *float64 `json:"minMargin,omitempty"`
	MaxMargin     *float64 `json:"maxMargin,omitempty"`
	Date          string   `json:"date,omitempty"` // Specific date or latest if empty
	StatementType string   `json:"statementType"`
	Frequency     string   `json:"frequency"`
}

// GetStocksWithMarginRange returns stocks that match margin criteria
func (s *FundamentalDataService) GetStocksWithMarginRange(filter MarginFilter) ([]FundamentalMetrics, error) {
	dbFilter := FundamentalDataFilter{
		StatementTypes: []string{filter.StatementType},
		Frequencies:    []string{filter.Frequency},
	}

	fundamentalData, err := s.FilterFundamentalData(dbFilter)
	if err != nil {
		return nil, err
	}

	var results []FundamentalMetrics

	for _, fd := range fundamentalData {
		metrics, err := s.calculateMetrics(&fd)
		if err != nil {
			continue
		}

		var marginMap map[string]float64
		switch strings.ToLower(filter.MarginType) {
		case "gross":
			marginMap = metrics.GrossProfitMargin
		case "operating":
			marginMap = metrics.OperatingMargin
		case "net":
			marginMap = metrics.NetMargin
		default:
			continue
		}

		var marginValue float64
		var found bool

		if filter.Date != "" {
			marginValue, found = marginMap[filter.Date]
		} else {
			// Get latest margin
			if len(marginMap) > 0 {
				dates := make([]string, 0, len(marginMap))
				for date := range marginMap {
					dates = append(dates, date)
				}
				sort.Strings(dates)
				if len(dates) > 0 {
					marginValue, found = marginMap[dates[0]]
				}
			}
		}

		if !found {
			continue
		}

		if filter.MinMargin != nil && marginValue < *filter.MinMargin {
			continue
		}
		if filter.MaxMargin != nil && marginValue > *filter.MaxMargin {
			continue
		}

		results = append(results, *metrics)
	}

	return results, nil
}

// SearchFundamentalData searches fundamental data by symbol
func (s *FundamentalDataService) SearchFundamentalData(searchTerm string) ([]model.FundamentalData, error) {
	if searchTerm == "" {
		return []model.FundamentalData{}, nil
	}

	var fundamentalData []model.FundamentalData
	searchPattern := "%" + strings.ToUpper(searchTerm) + "%"
	result := s.db.Where("symbol ILIKE ?", searchPattern).Find(&fundamentalData)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search fundamental data: %w", result.Error)
	}

	return fundamentalData, nil
}
