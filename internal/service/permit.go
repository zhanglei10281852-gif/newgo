package service

import (
	"context"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

type PermitService struct {
	Store repository.Store
	Clock func() time.Time
	Rule  domain.PermitRule
}

func (p PermitService) Preflight(ctx context.Context, wellID string, start, end time.Time, stationCount int) (domain.PermitDecision, error) {
	query := QueryService{Store: p.Store, Clock: p.Clock}
	envelope, e := query.RiskSnapshot(ctx, wellID, start, end)
	if e != nil {
		return domain.PermitDecision{}, e
	}
	rule := p.Rule
	if rule.Name == "" {
		rule = domain.DefaultPermitRule()
	}
	if d := domain.EvaluatePermit(rule, envelope, stationCount); !d.Allowed {
		return d, nil
	}
	return domain.EvaluatePermit(rule, envelope, stationCount), nil
}
func (p PermitService) Explain(ctx context.Context, wellID string, start, end time.Time, threshold float64) (string, error) {
	envelope, e := (QueryService{Store: p.Store, Clock: p.Clock}).RiskSnapshot(ctx, wellID, start, end)
	if e != nil {
		return "", e
	}
	if envelope.AllowsPermit(threshold) {
		return "eligible", nil
	}
	return fmt.Sprintf("blocked: %s", envelope.String()), nil
}
