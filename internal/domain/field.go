package domain

import "time"

type FieldState string

const (
	FieldDraft    FieldState = "draft"
	FieldActive   FieldState = "active"
	FieldArchived FieldState = "archived"
)

type Field struct {
	ID, Code, Name, Timezone string
	State                    FieldState
	RiskThreshold            float64
	Version                  int64
	CreatedAt, UpdatedAt     time.Time
}

func (f Field) Activate(now time.Time) (Field, error) {
	if f.State != FieldDraft {
		return f, TransitionError{"field", string(f.State), string(FieldActive)}
	}
	f.State = FieldActive
	f.UpdatedAt = now.UTC()
	return f, nil
}
func (f Field) Archive(now time.Time) (Field, error) {
	if f.State != FieldActive {
		return f, TransitionError{"field", string(f.State), string(FieldArchived)}
	}
	f.State = FieldArchived
	f.UpdatedAt = now.UTC()
	return f, nil
}
func (f Field) Validate() error {
	if f.ID == "" || f.Code == "" || f.Timezone == "" {
		return FieldError{"field", "id, code and timezone are required"}
	}
	if f.RiskThreshold <= 0 {
		return FieldError{"risk_threshold", "must be positive"}
	}
	return nil
}
