package domain

import (
	"math"
	"sort"
	"time"
)

type ForecastSample struct {
	At                                          time.Time
	EventRate, PeakMagnitude, InjectionPressure float64
}
type RiskForecast struct {
	GeneratedAt, HorizonEnd                       time.Time
	Samples                                       int
	ExpectedRate, ExpectedPeak, Confidence, Score float64
	Band                                          RiskBand
}

func BuildForecast(samples []ForecastSample, generatedAt time.Time, horizon time.Duration) (RiskForecast, error) {
	if horizon <= 0 {
		return RiskForecast{}, FieldError{"horizon", "must be positive"}
	}
	if len(samples) < 3 {
		return RiskForecast{}, ConflictError{"forecast", "at least three samples are required"}
	}
	sorted := append([]ForecastSample(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].At.Before(sorted[j].At) })
	rate, peak, pressure := 0.0, 0.0, 0.0
	for index, sample := range sorted {
		weight := float64(index + 1)
		rate += sample.EventRate * weight
		peak += sample.PeakMagnitude * weight
		pressure += sample.InjectionPressure * weight
	}
	divisor := float64(len(sorted)*(len(sorted)+1)) / 2
	rate /= divisor
	peak /= divisor
	pressure /= divisor
	score := rate*.35 + peak*.45 + math.Log1p(pressure)*.2
	band := RiskGreen
	if score >= 2 {
		band = RiskAmber
	}
	if score >= 4 {
		band = RiskRed
	}
	if score >= 7 {
		band = RiskBlack
	}
	confidence := math.Min(0.99, .5+float64(len(sorted))*.04)
	return RiskForecast{GeneratedAt: generatedAt.UTC(), HorizonEnd: generatedAt.Add(horizon).UTC(), Samples: len(sorted), ExpectedRate: rate, ExpectedPeak: peak, Confidence: confidence, Score: score, Band: band}, nil
}
func (f RiskForecast) AllowsOperation(maxBand RiskBand) bool {
	rank := map[RiskBand]int{RiskGreen: 0, RiskAmber: 1, RiskRed: 2, RiskBlack: 3}
	return rank[f.Band] <= rank[maxBand]
}
func ForecastTrend(samples []ForecastSample) float64 {
	if len(samples) < 2 {
		return 0
	}
	sorted := append([]ForecastSample(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].At.Before(sorted[j].At) })
	duration := sorted[len(sorted)-1].At.Sub(sorted[0].At).Hours()
	if duration <= 0 {
		return 0
	}
	return (sorted[len(sorted)-1].EventRate - sorted[0].EventRate) / duration
}
