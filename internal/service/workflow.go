package service

import (
	"context"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

type Workflow struct {
	Store     repository.Store
	Clock     func() time.Time
	PermitTTL time.Duration
}

func (w Workflow) OpenBatch(ctx context.Context, b domain.MonitoringBatch) (domain.MonitoringBatch, error) {
	b.ID = identity.New("batch")
	b.Version = 1
	b.CreatedAt = w.Clock()
	b.UpdatedAt = b.CreatedAt
	if b.State == "" {
		b.State = domain.BatchOpen
	}
	return b, w.Store.InsertBatch(ctx, b)
}
func (w Workflow) StartBatch(ctx context.Context, id string) (domain.MonitoringBatch, error) {
	b, e := w.Store.BatchByID(ctx, id)
	if e != nil {
		return b, e
	}
	next, e := b.Start(w.Clock())
	if e != nil {
		return b, e
	}
	return next, w.Store.UpdateBatch(ctx, next, b.Version)
}
func (w Workflow) RecordEvent(ctx context.Context, e domain.SeismicEvent) (domain.SeismicEvent, error) {
	if _, err := w.Store.WellByID(ctx, e.WellID); err != nil {
		return e, err
	}
	if _, err := w.Store.BatchByID(ctx, e.BatchID); err != nil {
		return e, err
	}
	e.ID = identity.New("event")
	e.Version = 1
	e.CreatedAt = w.Clock()
	e.UpdatedAt = e.CreatedAt
	if e.Status == "" {
		e.Status = domain.EventUnclassified
	}
	return e, w.Store.InsertEvent(ctx, e)
}
func (w Workflow) ClassifyEvent(ctx context.Context, id, label, notes string) (domain.SeismicEvent, error) {
	e, err := w.Store.EventByID(ctx, id)
	if err != nil {
		return e, err
	}
	next, err := e.Classify(label, notes, w.Clock())
	if err != nil {
		return e, err
	}
	return next, w.Store.UpdateEvent(ctx, next, e.Version)
}
func (w Workflow) SubmitPermit(ctx context.Context, p domain.StimulationPermit) (domain.StimulationPermit, error) {
	well, err := w.Store.WellByID(ctx, p.WellID)
	if err != nil {
		return p, err
	}
	if well.State != domain.WellMonitored {
		return p, fmt.Errorf("well is not monitoring: %w", domain.ErrConflict)
	}
	p.ID = identity.New("permit")
	p.Version = 1
	p.CreatedAt = w.Clock()
	p.UpdatedAt = p.CreatedAt
	expires := w.Clock().Add(w.PermitTTL)
	p.ExpiresAt = &expires
	if p.State == "" {
		p.State = domain.PermitDraft
	}
	next, err := p.Submit(w.Clock())
	if err != nil {
		return p, err
	}
	return next, w.Store.InsertPermit(ctx, next)
}
func (w Workflow) DecidePermit(ctx context.Context, id string, approved bool) (domain.StimulationPermit, error) {
	p, e := w.Store.PermitByID(ctx, id)
	if e != nil {
		return p, e
	}
	if p.ExpiresAt.Before(w.Clock()) {
		return p, domain.ErrExpired
	}
	next, e := p.Decide(approved, w.Clock())
	if e != nil {
		return p, e
	}
	return next, w.Store.UpdatePermit(ctx, next, p.Version)
}
func (w Workflow) ExecutePermit(ctx context.Context, id string) (domain.StimulationPermit, error) {
	p, e := w.Store.PermitByID(ctx, id)
	if e != nil {
		return p, e
	}
	if p.ExpiresAt.Before(w.Clock()) {
		return p, domain.ErrExpired
	}
	next, e := p.Execute(w.Clock())
	if e != nil {
		return p, e
	}
	return next, w.Store.UpdatePermit(ctx, next, p.Version)
}
