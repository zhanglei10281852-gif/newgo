package analytics

import (
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"sort"
	"time"
)

type EscalationRule struct {
	Name        string
	MinimumBand domain.RiskBand
	Cooldown    time.Duration
	Repeat      bool
}
type Escalation struct {
	Rule    string
	Target  string
	Message string
	At      time.Time
}

func EvaluateEscalations(score RiskScore, rules []EscalationRule, last map[string]time.Time, now time.Time) []Escalation {
	bandRank := func(b domain.RiskBand) int {
		switch b {
		case domain.RiskGreen:
			return 0
		case domain.RiskAmber:
			return 1
		case domain.RiskRed:
			return 2
		case domain.RiskBlack:
			return 3
		}
		return 4
	}
	out := []Escalation{}
	sort.Slice(rules, func(i, j int) bool { return bandRank(rules[i].MinimumBand) < bandRank(rules[j].MinimumBand) })
	for _, r := range rules {
		if bandRank(score.Band) < bandRank(r.MinimumBand) {
			continue
		}
		if t, ok := last[r.Name]; ok && !r.Repeat && now.Sub(t) < r.Cooldown {
			continue
		}
		out = append(out, Escalation{Rule: r.Name, Target: "safety-control", Message: fmt.Sprintf("risk score %.1f reached %s", score.Score, score.Band), At: now.UTC()})
	}
	return out
}

type TimelineEntry struct {
	At                time.Time
	Kind, ID, Summary string
}

func BuildTimeline(events []domain.SeismicEvent, permits []domain.StimulationPermit) []TimelineEntry {
	out := make([]TimelineEntry, 0, len(events)+len(permits))
	for _, e := range events {
		out = append(out, TimelineEntry{At: e.OccurredAt, Kind: "seismic", ID: e.ID, Summary: e.Classification})
	}
	for _, p := range permits {
		out = append(out, TimelineEntry{At: p.CreatedAt, Kind: "permit", ID: p.ID, Summary: string(p.State)})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].At.Before(out[j].At) })
	return out
}
