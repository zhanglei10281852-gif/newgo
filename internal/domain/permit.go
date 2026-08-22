package domain

import "time"

type PermitState string

const (
	PermitDraft    PermitState = "draft"
	PermitPending  PermitState = "pending_review"
	PermitApproved PermitState = "approved"
	PermitDenied   PermitState = "denied"
	PermitExpired  PermitState = "expired"
	PermitExecuted PermitState = "executed"
)

type StimulationPermit struct {
	ID, WellID, BatchID, RequesterID, ReviewerID, EquipmentSlot string
	State                                                       PermitState
	RequestedAt, ExpiresAt, ReviewedAt, ExecutedAt              *time.Time
	RiskScore                                                   float64
	Version                                                     int64
	CreatedAt, UpdatedAt                                        time.Time
}

func (p StimulationPermit) Submit(now time.Time) (StimulationPermit, error) {
	if p.State != PermitDraft || p.RequesterID == "" || p.ReviewerID == "" || p.RequesterID == p.ReviewerID {
		return p, ConflictError{"permit", "draft permit needs two different operators"}
	}
	p.State = PermitPending
	t := now.UTC()
	p.RequestedAt = &t
	p.UpdatedAt = t
	return p, nil
}
func (p StimulationPermit) Decide(approved bool, now time.Time) (StimulationPermit, error) {
	if p.State != PermitPending {
		return p, TransitionError{"permit", string(p.State), "decision"}
	}
	t := now.UTC()
	p.ReviewedAt = &t
	p.UpdatedAt = t
	if approved {
		p.State = PermitApproved
	} else {
		p.State = PermitDenied
	}
	return p, nil
}
func (p StimulationPermit) Execute(now time.Time) (StimulationPermit, error) {
	if p.State != PermitApproved {
		return p, TransitionError{"permit", string(p.State), string(PermitExecuted)}
	}
	t := now.UTC()
	p.ExecutedAt = &t
	p.State = PermitExecuted
	p.UpdatedAt = t
	return p, nil
}
func (p StimulationPermit) Validate() error {
	if p.ID == "" || p.WellID == "" || p.BatchID == "" || p.EquipmentSlot == "" {
		return FieldError{"permit", "well, batch and slot are required"}
	}
	if p.ExpiresAt == nil {
		return FieldError{"expires_at", "is required"}
	}
	return nil
}
