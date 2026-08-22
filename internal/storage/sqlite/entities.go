package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"time"
)

func insertUser(q execer, ctx context.Context, u domain.User) error {
	if err := u.Validate(); err != nil {
		return err
	}
	_, err := q.ExecContext(ctx, `INSERT INTO users(id,email,display_name,password_hash,role,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, u.ID, u.Email, u.DisplayName, u.PasswordHash, u.Role, u.Status, u.Version, formatTime(u.CreatedAt), formatTime(u.UpdatedAt))
	return constraint(err)
}
func scanUser(row *sql.Row) (domain.User, error) {
	var u domain.User
	var role, status string
	var created, updated string
	err := row.Scan(&u.ID, &u.Email, &u.DisplayName, &u.PasswordHash, &role, &status, &u.Version, &created, &updated)
	if err != nil {
		return u, notFound(err)
	}
	u.Role = domain.Role(role)
	u.Status = domain.UserStatus(status)
	u.CreatedAt, _ = parseTime(created)
	u.UpdatedAt, _ = parseTime(updated)
	return u, nil
}
func (s *Store) InsertUser(ctx context.Context, u domain.User) error { return insertUser(s.q, ctx, u) }
func (s *Store) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	return scanUser(s.q.QueryRowContext(ctx, `SELECT id,email,display_name,password_hash,role,status,version,created_at,updated_at FROM users WHERE email=?`, email))
}
func (s *Store) UserByID(ctx context.Context, id string) (domain.User, error) {
	return scanUser(s.q.QueryRowContext(ctx, `SELECT id,email,display_name,password_hash,role,status,version,created_at,updated_at FROM users WHERE id=?`, id))
}
func (s *Store) UpdateUser(ctx context.Context, u domain.User, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE users SET display_name=?,role=?,status=?,version=version+1,updated_at=? WHERE id=? AND version=?`, u.DisplayName, u.Role, u.Status, formatTime(u.UpdatedAt), u.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}
func (s *Store) InsertSession(ctx context.Context, v domain.Session) error {
	_, e := s.q.ExecContext(ctx, `INSERT INTO sessions(id,user_id,expires_at,revoked_at,created_at) VALUES(?,?,?,?,?)`, v.ID, v.UserID, formatTime(v.ExpiresAt), nullableTime(v.RevokedAt), formatTime(v.CreatedAt))
	return constraint(e)
}
func (s *Store) SessionByID(ctx context.Context, id string) (domain.Session, error) {
	var v domain.Session
	var exp, created string
	var revoked *string
	e := s.q.QueryRowContext(ctx, `SELECT id,user_id,expires_at,revoked_at,created_at FROM sessions WHERE id=?`, id).Scan(&v.ID, &v.UserID, &exp, &revoked, &created)
	if e != nil {
		return v, notFound(e)
	}
	v.ExpiresAt, _ = parseTime(exp)
	v.CreatedAt, _ = parseTime(created)
	v.RevokedAt, _ = parseNullable(revoked)
	return v, nil
}
func (s *Store) RevokeSession(ctx context.Context, id string, now time.Time) error {
	r, e := s.q.ExecContext(ctx, `UPDATE sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL`, formatTime(now), id)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (s *Store) InsertField(ctx context.Context, v domain.Field) error {
	if e := v.Validate(); e != nil {
		return e
	}
	_, e := s.q.ExecContext(ctx, `INSERT INTO fields(id,code,name,timezone,state,risk_threshold,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.Code, v.Name, v.Timezone, v.State, v.RiskThreshold, v.Version, formatTime(v.CreatedAt), formatTime(v.UpdatedAt))
	return constraint(e)
}
func scanField(row *sql.Row) (domain.Field, error) {
	var v domain.Field
	var state, created, updated string
	e := row.Scan(&v.ID, &v.Code, &v.Name, &v.Timezone, &state, &v.RiskThreshold, &v.Version, &created, &updated)
	if e != nil {
		return v, notFound(e)
	}
	v.State = domain.FieldState(state)
	v.CreatedAt, _ = parseTime(created)
	v.UpdatedAt, _ = parseTime(updated)
	return v, nil
}
func (s *Store) FieldByID(ctx context.Context, id string) (domain.Field, error) {
	return scanField(s.q.QueryRowContext(ctx, `SELECT id,code,name,timezone,state,risk_threshold,version,created_at,updated_at FROM fields WHERE id=?`, id))
}
func (s *Store) UpdateField(ctx context.Context, v domain.Field, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE fields SET name=?,timezone=?,state=?,risk_threshold=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Name, v.Timezone, v.State, v.RiskThreshold, formatTime(v.UpdatedAt), v.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) InsertWell(ctx context.Context, v domain.Well) error {
	if e := v.Validate(); e != nil {
		return e
	}
	_, e := s.q.ExecContext(ctx, `INSERT INTO wells(id,field_id,name,api_identifier,state,max_magnitude,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.FieldID, v.Name, v.APIIdentifier, v.State, v.MaxMagnitude, v.Version, formatTime(v.CreatedAt), formatTime(v.UpdatedAt))
	return constraint(e)
}
func scanWell(row *sql.Row) (domain.Well, error) {
	var v domain.Well
	var state, created, updated string
	e := row.Scan(&v.ID, &v.FieldID, &v.Name, &v.APIIdentifier, &state, &v.MaxMagnitude, &v.Version, &created, &updated)
	if e != nil {
		return v, notFound(e)
	}
	v.State = domain.WellState(state)
	v.CreatedAt, _ = parseTime(created)
	v.UpdatedAt, _ = parseTime(updated)
	return v, nil
}
func (s *Store) WellByID(ctx context.Context, id string) (domain.Well, error) {
	return scanWell(s.q.QueryRowContext(ctx, `SELECT id,field_id,name,api_identifier,state,max_magnitude,version,created_at,updated_at FROM wells WHERE id=?`, id))
}
func (s *Store) UpdateWell(ctx context.Context, v domain.Well, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE wells SET name=?,state=?,max_magnitude=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Name, v.State, v.MaxMagnitude, formatTime(v.UpdatedAt), v.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

func (s *Store) InsertStation(ctx context.Context, v domain.Station) error {
	if e := v.Validate(); e != nil {
		return e
	}
	_, e := s.q.ExecContext(ctx, `INSERT INTO stations(id,well_id,serial,location,state,calibration_due_at,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?)`, v.ID, v.WellID, v.Serial, v.Location, v.State, formatTime(v.CalibrationDueAt), v.Version, formatTime(v.CreatedAt), formatTime(v.UpdatedAt))
	return constraint(e)
}
func scanStation(row *sql.Row) (domain.Station, error) {
	var v domain.Station
	var state, due, created, updated string
	e := row.Scan(&v.ID, &v.WellID, &v.Serial, &v.Location, &state, &due, &v.Version, &created, &updated)
	if e != nil {
		return v, notFound(e)
	}
	v.State = domain.StationState(state)
	v.CalibrationDueAt, _ = parseTime(due)
	v.CreatedAt, _ = parseTime(created)
	v.UpdatedAt, _ = parseTime(updated)
	return v, nil
}
func (s *Store) StationByID(ctx context.Context, id string) (domain.Station, error) {
	return scanStation(s.q.QueryRowContext(ctx, `SELECT id,well_id,serial,location,state,calibration_due_at,version,created_at,updated_at FROM stations WHERE id=?`, id))
}
func (s *Store) UpdateStation(ctx context.Context, v domain.Station, version int64) error {
	r, e := s.q.ExecContext(ctx, `UPDATE stations SET location=?,state=?,calibration_due_at=?,version=version+1,updated_at=? WHERE id=? AND version=?`, v.Location, v.State, formatTime(v.CalibrationDueAt), formatTime(v.UpdatedAt), v.ID, version)
	if e != nil {
		return constraint(e)
	}
	n, _ := r.RowsAffected()
	if n != 1 {
		return domain.ErrConflict
	}
	return nil
}

func unsupported(name string) error { return fmt.Errorf("%s is not implemented", name) }
