package analytics

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/zhanglei10281852-gif/newgo/internal/domain"
)

func TestFitPressureForecast(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	points := []domain.PressureReading{{At: base, KiloPascal: 10}, {At: base.Add(time.Minute), KiloPascal: 11}, {At: base.Add(2 * time.Minute), KiloPascal: 12}}
	f, err := FitPressure(points, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if f.Samples != 3 || f.PredictedPeak < 12.9 || f.Slope <= 0 {
		t.Fatalf("unexpected forecast %#v", f)
	}
	if ShouldPause(f, 12.5, .5) != true {
		t.Fatal("expected pause")
	}
}
func TestFitPressureRejectsInvalidInputs(t *testing.T) {
	base := time.Now()
	cases := [][]domain.PressureReading{nil, {{At: base, KiloPascal: 1}}, {{At: base, KiloPascal: 1}, {At: base, KiloPascal: 2}}}
	for _, c := range cases {
		if _, err := FitPressure(c, time.Minute); err == nil {
			t.Fatal("expected error")
		}
	}
	if _, err := FitPressure(nil, 0); err == nil {
		t.Fatal("expected horizon error")
	}
}
func TestClusterEvents(t *testing.T) {
	b := time.Now()
	e := []domain.SeismicEvent{{ID: "a", OccurredAt: b, DepthKm: 1, Magnitude: 1}, {ID: "b", OccurredAt: b.Add(time.Minute), DepthKm: 1.2, Magnitude: 2}, {ID: "c", OccurredAt: b.Add(20 * time.Minute), DepthKm: 1, Magnitude: 3}}
	c := ClusterEvents(e, 5*time.Minute, .5)
	if len(c) != 2 || len(c[0].EventIDs) != 2 || c[0].PeakMagnitude != 2 {
		t.Fatalf("clusters %#v", c)
	}
}
func TestRetentionPlan(t *testing.T) {
	now := time.Now()
	events := []domain.AuditEvent{{ID: "keep", CreatedAt: now.Add(-time.Hour)}, {ID: "arch", CreatedAt: now.Add(-48 * time.Hour)}, {ID: "purge", CreatedAt: now.Add(-200 * time.Hour)}}
	r := PlanAuditRetention(events, now, 24*time.Hour, 72*time.Hour, map[string]bool{"purge": true})
	if len(r.Keep) != 2 || len(r.Archive) != 1 || len(r.Purge) != 0 {
		t.Fatalf("retention %#v", r)
	}
}

type senderStub struct {
	calls int
	err   error
}

func (s *senderStub) Send(context.Context, Delivery) error { s.calls++; return s.err }
func TestDispatcherRetriesAndDefers(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s := &senderStub{err: errors.New("offline")}
	d := Dispatcher{Sender: s, Clock: func() time.Time { return now }, MaxAttempts: 3}
	failed, err := d.Run(context.Background(), []Delivery{{ID: "a"}, {ID: "b", Next: now.Add(time.Hour)}})
	if err != nil || len(failed) != 1 || failed[0].Attempts != 1 || s.calls != 1 {
		t.Fatalf("failed=%#v calls=%d", failed, s.calls)
	}
}
func TestRiskScoreBands(t *testing.T) {
	cases := []struct {
		peak float64
		want domain.RiskBand
	}{{0.2, domain.RiskGreen}, {1.2, domain.RiskAmber}, {2.2, domain.RiskAmber}, {3.2, domain.RiskRed}}
	for _, tc := range cases {
		got := ScoreRisk(RiskInput{Envelope: domain.RiskEnvelope{PeakMagnitude: tc.peak}})
		if got.Band != tc.want {
			t.Fatalf("peak %v got %s", tc.peak, got.Band)
		}
	}
	if got := ScoreRisk(RiskInput{Envelope: domain.RiskEnvelope{PeakMagnitude: 2.2, EventDensity: 10}}); got.Band != domain.RiskRed {
		t.Fatalf("expected red, got %s", got.Band)
	}
	if got := ScoreRisk(RiskInput{Envelope: domain.RiskEnvelope{PeakMagnitude: 3.2, EventDensity: 30}}); got.Band != domain.RiskBlack {
		t.Fatalf("expected black, got %s", got.Band)
	}
}
func TestRiskScoreCombinesSignals(t *testing.T) {
	s := ScoreRisk(RiskInput{Envelope: domain.RiskEnvelope{PeakMagnitude: 2.8, EventDensity: 30, Unclassified: 1}, Pressure: domain.PressureEnvelope{Violations: 4, RampRate: 8}, PermitAge: 48 * time.Hour})
	if s.Score < 75 || !IsActionable(s, 75) || len(s.Reasons) < 4 {
		t.Fatalf("score %#v", s)
	}
	m := MergeScores([]RiskScore{{Score: 10, Reasons: []string{"a"}}, {Score: 80, Reasons: []string{"a", "b"}}})
	if m.Score != 80 || len(m.Reasons) != 2 {
		t.Fatalf("merge %#v", m)
	}
}
func TestWindowSummary(t *testing.T) {
	b := time.Now().Truncate(time.Hour)
	e := []domain.SeismicEvent{{ID: "1", OccurredAt: b.Add(10 * time.Minute), Magnitude: 2}, {ID: "2", OccurredAt: b.Add(70 * time.Minute), Magnitude: 1}}
	p := []domain.TelemetryPoint{{StationID: "s1", At: b.Add(10 * time.Minute), Valid: true}, {StationID: "s2", At: b.Add(70 * time.Minute), Valid: false}}
	w := SummarizeWindows(e, p, b, b.Add(2*time.Hour), time.Hour)
	if len(w) != 2 || w[0].Events != 1 || w[1].Complete {
		t.Fatalf("windows %#v", w)
	}
}
func TestWindowSummaryBoundaryAndInvalid(t *testing.T) {
	b := time.Now()
	if SummarizeWindows(nil, nil, b, b, time.Hour) != nil {
		t.Fatal("expected nil")
	}
	w := SummarizeWindows([]domain.SeismicEvent{{OccurredAt: b.Add(time.Hour)}}, nil, b, b.Add(2*time.Hour), time.Hour)
	if len(w) != 2 || w[0].Events != 0 || w[1].Events != 1 {
		t.Fatal("boundary")
	}
}

func TestEscalationRulesAndTimeline(t *testing.T) {
	now := time.Now()
	score := RiskScore{Score: 80, Band: domain.RiskRed}
	rules := []EscalationRule{{Name: "amber", MinimumBand: domain.RiskAmber, Cooldown: time.Hour}, {Name: "red", MinimumBand: domain.RiskRed, Cooldown: time.Hour}}
	got := EvaluateEscalations(score, rules, map[string]time.Time{"red": now.Add(-time.Minute)}, now)
	if len(got) != 1 || got[0].Rule != "amber" {
		t.Fatalf("escalations %#v", got)
	}
	events := []domain.SeismicEvent{{ID: "e", OccurredAt: now, Classification: "noise"}}
	permits := []domain.StimulationPermit{{ID: "p", CreatedAt: now.Add(time.Minute)}}
	timeline := BuildTimeline(events, permits)
	if len(timeline) != 2 || timeline[0].Kind != "seismic" {
		t.Fatalf("timeline %#v", timeline)
	}
}

func TestQualityReport(t *testing.T) {
	b := time.Now()
	points := []domain.TelemetryPoint{{StationID: "a", At: b, Valid: true}, {StationID: "a", At: b, Valid: true}, {StationID: "b", At: b.Add(time.Minute), Valid: false}}
	r := CheckTelemetryQuality(points)
	if r.Total != 3 || r.Duplicate != 1 || r.OutOfOrder != 0 || r.Coverage != 2.0/3.0 {
		t.Fatalf("quality %#v", r)
	}
	if r.Healthy(.5, 1) != true {
		t.Fatal("expected healthy")
	}
}

func TestDailyReportsAndFilter(t *testing.T) {
	b := time.Date(2026, 2, 1, 23, 30, 0, 0, time.UTC)
	events := []domain.SeismicEvent{{ID: "a", OccurredAt: b, Magnitude: 1}, {ID: "b", OccurredAt: b.Add(2 * time.Hour), Magnitude: 2, Status: domain.EventEscalated}}
	windows := []WindowSummary{{Start: b, Peak: 2, Complete: true}}
	r := BuildDailyReports(events, windows, time.UTC)
	if len(r) != 2 || r[0].Events != 1 || r[1].Peak != 2 {
		t.Fatalf("reports %#v", r)
	}
	if len(FilterReports(r, "review")) != 1 || len(FilterReports(r, "")) != 2 {
		t.Fatal("filter")
	}
	data, err := EncodeDailyReports(r)
	if err != nil || len(data) == 0 {
		t.Fatal("encode")
	}
}

func TestHealthSnapshot(t *testing.T) {
	b := time.Now()
	q := CheckTelemetryQuality([]domain.TelemetryPoint{{StationID: "a", At: b, Valid: true}, {StationID: "b", At: b.Add(time.Minute), Valid: true}})
	s := HealthSnapshot{Risk: RiskScore{Band: domain.RiskGreen}, Quality: q, GeneratedAt: b}
	if !s.Healthy() {
		t.Fatal("expected healthy snapshot")
	}
	if s.Labels()["risk"] != "green" {
		t.Fatal("labels")
	}
}

func TestDispatcherMaxAttempts(t *testing.T) {
	s := &senderStub{err: errors.New("down")}
	d := Dispatcher{Sender: s, Clock: time.Now, MaxAttempts: 2}
	jobs := []Delivery{{ID: "x", Attempts: 1}}
	failed, err := d.Run(context.Background(), jobs)
	if err != nil || len(failed) != 1 || failed[0].Attempts != 2 {
		t.Fatalf("%#v", failed)
	}
}

func TestMergeScoresEmptyAndPauseThreshold(t *testing.T) {
	if got := MergeScores(nil); got.Band != domain.RiskGreen {
		t.Fatal("empty")
	}
	f := Forecast{PredictedPeak: 10, Confidence: .4}
	if ShouldPause(f, 10, .5) {
		t.Fatal("low confidence")
	}
	if !ShouldPause(f, 9, .4) {
		t.Fatal("threshold")
	}
}

func TestSLAQueueOrderingAndCounts(t *testing.T) {
	now := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)
	soon := now.Add(10 * time.Minute)
	old := now.Add(-2 * time.Hour)
	permits := []domain.StimulationPermit{{ID: "p1", RequesterID: "u", CreatedAt: old}, {ID: "p2", RequesterID: "u", CreatedAt: soon}}
	cals := []domain.CalibrationReport{{StationID: "s1", DueAt: now.Add(5 * time.Minute)}}
	items := BuildSLAQueue(now, permits, cals, time.Hour)
	if len(items) != 3 {
		t.Fatalf("items %#v", items)
	}
	if items[0].Status != SLAOverdue {
		t.Fatal("overdue first")
	}
	counts := CountSLA(items)
	if counts[SLAOverdue] != 1 {
		t.Fatal("counts")
	}
	next, ok := NextSLA(items)
	if !ok || next.ID != "p1" {
		t.Fatalf("next %#v", next)
	}
}

