package domain

import (
	"sort"
	"time"
)

type RetentionClass string

const (
	RetentionRawTelemetry     RetentionClass = "raw_telemetry"
	RetentionClassifiedEvents RetentionClass = "classified_events"
	RetentionAudit            RetentionClass = "audit"
	RetentionPermits          RetentionClass = "permits"
)

type RetentionPolicy struct {
	Class           RetentionClass
	Duration        time.Duration
	LegalHold       bool
	DeleteBatchSize int
}
type RetentionObject struct {
	ID        string
	Class     RetentionClass
	CreatedAt time.Time
	LegalHold bool
	DeletedAt *time.Time
}

func (p RetentionPolicy) Validate() error {
	if p.Class == "" || p.Duration <= 0 || p.DeleteBatchSize < 1 {
		return FieldError{"retention", "class, duration and batch size are required"}
	}
	return nil
}
func (p RetentionPolicy) Expired(object RetentionObject, now time.Time) bool {
	if p.LegalHold || object.LegalHold || object.DeletedAt != nil || object.Class != p.Class {
		return false
	}
	return !object.CreatedAt.Add(p.Duration).After(now)
}
func SelectExpired(policy RetentionPolicy, objects []RetentionObject, now time.Time) []RetentionObject {
	out := make([]RetentionObject, 0)
	for _, object := range objects {
		if policy.Expired(object, now) {
			out = append(out, object)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	if len(out) > policy.DeleteBatchSize {
		out = out[:policy.DeleteBatchSize]
	}
	return out
}
func MarkDeleted(objects []RetentionObject, ids []string, now time.Time) []RetentionObject {
	wanted := map[string]bool{}
	for _, id := range ids {
		wanted[id] = true
	}
	out := append([]RetentionObject(nil), objects...)
	for index := range out {
		if wanted[out[index].ID] && out[index].DeletedAt == nil {
			deleted := now.UTC()
			out[index].DeletedAt = &deleted
		}
	}
	return out
}
func DefaultRetentionPolicies() []RetentionPolicy {
	return []RetentionPolicy{{Class: RetentionRawTelemetry, Duration: 90 * 24 * time.Hour, DeleteBatchSize: 500}, {Class: RetentionClassifiedEvents, Duration: 7 * 365 * 24 * time.Hour, DeleteBatchSize: 100}, {Class: RetentionAudit, Duration: 10 * 365 * 24 * time.Hour, DeleteBatchSize: 100}, {Class: RetentionPermits, Duration: 10 * 365 * 24 * time.Hour, DeleteBatchSize: 100}}
}
