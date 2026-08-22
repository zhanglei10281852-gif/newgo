package sqlite

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
)

func (s *Store) Summary(ctx context.Context) (domain.PlatformSummary, error) {
	var out domain.PlatformSummary
	queries := []struct {
		target *int
		sql    string
	}{
		{&out.Fields, `SELECT COUNT(*) FROM fields`}, {&out.Wells, `SELECT COUNT(*) FROM wells`},
		{&out.ActiveBatches, `SELECT COUNT(*) FROM monitoring_batches WHERE state IN ('open','collecting')`},
		{&out.OpenEvents, `SELECT COUNT(*) FROM seismic_events WHERE status <> 'classified'`},
		{&out.PendingPermits, `SELECT COUNT(*) FROM stimulation_permits WHERE state = 'pending_review'`},
		{&out.FailedJobs, `SELECT COUNT(*) FROM delivery_jobs WHERE state IN ('failed','dead')`},
	}
	for _, query := range queries {
		if err := s.q.QueryRowContext(ctx, query.sql).Scan(query.target); err != nil {
			return out, err
		}
	}
	return out, nil
}
