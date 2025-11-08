package indicators

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
