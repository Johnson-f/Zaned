package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"screener/backend/database"
	"screener/backend/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FetcherService fetches external market data and persists it.
type FetcherService struct {
	db          *gorm.DB
	httpClient  *http.Client
	baseURL     string
	histService *HistoricalService
}

// getBaseURLs returns primary and fallback base URLs for failover
// Both endpoints are hardcoded as defaults and can be overridden via environment variables
func getBaseURLs() (string, string) {
	// Hardcoded default endpoints
	defaultPrimary := "https://finance-query.onrender.com"
	defaultFallback := "https://finance-query-uzbi.onrender.com"

	// Get from environment variables if set, otherwise use hardcoded defaults
	primary := os.Getenv("FINANCE_QUERY_PRIMARY_URL")
	if primary == "" {
		primary = defaultPrimary
	}

	fallback := os.Getenv("FINANCE_QUERY_FALLBACK_URL")
	if fallback == "" {
		fallback = defaultFallback
	}

	return primary, fallback
}

// fetchWithFailover attempts to fetch from primary URL, falls back to secondary URL on immediate failure
// Returns the response body and which endpoint was used
// Fails over immediately if: network error, timeout, or HTTP error status (4xx, 5xx)
func (s *FetcherService) fetchWithFailover(ctx context.Context, primaryURL, fallbackURL string, jobID string, batchNum, totalBatches int) (*http.Response, string, error) {
	// Try primary endpoint first
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, primaryURL, nil)
	if err != nil {
		return nil, "", err
	}

	// Try primary endpoint (single attempt)
	resp, err := s.httpClient.Do(req)
	primarySuccess := err == nil && resp != nil && resp.StatusCode >= 200 && resp.StatusCode < 300

	if primarySuccess {
		if jobID != "" {
			fmt.Printf("[%s] Batch %d/%d: Successfully fetched from primary endpoint\n", jobID, batchNum, totalBatches)
		}
		return resp, primaryURL, nil
	}

	// If primary failed, capture error message and close response
	var primaryErrorMsg string
	var primaryStatusCode int
	if err != nil {
		primaryErrorMsg = err.Error()
	} else if resp != nil {
		primaryStatusCode = resp.StatusCode
		primaryErrorMsg = fmt.Sprintf("HTTP status %d", resp.StatusCode)
		resp.Body.Close()
	} else {
		primaryErrorMsg = "unknown error"
	}

	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Primary endpoint failed (%s), trying fallback: %s\n", jobID, batchNum, totalBatches, primaryErrorMsg, fallbackURL)
	}

	// Try fallback endpoint
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, fallbackURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create fallback request: %w", err)
	}

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("both endpoints failed. Primary: %s, Fallback: %w", primaryErrorMsg, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fallbackStatusCode := resp.StatusCode
		resp.Body.Close()
		return nil, "", fmt.Errorf("both endpoints failed. Primary: %s (status %d), Fallback: HTTP status %d", primaryErrorMsg, primaryStatusCode, fallbackStatusCode)
	}

	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Successfully fetched from fallback endpoint\n", jobID, batchNum, totalBatches)
	}
	return resp, fallbackURL, nil
}

// NewFetcherService constructs a FetcherService with sensible defaults.
func NewFetcherService() *FetcherService {
	// Get primary URL (fallback is handled in getBaseURLs)
	primaryBase, _ := getBaseURLs()

	timeoutStr := os.Getenv("FETCHER_HTTP_TIMEOUT_SECONDS")
	timeout := 15 * time.Second
	if timeoutStr != "" {
		if v, err := strconv.Atoi(timeoutStr); err == nil && v > 0 {
			timeout = time.Duration(v) * time.Second
		}
	}

	return &FetcherService{
		db:          database.GetDB(),
		httpClient:  &http.Client{Timeout: timeout},
		baseURL:     primaryBase,
		histService: NewHistoricalService(),
	}
}