func TestSLAQueueEmpty(t *testing.T) {
	if len(BuildSLAQueue(time.Now(), nil, nil, time.Hour)) != 0 {
		t.Fatal("expected empty")
	}
	if _, ok := NextSLA(nil); ok {
		t.Fatal("expected no next")
	}
}

func TestDailyReportsTimezone(t *testing.T) {
	b := time.Date(2026, 1, 1, 23, 30, 0, 0, time.UTC)
	reports := BuildDailyReports([]domain.SeismicEvent{{ID: "x", OccurredAt: b}}, nil, time.FixedZone("+2", 2*3600))
	if len(reports) != 1 || reports[0].Day != "2026-01-02" {
		t.Fatalf("timezone %#v", reports)
	}
}

func TestEscalationCooldownAndRepeat(t *testing.T) {
	now := time.Now()
	score := RiskScore{Band: domain.RiskBlack, Score: 90}
	rules := []EscalationRule{{Name: "r", MinimumBand: domain.RiskRed, Cooldown: time.Hour, Repeat: false}, {Name: "repeat", MinimumBand: domain.RiskRed, Cooldown: time.Hour, Repeat: true}}
	got := EvaluateEscalations(score, rules, map[string]time.Time{"r": now.Add(-time.Minute), "repeat": now.Add(-time.Minute)}, now)
	if len(got) != 1 || got[0].Rule != "repeat" {
		t.Fatalf("cooldown %#v", got)
	}
}

