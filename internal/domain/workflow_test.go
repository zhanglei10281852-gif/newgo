package domain

import (
	"testing"
	"time"
)

func operation(id, well, slot string, start time.Time) ScheduledOperation {
	window, _ := NewWindow(start, start.Add(time.Hour), time.UTC)
	return ScheduledOperation{ID: id, FieldID: "f", WellID: well, EquipmentSlot: slot, OwnerID: "u", Kind: OperationStimulation, Window: window, State: "scheduled", Version: 1}
}
func TestScheduleConflicts(t *testing.T) {
	now := testTime()
	a := operation("a", "w1", "slot", now)
	b := operation("b", "w2", "slot", now.Add(30*time.Minute))
	conflicts := DetectScheduleConflicts(b, []ScheduledOperation{a})
	if len(conflicts) != 1 || conflicts[0].ID != "a" {
		t.Fatalf("conflicts=%+v", conflicts)
	}
	if e := ValidateNoConflicts([]ScheduledOperation{a, b}); e == nil {
		t.Fatal("conflicting schedule accepted")
	}
}
func TestFindAvailableWindow(t *testing.T) {
	now := testTime()
	day, _ := NewWindow(now, now.Add(8*time.Hour), time.UTC)
	busy := operation("busy", "w", "slot", now.Add(2*time.Hour))
	window, e := FindAvailableWindow(day, time.Hour, []ScheduledOperation{busy})
	if e != nil {
		t.Fatal(e)
	}
	if !window.Start.Equal(now) {
		t.Fatalf("start=%v", window.Start)
	}
}
func TestOperationLifecycle(t *testing.T) {
	now := testTime()
	op := operation("a", "w", "slot", now)
	running, e := op.Start(now.Add(time.Minute))
	if e != nil || running.State != "running" {
		t.Fatal(e)
	}
	done, e := running.Complete(now.Add(time.Hour))
	if e != nil || done.State != "completed" {
		t.Fatal(e)
	}
	if _, e = done.Cancel(); e == nil {
		t.Fatal("completed operation cancelled")
	}
}
func TestIdempotencyRoundTrip(t *testing.T) {
	now := testTime()
	hash, e := IdempotencyHash(map[string]any{"well": "w", "value": 2})
	if e != nil {
		t.Fatal(e)
	}
	record, e := NewIdempotencyRecord("tenant", "key", "post", "/permit", hash, 201, []byte(`{"id":"p"}`), now, now.Add(time.Hour))
	if e != nil {
		t.Fatal(e)
	}
	if e = record.Match("tenant", "key", "POST", "/permit", hash, now); e != nil {
		t.Fatal(e)
	}
	replay := record.Replay()
	replay[0] = 'x'
	if string(record.Response) == string(replay) {
		t.Fatal("replay shared backing array")
	}
}
func TestIdempotencyRejectsReuse(t *testing.T) {
	now := testTime()
	record, _ := NewIdempotencyRecord("t", "k", "POST", "/one", "hash", 200, []byte("ok"), now, now.Add(time.Hour))
	if e := record.Match("t", "k", "POST", "/two", "hash", now); e == nil {
		t.Fatal("path reuse accepted")
	}
	if e := record.Match("t", "k", "POST", "/one", "other", now); e == nil {
		t.Fatal("payload reuse accepted")
	}
	if e := record.Match("t", "k", "POST", "/one", "hash", now.Add(2*time.Hour)); e != ErrExpired {
		t.Fatalf("expired=%v", e)
	}
}
func TestNotifications(t *testing.T) {
	alerts := []Alert{{Code: "HIGH", Severity: "critical", Message: "stop"}, {Code: "CHECK", Severity: "warning", Message: "review"}}
	items := BuildRiskNotifications(alerts, map[string][]NotificationChannel{"ops": {ChannelPager, ChannelEmail}}, testTime())
	if len(items) != 4 || items[0].Priority != 10 {
		t.Fatalf("items=%+v", items)
	}
	dupes := append(items, items[0])
	if len(DeduplicateNotifications(dupes)) != 4 {
		t.Fatal("dedup failed")
	}
	if len(PartitionNotifications(items, 3)) != 2 {
		t.Fatal("partition failed")
	}
}
func TestIncidentLifecycle(t *testing.T) {
	now := testTime()
	incident := RiskIncident{ID: "i", FieldID: "f", WellID: "w", Title: "risk", State: IncidentOpen, Severity: 4, OpenedAt: now, UpdatedAt: now, Version: 1}
	assigned, e := incident.Assign("owner", now)
	if e != nil {
		t.Fatal(e)
	}
	assigned, e = assigned.AddEvent("event-1", now)
	if e != nil {
		t.Fatal(e)
	}
	mitigated, e := assigned.Mitigate("pressure reduced", now)
	if e != nil || mitigated.State != IncidentMitigated {
		t.Fatal(e)
	}
	closed, e := mitigated.Close(now)
	if e != nil || closed.ClosedAt == nil {
		t.Fatal(e)
	}
}
func TestIncidentFilter(t *testing.T) {
	now := testTime()
	items := []RiskIncident{{ID: "a", FieldID: "f", WellID: "w", State: IncidentOpen, Severity: 2, OpenedAt: now}, {ID: "b", FieldID: "f", WellID: "w", State: IncidentOpen, Severity: 5, OpenedAt: now.Add(time.Hour)}, {ID: "c", FieldID: "other", State: IncidentClosed, Severity: 5}}
	out := FilterIncidents(items, IncidentFilter{FieldID: "f", MinimumSeverity: 3})
	if len(out) != 1 || out[0].ID != "b" {
		t.Fatalf("out=%+v", out)
	}
}
func TestCalibrationReport(t *testing.T) {
	now := testTime()
	samples := []CalibrationSample{{StationID: "s", Expected: 1, Observed: 1.01}, {StationID: "s", Expected: 2, Observed: 2.02}, {StationID: "s", Expected: 3, Observed: 3.01}}
	report, e := BuildCalibrationReport("s", samples, .1, now, 24*time.Hour)
	if e != nil || !report.Passed || !CalibrationValid(report, now) {
		t.Fatalf("report=%+v err=%v", report, e)
	}
	latest, e := LatestCalibration("s", []CalibrationReport{report})
	if e != nil || latest.StationID != "s" {
		t.Fatal(e)
	}
}
func TestCalibrationRequiresSamples(t *testing.T) {
	if _, e := BuildCalibrationReport("s", []CalibrationSample{{StationID: "s"}}, .1, testTime(), time.Hour); e == nil {
		t.Fatal("insufficient calibration accepted")
	}
}
func TestEvidenceFingerprintStable(t *testing.T) {
	a := EvidenceItem{Kind: "events", Identifier: "a", Checksum: Checksum([]byte("a")), Size: 1}
	b := EvidenceItem{Kind: "report", Identifier: "b", Checksum: Checksum([]byte("b")), Size: 2}
	first, e := EvidenceFingerprint([]EvidenceItem{a, b})
	if e != nil {
		t.Fatal(e)
	}
	second, e := EvidenceFingerprint([]EvidenceItem{b, a})
	if e != nil || first != second {
		t.Fatalf("first=%s second=%s err=%v", first, second, e)
	}
	missing, ok := EvidenceComplete([]string{"events", "report", "permit"}, []EvidenceItem{a, b})
	if ok || len(missing) != 1 {
		t.Fatalf("missing=%v", missing)
	}
}
