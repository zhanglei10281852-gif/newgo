package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type Auth struct {
	Store      repository.Store
	Clock      clock.Clock
	SessionTTL time.Duration
}

func (a Auth) Login(ctx context.Context, email, password string) (domain.Session, domain.User, error) {
	u, e := a.Store.UserByEmail(ctx, email)
	if e != nil {
		return domain.Session{}, u, e
	}
	if u.Status != domain.UserActive || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return domain.Session{}, u, fmt.Errorf("invalid credentials: %w", domain.ErrForbidden)
	}
	s := domain.Session{ID: identity.New("session"), UserID: u.ID, ExpiresAt: a.Clock.Now().Add(a.SessionTTL), CreatedAt: a.Clock.Now()}
	return s, u, a.Store.InsertSession(ctx, s)
}
func (a Auth) Authenticate(ctx context.Context, id string) (domain.User, error) {
	s, e := a.Store.SessionByID(ctx, id)
	if e != nil {
		return domain.User{}, e
	}
	if !s.Active(a.Clock.Now()) {
		return domain.User{}, fmt.Errorf("session inactive: %w", domain.ErrForbidden)
	}
	u, e := a.Store.UserByID(ctx, s.UserID)
	if e != nil {
		return u, e
	}
	return u, nil
}
func (a Auth) Logout(ctx context.Context, id string) error {
	return a.Store.RevokeSession(ctx, id, a.Clock.Now())
}
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password required")
	}
	b, e := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(b), e
}
