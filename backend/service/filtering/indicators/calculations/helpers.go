package calculations

import "screener/backend/model"

// SimpleMovingAverage returns SMA over the last N values of the series.
// If there are fewer than N points, it averages available points; if series is empty returns 0.
func SimpleMovingAverage(series []float64, n int) float64 {
	if n <= 0 || len(series) == 0 {
		return 0
	}
	if len(series) < n {
		// average over available
		var sum float64
		for _, v := range series {
			sum += v
		}
		return sum / float64(len(series))
	}
	var sum float64
	for i := len(series) - n; i < len(series); i++ {
		sum += series[i]
	}
	return sum / float64(n)
}

// AverageTrueRange computes ATR over the last N bars using Wilder's SMA of True Range.
// If fewer than N bars, it averages available TRs.
func AverageTrueRange(rows []model.Historical, n int) float64 {
	if n <= 0 || len(rows) == 0 {
		return 0
	}
	// Build TR series
	trs := make([]float64, 0, len(rows))
	for i := range rows {
		cur := rows[i]
		var prevClose float64
		if i == 0 {
			prevClose = cur.Close
		} else {
			prevClose = rows[i-1].Close
		}
		hl := cur.High - cur.Low
		hc := abs(cur.High - prevClose)
		lc := abs(cur.Low - prevClose)
		tr := hl
		if hc > tr {
			tr = hc
		}
		if lc > tr {
			tr = lc
		}
		trs = append(trs, tr)
	}
	return SimpleMovingAverage(trs, n)
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

