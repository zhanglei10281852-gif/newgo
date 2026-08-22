package service

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

type Retention struct {
	Store repository.Store
	Clock func() time.Time
}

func (r Retention) ExpiredSession(ctx context.Context, s domain.Session) bool {
	return !s.Active(r.Clock())
}
func (r Retention) PermitExpired(p domain.StimulationPermit) bool {
	return p.ExpiresAt == nil || !p.ExpiresAt.After(r.Clock())
}
func (r Retention) ReviewWindow(p domain.StimulationPermit) time.Duration {
	if p.RequestedAt == nil || p.ExpiresAt == nil {
		return 0
	}
	return p.ExpiresAt.Sub(*p.RequestedAt)
}
func (r Retention) IsWithinGrace(p domain.StimulationPermit, grace time.Duration) bool {
	if p.ExpiresAt == nil {
		return false
	}
	return r.Clock().Sub(*p.ExpiresAt) <= grace
}