// RunIngestion fetches and stores data for all symbols concurrently. Suitable for cron trigger.
func (s *FetcherService) RunIngestion(ctx context.Context, concurrency int) (string, error) {
	if concurrency <= 0 {
		concurrency = 8
	}

	// Load all symbols from screener table
	var symbols []string
	if err := s.db.Model(&model.Screener{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return "", fmt.Errorf("failed to load screener symbols: %w", err)
	}
	if len(symbols) == 0 {
		return fmt.Sprintf("job-%d", time.Now().UnixNano()), nil
	}

	// Worker pool
	jobs := make(chan string)
	wg := sync.WaitGroup{}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for symbol := range jobs {
				_ = s.processSymbol(ctx, symbol)
			}
		}()
	}

	for _, sym := range symbols {
		select {
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return "", ctx.Err()
		case jobs <- sym:
		}
	}
	close(jobs)
	wg.Wait()

	return fmt.Sprintf("job-%d", time.Now().UnixNano()), nil
}

// processSymbol fetches 1d/1m, aggregates to daily and updates Screener, then fetches 1d/30m into Historical.
func (s *FetcherService) processSymbol(ctx context.Context, symbol string) error {
	// 0) Daily backfill for last 10 years (1d interval)
	_ = s.fetchAndUpsertDaily10y(ctx, symbol)
	// 1) Screener update from 1d/1m aggregated to daily
	bars1m, err := s.fetchBars(ctx, symbol, "1d", "1m")
	if err == nil && len(bars1m) > 0 {
		daily := aggregateDailyFromIntraday(bars1m)
		if daily != nil {
			// Upsert/update Screener row
			// Only update price fields; keep other fields intact (e.g., logo)
			updates := map[string]interface{}{
				"open":   daily.Open,
				"high":   daily.High,
				"low":    daily.Low,
				"close":  daily.Close,
				"volume": daily.Volume,
			}
			_ = s.db.Model(&model.Screener{}).Where("symbol = ?", symbol).Updates(updates).Error
		}
	}

	// 2) Historical: store 1d/30m into historical table
	bars30m, err := s.fetchBars(ctx, symbol, "1d", "30m")
	if err != nil || len(bars30m) == 0 {
		return err
	}

	// Prepare upsert batch
	batch := make([]model.Historical, 0, len(bars30m))
	for _, b := range bars30m {
		batch = append(batch, model.Historical{
			Symbol:   symbol,
			Epoch:    b.Epoch,
			Range:    "1d",
			Interval: "30m",
			Open:     b.Open,
			High:     b.High,
			Low:      b.Low,
			Close:    b.Close,
			AdjClose: b.AdjClose,
			Volume:   b.Volume,
		})
	}
	return s.histService.UpsertHistoricalBatch(batch)
}

// fetchAndUpsertDaily10y fetches last 10 years of daily bars and upserts them
func (s *FetcherService) fetchAndUpsertDaily10y(ctx context.Context, symbol string) error {
	bars, err := s.fetchBars(ctx, symbol, "10y", "1d")
	if err != nil || len(bars) == 0 {
		return err
	}

	batch := make([]model.Historical, 0, len(bars))
	for _, b := range bars {
		batch = append(batch, model.Historical{
			Symbol:   symbol,
			Epoch:    b.Epoch,
			Range:    "10y",
			Interval: "1d",
			Open:     b.Open,
			High:     b.High,
			Low:      b.Low,
			Close:    b.Close,
			AdjClose: b.AdjClose,
			Volume:   b.Volume,
		})
	}

	return s.histService.UpsertHistoricalBatch(batch)
}

// externalBar represents a single bar returned by the external API after normalization
type externalBar struct {
	Epoch    int64
	Open     float64
	High     float64
	Low      float64
	Close    float64
	AdjClose *float64
	Volume   int64
}

