package worker

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"log/slog"
	"time"
)

type Worker struct {
	Store     repository.Store
	Clock     clock.Clock
	Interval  time.Duration
	BatchSize int
	Logger    *slog.Logger
}

func (w Worker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := w.RunOnce(ctx); err != nil && w.Logger != nil {
				w.Logger.Error("worker cycle failed", "error", err)
			}
		}
	}
}
func (w Worker) RunOnce(ctx context.Context) error {
	jobs, e := w.Store.ClaimJobs(ctx, w.Clock.Now(), w.BatchSize)
	if e != nil {
		return e
	}
	for _, job := range jobs {
		next, e := job.Complete(w.Clock.Now())
		if e != nil {
			continue
		}
		if e = w.Store.UpdateJob(ctx, next, int64(job.Attempts)); e != nil {
			return e
		}
	}
	return nil
}
