package domain

import (
	"math"
	"sort"
	"time"
)

type TelemetryPoint struct {
	StationID                   string
	At                          time.Time
	Amplitude, Frequency, Noise float64
	Valid                       bool
}
type TelemetryWindow struct {
	Start, End  time.Time
	Points      []TelemetryPoint
	Stations    int
	ValidPoints int
	NoiseFloor  float64
}

func BuildTelemetryWindow(points []TelemetryPoint, start, end time.Time) TelemetryWindow {
	window := TelemetryWindow{Start: start, End: end, Points: make([]TelemetryPoint, 0)}
	stations := map[string]struct{}{}
	noise := 0.0
	for _, point := range points {
		if point.At.Before(start) || !point.At.Before(end) {
			continue
		}
		window.Points = append(window.Points, point)
		stations[point.StationID] = struct{}{}
		if point.Valid {
			window.ValidPoints++
			noise += point.Noise
		}
	}
	window.Stations = len(stations)
	if window.ValidPoints > 0 {
		window.NoiseFloor = noise / float64(window.ValidPoints)
	}
	sort.Slice(window.Points, func(i, j int) bool { return window.Points[i].At.Before(window.Points[j].At) })
	return window
}
func (w TelemetryWindow) Coverage() float64 {
	if len(w.Points) == 0 {
		return 0
	}
	return float64(w.ValidPoints) / float64(len(w.Points))
}
func (w TelemetryWindow) Stable() bool {
	return w.Stations >= 2 && w.ValidPoints >= 3 && w.Coverage() >= 0.8 && w.NoiseFloor < 0.5
}
func (w TelemetryWindow) AmplitudeRange() (float64, float64) {
	if len(w.Points) == 0 {
		return 0, 0
	}
	low, high := math.MaxFloat64, -math.MaxFloat64
	for _, point := range w.Points {
		if !point.Valid {
			continue
		}
		if point.Amplitude < low {
			low = point.Amplitude
		}
		if point.Amplitude > high {
			high = point.Amplitude
		}
	}
	if low == math.MaxFloat64 {
		return 0, 0
	}
	return low, high
}
func (w TelemetryWindow) FrequencyMean() float64 {
	sum, count := 0.0, 0
	for _, point := range w.Points {
		if point.Valid {
			sum += point.Frequency
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return sum / float64(count)
}
func ValidateTelemetry(point TelemetryPoint) error {
	if point.StationID == "" {
		return FieldError{"station_id", "is required"}
	}
	if point.Amplitude < 0 || point.Frequency < 0 || point.Noise < 0 {
		return FieldError{"telemetry", "measurements cannot be negative"}
	}
	return nil
}
func DeduplicateTelemetry(points []TelemetryPoint) []TelemetryPoint {
	seen := map[string]TelemetryPoint{}
	for _, point := range points {
		key := point.StationID + "/" + point.At.UTC().Format(time.RFC3339Nano)
		seen[key] = point
	}
	out := make([]TelemetryPoint, 0, len(seen))
	for _, point := range seen {
		out = append(out, point)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}
