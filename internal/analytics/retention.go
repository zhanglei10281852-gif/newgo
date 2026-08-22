package analytics

import (
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"sort"
	"time"
)

type RetentionDecision struct {
	Keep    []domain.AuditEvent
	Archive []domain.AuditEvent
	Purge   []domain.AuditEvent
}

// PlanAuditRetention preserves legal holds, archives aged records, and purges expired archives.
func PlanAuditRetention(events []domain.AuditEvent, now time.Time, active, archive time.Duration, holds map[string]bool) RetentionDecision {
	r := RetentionDecision{}
	ordered := append([]domain.AuditEvent(nil), events...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].CreatedAt.Before(ordered[j].CreatedAt) })
	for _, e := range ordered {
		age := now.Sub(e.CreatedAt)
		if holds[e.ID] || age < active {
			r.Keep = append(r.Keep, e)
		} else if age < active+archive {
			r.Archive = append(r.Archive, e)
		} else {
			r.Purge = append(r.Purge, e)
		}
	}
	return r
}
