package service

import (
	"context"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"sort"
	"time"
)

type ReportService struct {
	Store repository.Store
	Clock func() time.Time
}
type WellReport struct {
	Well         domain.Well
	Envelope     domain.RiskEnvelope
	RecentEvents []domain.SeismicEvent
	OpenPermits  int
}

func (r ReportService) BuildWellReport(ctx context.Context, wellID string, start, end time.Time) (WellReport, error) {
	well, e := r.Store.WellByID(ctx, wellID)
	if e != nil {
		return WellReport{}, e
	}
	events, e := r.Store.ListEvents(ctx, repository.EventFilter{Page: repository.PageRequest{Page: 1, Size: 500}})
	if e != nil {
		return WellReport{}, e
	}
	observations := make([]domain.RiskObservation, 0)
	recent := make([]domain.SeismicEvent, 0)
	for _, event := range events.Items {
		if event.WellID != wellID {
			continue
		}
		recent = append(recent, event)
		observations = append(observations, domain.RiskObservation{EventID: event.ID, OccurredAt: event.OccurredAt, Magnitude: event.Magnitude, DepthKm: event.DepthKm, Classified: event.Status == domain.EventClassified})
	}
	sort.Slice(recent, func(i, j int) bool { return recent[i].OccurredAt.After(recent[j].OccurredAt) })
	if len(recent) > 20 {
		recent = recent[:20]
	}
	envelope, e := domain.BuildRiskEnvelope(observations, start, end)
	if e != nil {
		return WellReport{}, e
	}
	return WellReport{Well: well, Envelope: envelope, RecentEvents: recent}, nil
}
func (r ReportService) ExplainPermit(envelope domain.RiskEnvelope, threshold float64) string {
	if envelope.AllowsPermit(threshold) {
		return "permit may be reviewed"
	}
	if envelope.Unclassified > 0 {
		return fmt.Sprintf("%d events still need classification", envelope.Unclassified)
	}
	return fmt.Sprintf("risk envelope %s exceeds threshold", envelope.Band)
}
