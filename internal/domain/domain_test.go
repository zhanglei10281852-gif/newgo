package domain

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func testTime() time.Time { return time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC) }
func TestFieldLifecycle(t *testing.T) {
	f := Field{ID: "f", Code: "GEO-1", Name: "Field", Timezone: "Asia/Shanghai", State: FieldDraft, RiskThreshold: 2}
	next, e := f.Activate(testTime())
	if e != nil || next.State != FieldActive {
		t.Fatalf("activate=%+v %v", next, e)
	}
	next, e = next.Archive(testTime())
	if e != nil || next.State != FieldArchived {
		t.Fatalf("archive=%+v %v", next, e)
	}
	if _, e = next.Activate(testTime()); e == nil {
		t.Fatal("archived field activated")
	}
}
func TestWellLifecycle(t *testing.T) {
	w := Well{ID: "w", FieldID: "f", APIIdentifier: "API", State: WellDrilled, MaxMagnitude: 2}
	next, e := w.BeginMonitoring(testTime())
	if e != nil || next.State != WellMonitored {
		t.Fatal(e)
	}
	next, e = next.Suspend(testTime())
	if e != nil || next.State != WellSuspended {
		t.Fatal(e)
	}
	if _, e = next.BeginMonitoring(testTime()); e == nil {
		t.Fatal("suspended well resumed")
	}
}
func TestStationCalibration(t *testing.T) {
	s := Station{ID: "s", WellID: "w", Serial: "SN", State: StationCalibrating}
	due := testTime().Add(time.Hour)
	next, e := s.Calibrate(testTime(), due)
	if e != nil || !next.Ready(testTime()) {
		t.Fatalf("station=%+v err=%v", next, e)
	}
	if next.Ready(due) {
		t.Fatal("station remained ready after due")
	}
}
func TestBatchRequiresEventsToClose(t *testing.T) {
	b := MonitoringBatch{ID: "b", WellID: "w", Reference: "r", State: BatchCollecting, StartsAt: testTime(), EndsAt: testTime().Add(time.Hour)}
	if _, e := b.Close(testTime()); e == nil {
		t.Fatal("empty batch closed")
	}
	b.EventCount = 1
	next, e := b.Close(testTime())
	if e != nil || next.State != BatchClosed {
		t.Fatal(e)
	}
}
func TestEventClassificationIsOneWay(t *testing.T) {
	e := SeismicEvent{ID: "e", BatchID: "b", WellID: "w", Status: EventUnclassified}
	next, err := e.Classify("tectonic", "note", testTime())
	if err != nil || next.Status != EventClassified {
		t.Fatal(err)
	}
	if _, err = next.Classify("induced", "", testTime()); err == nil {
		t.Fatal("classified event changed")
	}
}
func TestPermitNeedsIndependentReviewer(t *testing.T) {
	p := StimulationPermit{ID: "p", WellID: "w", BatchID: "b", RequesterID: "u", ReviewerID: "u", EquipmentSlot: "slot", State: PermitDraft}
	if _, e := p.Submit(testTime()); e == nil {
		t.Fatal("self approval allowed")
	}
	p.ReviewerID = "reviewer"
	next, e := p.Submit(testTime())
	if e != nil || next.State != PermitPending {
		t.Fatal(e)
	}
}
func TestPermitDecisionAndExecution(t *testing.T) {
	expires := testTime().Add(time.Hour)
	p := StimulationPermit{ID: "p", WellID: "w", BatchID: "b", RequesterID: "u", ReviewerID: "r", EquipmentSlot: "slot", State: PermitPending, ExpiresAt: &expires}
	next, e := p.Decide(true, testTime())
	if e != nil || next.State != PermitApproved {
		t.Fatal(e)
	}
	next, e = next.Execute(testTime())
	if e != nil || next.State != PermitExecuted {
		t.Fatal(e)
	}
}
func TestJobRetryBackoff(t *testing.T) {
	now := testTime()
	j := DeliveryJob{ID: "j", State: JobPending, Attempts: 0, MaxAttempts: 3}
	next, e := j.Claim(now)
	if e != nil || next.State != JobRunning || next.Attempts != 1 {
		t.Fatal(e)
	}
	next, e = next.Retry(errors.New("temporary"), now)
	if e != nil || next.State != JobFailed || next.AvailableAt.Sub(now) != time.Second {
		t.Fatal(e)
	}
}
func TestRiskEnvelopeBands(t *testing.T) {
	start, end := testTime(), testTime().Add(time.Hour)
	green, e := BuildRiskEnvelope(nil, start, end)
	if e != nil || green.Band != RiskGreen {
		t.Fatal(e)
	}
	events := []RiskObservation{{EventID: "e", OccurredAt: start.Add(time.Minute), Magnitude: 2.5, DepthKm: 3, Classified: true}}
	red, e := BuildRiskEnvelope(events, start, end)
	if e != nil || red.Band != RiskRed {
		t.Fatalf("%+v %v", red, e)
	}
	if red.AllowsPermit(3) {
		t.Fatal("red risk allowed")
	}
}
func TestRiskIgnoresOutsideWindow(t *testing.T) {
	start, end := testTime(), testTime().Add(time.Hour)
	events := []RiskObservation{{OccurredAt: start.Add(-time.Hour), Magnitude: 4}}
	envelope, e := BuildRiskEnvelope(events, start, end)
	if e != nil || envelope.EventCount != 0 {
		t.Fatalf("%+v %v", envelope, e)
	}
}
func TestWindowSplitAndOverlap(t *testing.T) {
	start, end := testTime(), testTime().Add(3*time.Hour)
	w, e := NewWindow(start, end, time.FixedZone("field", 8*3600))
	if e != nil {
		t.Fatal(e)
	}
	parts := w.Split(time.Hour)
	if len(parts) != 3 || !parts[0].Contains(start) {
		t.Fatalf("parts=%d", len(parts))
	}
	other, _ := NewWindow(start.Add(2*time.Hour), end.Add(time.Hour), w.Location)
	if !w.Overlaps(other) {
		t.Fatal("windows should overlap")
	}
}
func TestCapacityReserveRelease(t *testing.T) {
	s := EquipmentSlot{ID: "s", Label: "slot", Capacity: 3}
	next, e := s.Reserve(2)
	if e != nil || next.Available() != 1 {
		t.Fatal(e)
	}
	if _, e = next.Reserve(2); e == nil {
		t.Fatal("oversubscription")
	}
	next, e = next.Release(1)
	if e != nil || next.Reserved != 1 {
		t.Fatal(e)
	}
}
func TestPaginationClamp(t *testing.T) {
	page, size := ClampPage(-2, 999)
	if page != 1 || size != 500 {
		t.Fatalf("%d %d", page, size)
	}
	c := NewCursor(0, 20)
	p, pages := c.Page(41)
	if p != 1 || pages != 3 {
		t.Fatalf("%d %d", p, pages)
	}
}
func TestCSVImportValidation(t *testing.T) {
	rows := []EventImportRow{{ExternalID: "a", Magnitude: 1, DepthKm: 2}, {ExternalID: "a", Magnitude: 1, DepthKm: 2}, {ExternalID: "", Magnitude: -1}}
	result := ValidateImport(rows)
	if result.Accepted != 1 || result.Rejected != 2 || len(result.Errors) != 2 {
		t.Fatalf("%+v", result)
	}
}
func TestRoles(t *testing.T) {
	if !Allowed(RoleSafetyReviewer, ActionReviewPermit) {
		t.Fatal("review role denied")
	}
	if Allowed(RoleFieldEngineer, ActionReviewPermit) {
		t.Fatal("field engineer reviewed")
	}
	if !Allowed(RoleAuditor, ActionReadAudit) {
		t.Fatal("auditor denied")
	}
}
func TestErrorChains(t *testing.T) {
	if !strings.Contains((ConflictError{"permit", "occupied"}).Error(), "occupied") {
		t.Fatal("missing reason")
	}
	if !IsNotFound(ErrNotFound) {
		t.Fatal("not found lost")
	}
}
