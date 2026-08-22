package audit

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"sort"
)

type Query struct{ Store repository.Store }

func (q Query) List(ctx context.Context, filter domain.AuditQuery) ([]domain.AuditEvent, error) {
	if e := filter.Validate(); e != nil {
		return nil, e
	}
	events, e := q.Store.ListAudit(ctx, filter.Limit)
	if e != nil {
		return nil, e
	}
	out := events[:0]
	for _, event := range events {
		if filter.Matches(event) {
			out = append(out, event)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func GroupByAction(events []domain.AuditEvent) map[string]int {
	out := map[string]int{}
	for _, event := range events {
		out[event.Action]++
	}
	return out
}
func OutcomeCount(events []domain.AuditEvent, outcome string) int {
	count := 0
	for _, event := range events {
		if event.Outcome == outcome {
			count++
		}
	}
	return count
}
