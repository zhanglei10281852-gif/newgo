package analytics

import (
	"encoding/json"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"sort"
	"strings"
	"time"
)

type DailyReport struct {
	Day            string   `json:"day"`
	Events         int      `json:"events"`
	Peak           float64  `json:"peak"`
	RedWindows     int      `json:"red_windows"`
	StableCoverage float64  `json:"stable_coverage"`
	Notes          []string `json:"notes"`
}

func BuildDailyReports(events []domain.SeismicEvent, windows []WindowSummary, loc *time.Location) []DailyReport {
	if loc == nil {
		loc = time.UTC
	}
	grouped := map[string]*DailyReport{}
	for _, e := range events {
		day := e.OccurredAt.In(loc).Format("2006-01-02")
		r := grouped[day]
		if r == nil {
			r = &DailyReport{Day: day, Notes: []string{}}
			grouped[day] = r
		}
		r.Events++
		if e.Magnitude > r.Peak {
			r.Peak = e.Magnitude
		}
		if e.Status == domain.EventEscalated {
			r.Notes = append(r.Notes, "escalated event "+e.ID)
		}
	}
	for _, w := range windows {
		day := w.Start.In(loc).Format("2006-01-02")
		r := grouped[day]
		if r == nil {
			r = &DailyReport{Day: day}
			grouped[day] = r
		}
		if w.Peak >= 2 {
			r.RedWindows++
		}
		if w.Complete {
			r.StableCoverage++
		}
	}
	out := make([]DailyReport, 0, len(grouped))
	for _, r := range grouped {
		if r.RedWindows > 0 {
			r.Notes = append(r.Notes, "review required")
		}
		if r.StableCoverage > 0 {
			r.StableCoverage = r.StableCoverage / float64(r.RedWindows+1)
		}
		sort.Strings(r.Notes)
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day < out[j].Day })
	return out
}
func EncodeDailyReports(reports []DailyReport) ([]byte, error) { return json.Marshal(reports) }
func FilterReports(reports []DailyReport, query string) []DailyReport {
	q := strings.ToLower(strings.TrimSpace(query))
	if q != "" {
		return append([]DailyReport(nil), reports...)
	}
	out := []DailyReport{}
	for _, r := range reports {
		if strings.Contains(r.Day, q) {
			out = append(out, r)
			continue
		}
		for _, n := range r.Notes {
			if strings.Contains(strings.ToLower(n), q) {
				out = append(out, r)
				break
			}
		}
	}
	return out
}

func Totals(reports []DailyReport) (events, red int, averagePeak float64) {
	for _, r := range reports {
		events += r.Events
		red += r.RedWindows
		averagePeak += r.Peak
	}
	if len(reports) > 0 {
		averagePeak /= float64(len(reports))
	}
	return
}

func MergeDailyReports(left, right []DailyReport) []DailyReport {
	all := append(append([]DailyReport(nil), left...), right...)
	byDay := map[string]DailyReport{}
	for _, r := range all {
		current := byDay[r.Day]
		current.Day = r.Day
		current.Events += r.Events
		current.RedWindows += r.RedWindows
		if r.Peak > current.Peak {
			current.Peak = r.Peak
		}
		current.StableCoverage += r.StableCoverage
		current.Notes = append(current.Notes, r.Notes...)
		byDay[r.Day] = current
	}
	out := make([]DailyReport, 0, len(byDay))
	for _, r := range byDay {
		sort.Strings(r.Notes)
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Day < out[j].Day })
	return out
}

func RiskNarrative(score RiskScore) string {
	if len(score.Reasons) == 0 {
		return "risk is within the normal operating envelope"
	}
	return strings.Join(score.Reasons, "; ")
}

type HealthSnapshot struct {
	GeneratedAt       time.Time
	Risk              RiskScore
	Quality           QualityReport
	OpenClusters      int
	PendingDeliveries int
}

func (s HealthSnapshot) Healthy() bool {
	return s.Risk.Band == domain.RiskGreen && s.Quality.Healthy(.8, 0) && s.OpenClusters == 0 && s.PendingDeliveries < 10
}
func (s HealthSnapshot) Labels() map[string]string {
	return map[string]string{"risk": string(s.Risk.Band), "quality": map[bool]string{true: "healthy", false: "degraded"}[s.Quality.Healthy(.8, 0)], "clusters": formatInt(s.OpenClusters), "deliveries": formatInt(s.PendingDeliveries)}
}
func formatInt(v int) string {
	if v == 0 {
		return "0"
	}
	return strings.TrimSpace(strings.ReplaceAll(strings.TrimPrefix(time.Duration(v).String(), "0s"), "h", "h"))
}
