package domain

import (
	"testing"
	"time"
)

func TestTenantLifecycleAndUsage(t *testing.T) {
	tenant := Tenant{ID: "t", Code: "T", Name: "Tenant", Timezone: "UTC", State: TenantTrial, WellLimit: 2, StationLimit: 4, MonthlyEventLimit: 10}
	if e := tenant.Validate(); e != nil {
		t.Fatal(e)
	}
	active, e := tenant.Activate(testTime())
	if e != nil || active.State != TenantActive {
		t.Fatal(e)
	}
	if e = active.CheckUsage(TenantUsage{TenantID: "t", Month: "2026-08", Wells: 3}); e == nil {
		t.Fatal("quota ignored")
	}
	merged := MergeUsage([]TenantUsage{{TenantID: "t", Month: "2026-08", Wells: 1}, {TenantID: "t", Month: "2026-08", Wells: 1}})
	if len(merged) != 1 || merged[0].Wells != 2 {
		t.Fatalf("merged=%+v", merged)
	}
}
func TestTenantClosure(t *testing.T) {
	tenant := Tenant{ID: "t", Code: "T", Name: "Tenant", Timezone: "UTC", State: TenantActive, WellLimit: 1, StationLimit: 1, MonthlyEventLimit: 1}
	if _, e := tenant.Close(testTime()); e == nil {
		t.Fatal("active tenant closed")
	}
	suspended, e := tenant.Suspend(testTime())
	if e != nil {
		t.Fatal(e)
	}
	closed, e := suspended.Close(testTime())
	if e != nil || closed.State != TenantClosed {
		t.Fatal(e)
	}
}
func TestRetentionSelection(t *testing.T) {
	now := testTime()
	policy := RetentionPolicy{Class: RetentionRawTelemetry, Duration: time.Hour, DeleteBatchSize: 1}
	objects := []RetentionObject{{ID: "old", Class: RetentionRawTelemetry, CreatedAt: now.Add(-2 * time.Hour)}, {ID: "hold", Class: RetentionRawTelemetry, CreatedAt: now.Add(-2 * time.Hour), LegalHold: true}}
	expired := SelectExpired(policy, objects, now)
	if len(expired) != 1 || expired[0].ID != "old" {
		t.Fatalf("expired=%+v", expired)
	}
	marked := MarkDeleted(objects, []string{"old"}, now)
	if marked[0].DeletedAt == nil {
		t.Fatal("not marked")
	}
}
func TestForecastAndPressure(t *testing.T) {
	now := testTime()
	samples := []ForecastSample{{At: now, EventRate: 1, PeakMagnitude: 1, InjectionPressure: 10}, {At: now.Add(time.Hour), EventRate: 2, PeakMagnitude: 1, InjectionPressure: 10}, {At: now.Add(2 * time.Hour), EventRate: 3, PeakMagnitude: 2, InjectionPressure: 10}}
	forecast, e := BuildForecast(samples, now, time.Hour)
	if e != nil || forecast.Samples != 3 {
		t.Fatal(e)
	}
	if ForecastTrend(samples) != 1 {
		t.Fatalf("trend=%v", ForecastTrend(samples))
	}
	readings := []PressureReading{{At: now, KiloPascal: 100}, {At: now.Add(time.Hour), KiloPascal: 101}, {At: now.Add(2 * time.Hour), KiloPascal: 100.5}}
	envelope, e := BuildPressureEnvelope(readings, now, now.Add(3*time.Hour), 90, 110)
	if e != nil || !envelope.Stable(2) {
		t.Fatalf("pressure=%+v err=%v", envelope, e)
	}
}
