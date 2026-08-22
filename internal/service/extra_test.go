package service

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/storage/sqlite"
	"testing"
	"time"
)

func TestExportRiskReport(t *testing.T) {
	report := ExportRiskReport(domain.RiskEnvelope{Band: domain.RiskGreen, WindowStart: time.Unix(0, 0), WindowEnd: time.Unix(3600, 0), PeakMagnitude: 1, EventCount: 3})
	if report == "" || len(report) < 20 {
		t.Fatal(report)
	}
}
func TestMetricsSnapshot(t *testing.T) {
	m := NewMetrics()
	m.Observe("a", time.Second)
	m.Observe("a", 3*time.Second)
	snapshot := m.Snapshot()
	if snapshot["a"].Count != 2 || snapshot["a"].Average != 2*time.Second {
		t.Fatalf("snapshot=%+v", snapshot)
	}
}
func TestBulkValidation(t *testing.T) {
	if e := ValidateBatchInput(domain.MonitoringBatch{ID: "b", WellID: "w", Reference: "r", StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour), EventCount: -1}); e == nil {
		t.Fatal("negative count accepted")
	}
}
func TestQueryMissingBatch(t *testing.T) {
	store, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	page, e := FilterPage(context.Background(), store, "missing", "", 1, 10)
	if e != nil || page.Total != 0 {
		t.Fatalf("page=%+v err=%v", page, e)
	}
}
