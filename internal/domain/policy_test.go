package domain

import (
	"testing"
	"time"
)

func TestPermitRuleEvaluation(t *testing.T) {
	rule := DefaultPermitRule()
	cases := []struct {
		name     string
		envelope RiskEnvelope
		stations int
		allowed  bool
	}{
		{"green", RiskEnvelope{Band: RiskGreen, PeakMagnitude: 1, Unclassified: 0}, 2, true},
		{"one station", RiskEnvelope{Band: RiskGreen, PeakMagnitude: 1}, 1, false},
		{"high peak", RiskEnvelope{Band: RiskGreen, PeakMagnitude: 2}, 2, false},
		{"pending", RiskEnvelope{Band: RiskAmber, PeakMagnitude: 1, Unclassified: 1}, 2, false},
		{"red", RiskEnvelope{Band: RiskRed, PeakMagnitude: 1}, 2, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EvaluatePermit(rule, tc.envelope, tc.stations)
			if got.Allowed != tc.allowed {
				t.Fatalf("decision=%+v", got)
			}
			if tc.allowed && got.Message() != "permit eligible" {
				t.Fatal(got.Message())
			}
		})
	}
}
func TestClassificationRules(t *testing.T) {
	rules := DefaultClassificationRules()
	tests := []struct {
		value float64
		label string
	}{{0.1, "background"}, {0.8, "induced"}, {1.9, "induced"}, {2, "tectonic"}}
	for _, tc := range tests {
		rule, e := ClassifyMagnitude(tc.value, rules)
		if e != nil || rule.Label != tc.label {
			t.Fatalf("value %.2f got=%+v err=%v", tc.value, rule, e)
		}
	}
	if _, e := ClassifyMagnitude(11, rules); e == nil {
		t.Fatal("out of range classified")
	}
}
func TestAlertsSeverity(t *testing.T) {
	alerts := BuildAlerts(RiskEnvelope{Band: RiskBlack, PeakMagnitude: 3, Unclassified: 2}, 2)
	if len(alerts) != 3 || HighestSeverity(alerts) != "critical" {
		t.Fatalf("alerts=%+v", alerts)
	}
}
func TestChecklist(t *testing.T) {
	c := NewChecklist([]string{"station", "events", "review"})
	if c.Ready() {
		t.Fatal("new checklist ready")
	}
	if e := c.Complete("unknown"); e == nil {
		t.Fatal("unknown checklist accepted")
	}
	for _, item := range c.Items {
		if e := c.Complete(item); e != nil {
			t.Fatal(e)
		}
	}
	if !c.Ready() || len(c.Missing()) != 0 {
		t.Fatal("checklist not complete")
	}
}
func TestAuditQuery(t *testing.T) {
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	query := AuditQuery{ActorID: "u", Action: "classify", Since: &at, Limit: 10}
	if e := query.Validate(); e != nil {
		t.Fatal(e)
	}
	event := AuditEvent{ActorID: "u", Action: "classify", EntityType: "event", CreatedAt: at.Add(time.Hour)}
	if !query.Matches(event) {
		t.Fatal("matching audit rejected")
	}
	event.ActorID = "other"
	if query.Matches(event) {
		t.Fatal("foreign audit matched")
	}
}
func TestBatchCloseReport(t *testing.T) {
	b := MonitoringBatch{ID: "b", State: BatchCollecting}
	events := []SeismicEvent{{Status: EventClassified}, {Status: EventEscalated}}
	report := BuildBatchCloseReport(b, events)
	if !report.Closed || report.Classified != 1 || report.Escalated != 1 {
		t.Fatalf("report=%+v", report)
	}
	report = BuildBatchCloseReport(b, []SeismicEvent{{Status: EventUnclassified}})
	if report.Closed || len(report.Reasons) == 0 {
		t.Fatal("unclassified batch closed")
	}
}
func TestNormalization(t *testing.T) {
	if NormalizeReference("  run-7 ") != "RUN-7" {
		t.Fatal("reference")
	}
	if NormalizeSerial(" SN 77 ") != "SN-77" {
		t.Fatal("serial")
	}
	if len(SafeNote("x")) != 1 {
		t.Fatal("note")
	}
}
func TestOperatorAndExpiryGuards(t *testing.T) {
	if e := RequireDistinct("u", "u"); e == nil {
		t.Fatal("same operator")
	}
	now := time.Now()
	if e := RequireFuture(now.Add(time.Minute), now); e != nil {
		t.Fatal(e)
	}
	if e := RequireFuture(now.Add(-time.Minute), now); e == nil {
		t.Fatal("past time accepted")
	}
}
func TestRiskComparison(t *testing.T) {
	if CompareRisk(RiskEnvelope{Band: RiskGreen}, RiskEnvelope{Band: RiskRed}) != 1 {
		t.Fatal("risk did not increase")
	}
	if CompareRisk(RiskEnvelope{Band: RiskBlack}, RiskEnvelope{Band: RiskAmber}) != -1 {
		t.Fatal("risk did not decrease")
	}
	if CompareRisk(RiskEnvelope{Band: RiskAmber}, RiskEnvelope{Band: RiskAmber}) != 0 {
		t.Fatal("equal risk differs")
	}
}
