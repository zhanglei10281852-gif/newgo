package domain

import "time"

type BatchState string

const (
	BatchOpen       BatchState = "open"
	BatchCollecting BatchState = "collecting"
	BatchClosed     BatchState = "closed"
	BatchRejected   BatchState = "rejected"
)

type MonitoringBatch struct {
	ID, WellID, Reference string
	State                 BatchState
	StartsAt, EndsAt      time.Time
	EventCount            int
	Version               int64
	CreatedAt, UpdatedAt  time.Time
}

func (b MonitoringBatch) Start(now time.Time) (MonitoringBatch, error) {
	if b.State != BatchOpen {
		return b, TransitionError{"batch", string(b.State), string(BatchCollecting)}
	}
	b.State = BatchCollecting
	b.UpdatedAt = now.UTC()
	return b, nil
}
func (b MonitoringBatch) Close(now time.Time) (MonitoringBatch, error) {
	if b.State != BatchCollecting || b.EventCount == 0 {
		return b, ConflictError{"batch", "a collecting batch with events is required"}
	}
	b.State = BatchClosed
	b.UpdatedAt = now.UTC()
	return b, nil
}
func (b MonitoringBatch) Validate() error {
	if b.ID == "" || b.WellID == "" || b.Reference == "" {
		return FieldError{"batch", "well, id and reference are required"}
	}
	if !b.EndsAt.After(b.StartsAt) {
		return FieldError{"window", "end must be after start"}
	}
	return nil
}
