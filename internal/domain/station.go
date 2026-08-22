package domain

import "time"

type StationState string

const (
	StationCalibrating StationState = "calibrating"
	StationReady       StationState = "ready"
	StationOffline     StationState = "offline"
)

type Station struct {
	ID, WellID, Serial, Location string
	State                        StationState
	CalibrationDueAt             time.Time
	Version                      int64
	CreatedAt, UpdatedAt         time.Time
}

func (s Station) Ready(now time.Time) bool {
	return s.State == StationReady && s.CalibrationDueAt.After(now)
}
func (s Station) Calibrate(now, due time.Time) (Station, error) {
	if s.State != StationCalibrating && s.State != StationOffline {
		return s, TransitionError{"station", string(s.State), string(StationReady)}
	}
	s.State = StationReady
	s.CalibrationDueAt = due.UTC()
	s.UpdatedAt = now.UTC()
	return s, nil
}
func (s Station) Validate() error {
	if s.ID == "" || s.WellID == "" || s.Serial == "" {
		return FieldError{"station", "well, id and serial are required"}
	}
	return nil
}
