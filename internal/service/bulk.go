package service

import (
	"context"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
)

type BulkService struct {
	Store    repository.Store
	Workflow Workflow
}
type BulkEventResult struct {
	ExternalID, EventID string
	Accepted            bool
	Error               string
}

func (b BulkService) Record(ctx context.Context, wellID, batchID string, rows []domain.EventImportRow) []BulkEventResult {
	out := make([]BulkEventResult, 0, len(rows))
	for _, row := range rows {
		result := BulkEventResult{ExternalID: row.ExternalID}
		event := domain.SeismicEvent{ID: identity.New("event"), BatchID: batchID, WellID: wellID, Magnitude: row.Magnitude, DepthKm: row.DepthKm}
		if value, e := b.Workflow.RecordEvent(ctx, event); e != nil {
			result.Error = e.Error()
		} else {
			result.Accepted = true
			result.EventID = value.ID
		}
		out = append(out, result)
	}
	return out
}
func SummarizeBulk(results []BulkEventResult) string {
	accepted := 0
	for _, result := range results {
		if result.Accepted {
			accepted++
		}
	}
	return fmt.Sprintf("accepted=%d rejected=%d", accepted, len(results)-accepted)
}
func ValidateBatchInput(batch domain.MonitoringBatch) error {
	if err := batch.Validate(); err != nil {
		return err
	}
	if batch.EventCount < 0 {
		return domain.FieldError{Field: "event_count", Message: "cannot be negative"}
	}
	return nil
}

var _ repository.Store
