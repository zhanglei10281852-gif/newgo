package domain

import "time"

type AuditEvent struct {
	ID, ActorID, Action, EntityType, EntityID, Outcome, RequestID, Metadata string
	CreatedAt                                                               time.Time
}
