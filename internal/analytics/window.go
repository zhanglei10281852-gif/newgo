package analytics

import (
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"sort"
	"time"
)

type WindowSummary struct {
	Start, End       time.Time
	Events, Stations int
	Peak, Mean       float64
	Complete         bool
}

func SummarizeWindows(events []domain.SeismicEvent, telemetry []domain.TelemetryPoint, start, end time.Time, step time.Duration) []WindowSummary {
	if !end.After(start) || step <= 0 {
		return nil
	}
	out := []WindowSummary{}
	for cursor := start; cursor.Before(end); cursor = cursor.Add(step) {
		next := cursor.Add(step)
		if next.After(end) {
			next = end
		}
		s := WindowSummary{Start: cursor, End: next, Complete: true}
		station := map[string]bool{}
		sum := 0.0
		for _, e := range events {
			if !e.OccurredAt.Before(cursor) && e.OccurredAt.Before(next) {
				s.Events++
				sum += e.Magnitude
				if e.Magnitude > s.Peak {
					s.Peak = e.Magnitude
				}
			}
		}
		for _, p := range telemetry {
			if !p.At.Before(cursor) && p.At.Before(next) {
				station[p.StationID] = true
				if !p.Valid {
					s.Complete = false
				}
			}
		}
		s.Stations = len(station)
		if s.Events > 0 {
			s.Mean = sum / float64(s.Events)
		}
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Start.Before(out[j].Start) })
	return out
}
