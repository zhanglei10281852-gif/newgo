package domain

import (
	"sort"
	"time"
)

type MaintenanceKind string

const (
	MaintenanceCalibration MaintenanceKind = "calibration"
	MaintenanceFirmware    MaintenanceKind = "firmware"
	MaintenanceBattery     MaintenanceKind = "battery"
	MaintenanceInspection  MaintenanceKind = "inspection"
)

type MaintenanceTicket struct {
	ID, StationID, AssignedTo string
	Kind                      MaintenanceKind
	State                     string
	DueAt, CompletedAt        *time.Time
	Priority                  int
	Note                      string
}

func (t MaintenanceTicket) Open() bool { return t.State == "open" || t.State == "assigned" }
func (t MaintenanceTicket) Assign(operator string, now time.Time) (MaintenanceTicket, error) {
	if operator == "" {
		return t, FieldError{"assigned_to", "is required"}
	}
	if !t.Open() {
		return t, TransitionError{"maintenance", t.State, "assigned"}
	}
	t.AssignedTo = operator
	t.State = "assigned"
	if t.DueAt == nil {
		due := now.Add(24 * time.Hour)
		t.DueAt = &due
	}
	return t, nil
}
func (t MaintenanceTicket) Complete(note string, now time.Time) (MaintenanceTicket, error) {
	if !t.Open() {
		return t, TransitionError{"maintenance", t.State, "completed"}
	}
	t.Note = SafeNote(note)
	t.State = "completed"
	t.CompletedAt = &now
	return t, nil
}
func OrderMaintenance(tickets []MaintenanceTicket) []MaintenanceTicket {
	out := append([]MaintenanceTicket(nil), tickets...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority == out[j].Priority {
			return out[i].DueAt != nil && out[j].DueAt != nil && out[i].DueAt.Before(*out[j].DueAt)
		}
		return out[i].Priority > out[j].Priority
	})
	return out
}
func ExpiredTickets(tickets []MaintenanceTicket, now time.Time) []MaintenanceTicket {
	out := make([]MaintenanceTicket, 0)
	for _, ticket := range tickets {
		if ticket.Open() && ticket.DueAt != nil && ticket.DueAt.Before(now) {
			out = append(out, ticket)
		}
	}
	return out
}
