package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"screener/backend/database"
	"screener/backend/model"

	"gorm.io/gorm"
)

// FetcherService fetches external market data and persists it.
type FetcherService struct {
	db          *gorm.DB
	httpClient  *http.Client
	baseURL     string
	histService *HistoricalService
}

// NewFetcherService constructs a FetcherService with sensible defaults.
func NewFetcherService() *FetcherService {
	base := os.Getenv("FINANCE_QUERY_BASE_URL")
	if base == "" {
		base = "https://finance-query.onrender.com"
	}

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
		baseURL:     base,
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
	url := fmt.Sprintf("%s/v1/historical?symbol=%s&range=%s&interval=%s&epoch=true", s.baseURL, symbol, rangeParam, interval)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// simple retry: up to 3 attempts with backoff
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		resp, err = s.httpClient.Do(req)
		if err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Duration(300*(attempt+1)) * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

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
