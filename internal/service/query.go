package service

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"sort"
	"time"
)

type QueryService struct {
	Store repository.Store
	Clock func() time.Time
}
type EventSearch struct {
	BatchID, Status string
	From, Until     *time.Time
	Page, Size      int
	Ascending       bool
}

func (q QueryService) SearchEvents(ctx context.Context, search EventSearch) (repository.EventPage, error) {
	page, e := FilterPage(ctx, q.Store, search.BatchID, search.Status, search.Page, search.Size)
	if e != nil {
		return page, e
	}
	if search.From == nil && search.Until == nil {
		return page, nil
	}
	filtered := page.Items[:0]
	for _, event := range page.Items {
		if search.From != nil && event.OccurredAt.Before(*search.From) {
			continue
		}
		if search.Until != nil && !event.OccurredAt.Before(*search.Until) {
			continue
		}
		filtered = append(filtered, event)
	}
	page.Items = filtered
	return page, nil
}
func (q QueryService) RecentEvents(ctx context.Context, wellID string, limit int) ([]domain.SeismicEvent, error) {
	page, e := q.Store.ListEvents(ctx, repository.EventFilter{Page: repository.PageRequest{Page: 1, Size: 500}})
	if e != nil {
		return nil, e
	}
	events := make([]domain.SeismicEvent, 0)
	for _, event := range page.Items {
		if event.WellID == wellID {
			events = append(events, event)
		}
	}
	sort.Slice(events, func(i, j int) bool { return events[i].OccurredAt.After(events[j].OccurredAt) })
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}
func (q QueryService) OpenEventCount(ctx context.Context, batchID string) (int, error) {
	page, e := q.Store.ListEvents(ctx, repository.EventFilter{BatchID: batchID, Status: "unclassified", Page: repository.PageRequest{Page: 1, Size: 1}})
	return page.Total, e
}
func (q QueryService) RiskSnapshot(ctx context.Context, wellID string, start, end time.Time) (domain.RiskEnvelope, error) {
	events, e := q.RecentEvents(ctx, wellID, 500)
	if e != nil {
		return domain.RiskEnvelope{}, e
	}
	observations := make([]domain.RiskObservation, 0, len(events))
	for _, event := range events {
		observations = append(observations, domain.RiskObservation{EventID: event.ID, Magnitude: event.Magnitude, DepthKm: event.DepthKm, OccurredAt: event.OccurredAt, Classified: event.Status == domain.EventClassified})
	}
	return domain.BuildRiskEnvelope(observations, start, end)
}
