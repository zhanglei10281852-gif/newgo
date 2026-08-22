package domain

import "time"

type EventStatus string

const (
	EventUnclassified EventStatus = "unclassified"
	EventClassified   EventStatus = "classified"
	EventEscalated    EventStatus = "escalated"
)

type SeismicEvent struct {
	ID, BatchID, WellID   string
	OccurredAt            time.Time
	Magnitude, DepthKm    float64
	Status                EventStatus
	Classification, Notes string
	Version               int64
	CreatedAt, UpdatedAt  time.Time
}

func (e SeismicEvent) Classify(label, notes string, now time.Time) (SeismicEvent, error) {
	if e.Status != EventUnclassified {
		return e, TransitionError{"event", string(e.Status), string(EventClassified)}
	}
	if label == "" {
		return e, FieldError{"classification", "is required"}
	}
	e.Status = EventClassified
	e.Classification = label
	e.Notes = notes
	e.UpdatedAt = now.UTC()
	return e, nil
}
func (e SeismicEvent) Escalate(reason string, now time.Time) (SeismicEvent, error) {
	if e.Status == EventEscalated {
		return e, nil
	}
	e.Status = EventEscalated
	e.Notes = reason
	e.UpdatedAt = now.UTC()
	return e, nil
}
func (e SeismicEvent) Validate() error {
	if e.ID == "" || e.BatchID == "" || e.WellID == "" {
		return FieldError{"event", "batch and well are required"}
	}
	if e.Magnitude < 0 || e.DepthKm < 0 {
		return FieldError{"event", "measurements cannot be negative"}
	}
	return nil
}
