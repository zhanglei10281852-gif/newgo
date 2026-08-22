package analytics

import (
	"math"
	"sort"
	"time"

	"github.com/zhanglei10281852-gif/newgo/internal/domain"
)

type RiskScore struct {
	Score      float64
	Band       domain.RiskBand
	Reasons    []string
	ComputedAt time.Time
}

type RiskInput struct {
	Envelope  domain.RiskEnvelope
	Telemetry domain.TelemetryWindow
	Pressure  domain.PressureEnvelope
	PermitAge time.Duration
	Now       time.Time
}

func ScoreRisk(in RiskInput) RiskScore {
	score := 0.0
	reasons := []string{}
	if in.Envelope.PeakMagnitude >= 3 {
		score += 45
		reasons = append(reasons, "large seismic event")
	} else if in.Envelope.PeakMagnitude >= 2 {
		score += 25
		reasons = append(reasons, "elevated seismic magnitude")
	} else if in.Envelope.PeakMagnitude >= 1 {
		score += 10
	}
	if in.Envelope.Unclassified > 0 {
		score += 15
		reasons = append(reasons, "events awaiting classification")
	}
	if in.Envelope.EventDensity >= 25 {
		score += 25
		reasons = append(reasons, "dense event sequence")
	} else if in.Envelope.EventDensity >= 10 {
		score += 12
	}
	if !in.Telemetry.Stable() {
		score += 15
		reasons = append(reasons, "unstable station telemetry")
	}
	if in.Pressure.Violations > 0 {
		score += math.Min(20, float64(in.Pressure.Violations)*4)
		reasons = append(reasons, "pressure envelope violation")
	}
	if math.Abs(in.Pressure.RampRate) > 5 {
		score += 10
		reasons = append(reasons, "rapid pressure ramp")
	}
	if in.PermitAge > 24*time.Hour {
		score += 5
		reasons = append(reasons, "permit review overdue")
	}
	if score > 100 {
		score = 100
	}
	band := domain.RiskGreen
	if score >= 75 {
		band = domain.RiskBlack
	} else if score >= 50 {
		band = domain.RiskRed
	} else if score >= 25 {
		band = domain.RiskAmber
	}
	sort.Strings(reasons)
	return RiskScore{Score: score, Band: band, Reasons: reasons, ComputedAt: in.Now.UTC()}
}

func MergeScores(scores []RiskScore) RiskScore {
	if len(scores) == 0 {
		return RiskScore{Band: domain.RiskGreen}
	}
	best := scores[0]
	for _, s := range scores[1:] {
		if s.Score > best.Score {
			best = s
		}
	}
	seen := map[string]bool{}
	reasons := []string{}
	for _, s := range scores {
		for _, r := range s.Reasons {
			if !seen[r] {
				seen[r] = true
				reasons = append(reasons, r)
			}
		}
	}
	sort.Strings(reasons)
	best.Reasons = reasons
	return best
}

func IsActionable(score RiskScore, minimum float64) bool {
	return score.Score > minimum && score.Band != domain.RiskGreen
}