// fetchBars calls the external API and parses the map epoch->bar payload.
func (s *FetcherService) fetchBars(ctx context.Context, symbol, rangeParam, interval string) ([]externalBar, error) {
	if symbol == "" {
		return nil, errors.New("symbol is required")
	}

	// Get primary and fallback URLs
	primaryBase, fallbackBase := getBaseURLs()
	primaryURL := fmt.Sprintf("%s/v1/historical?symbol=%s&range=%s&interval=%s&epoch=true", primaryBase, symbol, rangeParam, interval)
	fallbackURL := fmt.Sprintf("%s/v1/historical?symbol=%s&range=%s&interval=%s&epoch=true", fallbackBase, symbol, rangeParam, interval)

	// Try with failover
	resp, _, err := s.fetchWithFailover(ctx, primaryURL, fallbackURL, "", 0, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The response is a JSON object keyed by epoch strings
	var raw map[string]struct {
		Open     float64  `json:"open"`
		High     float64  `json:"high"`
		Low      float64  `json:"low"`
		Close    float64  `json:"close"`
		AdjClose *float64 `json:"adjClose"`
		Volume   int64    `json:"volume"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	out := make([]externalBar, 0, len(raw))
	for k, v := range raw {
		epoch, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			continue
		}
		out = append(out, externalBar{
			Epoch:    epoch,
			Open:     v.Open,
			High:     v.High,
			Low:      v.Low,
			Close:    v.Close,
			AdjClose: v.AdjClose,
			Volume:   v.Volume,
		})
	}
	// sort ascending by epoch for consistent aggregation
	sort.Slice(out, func(i, j int) bool { return out[i].Epoch < out[j].Epoch })
	return out, nil
}

// dailyOHLCV represents a single aggregated day
type dailyOHLCV struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int64
}

// aggregateDailyFromIntraday aggregates intraday bars (e.g., 1m) into daily OHLCV.
func aggregateDailyFromIntraday(bars []externalBar) *dailyOHLCV {
	if len(bars) == 0 {
		return nil
	}
	// assume all bars are from the same day since range=1d
	open := bars[0].Open
	high := bars[0].High
	low := bars[0].Low
	closeP := bars[len(bars)-1].Close
	var volume int64 = 0
	for _, b := range bars {
		if b.High > high {
			high = b.High
		}
		if b.Low < low {
			low = b.Low
		}
		volume += b.Volume
	}
	return &dailyOHLCV{Open: open, High: high, Low: low, Close: closeP, Volume: volume}
}

// simpleQuote represents a stock quote from the simple-quotes API
type simpleQuote struct {
	Symbol          string `json:"symbol"`
	Name            string `json:"name"`
	Price           string `json:"price"`
	AfterHoursPrice string `json:"afterHoursPrice"`
	Change          string `json:"change"`
	PercentChange   string `json:"percentChange"`
	Logo            string `json:"logo"`
}

// RunWatchlistPriceUpdate fetches price data for all unique stocks in watchlists and updates them.
// It avoids duplicate fetches by processing unique symbols only.
func (s *FetcherService) RunWatchlistPriceUpdate(ctx context.Context) (string, error) {
	// Get all unique symbols from watchlist_items (where symbol is not empty)
	var symbols []string
	if err := s.db.Model(&model.WatchlistItem{}).
		Where("symbol IS NOT NULL AND symbol != ''").
		Distinct("symbol").
		Pluck("symbol", &symbols).Error; err != nil {
		return "", fmt.Errorf("failed to load watchlist symbols: %w", err)
	}

	if len(symbols) == 0 {
		return fmt.Sprintf("watchlist-price-update-%d", time.Now().UnixNano()), nil
	}

	// Fetch quotes for all symbols in batches (API may have limits, so we'll do in chunks of 50)
	batchSize := 50
	totalUpdated := 0

	for i := 0; i < len(symbols); i += batchSize {
		end := i + batchSize
		if end > len(symbols) {
			end = len(symbols)
		}
		batch := symbols[i:end]

		quotes, err := s.fetchSimpleQuotes(ctx, batch)
		if err != nil {
			// Log error but continue with next batch
			continue
		}

		// Update watchlist items with the fetched data
		updated, err := s.updateWatchlistItemsFromQuotes(quotes)
		if err != nil {
			// Log error but continue
			continue
		}
		totalUpdated += updated
	}

	return fmt.Sprintf("watchlist-price-update-%d", time.Now().UnixNano()), nil
}

// fetchSimpleQuotes calls the simple-quotes API for a batch of symbols
func (s *FetcherService) fetchSimpleQuotes(ctx context.Context, symbols []string) ([]simpleQuote, error) {
	return s.fetchSimpleQuotesWithLogging(ctx, symbols, "", 0, 0)
}

// fetchSimpleQuotesWithLogging calls the simple-quotes API with detailed logging
func (s *FetcherService) fetchSimpleQuotesWithLogging(ctx context.Context, symbols []string, jobID string, batchNum, totalBatches int) ([]simpleQuote, error) {
	if len(symbols) == 0 {
		return nil, errors.New("symbols cannot be empty")
	}

	// Build URL with comma-separated symbols (URL encoded)
	symbolsParam := strings.Join(symbols, ", ")
	encodedSymbols := url.QueryEscape(symbolsParam)

	// Get primary and fallback URLs
	primaryBase, fallbackBase := getBaseURLs()
	primaryURL := fmt.Sprintf("%s/v1/simple-quotes?symbols=%s", primaryBase, encodedSymbols)
	fallbackURL := fmt.Sprintf("%s/v1/simple-quotes?symbols=%s", fallbackBase, encodedSymbols)

	// Only log detailed info if jobID is provided (for market aggregation)
	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Calling API: %s\n", jobID, batchNum, totalBatches, primaryURL)
	}

	startTime := time.Now()
	resp, usedURL, err := s.fetchWithFailover(ctx, primaryURL, fallbackURL, jobID, batchNum, totalBatches)
	requestDuration := time.Since(startTime)

	if err != nil {
		if jobID != "" {
			fmt.Printf("[%s] Batch %d/%d: ERROR: %v (duration: %v)\n", jobID, batchNum, totalBatches, err, requestDuration)
		}
		return nil, err
	}
	defer resp.Body.Close()

	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Successfully fetched from %s in %v\n", jobID, batchNum, totalBatches, usedURL, requestDuration)
	}

	// Read response body
	var quotes []simpleQuote
	decoder := json.NewDecoder(resp.Body)
	if err := decoder.Decode(&quotes); err != nil {
		if jobID != "" {
			fmt.Printf("[%s] Batch %d/%d: ERROR decoding JSON response: %v\n", jobID, batchNum, totalBatches, err)
		}
		return nil, err
	}

	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Successfully fetched %d quotes in %v\n", jobID, batchNum, totalBatches, len(quotes), requestDuration)
	}

	// Log which symbols were successfully retrieved
	if len(quotes) < len(symbols) && jobID != "" {
		retrievedSymbols := make(map[string]bool)
		for _, q := range quotes {
			retrievedSymbols[q.Symbol] = true
		}
		missingSymbols := make([]string, 0)
		for _, sym := range symbols {
			if !retrievedSymbols[sym] {
				missingSymbols = append(missingSymbols, sym)
			}
		}
		if len(missingSymbols) > 0 {
			fmt.Printf("[%s] Batch %d/%d: WARNING: %d symbols not found in response: %v\n", jobID, batchNum, totalBatches, len(missingSymbols), missingSymbols)
		}
	}

	return quotes, nil
}

// updateWatchlistItemsFromQuotes updates watchlist items with quote data, matching by symbol
func (s *FetcherService) updateWatchlistItemsFromQuotes(quotes []simpleQuote) (int, error) {
	updated := 0

	for _, quote := range quotes {
		if quote.Symbol == "" {
			continue
		}

		// Parse price strings to float64
		var price *float64
		if quote.Price != "" {
			if p, err := strconv.ParseFloat(quote.Price, 64); err == nil {
				price = &p
			}
		}

		var afterHoursPrice *float64
		if quote.AfterHoursPrice != "" {
			if p, err := strconv.ParseFloat(quote.AfterHoursPrice, 64); err == nil {
				afterHoursPrice = &p
			}
		}

		var change *float64
		if quote.Change != "" {
			if c, err := strconv.ParseFloat(quote.Change, 64); err == nil {
				change = &c
			}
		}

		// Update all watchlist items with this symbol (or name if symbol not set)
		updates := map[string]interface{}{
			"symbol":            quote.Symbol,
			"price":             price,
			"after_hours_price": afterHoursPrice,
			"change":            change,
			"percent_change":    quote.PercentChange,
			"logo":              quote.Logo,
		}

		// Also update name if it's different (in case it changed)
		if quote.Name != "" {
			updates["name"] = quote.Name
		}

		// Try to update by symbol first, then by name as fallback
		result := s.db.Model(&model.WatchlistItem{}).
			Where("symbol = ? OR (symbol IS NULL OR symbol = '') AND name = ?", quote.Symbol, quote.Name).
			Updates(updates)

		if result.Error != nil {
			continue
		}
		updated += int(result.RowsAffected)
	}

	return updated, nil
}

// detailedQuote represents a detailed quote from the quotes API
type detailedQuote struct {
	Symbol           string `json:"symbol"`
	Name             string `json:"name"`
	Price            string `json:"price"`
	AfterHoursPrice  string `json:"afterHoursPrice"`
	Change           string `json:"change"`
	PercentChange    string `json:"percentChange"`
	Open             string `json:"open"`
	High             string `json:"high"`
	Low              string `json:"low"`
	YearHigh         string `json:"yearHigh"`
	YearLow          string `json:"yearLow"`
	Volume           int64  `json:"volume"`
	AvgVolume        int64  `json:"avgVolume"`
	MarketCap        string `json:"marketCap"`
	Beta             string `json:"beta"`
	PE               string `json:"pe"`
	EarningsDate     string `json:"earningsDate"`
	Sector           string `json:"sector"`
	Industry         string `json:"industry"`
	About            string `json:"about"`
	Employees        string `json:"employees"`
	FiveDaysReturn   string `json:"fiveDaysReturn"`
	OneMonthReturn   string `json:"oneMonthReturn"`
	ThreeMonthReturn string `json:"threeMonthReturn"`
	SixMonthReturn   string `json:"sixMonthReturn"`
	YtdReturn        string `json:"ytdReturn"`
	YearReturn       string `json:"yearReturn"`
	ThreeYearReturn  string `json:"threeYearReturn"`
	FiveYearReturn   string `json:"fiveYearReturn"`
	TenYearReturn    string `json:"tenYearReturn"`
	MaxReturn        string `json:"maxReturn"`
	Logo             string `json:"logo"`
}

// RunCompanyInfoIngestion fetches company info for all symbols from screener table and upserts them.
// It avoids duplicate data by using ON CONFLICT (upsert) based on symbol primary key.
func (s *FetcherService) RunCompanyInfoIngestion(ctx context.Context) (string, error) {
	// Get all unique symbols from screener table
	var symbols []string
	if err := s.db.Model(&model.Screener{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return "", fmt.Errorf("failed to load screener symbols: %w", err)
	}

	if len(symbols) == 0 {
		return fmt.Sprintf("company-info-ingestion-%d", time.Now().UnixNano()), nil
	}

	// Fetch company info for all symbols in batches (API may have limits, so we'll do in chunks of 50)
	batchSize := 50
	totalUpserted := 0

	for i := 0; i < len(symbols); i += batchSize {
		end := i + batchSize
		if end > len(symbols) {
			end = len(symbols)
		}
		batch := symbols[i:end]

		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		quotes, err := s.fetchDetailedQuotes(ctx, batch, "", 0, 0)
		if err != nil {
			// Log error but continue with next batch
			continue
		}

		// Upsert company info with the fetched data
		upserted, err := s.upsertCompanyInfoFromQuotes(quotes)
		if err != nil {
			// Log error but continue
			continue
		}
		totalUpserted += upserted
	}

	return fmt.Sprintf("company-info-ingestion-%d", time.Now().UnixNano()), nil
}

// RunMarketAggregation fetches quotes for all stocks from screener table and aggregates them
// for market statistics (up/down/unchanged counts). Suitable for cron trigger every 5 minutes.
func (s *FetcherService) RunMarketAggregation(ctx context.Context) (string, error) {
	jobID := fmt.Sprintf("market-aggregation-%d", time.Now().UnixNano())
	startTime := time.Now()

	fmt.Printf("[%s] Starting market aggregation...\n", jobID)

	// Get all unique symbols from screener table
	var symbols []string
	if err := s.db.Model(&model.Screener{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		fmt.Printf("[%s] ERROR: Failed to load screener symbols: %v\n", jobID, err)
		return "", fmt.Errorf("failed to load screener symbols: %w", err)
	}

	totalSymbols := len(symbols)
	fmt.Printf("[%s] Loaded %d symbols from screener table\n", jobID, totalSymbols)

	if totalSymbols == 0 {
		fmt.Printf("[%s] No symbols found, skipping aggregation\n", jobID)
		return jobID, nil
	}

	// Initialize market statistics service
	statsService := NewMarketStatisticsService()

	// Fetch quotes for all symbols in batches (API may have limits, so we'll do in chunks of 50)
	batchSize := 50
	totalBatches := (totalSymbols + batchSize - 1) / batchSize
	successfulBatches := 0
	failedBatches := 0
	totalQuotesProcessed := 0

	fmt.Printf("[%s] Processing %d batches of %d symbols each\n", jobID, totalBatches, batchSize)

	for i := 0; i < totalSymbols; i += batchSize {
		batchNum := (i / batchSize) + 1
		end := i + batchSize
		if end > totalSymbols {
			end = totalSymbols
		}
		batch := symbols[i:end]

		// Check for context cancellation
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] Cancelled: context deadline exceeded at batch %d/%d\n", jobID, batchNum, totalBatches)
			return "", ctx.Err()
		default:
		}

		fmt.Printf("[%s] Processing batch %d/%d (%d symbols): %v\n", jobID, batchNum, totalBatches, len(batch), batch)

		quotes, err := s.fetchSimpleQuotesWithLogging(ctx, batch, jobID, batchNum, totalBatches)
		if err != nil {
			failedBatches++
			fmt.Printf("[%s] ERROR: Failed to fetch quotes for batch %d/%d: %v\n", jobID, batchNum, totalBatches, err)
			continue
		}

		if len(quotes) == 0 {
			fmt.Printf("[%s] WARNING: Batch %d/%d returned 0 quotes (all symbols may be invalid)\n", jobID, batchNum, totalBatches)
			failedBatches++
			continue
		}

		// Aggregate the quotes
		if err := statsService.AggregateQuotes(ctx, quotes); err != nil {
			failedBatches++
			fmt.Printf("[%s] ERROR: Failed to aggregate quotes for batch %d/%d: %v\n", jobID, batchNum, totalBatches, err)
			continue
		}

		successfulBatches++
		totalQuotesProcessed += len(quotes)
		fmt.Printf("[%s] Batch %d/%d completed: %d quotes processed (expected %d symbols)\n", jobID, batchNum, totalBatches, len(quotes), len(batch))
	}

	// Get final stats
	finalStats, err := statsService.GetCurrentDayStats()
	if err != nil {
		fmt.Printf("[%s] WARNING: Failed to get final stats: %v\n", jobID, err)
	} else {
		fmt.Printf("[%s] Final statistics - Up: %d, Down: %d, Unchanged: %d, Total: %d\n",
			jobID, finalStats["up"], finalStats["down"], finalStats["unchanged"], finalStats["total"])
	}

	duration := time.Since(startTime)
	fmt.Printf("[%s] Aggregation completed in %v - Successful batches: %d/%d, Failed: %d, Quotes processed: %d\n",
		jobID, duration, successfulBatches, totalBatches, failedBatches, totalQuotesProcessed)

	return jobID, nil
}

// fetchDetailedQuotes calls the quotes API for a batch of symbols
func (s *FetcherService) fetchDetailedQuotes(ctx context.Context, symbols []string, jobID string, batchNum, totalBatches int) ([]detailedQuote, error) {
	if len(symbols) == 0 {
		return nil, errors.New("symbols cannot be empty")
	}

	// Build URL with comma-separated symbols (URL encoded)
	symbolsParam := strings.Join(symbols, ", ")
	encodedSymbols := url.QueryEscape(symbolsParam)

	// Get primary and fallback URLs for quotes API
	primaryBase, fallbackBase := getBaseURLs()
	primaryURL := fmt.Sprintf("%s/v1/quotes?symbols=%s", primaryBase, encodedSymbols)
	fallbackURL := fmt.Sprintf("%s/v1/quotes?symbols=%s", fallbackBase, encodedSymbols)

	// Only log detailed info if jobID is provided (for market aggregation)
	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Calling API: %s\n", jobID, batchNum, totalBatches, primaryURL)
	}

	startTime := time.Now()
	resp, usedURL, err := s.fetchWithFailover(ctx, primaryURL, fallbackURL, jobID, batchNum, totalBatches)
	requestDuration := time.Since(startTime)

	if err != nil {
		if jobID != "" {
			fmt.Printf("[%s] Batch %d/%d: ERROR: %v (duration: %v)\n", jobID, batchNum, totalBatches, err, requestDuration)
		}
		return nil, err
	}
	defer resp.Body.Close()

	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Successfully fetched from %s in %v\n", jobID, batchNum, totalBatches, usedURL, requestDuration)
	}

	// Read response body
	bodyReader := resp.Body
	var quotes []detailedQuote
	decoder := json.NewDecoder(bodyReader)
	if err := decoder.Decode(&quotes); err != nil {
		if jobID != "" {
			fmt.Printf("[%s] Batch %d/%d: ERROR decoding JSON response: %v\n", jobID, batchNum, totalBatches, err)
		}
		return nil, err
	}

	if jobID != "" {
		fmt.Printf("[%s] Batch %d/%d: Successfully fetched %d quotes in %v\n", jobID, batchNum, totalBatches, len(quotes), requestDuration)
	}

	// Log which symbols were successfully retrieved
	if len(quotes) < len(symbols) && jobID != "" {
		retrievedSymbols := make(map[string]bool)
		for _, q := range quotes {
			retrievedSymbols[q.Symbol] = true
		}
		missingSymbols := make([]string, 0)
		for _, sym := range symbols {
			if !retrievedSymbols[sym] {
				missingSymbols = append(missingSymbols, sym)
			}
		}
		if len(missingSymbols) > 0 {
			fmt.Printf("[%s] Batch %d/%d: WARNING: %d symbols not found in response: %v\n", jobID, batchNum, totalBatches, len(missingSymbols), missingSymbols)
		}
	}

	return quotes, nil
}

// upsertCompanyInfoFromQuotes upserts company info records in batch, avoiding duplicates by symbol (primary key)
func (s *FetcherService) upsertCompanyInfoFromQuotes(quotes []detailedQuote) (int, error) {
	if len(quotes) == 0 {
		return 0, nil
	}

	// Convert quotes to CompanyInfo models
	companyInfoList := make([]model.CompanyInfo, 0, len(quotes))
	for _, quote := range quotes {
		if quote.Symbol == "" {
			continue
		}

		companyInfo := model.CompanyInfo{
			Symbol:           quote.Symbol,
			Name:             quote.Name,
			Price:            quote.Price,
			AfterHoursPrice:  quote.AfterHoursPrice,
			Change:           quote.Change,
			PercentChange:    quote.PercentChange,
			Open:             quote.Open,
			High:             quote.High,
			Low:              quote.Low,
			YearHigh:         quote.YearHigh,
			YearLow:          quote.YearLow,
			Volume:           quote.Volume,
			AvgVolume:        quote.AvgVolume,
			MarketCap:        quote.MarketCap,
			Beta:             quote.Beta,
			PE:               quote.PE,
			EarningsDate:     quote.EarningsDate,
			Sector:           quote.Sector,
			Industry:         quote.Industry,
			About:            quote.About,
			Employees:        quote.Employees,
			FiveDaysReturn:   quote.FiveDaysReturn,
			OneMonthReturn:   quote.OneMonthReturn,
			ThreeMonthReturn: quote.ThreeMonthReturn,
			SixMonthReturn:   quote.SixMonthReturn,
			YtdReturn:        quote.YtdReturn,
			YearReturn:       quote.YearReturn,
			ThreeYearReturn:  quote.ThreeYearReturn,
			FiveYearReturn:   quote.FiveYearReturn,
			TenYearReturn:    quote.TenYearReturn,
			MaxReturn:        quote.MaxReturn,
			Logo:             quote.Logo,
		}
		companyInfoList = append(companyInfoList, companyInfo)
	}

	if len(companyInfoList) == 0 {
		return 0, nil
	}

	// Batch upsert using GORM's Clauses with OnConflict
	// Since symbol is the primary key, this will update existing records or insert new ones
	result := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "symbol"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "price", "after_hours_price", "change", "percent_change",
			"open", "high", "low", "year_high", "year_low",
			"volume", "avg_volume", "market_cap", "beta", "pe",
			"earnings_date", "sector", "industry", "about", "employees",
			"five_days_return", "one_month_return", "three_month_return",
			"six_month_return", "ytd_return", "year_return",
			"three_year_return", "five_year_return", "ten_year_return",
			"max_return", "logo", "updated_at",
		}),
	}).CreateInBatches(companyInfoList, 100)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to upsert company info: %w", result.Error)
	}

	return len(companyInfoList), nil
}

// financialsResponse represents the response from the financials API
type financialsResponse struct {
	Symbol        string                 `json:"symbol"`
	StatementType string                 `json:"statement_type"`
	Frequency     string                 `json:"frequency"`
	Statement     map[string]interface{} `json:"statement"`
}

// RunFundamentalDataIngestion fetches fundamental data (income, balance, cashflow) for all symbols from screener table.
// It fetches all three statement types and both annual and quarterly frequencies.
// It avoids duplicate data by using ON CONFLICT (upsert) based on unique constraint (symbol, statement_type, frequency).
func (s *FetcherService) RunFundamentalDataIngestion(ctx context.Context) (string, error) {
	// Get all unique symbols from screener table
	var symbols []string
	if err := s.db.Model(&model.Screener{}).Distinct("symbol").Pluck("symbol", &symbols).Error; err != nil {
		return "", fmt.Errorf("failed to load screener symbols: %w", err)
	}

	if len(symbols) == 0 {
		return fmt.Sprintf("fundamental-data-ingestion-%d", time.Now().UnixNano()), nil
	}

	// Statement types to fetch
	statementTypes := []string{"income", "balance", "cashflow"}
	// Frequencies to fetch
	frequencies := []string{"annual", "quarterly"}

	totalUpserted := 0

	// Process each symbol
	for _, symbol := range symbols {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		// Fetch all statement types and frequencies for this symbol
		for _, statementType := range statementTypes {
			for _, frequency := range frequencies {
				// Fetch financial data
				financialData, err := s.fetchFinancials(ctx, symbol, statementType, frequency)
				if err != nil {
					// Log error but continue with next combination
					continue
				}

				// Upsert into database
				if err := s.upsertFundamentalData(financialData); err != nil {
					// Log error but continue
					continue
				}
				totalUpserted++
			}
		}
	}

	return fmt.Sprintf("fundamental-data-ingestion-%d", time.Now().UnixNano()), nil
}

// fetchFinancials calls the financials API for a specific symbol, statement type, and frequency
func (s *FetcherService) fetchFinancials(ctx context.Context, symbol, statementType, frequency string) (*financialsResponse, error) {
	if symbol == "" || statementType == "" || frequency == "" {
		return nil, errors.New("symbol, statement type, and frequency are required")
	}

	// Get primary and fallback URLs for financials API
	primaryBase, fallbackBase := getBaseURLs()
	primaryURL := fmt.Sprintf("%s/v1/financials/%s?statement=%s&frequency=%s", primaryBase, symbol, statementType, frequency)
	fallbackURL := fmt.Sprintf("%s/v1/financials/%s?statement=%s&frequency=%s", fallbackBase, symbol, statementType, frequency)

	// Try with failover
	resp, _, err := s.fetchWithFailover(ctx, primaryURL, fallbackURL, "", 0, 0)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var financialData financialsResponse
	if err := json.NewDecoder(resp.Body).Decode(&financialData); err != nil {
		return nil, err
	}

	// Validate response
	if financialData.Symbol == "" || financialData.StatementType == "" || financialData.Frequency == "" {
		return nil, errors.New("invalid response: missing required fields")
	}

	return &financialData, nil
}

// upsertFundamentalData upserts a fundamental data record, avoiding duplicates by unique constraint
func (s *FetcherService) upsertFundamentalData(financialData *financialsResponse) error {
	if financialData == nil {
		return errors.New("financial data is nil")
	}

	// Convert statement map to JSON string
	statementJSON, err := json.Marshal(financialData.Statement)
	if err != nil {
		return fmt.Errorf("failed to marshal statement: %w", err)
	}

	// Create or update fundamental data record
	fundamentalDataRecord := model.FundamentalData{
		Symbol:        financialData.Symbol,
		StatementType: financialData.StatementType,
		Frequency:     financialData.Frequency,
		Statement:     string(statementJSON),
	}

	// Use upsert with ON CONFLICT based on unique constraint (symbol, statement_type, frequency)
	result := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "symbol"},
			{Name: "statement_type"},
			{Name: "frequency"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"statement", "updated_at",
		}),
	}).Create(&fundamentalDataRecord)

	if result.Error != nil {
		return fmt.Errorf("failed to upsert fundamental data: %w", result.Error)
	}

	return nil
}
