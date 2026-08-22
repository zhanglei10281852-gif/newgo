package analytics

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Delivery struct {
	ID       string
	Attempts int
	Next     time.Time
	Payload  []byte
}
type Sender interface {
	Send(context.Context, Delivery) error
}
type Dispatcher struct {
	Sender      Sender
	Clock       func() time.Time
	MaxAttempts int
	mu          sync.Mutex
}

func (d *Dispatcher) Run(ctx context.Context, jobs []Delivery) ([]Delivery, error) {
	if d.Sender == nil {
		return nil, errors.New("sender is required")
	}
	if d.Clock == nil {
		d.Clock = time.Now
	}
	if d.MaxAttempts <= 0 {
		d.MaxAttempts = 3
	}
	failed := []Delivery{}
	for _, job := range jobs {
		if !job.Next.IsZero() && job.Next.After(d.Clock()) {
			continue
		}
		err := d.Sender.Send(ctx, job)
		if err == nil {
			continue
		}
		job.Attempts++
		if job.Attempts < d.MaxAttempts {
			job.Next = d.Clock().Add(time.Duration(job.Attempts) * time.Minute)
			failed = append(failed, job)
		} else {
			failed = append(failed, job)
		}
	}
	return failed, nil
}
