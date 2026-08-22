package domain

import (
	"sort"
	"strings"
	"time"
)

type IncidentState string

const (
	IncidentOpen          IncidentState = "open"
	IncidentInvestigating IncidentState = "investigating"
	IncidentMitigated     IncidentState = "mitigated"
	IncidentClosed        IncidentState = "closed"
)

type RiskIncident struct {
	ID, FieldID, WellID, OwnerID, Title, Description string
	State                                            IncidentState
	Severity                                         int
	EventIDs                                         []string
	OpenedAt, UpdatedAt                              time.Time
	ClosedAt                                         *time.Time
	Version                                          int64
}

func (i RiskIncident) Validate() error {
	if i.ID == "" || i.FieldID == "" || i.WellID == "" || strings.TrimSpace(i.Title) == "" {
		return FieldError{"incident", "identity, field, well and title are required"}
	}
	if i.Severity < 1 || i.Severity > 5 {
		return FieldError{"severity", "must be between one and five"}
	}
	return nil
}
func (i RiskIncident) Assign(owner string, now time.Time) (RiskIncident, error) {
	if i.State != IncidentOpen && i.State != IncidentInvestigating {
		return i, TransitionError{"incident", string(i.State), string(IncidentInvestigating)}
	}
	if strings.TrimSpace(owner) == "" {
		return i, FieldError{"owner", "is required"}
	}
	i.OwnerID = owner
	i.State = IncidentInvestigating
	i.UpdatedAt = now.UTC()
	i.Version++
	return i, nil
}
func (i RiskIncident) AddEvent(eventID string, now time.Time) (RiskIncident, error) {
	if i.State == IncidentClosed {
		return i, ConflictError{"incident", "closed incident cannot accept evidence"}
	}
	for _, existing := range i.EventIDs {
		if existing == eventID {
			return i, nil
		}
	}
	i.EventIDs = append(append([]string(nil), i.EventIDs...), eventID)
	sort.Strings(i.EventIDs)
	i.UpdatedAt = now.UTC()
	i.Version++
	return i, nil
}
func (i RiskIncident) Mitigate(note string, now time.Time) (RiskIncident, error) {
	if i.State != IncidentInvestigating || len(i.EventIDs) == 0 {
		return i, ConflictError{"incident", "investigation and evidence are required"}
	}
	if strings.TrimSpace(note) == "" {
		return i, FieldError{"mitigation", "note is required"}
	}
	i.State = IncidentMitigated
	i.Description = SafeNote(note)
	i.UpdatedAt = now.UTC()
	i.Version++
	return i, nil
}
func (i RiskIncident) Close(now time.Time) (RiskIncident, error) {
	if i.State != IncidentMitigated {
		return i, TransitionError{"incident", string(i.State), string(IncidentClosed)}
	}
	i.State = IncidentClosed
	closed := now.UTC()
	i.ClosedAt = &closed
	i.UpdatedAt = closed
	i.Version++
	return i, nil
}

type IncidentFilter struct {
	FieldID, WellID, OwnerID string
	States                   []IncidentState
	MinimumSeverity          int
}

func (f IncidentFilter) Match(i RiskIncident) bool {
	if f.FieldID != "" && f.FieldID != i.FieldID {
		return false
	}
	if f.WellID != "" && f.WellID != i.WellID {
		return false
	}
	if f.OwnerID != "" && f.OwnerID != i.OwnerID {
		return false
	}
	if i.Severity < f.MinimumSeverity {
		return false
	}
	if len(f.States) > 0 {
		found := false
		for _, state := range f.States {
			if state == i.State {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
func FilterIncidents(items []RiskIncident, filter IncidentFilter) []RiskIncident {
	out := make([]RiskIncident, 0)
	for _, item := range items {
		if filter.Match(item) {
			out = append(out, item)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Severity == out[j].Severity {
			return out[i].OpenedAt.After(out[j].OpenedAt)
		}
		return out[i].Severity > out[j].Severity
	})
	return out
}
