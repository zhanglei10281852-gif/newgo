package analytics

import (
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"sort"
	"time"
)

type SLAStatus string

const (
	SLAOnTrack SLAStatus = "on_track"
	SLADueSoon SLAStatus = "due_soon"
	SLAOverdue SLAStatus = "overdue"
)

type SLAItem struct {
	ID, Kind string
	DueAt    time.Time
	Status   SLAStatus
	Owner    string
}

func BuildSLAQueue(now time.Time, permits []domain.StimulationPermit, calibrations []domain.CalibrationReport, reviewWindow time.Duration) []SLAItem {
	out := []SLAItem{}
	for _, p := range permits {
		due := p.CreatedAt.Add(reviewWindow)
		status := SLAOnTrack
		if now.After(due) {
			status = SLAOverdue
		} else if due.Sub(now) <= reviewWindow/3 {
			status = SLADueSoon
		}
		out = append(out, SLAItem{ID: p.ID, Kind: "permit", DueAt: due, Status: status, Owner: p.RequesterID})
	}
	for _, c := range calibrations {
		due := c.DueAt
		status := SLAOnTrack
		if now.After(due) {
			status = SLAOverdue
		} else if due.Sub(now) <= reviewWindow/3 {
			status = SLADueSoon
		}
		out = append(out, SLAItem{ID: c.StationID, Kind: "calibration", DueAt: due, Status: status, Owner: c.StationID})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DueAt.Before(out[j].DueAt) })
	return out
}
func CountSLA(items []SLAItem) map[SLAStatus]int {
	out := map[SLAStatus]int{}
	for _, i := range items {
		out[i.Status]++
	}
	return out
}
func NextSLA(items []SLAItem) (SLAItem, bool) {
	if len(items) == 0 {
		return SLAItem{}, false
	}
	best := items[0]
	for _, i := range items[1:] {
		if i.DueAt.Before(best.DueAt) {
			best = i
		}
	}
	return best, true
}
