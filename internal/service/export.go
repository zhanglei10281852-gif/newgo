package service

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"strconv"
	"time"
)

func ExportEvents(events []domain.SeismicEvent) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)
	if err := writer.Write([]string{"id", "batch_id", "well_id", "occurred_at", "magnitude", "depth_km", "status", "classification"}); err != nil {
		return nil, err
	}
	for _, event := range events {
		if err := writer.Write([]string{event.ID, event.BatchID, event.WellID, event.OccurredAt.Format(time.RFC3339), strconv.FormatFloat(event.Magnitude, 'f', 3, 64), strconv.FormatFloat(event.DepthKm, 'f', 3, 64), string(event.Status), event.Classification}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
func ExportRiskReport(envelope domain.RiskEnvelope) string {
	return fmt.Sprintf("window=%s..%s band=%s peak=%.3f depth_spread=%.3f density=%.3f classified=%d pending=%d", envelope.WindowStart.Format(time.RFC3339), envelope.WindowEnd.Format(time.RFC3339), envelope.Band, envelope.PeakMagnitude, envelope.DepthSpread, envelope.EventDensity, envelope.EventCount-envelope.Unclassified, envelope.Unclassified)
}