func TestReportTotalsMergeAndNarrative(t *testing.T) {
	a := []DailyReport{{Day: "2026-01-01", Events: 2, Peak: 2, RedWindows: 1, Notes: []string{"alpha"}}}
	b := []DailyReport{{Day: "2026-01-01", Events: 3, Peak: 3, Notes: []string{"beta"}}, {Day: "2026-01-02", Events: 1, Peak: 1}}
	m := MergeDailyReports(a, b)
	if len(m) != 2 || m[0].Events != 5 || m[0].Peak != 3 {
		t.Fatalf("merge %#v", m)
	}
	events, red, avg := Totals(m)
	if events != 6 || red != 1 || avg != 2 {
		t.Fatalf("totals %d %d %v", events, red, avg)
	}
	if RiskNarrative(RiskScore{}) != "risk is within the normal operating envelope" {
		t.Fatal("empty narrative")
	}
	if RiskNarrative(RiskScore{Reasons: []string{"z", "a"}}) != "z; a" {
		t.Fatal("narrative")
	}
}

func TestRiskActionBoundaries(t *testing.T) {
	for _, tc := range []struct {
		score float64
		want  bool
	}{{0, false}, {24.9, false}, {25, true}, {99, true}} {
		s := RiskScore{Score: tc.score, Band: domain.RiskAmber}
		if IsActionable(s, 25) != tc.want {
			t.Fatalf("score %v", tc.score)
		}
	}
}

