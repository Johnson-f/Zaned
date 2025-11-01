<!-- 8d96df94-010a-48cc-bf76-a9d5d3ebdc9e a11a3b0f-a079-42d2-a806-261818c2bcae -->
# Volume Metrics Calculation Implementation

## Overview

Add a new standalone function `GetStocksVolumeMetrics()` to `backend/service/historical.go` that:

1. Scans all stocks in the historical table
2. Fetches historical data with `range="1d"` and `interval="30m"` for each stock
3. Aggregates 30-minute intervals into daily volumes
4. Calculates three metrics: highest volume in a year (365 days), quarter (90 days), and ever

## Implementation Details

### 1. Create Volume Metrics Result Type

Add to `backend/service/historical.go`:

```go
type VolumeMetricsResult struct {
    Symbol                string `json:"symbol"`
    HighestVolumeInYear   int64  `json:"highest_volume_in_year"`
    HighestVolumeInQuarter int64  `json:"highest_volume_in_quarter"`
    HighestVolumeEver     int64  `json:"highest_volume_ever"`
}
```

### 2. Implement GetStocksVolumeMetrics Function

- Get all unique symbols from historical table (distinct query)
- For each symbol:
  - Query historical records where `range="1d"` AND `interval="30m"`
  - Group records by day (convert epoch to date)
  - Sum volumes within each day to get daily volume
  - Calculate:
    - Highest daily volume in last 365 days from current date
    - Highest daily volume in last 90 days from current date  
    - Highest daily volume across all time
- Return slice of `VolumeMetricsResult`

### 3. Helper Functions

- `groupByDay()`: Groups historical records by date (epoch -> date conversion)
- `sumDailyVolumes()`: Sums volumes for records within the same day
- `calculateMetrics()`: Computes the three volume metrics

### 4. Error Handling

- Handle cases where stock has no data
- Handle database query errors
- Return empty metrics (0) if no data available

## Files to Modify

- `backend/service/historical.go`: Add result type and function implementation

### To-dos

- [ ] Add VolumeMetricsResult struct to historical.go service file
- [ ] Implement GetStocksVolumeMetrics() function that gets all unique symbols from database
- [ ] For each symbol, fetch historical records where range='1d' AND interval='30m'
- [ ] Create helper function to group records by day and sum daily volumes
- [ ] Calculate highest volume in year (365 days lookback from current date)
- [ ] Calculate highest volume in quarter (90 days lookback from current date)
- [ ] Calculate highest volume ever (all time)
- [ ] Return VolumeMetricsResult slice with all metrics for each stock