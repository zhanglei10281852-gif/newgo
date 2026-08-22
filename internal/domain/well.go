package domain

import "time"

type WellState string

const (
	WellDrilled   WellState = "drilled"
	WellMonitored WellState = "monitored"
	WellSuspended WellState = "suspended"
	WellRetired   WellState = "retired"
)

type Well struct {
	ID, FieldID, Name, APIIdentifier string
	State                            WellState
	MaxMagnitude                     float64
	Version                          int64
	CreatedAt, UpdatedAt             time.Time
}

func (w Well) BeginMonitoring(now time.Time) (Well, error) {
	if w.State != WellDrilled {
		return w, TransitionError{"well", string(w.State), string(WellMonitored)}
	}
	w.State = WellMonitored
	w.UpdatedAt = now.UTC()
	return w, nil
}
func (w Well) Suspend(now time.Time) (Well, error) {
	if w.State != WellMonitored {
		return w, TransitionError{"well", string(w.State), string(WellSuspended)}
	}
	w.State = WellSuspended
	w.UpdatedAt = now.UTC()
	return w, nil
}
func (w Well) Validate() error {
	if w.ID == "" || w.FieldID == "" || w.APIIdentifier == "" {
		return FieldError{"well", "field, id and API identifier are required"}
	}
	if w.MaxMagnitude <= 0 {
		return FieldError{"max_magnitude", "must be positive"}
	}
	return nil
}