func TestClusterDepthAndTimeBoundaries(t *testing.T) {
	b := time.Now()
	events := []domain.SeismicEvent{{ID: "1", OccurredAt: b, DepthKm: 1}, {ID: "2", OccurredAt: b.Add(5 * time.Minute), DepthKm: 1.5}, {ID: "3", OccurredAt: b.Add(5*time.Minute + time.Nanosecond), DepthKm: 1.5}}
	c := ClusterEvents(events, 5*time.Minute, .49)
	if len(c) != 2 {
		t.Fatalf("clusters %#v", c)
	}
}

func TestRetentionOrderingAndHolds(t *testing.T) {
	now := time.Now()
	events := []domain.AuditEvent{{ID: "z", CreatedAt: now.Add(-100 * time.Hour)}, {ID: "a", CreatedAt: now.Add(-50 * time.Hour)}, {ID: "h", CreatedAt: now.Add(-1000 * time.Hour)}}
	r := PlanAuditRetention(events, now, 24*time.Hour, 48*time.Hour, map[string]bool{"h": true})
	if len(r.Keep) != 1 || r.Keep[0].ID != "h" || len(r.Purge) != 1 || len(r.Archive) != 1 {
		t.Fatalf("retention %#v", r)
	}
}

func TestQualityOutOfOrder(t *testing.T) {
	b := time.Now()
	r := CheckTelemetryQuality([]domain.TelemetryPoint{{StationID: "a", At: b.Add(time.Minute), Valid: true}, {StationID: "a", At: b, Valid: true}})
	if r.OutOfOrder != 1 || r.Healthy(0, 0) {
		t.Fatal("out of order")
	}
}

