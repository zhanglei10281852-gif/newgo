package service

import (
	"context"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/audit"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

type Operations struct {
	Store     repository.Store
	Clock     clock.Clock
	Audit     audit.Recorder
	PermitTTL time.Duration
}

func (o Operations) CreateField(ctx context.Context, actor string, f domain.Field) (domain.Field, error) {
	if !domain.Allowed(domain.Role(actor), domain.ActionManageField) {
		return f, domain.ErrForbidden
	}
	f.ID = identity.New("field")
	f.Version = 1
	f.CreatedAt = o.Clock.Now()
	f.UpdatedAt = f.CreatedAt
	if f.State == "" {
		f.State = domain.FieldDraft
	}
	if err := o.Store.InsertField(ctx, f); err != nil {
		return f, err
	}
	if o.Audit.Store == nil {
		return f, nil
	}
	return f, o.Audit.Record(ctx, actor, "field_created", "field", f.ID, "success", nil)
}
func (o Operations) ActivateField(ctx context.Context, actor, id string) (domain.Field, error) {
	f, e := o.Store.FieldByID(ctx, id)
	if e != nil {
		return f, e
	}
	next, e := f.Activate(o.Clock.Now())
	if e != nil {
		return f, e
	}
	if e = o.Store.UpdateField(ctx, next, f.Version); e != nil {
		return f, e
	}
	return next, nil
}
func (o Operations) CreateWell(ctx context.Context, actor string, w domain.Well) (domain.Well, error) {
	if !domain.Allowed(domain.Role(actor), domain.ActionManageField) {
		return w, domain.ErrForbidden
	}
	w.ID = identity.New("well")
	w.Version = 1
	w.CreatedAt = o.Clock.Now()
	w.UpdatedAt = w.CreatedAt
	if w.State == "" {
		w.State = domain.WellDrilled
	}
	if e := o.Store.InsertWell(ctx, w); e != nil {
		return w, e
	}
	return w, nil
}
func (o Operations) AddStation(ctx context.Context, s domain.Station) (domain.Station, error) {
	s.ID = identity.New("station")
	s.Version = 1
	s.CreatedAt = o.Clock.Now()
	s.UpdatedAt = s.CreatedAt
	if s.State == "" {
		s.State = domain.StationCalibrating
	}
	return s, o.Store.InsertStation(ctx, s)
}
func (o Operations) CalibrateStation(ctx context.Context, id string, due time.Time) (domain.Station, error) {
	s, e := o.Store.StationByID(ctx, id)
	if e != nil {
		return s, e
	}
	next, e := s.Calibrate(o.Clock.Now(), due)
	if e != nil {
		return s, e
	}
	e = o.Store.UpdateStation(ctx, next, s.Version)
	return next, e
}
func (o Operations) RequireActiveWell(ctx context.Context, id string) (domain.Well, error) {
	w, e := o.Store.WellByID(ctx, id)
	if e != nil {
		return w, e
	}
	if w.State != domain.WellMonitored {
		return w, fmt.Errorf("well not monitored: %w", domain.ErrConflict)
	}
	return w, nil
}
