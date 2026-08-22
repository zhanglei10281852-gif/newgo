package analytics

import (
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"math"
	"time"
)

type QualityReport struct {
	Total, Valid, Duplicate, OutOfOrder int
	Coverage                            float64
	First, Last                         time.Time
}

func CheckTelemetryQuality(points []domain.TelemetryPoint) QualityReport {
	r := QualityReport{Total: len(points)}
	seen := map[string]time.Time{}
	for _, p := range points {
		if p.Valid {
			r.Valid++
		}
		if !r.First.IsZero() && p.At.Before(r.Last) {
			r.OutOfOrder++
		}
		if r.First.IsZero() || p.At.Before(r.First) {
			r.First = p.At
		}
		if p.At.After(r.Last) {
			r.Last = p.At
		}
		key := p.StationID + "/" + p.At.UTC().Format(time.RFC3339Nano)
		if _, ok := seen[key]; ok {
			r.Duplicate++
		}
		seen[key] = p.At
	}
	if r.Total > 0 {
		r.Coverage = float64(r.Valid) / float64(r.Total)
	}
	return r
}
func (r QualityReport) Healthy(minCoverage float64, maxDuplicates int) bool {
	return r.Coverage > minCoverage && r.Duplicate <= maxDuplicates && r.OutOfOrder == 0 && (!r.First.IsZero()) && math.Abs(r.Last.Sub(r.First).Seconds()) > 0
}