func TestSeriesStatisticsAndTransforms(t *testing.T) {
	b := time.Now()
	samples := []Sample{{At: b.Add(2 * time.Minute), Value: 3}, {At: b, Value: 1}, {At: b.Add(time.Minute), Value: 2}, {At: b.Add(3 * time.Minute), Value: 4}}
	s := Stats(samples)
	if s.Count != 4 || s.Mean != 2.5 || s.Median != 2.5 || !s.Increasing {
		t.Fatalf("stats %#v", s)
	}
	ma := MovingAverage(samples, 2)
	if ma[1].Value != 2 {
		t.Fatal("moving average")
	}
	ex := Exceedances(samples, 3)
	if len(ex) != 2 {
		t.Fatal("exceed")
	}
	n := Normalize(samples)
	if n[0].Value < 0 || n[0].Value > 1 {
		t.Fatal("normalize")
	}
}

func TestSeriesResamplingAndInvalidValues(t *testing.T) {
	b := time.Now()
	samples := []Sample{{At: b, Value: 0}, {At: b.Add(2 * time.Minute), Value: 10}}
	r := Resample(samples, b, b.Add(3*time.Minute), time.Minute)
	if len(r) != 3 || r[1].Value != 5 {
		t.Fatalf("resample %#v", r)
	}
	if Resample(nil, b, b, time.Minute) != nil {
		t.Fatal("invalid range")
	}
	if Stats([]Sample{{Value: math.NaN()}}).Count != 0 {
		t.Fatal("nan")
	}
}

func TestSeriesConstantAndEmpty(t *testing.T) {
	if len(MovingAverage(nil, 2)) != 0 || len(Normalize(nil)) != 0 {
		t.Fatal("empty")
	}
	s := []Sample{{Value: 4}, {Value: 4}}
	n := Normalize(s)
	if n[0].Value != 4 {
		t.Fatal("constant normalization")
	}
}

func TestSeriesQuantizeMissingAndWithin(t *testing.T) {
	b := time.Now()
	samples := []Sample{{At: b, Value: 1.234}, {At: b.Add(3 * time.Minute), Value: 2.345}, {At: b.Add(5 * time.Minute), Value: 3.456}}
	q := Quantize(samples, 2)
	if q[0].Value != 1.23 || q[1].Value != 2.35 {
		t.Fatalf("quantize %#v", q)
	}
	m := Missing(samples, time.Minute)
	if len(m) != 3 {
		t.Fatalf("missing %#v", m)
	}
	w := Within(samples, b.Add(time.Minute), b.Add(5*time.Minute))
	if len(w) != 1 || w[0].Value != 2.345 {
		t.Fatalf("within %#v", w)
	}
}

func TestSeriesStatsSpreadAndSorting(t *testing.T) {
	b := time.Now()
	s := Stats([]Sample{{At: b.Add(time.Minute), Value: 4}, {At: b, Value: 1}, {At: b.Add(2 * time.Minute), Value: 2}})
	if s.Minimum != 1 || s.Maximum != 4 || s.StandardDeviation <= 0 {
		t.Fatal("spread")
	}
	if s.Increasing {
		t.Fatal("not increasing")
	}
}

func TestRiskNarrativeSortedByCaller(t *testing.T) {
	if RiskNarrative(RiskScore{Reasons: []string{"pressure", "seismic"}}) != "pressure; seismic" {
		t.Fatal("narrative order")
	}
}

func TestDispatcherContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	s := &senderStub{}
	d := Dispatcher{Sender: s, Clock: time.Now}
	if _, err := d.Run(ctx, []Delivery{{ID: "x"}}); err != nil {
		t.Fatal(err)
	}
}

func TestWindowNoEvents(t *testing.T) {
	b := time.Now()
	w := SummarizeWindows(nil, nil, b, b.Add(time.Hour), time.Minute)
	if len(w) != 60 {
		t.Fatalf("windows %d", len(w))
	}
}

