package domain

import "time"

type Role string

const (
	RoleFieldEngineer  Role = "field_engineer"
	RoleGeophysicist   Role = "geophysicist"
	RoleSafetyReviewer Role = "safety_reviewer"
	RoleAuditor        Role = "auditor"
)

type UserStatus string

const (
	UserActive    UserStatus = "active"
	UserSuspended UserStatus = "suspended"
)

type User struct {
	ID, Email, DisplayName, PasswordHash string
	Role                                 Role
	Status                               UserStatus
	Version                              int64
	CreatedAt, UpdatedAt                 time.Time
}
type Session struct {
	ID, UserID string
	ExpiresAt  time.Time
	RevokedAt  *time.Time
	CreatedAt  time.Time
}

func (u User) Can(role Role) bool           { return u.Role == role || u.Role == RoleAuditor }
func (s Session) Active(now time.Time) bool { return s.RevokedAt == nil && s.ExpiresAt.After(now) }
func (u User) Validate() error {
	if u.ID == "" || u.Email == "" || u.PasswordHash == "" {
		return FieldError{"user", "id, email and password are required"}
	}
	if u.Status == "" {
		return FieldError{"status", "is required"}
	}
	return nil
}
