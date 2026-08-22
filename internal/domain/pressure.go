package domain

import (
	"math"
	"sort"
	"time"
)

type PressureReading struct {
	At                              time.Time
	WellID                          string
	KiloPascal, FlowLitersPerSecond float64
	Source                          string
}
type PressureEnvelope struct {
	Start, End                                 time.Time
	Minimum, Maximum, Mean, FlowMean, RampRate float64
	Samples                                    int
	Violations                                 int
}

func BuildPressureEnvelope(readings []PressureReading, start, end time.Time, minimum, maximum float64) (PressureEnvelope, error) {
	if !end.After(start) || minimum >= maximum {
		return PressureEnvelope{}, FieldError{"pressure_window", "bounds are invalid"}
	}
	selected := make([]PressureReading, 0)
	for _, reading := range readings {
		if reading.At.Before(start) || !reading.At.Before(end) {
			continue
		}
		selected = append(selected, reading)
	}
	if len(selected) == 0 {
		return PressureEnvelope{Start: start, End: end}, nil
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].At.Before(selected[j].At) })
	out := PressureEnvelope{Start: start, End: end, Minimum: math.MaxFloat64, Maximum: -math.MaxFloat64, Samples: len(selected)}
	for _, reading := range selected {
		out.Mean += reading.KiloPascal
		out.FlowMean += reading.FlowLitersPerSecond
		if reading.KiloPascal < out.Minimum {
			out.Minimum = reading.KiloPascal
		}
		if reading.KiloPascal > out.Maximum {
			out.Maximum = reading.KiloPascal
		}
		if reading.KiloPascal < minimum || reading.KiloPascal > maximum {
			out.Violations++
		}
	}
	out.Mean /= float64(len(selected))
	out.FlowMean /= float64(len(selected))
	hours := selected[len(selected)-1].At.Sub(selected[0].At).Hours()
	if hours > 0 {
		out.RampRate = (selected[len(selected)-1].KiloPascal - selected[0].KiloPascal) / hours
	}
	return out, nil
}
func (e PressureEnvelope) Stable(maxRamp float64) bool {
	return e.Samples >= 3 && e.Violations == 0 && math.Abs(e.RampRate) <= maxRamp
}
func (e PressureEnvelope) Margin(minimum, maximum float64) float64 {
	lower := e.Minimum - minimum
	upper := maximum - e.Maximum
	if lower < upper {
		return lower
	}
	return upper
}
