package domain

import "time"

type TimeWindow struct {
	Start, End time.Time
	Location   *time.Location
}

func NewWindow(start, end time.Time, loc *time.Location) (TimeWindow, error) {
	if loc == nil {
		loc = time.UTC
	}
	if !end.After(start) {
		return TimeWindow{}, FieldError{"window", "end must be after start"}
	}
	return TimeWindow{Start: start.In(loc), End: end.In(loc), Location: loc}, nil
}
func (w TimeWindow) Contains(value time.Time) bool {
	return !value.Before(w.Start) && value.Before(w.End)
}
func (w TimeWindow) Duration() time.Duration { return w.End.Sub(w.Start) }
func (w TimeWindow) Overlaps(other TimeWindow) bool {
	return w.Start.Before(other.End) && other.Start.Before(w.End)
}
func (w TimeWindow) LocalDate() string { return w.Start.In(w.Location).Format("2006-01-02") }
func (w TimeWindow) Split(step time.Duration) []TimeWindow {
	if step <= 0 {
		return nil
	}
	out := make([]TimeWindow, 0)
	for start := w.Start; start.Before(w.End); {
		end := start.Add(step)
		if end.After(w.End) {
			end = w.End
		}
		part, _ := NewWindow(start, end, w.Location)
		out = append(out, part)
		start = end
	}
	return out
}
