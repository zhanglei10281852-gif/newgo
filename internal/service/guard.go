package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

type Guard struct {
	Store repository.Store
	Now   func() time.Time
}

func (g Guard) EnsureRole(ctx context.Context, user domain.User, action domain.Action) error {
	if user.Status != domain.UserActive {
		return fmt.Errorf("user inactive: %w", domain.ErrForbidden)
	}
	if !domain.Allowed(user.Role, action) {
		return fmt.Errorf("role %s cannot %s: %w", user.Role, action, domain.ErrForbidden)
	}
	return nil
}
func (g Guard) EnsureStationCoverage(ctx context.Context, wellID string, at time.Time, stations []domain.Station) error {
	if len(stations) == 0 {
		return domain.ConflictError{Resource: "station", Reason: "at least one monitoring station is required"}
	}
	for _, station := range stations {
		if station.WellID == wellID && station.Ready(at) {
			return nil
		}
	}
	return domain.ConflictError{Resource: "station", Reason: "no calibrated station covers the requested time"}
}
func (g Guard) EnsureNoActivePermit(ctx context.Context, slot string) error {
	if slot == "" {
		return domain.FieldError{Field: "equipment_slot", Message: "is required"}
	}
	return nil
}
func IsRetryable(err error) bool {
	return !errors.Is(err, domain.ErrInvalid) && !errors.Is(err, domain.ErrForbidden) && !errors.Is(err, domain.ErrConflict)
}
func Backoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > 8 {
		attempt = 8
	}
	return time.Duration(1<<uint(attempt-1)) * time.Second
}
func FilterPage(ctx context.Context, store repository.Store, batch, status string, page, size int) (repository.EventPage, error) {
	page, size = domain.ClampPage(page, size)
	return store.ListEvents(ctx, repository.EventFilter{BatchID: batch, Status: status, Page: repository.PageRequest{Page: page, Size: size}})
}
