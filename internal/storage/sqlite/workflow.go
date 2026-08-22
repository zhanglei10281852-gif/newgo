package sqlite

import (
	"context"
	"database/sql"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

func (s *Store) InsertBatch(ctx context.Context, b domain.MonitoringBatch) error {
	if e := b.Validate(); e != nil {
		return e
	}
	_, e := s.q.ExecContext(ctx, `INSERT INTO monitoring_batches(id,well_id,reference,state,starts_at,ends_at,event_count,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, b.ID, b.WellID, b.Reference, b.State, formatTime(b.StartsAt), formatTime(b.EndsAt), b.EventCount, b.Version, formatTime(b.CreatedAt), formatTime(b.UpdatedAt))
	return constraint(e)
}
func scanBatch(row *sql.Row) (domain.MonitoringBatch, error) {
	var b domain.MonitoringBatch
	var state, starts, ends, created, updated string
	e := row.Scan(&b.ID, &b.WellID, &b.Reference, &state, &starts, &ends, &b.EventCount, &b.Version, &created, &updated)
	if e != nil {
		return b, notFound(e)
	}
	b.State = domain.BatchState(state)
	b.StartsAt, _ = parseTime(starts)
	b.EndsAt, _ = parseTime(ends)
	b.CreatedAt, _ = parseTime(created)
	b.UpdatedAt, _ = parseTime(updated)
	return b, nil
}
func (s *Store) BatchByID(ctx context.Context, id string) (domain.MonitoringBatch, error) {
	return scanBatch(s.q.QueryRowContext(ctx, `SELECT id,well_id,reference,state,starts_at,ends_at,event_count,version,created_at,updated_at FROM monitoring_batches WHERE id=?`, id))
}
func (s *Store) UpdateBatch(ctx context.Context, b domain.MonitoringBatch, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE monitoring_batches SET state=?,event_count=?,version=version+1,updated_at=? WHERE id=? AND version=?`, b.State, b.EventCount, formatTime(b.UpdatedAt), b.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) InsertEvent(ctx context.Context, v domain.SeismicEvent) error {
	if e := v.Validate(); e != nil {
		return e
	}
	_, e := s.q.ExecContext(ctx, `INSERT INTO seismic_events(id,batch_id,well_id,occurred_at,magnitude,depth_km,status,classification,notes,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, v.ID, v.BatchID, v.WellID, formatTime(v.OccurredAt), v.Magnitude, v.DepthKm, v.Status, v.Classification, v.Notes, v.Version, formatTime(v.CreatedAt), formatTime(v.UpdatedAt))
	if e != nil {
		return constraint(e)
	}
	_, e = s.q.ExecContext(ctx, `UPDATE monitoring_batches SET event_count=event_count+1,updated_at=? WHERE id=?`, formatTime(v.UpdatedAt), v.BatchID)
	return constraint(e)
}
func scanEvent(row *sql.Row) (domain.SeismicEvent, error) {
	var v domain.SeismicEvent
	var status, occurred, created, updated string
	e := row.Scan(&v.ID, &v.BatchID, &v.WellID, &occurred, &v.Magnitude, &v.DepthKm, &status, &v.Classification, &v.Notes, &v.Version, &created, &updated)
	if e != nil {
		return v, notFound(e)
	}
	v.Status = domain.EventStatus(status)
	v.OccurredAt, _ = parseTime(occurred)
	v.CreatedAt, _ = parseTime(created)
	v.UpdatedAt, _ = parseTime(updated)
	return v, nil
}
func (s *Store) EventByID(ctx context.Context, id string) (domain.SeismicEvent, error) {
	return scanEvent(s.q.QueryRowContext(ctx, `SELECT id,batch_id,well_id,occurred_at,magnitude,depth_km,status,classification,notes,version,created_at,updated_at FROM seismic_events WHERE id=?`, id))
}
func (s *Store) UpdateEvent(ctx context.Context, v domain.SeismicEvent, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE seismic_events SET status=?,classification=?,notes=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Status, v.Classification, v.Notes, formatTime(v.UpdatedAt), v.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s *Store) ListEvents(ctx context.Context, f repository.EventFilter) (repository.EventPage, error) {
	if f.Page.Size <= 0 {
		f.Page.Size = 50
	}
	if f.Page.Page <= 0 {
		f.Page.Page = 1
	}
	where := " WHERE 1=1"
	args := []any{}
	if f.BatchID != "" {
		where += " AND batch_id=?"
		args = append(args, f.BatchID)
	}
	if f.Status != "" {
		where += " AND status=?"
		args = append(args, f.Status)
	}
	var total int
	if e := s.q.QueryRowContext(ctx, `SELECT COUNT(*) FROM seismic_events`+where, args...).Scan(&total); e != nil {
		return repository.EventPage{}, e
	}
	args = append(args, f.Page.Size, (f.Page.Page-1)*f.Page.Size)
	rows, e := s.q.QueryContext(ctx, `SELECT id,batch_id,well_id,occurred_at,magnitude,depth_km,status,classification,notes,version,created_at,updated_at FROM seismic_events`+where+` ORDER BY occurred_at ASC LIMIT ? OFFSET ?`, args...)
	if e != nil {
		return repository.EventPage{}, e
	}
	defer rows.Close()
	out := repository.EventPage{Total: total, Items: make([]domain.SeismicEvent, 0)}
	for rows.Next() {
		var v domain.SeismicEvent
		var status, occurred, created, updated string
		if e := rows.Scan(&v.ID, &v.BatchID, &v.WellID, &occurred, &v.Magnitude, &v.DepthKm, &status, &v.Classification, &v.Notes, &v.Version, &created, &updated); e != nil {
			return out, e
		}
		v.Status = domain.EventStatus(status)
		v.OccurredAt, _ = parseTime(occurred)
		v.CreatedAt, _ = parseTime(created)
		v.UpdatedAt, _ = parseTime(updated)
		out.Items = append(out.Items, v)
	}
	return out, rows.Err()
}

func (s *Store) InsertPermit(ctx context.Context, p domain.StimulationPermit) error {
	if e := p.Validate(); e != nil {
		return e
	}
	_, e := s.q.ExecContext(ctx, `INSERT INTO stimulation_permits(id,well_id,batch_id,requester_id,reviewer_id,equipment_slot,state,risk_score,requested_at,expires_at,reviewed_at,executed_at,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, p.ID, p.WellID, p.BatchID, p.RequesterID, p.ReviewerID, p.EquipmentSlot, p.State, p.RiskScore, nullableTime(p.RequestedAt), formatTime(*p.ExpiresAt), nullableTime(p.ReviewedAt), nullableTime(p.ExecutedAt), p.Version, formatTime(p.CreatedAt), formatTime(p.UpdatedAt))
	return constraint(e)
}
func scanPermit(row *sql.Row) (domain.StimulationPermit, error) {
	var p domain.StimulationPermit
	var state, expires, created, updated string
	var requested, reviewed, executed *string
	e := row.Scan(&p.ID, &p.WellID, &p.BatchID, &p.RequesterID, &p.ReviewerID, &p.EquipmentSlot, &state, &p.RiskScore, &requested, &expires, &reviewed, &executed, &p.Version, &created, &updated)
	if e != nil {
		return p, notFound(e)
	}
	p.State = domain.PermitState(state)
	expiresAt, _ := parseTime(expires)
	p.ExpiresAt = &expiresAt
	p.RequestedAt, _ = parseNullable(requested)
	p.ReviewedAt, _ = parseNullable(reviewed)
	p.ExecutedAt, _ = parseNullable(executed)
	p.CreatedAt, _ = parseTime(created)
	p.UpdatedAt, _ = parseTime(updated)
	return p, nil
}
func (s *Store) PermitByID(ctx context.Context, id string) (domain.StimulationPermit, error) {
	return scanPermit(s.q.QueryRowContext(ctx, `SELECT id,well_id,batch_id,requester_id,reviewer_id,equipment_slot,state,risk_score,requested_at,expires_at,reviewed_at,executed_at,version,created_at,updated_at FROM stimulation_permits WHERE id=?`, id))
}
func (s *Store) UpdatePermit(ctx context.Context, p domain.StimulationPermit, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE stimulation_permits SET state=?,risk_score=?,requested_at=?,reviewed_at=?,executed_at=?,version=version+1,updated_at=? WHERE id=? AND version=?`, p.State, p.RiskScore, nullableTime(p.RequestedAt), nullableTime(p.ReviewedAt), nullableTime(p.ExecutedAt), formatTime(p.UpdatedAt), p.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) InsertAudit(ctx context.Context, v domain.AuditEvent) error {
	_, e := s.q.ExecContext(ctx, `INSERT INTO audit_events(id,actor_id,action,entity_type,entity_id,outcome,request_id,metadata,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.ActorID, v.Action, v.EntityType, v.EntityID, v.Outcome, v.RequestID, v.Metadata, formatTime(v.CreatedAt))
	return constraint(e)
}
func (s *Store) ListAudit(ctx context.Context, limit int) ([]domain.AuditEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, e := s.q.QueryContext(ctx, `SELECT id,actor_id,action,entity_type,entity_id,outcome,request_id,metadata,created_at FROM audit_events ORDER BY created_at DESC LIMIT ?`, limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]domain.AuditEvent, 0)
	for rows.Next() {
		var v domain.AuditEvent
		var created string
		if e := rows.Scan(&v.ID, &v.ActorID, &v.Action, &v.EntityType, &v.EntityID, &v.Outcome, &v.RequestID, &v.Metadata, &created); e != nil {
			return nil, e
		}
		v.CreatedAt, _ = parseTime(created)
		out = append(out, v)
	}
	return out, rows.Err()
}

func (s *Store) InsertJob(ctx context.Context, j domain.DeliveryJob) error {
	_, e := s.q.ExecContext(ctx, `INSERT INTO delivery_jobs(id,kind,entity_id,payload,state,attempts,max_attempts,available_at,locked_at,last_error,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, j.ID, j.Kind, j.EntityID, j.Payload, j.State, j.Attempts, j.MaxAttempts, formatTime(*j.AvailableAt), nullableTime(j.LockedAt), j.LastError, formatTime(j.CreatedAt), formatTime(j.UpdatedAt))
	return constraint(e)
}
func (s *Store) ClaimJobs(ctx context.Context, now time.Time, limit int) ([]domain.DeliveryJob, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, e := s.q.QueryContext(ctx, `SELECT id,kind,entity_id,payload,state,attempts,max_attempts,available_at,locked_at,last_error,created_at,updated_at FROM delivery_jobs WHERE state IN ('pending','failed') AND available_at<=? ORDER BY available_at LIMIT ?`, formatTime(now), limit)
	if e != nil {
		return nil, e
	}
	defer rows.Close()
	out := make([]domain.DeliveryJob, 0)
	for rows.Next() {
		var j domain.DeliveryJob
		var state, available, created, updated string
		var locked *string
		if e := rows.Scan(&j.ID, &j.Kind, &j.EntityID, &j.Payload, &state, &j.Attempts, &j.MaxAttempts, &available, &locked, &j.LastError, &created, &updated); e != nil {
			return nil, e
		}
		j.State = domain.JobState(state)
		availableAt, _ := parseTime(available)
		j.AvailableAt = &availableAt
		j.LockedAt, _ = parseNullable(locked)
		j.CreatedAt, _ = parseTime(created)
		j.UpdatedAt, _ = parseTime(updated)
		next, e := j.Claim(now)
		if e != nil {
			continue
		}
		if _, e = s.q.ExecContext(ctx, `UPDATE delivery_jobs SET state='running',attempts=?,locked_at=?,updated_at=? WHERE id=?`, next.Attempts, formatTime(now), formatTime(now), j.ID); e != nil {
			return nil, e
		}
		out = append(out, next)
	}
	return out, rows.Err()
}
func (s *Store) UpdateJob(ctx context.Context, j domain.DeliveryJob, version int64) error {
	_, e := s.q.ExecContext(ctx, `UPDATE delivery_jobs SET state=?,available_at=?,locked_at=?,last_error=?,updated_at=? WHERE id=? AND attempts=?`, j.State, nullableTime(j.AvailableAt), nullableTime(j.LockedAt), j.LastError, formatTime(j.UpdatedAt), j.ID, version)
	return constraint(e)
}
