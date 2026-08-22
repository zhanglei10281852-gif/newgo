package repository

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"time"
)

type PageRequest struct{ Page, Size int }
type EventFilter struct {
	BatchID, Status string
	Page            PageRequest
}
type EventPage struct {
	Items []domain.SeismicEvent
	Total int
}

type Store interface {
	Close() error
	WithTx(context.Context, func(Store) error) error
	InsertUser(context.Context, domain.User) error
	UserByEmail(context.Context, string) (domain.User, error)
	UserByID(context.Context, string) (domain.User, error)
	UpdateUser(context.Context, domain.User, int64) error
	InsertSession(context.Context, domain.Session) error
	SessionByID(context.Context, string) (domain.Session, error)
	RevokeSession(context.Context, string, time.Time) error
	InsertField(context.Context, domain.Field) error
	FieldByID(context.Context, string) (domain.Field, error)
	UpdateField(context.Context, domain.Field, int64) error
	InsertWell(context.Context, domain.Well) error
	WellByID(context.Context, string) (domain.Well, error)
	UpdateWell(context.Context, domain.Well, int64) error
	InsertStation(context.Context, domain.Station) error
	StationByID(context.Context, string) (domain.Station, error)
	UpdateStation(context.Context, domain.Station, int64) error
	InsertBatch(context.Context, domain.MonitoringBatch) error
	BatchByID(context.Context, string) (domain.MonitoringBatch, error)
	UpdateBatch(context.Context, domain.MonitoringBatch, int64) error
	InsertEvent(context.Context, domain.SeismicEvent) error
	EventByID(context.Context, string) (domain.SeismicEvent, error)
	UpdateEvent(context.Context, domain.SeismicEvent, int64) error
	ListEvents(context.Context, EventFilter) (EventPage, error)
	InsertPermit(context.Context, domain.StimulationPermit) error
	PermitByID(context.Context, string) (domain.StimulationPermit, error)
	UpdatePermit(context.Context, domain.StimulationPermit, int64) error
	InsertAudit(context.Context, domain.AuditEvent) error
	ListAudit(context.Context, int) ([]domain.AuditEvent, error)
	InsertJob(context.Context, domain.DeliveryJob) error
	ClaimJobs(context.Context, time.Time, int) ([]domain.DeliveryJob, error)
	UpdateJob(context.Context, domain.DeliveryJob, int64) error
	Summary(context.Context) (domain.PlatformSummary, error)
}
