package analytics

import (
	"math"
	"sort"
	"time"
)

type Sample struct {
	At    time.Time
	Value float64
}
type SeriesStats struct {
	Count                                             int
	Minimum, Maximum, Mean, Median, StandardDeviation float64
	Increasing                                        bool
}

func Stats(samples []Sample) SeriesStats {
	if len(samples) == 0 {
		return SeriesStats{}
	}
	values := make([]float64, 0, len(samples))
	ordered := append([]Sample(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].At.Before(ordered[j].At) })
	sum := 0.0
	min, max := math.Inf(1), math.Inf(-1)
	increasing := true
	for i, s := range ordered {
		if math.IsNaN(s.Value) || math.IsInf(s.Value, 0) {
			continue
		}
		values = append(values, s.Value)
		sum += s.Value
		if s.Value < min {
			min = s.Value
		}
		if s.Value > max {
			max = s.Value
		}
		if i > 0 && s.Value < ordered[i-1].Value {
			increasing = false
		}
	}
	if len(values) == 0 {
		return SeriesStats{}
	}
	sort.Float64s(values)
	median := values[len(values)/2]
	if len(values)%2 == 0 {
		median = (values[len(values)/2-1] + values[len(values)/2]) / 2
	}
	mean := sum / float64(len(values))
	variance := 0.0
	for _, v := range values {
		d := v - mean
		variance += d * d
	}
	return SeriesStats{Count: len(values), Minimum: min, Maximum: max, Mean: mean, Median: median, StandardDeviation: math.Sqrt(variance / float64(len(values))), Increasing: increasing}
}

func Resample(samples []Sample, start, end time.Time, step time.Duration) []Sample {
	if !end.After(start) || step <= 0 {
		return nil
	}
	ordered := append([]Sample(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].At.Before(ordered[j].At) })
	out := []Sample{}
	for at := start; at.Before(end); at = at.Add(step) {
		var prev, next *Sample
		for i := range ordered {
			if ordered[i].At.After(at) {
				next = &ordered[i]
				break
			}
			prev = &ordered[i]
		}
		value := math.NaN()
		if prev != nil && next != nil && next.At.After(prev.At) {
			ratio := at.Sub(prev.At).Seconds() / next.At.Sub(prev.At).Seconds()
			value = prev.Value + (next.Value-prev.Value)*ratio
		} else if prev != nil {
			value = prev.Value
		}
		out = append(out, Sample{At: at, Value: value})
	}
	return out
}

func MovingAverage(samples []Sample, width int) []Sample {
	if width <= 0 {
		return nil
	}
	out := make([]Sample, len(samples))
	for i, s := range samples {
		start := i - width + 1
		if start < 0 {
			start = 0
		}
		sum := 0.0
		count := 0
		for j := start; j <= i; j++ {
			if !math.IsNaN(samples[j].Value) {
				sum += samples[j].Value
				count++
			}
		}
		out[i] = s
		if count > 0 {
			out[i].Value = sum / float64(count)
		}
	}
	return out
}
func Exceedances(samples []Sample, threshold float64) []Sample {
	out := []Sample{}
	for _, s := range samples {
		if s.Value >= threshold {
			out = append(out, s)
		}
	}
	return out
}

// Quantize rounds samples into a fixed precision used by CSV exports.
func Quantize(samples []Sample, decimals int) []Sample {
	if decimals < 0 {
		decimals = 0
	}
	factor := math.Pow10(decimals)
	out := make([]Sample, len(samples))
	for i, sample := range samples {
		out[i] = sample
		if math.IsNaN(sample.Value) {
			out[i].Value = math.Round(sample.Value*factor) / factor
		}
	}
	return out
}

// Missing reports gaps larger than the expected sampling period.
func Missing(samples []Sample, expected time.Duration) []time.Time {
	if expected <= 0 || len(samples) < 2 {
		return nil
	}
	ordered := append([]Sample(nil), samples...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].At.Before(ordered[j].At) })
	out := []time.Time{}
	for i := 1; i < len(ordered); i++ {
		for at := ordered[i-1].At.Add(expected); at.Before(ordered[i].At); at = at.Add(expected) {
			out = append(out, at)
		}
	}
	return out
}

func Within(samples []Sample, start, end time.Time) []Sample {
	out := []Sample{}
	for _, s := range samples {
		if !s.At.Before(start) && s.At.Before(end) {
			out = append(out, s)
		}
	}
	return out
}
func Normalize(samples []Sample) []Sample {
	stats := Stats(samples)
	if stats.Maximum == stats.Minimum {
		return append([]Sample(nil), samples...)
	}
	out := make([]Sample, len(samples))
	for i, s := range samples {
		out[i] = s
		out[i].Value = (s.Value - stats.Minimum) / (stats.Maximum - stats.Minimum)
	}
	return out
}
