package main

import (
	"context"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/config"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/service"
	"github.com/zhanglei10281852-gif/newgo/internal/storage/sqlite"
	"os"
	"time"
)

func main() {
	cfg, e := config.Load()
	if e != nil {
		panic(e)
	}
	s, e := sqlite.Open(context.Background(), cfg.DatabasePath)
	if e != nil {
		panic(e)
	}
	defer s.Close()
	hash, e := service.HashPassword(os.Getenv("BOOTSTRAP_PASSWORD"))
	if e != nil {
		panic(e)
	}
	now := clock.System{}.Now()
	u := domain.User{ID: identity.New("user"), Email: os.Getenv("BOOTSTRAP_EMAIL"), DisplayName: os.Getenv("BOOTSTRAP_DISPLAY_NAME"), PasswordHash: hash, Role: domain.RoleFieldEngineer, Status: domain.UserActive, Version: 1, CreatedAt: now, UpdatedAt: now}
	if e = s.InsertUser(context.Background(), u); e != nil {
		panic(e)
	}
	fmt.Printf("created %s at %s\n", u.Email, now.Format(time.RFC3339))
}