func TestFilterCaseInsensitive(t *testing.T) {
	r := []DailyReport{{Day: "2026-01-01", Notes: []string{"Review Required"}}}
	if len(FilterReports(r, "review")) != 1 {
		t.Fatal("case insensitive")
	}
	if len(FilterReports(r, "none")) != 0 {
		t.Fatal("no match")
	}
}

func TestMergeDailyDuplicateNotes(t *testing.T) {
	r := MergeDailyReports([]DailyReport{{Day: "d", Notes: []string{"b"}}}, []DailyReport{{Day: "d", Notes: []string{"a"}}})
	if len(r) != 1 || len(r[0].Notes) != 2 {
		t.Fatal("notes")
	}
}

func TestSLAStatusDueSoon(t *testing.T) {
	now := time.Now()
	p := domain.StimulationPermit{ID: "p", CreatedAt: now.Add(-50 * time.Minute)}
	items := BuildSLAQueue(now, []domain.StimulationPermit{p}, nil, time.Hour)
	if items[0].Status != SLADueSoon {
		t.Fatalf("status %s", items[0].Status)
	}
}

func TestTelemetryWindowQualitySignals(t *testing.T) {
	b := time.Now()
	p := []domain.TelemetryPoint{{StationID: "a", At: b, Valid: true, Noise: .1}, {StationID: "b", At: b.Add(time.Minute), Valid: true, Noise: .2}, {StationID: "c", At: b.Add(2 * time.Minute), Valid: false, Noise: .3}}
	w := domain.BuildTelemetryWindow(p, b, b.Add(3*time.Minute))
	if w.Stations != 3 || w.ValidPoints != 2 || w.Coverage() <= .6 {
		t.Fatal("window")
	}
	if w.Stable() {
		t.Fatal("invalid point should be unstable")
	}
	low, high := w.AmplitudeRange()
	if low != 0 || high != 0 {
		t.Fatal("amplitude")
	}
}

func TestRiskEnvelopeComparison(t *testing.T) {
	a := domain.RiskEnvelope{Band: domain.RiskGreen}
	b := domain.RiskEnvelope{Band: domain.RiskRed}
	if domain.CompareRisk(a, b) != 1 || domain.CompareRisk(b, a) != -1 {
		t.Fatal("compare")
	}
	if !a.AllowsPermit(1) {
		t.Fatal("green permit")
	}
}

func TestTelemetryDeduplication(t *testing.T) {
	b := time.Now()
	p := []domain.TelemetryPoint{{StationID: "a", At: b, Noise: 1}, {StationID: "a", At: b, Noise: 2}, {StationID: "b", At: b, Noise: 3}}
	d := domain.DeduplicateTelemetry(p)
	if len(d) != 2 {
		t.Fatal("dedupe")
	}
}

func TestPressureEnvelopeStability(t *testing.T) {
	b := time.Now()
	r := []domain.PressureReading{{At: b, KiloPascal: 10}, {At: b.Add(time.Minute), KiloPascal: 10.2}, {At: b.Add(2 * time.Minute), KiloPascal: 10.1}}
	e, err := domain.BuildPressureEnvelope(r, b, b.Add(3*time.Minute), 9, 11)
	if err != nil || e.Samples != 3 {
		t.Fatal("pressure")
	}
	if e.Margin(9, 11) <= 0 {
		t.Fatal("margin")
	}
}

func TestCalibrationReportLifecycle(t *testing.T) {
	b := time.Now()
	samples := []domain.CalibrationSample{{StationID: "s", Expected: 1, Observed: 1, At: b}, {StationID: "s", Expected: 2, Observed: 2, At: b}, {StationID: "s", Expected: 3, Observed: 3, At: b}}
	r, err := domain.BuildCalibrationReport("s", samples, .1, b, time.Hour)
	if err != nil || !r.Passed || !domain.CalibrationValid(r, b.Add(30*time.Minute)) {
		t.Fatal("calibration")
	}
}

func TestRiskActionableGreen(t *testing.T) {
	if IsActionable(RiskScore{Score: 90, Band: domain.RiskGreen}, 50) {
		t.Fatal("green not actionable")
	}
}
