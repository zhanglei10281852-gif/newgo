package domain

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type RiskBand string

const (
	RiskGreen RiskBand = "green"
	RiskAmber RiskBand = "amber"
	RiskRed   RiskBand = "red"
	RiskBlack RiskBand = "black"
)

type RiskObservation struct {
	EventID            string
	Magnitude, DepthKm float64
	OccurredAt         time.Time
	Classified         bool
}
type RiskEnvelope struct {
	Band                                     RiskBand
	PeakMagnitude, DepthSpread, EventDensity float64
	EventCount, Unclassified                 int
	WindowStart, WindowEnd                   time.Time
}

func (r RiskEnvelope) AllowsPermit(threshold float64) bool {
	return r.Band == RiskGreen && r.Unclassified == 0 && r.PeakMagnitude < threshold
}
func (r RiskEnvelope) String() string {
	return fmt.Sprintf("%s peak=%.3f density=%.3f events=%d", r.Band, r.PeakMagnitude, r.EventDensity, r.EventCount)
}
func BuildRiskEnvelope(events []RiskObservation, start, end time.Time) (RiskEnvelope, error) {
	if !end.After(start) {
		return RiskEnvelope{}, FieldError{"window", "end must be after start"}
	}
	filtered := make([]RiskObservation, 0, len(events))
	for _, event := range events {
		if event.OccurredAt.Before(start) || event.OccurredAt.After(end) {
			continue
		}
		filtered = append(filtered, event)
	}
	if len(filtered) == 0 {
		return RiskEnvelope{Band: RiskGreen, WindowStart: start, WindowEnd: end}, nil
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i].OccurredAt.Before(filtered[j].OccurredAt) })
	peak, lowDepth, highDepth := 0.0, math.MaxFloat64, -math.MaxFloat64
	unclassified := 0
	for _, event := range filtered {
		if event.Magnitude > peak {
			peak = event.Magnitude
		}
		if event.DepthKm < lowDepth {
			lowDepth = event.DepthKm
		}
		if event.DepthKm > highDepth {
			highDepth = event.DepthKm
		}
		if !event.Classified {
			unclassified++
		}
	}
	duration := end.Sub(start).Hours()
	if duration < 1 {
		duration = 1
	}
	density := float64(len(filtered)) / duration
	band := RiskGreen
	if unclassified > 0 {
		band = RiskAmber
	}
	if peak >= 2.0 || density >= 10 {
		band = RiskRed
	}
	if peak >= 3.0 || density >= 25 {
		band = RiskBlack
	}
	return RiskEnvelope{Band: band, PeakMagnitude: peak, DepthSpread: highDepth - lowDepth, EventDensity: density, EventCount: len(filtered), Unclassified: unclassified, WindowStart: start, WindowEnd: end}, nil
}
func CompareRisk(before, after RiskEnvelope) int {
	rank := func(b RiskBand) int {
		switch b {
		case RiskGreen:
			return 0
		case RiskAmber:
			return 1
		case RiskRed:
			return 2
		case RiskBlack:
			return 3
		default:
			return 4
		}
	}
	a, c := rank(before.Band), rank(after.Band)
	if a < c {
		return 1
	}
	if a > c {
		return -1
	}
	return 0
}
