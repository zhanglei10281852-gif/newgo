package analytics

import (
	"errors"
	"math"
	"sort"
	"time"

	"github.com/zhanglei10281852-gif/newgo/internal/domain"
)

// Forecast describes a short-horizon pressure response used by the permit gate.
type Forecast struct {
	Horizon       time.Duration
	Slope         float64
	Intercept     float64
	PredictedPeak float64
	Confidence    float64
	Samples       int
}

func FitPressure(points []domain.PressureReading, horizon time.Duration) (Forecast, error) {
	if horizon <= 0 {
		return Forecast{}, errors.New("horizon must be positive")
	}
	clean := make([]domain.PressureReading, 0, len(points))
	for _, p := range points {
		if !p.At.IsZero() && !math.IsNaN(p.KiloPascal) && !math.IsInf(p.KiloPascal, 0) {
			clean = append(clean, p)
		}
	}
	if len(clean) < 2 {
		return Forecast{}, errors.New("at least two pressure samples are required")
	}
	sort.Slice(clean, func(i, j int) bool { return clean[i].At.Before(clean[j].At) })
	base := clean[0].At
	var sx, sy, sxx, sxy float64
	for _, p := range clean {
		x := p.At.Sub(base).Seconds()
		sx += x
		sy += p.KiloPascal
		sxx += x * x
		sxy += x * p.KiloPascal
	}
	n := float64(len(clean))
	den := n*sxx - sx*sx
	if math.Abs(den) < 1e-9 {
		return Forecast{}, errors.New("pressure samples have no time spread")
	}
	slope := (n*sxy - sx*sy) / den
	intercept := (sy - slope*sx) / n
	last := clean[len(clean)-1].At.Sub(base).Seconds()
	peak := slope*(last+horizon.Seconds()) + intercept
	residual := 0.0
	for _, p := range clean {
		d := p.KiloPascal - (slope*p.At.Sub(base).Seconds() + intercept)
		residual += d * d
	}
	confidence := 1 / (1 + math.Sqrt(residual/n))
	if confidence > 1 {
		confidence = 1
	}
	return Forecast{Horizon: horizon, Slope: slope, Intercept: intercept, PredictedPeak: peak, Confidence: confidence, Samples: len(clean)}, nil
}

func ShouldPause(f Forecast, limit, minConfidence float64) bool {
	return f.PredictedPeak > limit && f.Confidence >= minConfidence
}
