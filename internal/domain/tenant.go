package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type TenantState string

const (
	TenantTrial     TenantState = "trial"
	TenantActive    TenantState = "active"
	TenantSuspended TenantState = "suspended"
	TenantClosed    TenantState = "closed"
)

type Tenant struct {
	ID, Code, Name, Timezone                   string
	State                                      TenantState
	WellLimit, StationLimit, MonthlyEventLimit int
	Version                                    int64
	CreatedAt, UpdatedAt                       time.Time
}

func (t Tenant) Validate() error {
	if t.ID == "" || strings.TrimSpace(t.Code) == "" || strings.TrimSpace(t.Name) == "" {
		return FieldError{"tenant", "id, code and name are required"}
	}
	if t.WellLimit < 1 || t.StationLimit < 1 || t.MonthlyEventLimit < 1 {
		return FieldError{"tenant_limits", "must be positive"}
	}
	if _, err := time.LoadLocation(t.Timezone); err != nil {
		return fmt.Errorf("tenant timezone: %w", err)
	}
	return nil
}
func (t Tenant) Activate(now time.Time) (Tenant, error) {
	if t.State != TenantTrial && t.State != TenantSuspended {
		return t, TransitionError{"tenant", string(t.State), string(TenantActive)}
	}
	t.State = TenantActive
	t.UpdatedAt = now.UTC()
	t.Version++
	return t, nil
}
func (t Tenant) Suspend(now time.Time) (Tenant, error) {
	if t.State != TenantActive {
		return t, TransitionError{"tenant", string(t.State), string(TenantSuspended)}
	}
	t.State = TenantSuspended
	t.UpdatedAt = now.UTC()
	t.Version++
	return t, nil
}
func (t Tenant) Close(now time.Time) (Tenant, error) {
	if t.State == TenantClosed {
		return t, nil
	}
	if t.State == TenantActive {
		return t, ConflictError{"tenant", "active tenant must be suspended first"}
	}
	t.State = TenantClosed
	t.UpdatedAt = now.UTC()
	t.Version++
	return t, nil
}

type TenantUsage struct {
	TenantID                string
	Month                   string
	Wells, Stations, Events int
}

func (u TenantUsage) Validate() error {
	if u.TenantID == "" || len(u.Month) != 7 {
		return FieldError{"usage", "tenant and YYYY-MM month are required"}
	}
	if u.Wells < 0 || u.Stations < 0 || u.Events < 0 {
		return FieldError{"usage", "counts cannot be negative"}
	}
	return nil
}
func (t Tenant) CheckUsage(u TenantUsage) error {
	if u.TenantID != t.ID {
		return ConflictError{"tenant", "usage belongs to another tenant"}
	}
	if u.Wells > t.WellLimit {
		return ConflictError{"tenant", "well quota exceeded"}
	}
	if u.Stations > t.StationLimit {
		return ConflictError{"tenant", "station quota exceeded"}
	}
	if u.Events > t.MonthlyEventLimit {
		return ConflictError{"tenant", "monthly event quota exceeded"}
	}
	return nil
}
func MergeUsage(items []TenantUsage) []TenantUsage {
	byKey := map[string]TenantUsage{}
	for _, item := range items {
		key := item.TenantID + "/" + item.Month
		current := byKey[key]
		current.TenantID = item.TenantID
		current.Month = item.Month
		current.Wells += item.Wells
		current.Stations += item.Stations
		current.Events += item.Events
		byKey[key] = current
	}
	out := make([]TenantUsage, 0, len(byKey))
	for _, item := range byKey {
		out = append(out, item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TenantID == out[j].TenantID {
			return out[i].Month < out[j].Month
		}
		return out[i].TenantID < out[j].TenantID
	})
	return out
}
