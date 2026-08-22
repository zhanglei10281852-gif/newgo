package service

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"time"
)

type Services struct {
	Auth       Auth
	Operations Operations
}

func New(store repository.Store, c clock.Clock, sessionTTL, permitTTL time.Duration) *Services {
	return &Services{Auth: Auth{Store: store, Clock: c, SessionTTL: sessionTTL}, Operations: Operations{Store: store, Clock: c, PermitTTL: permitTTL}}
}
func (s *Services) Ready(ctx context.Context) error {
	_, e := s.Operations.Store.Summary(ctx)
	return e
}
