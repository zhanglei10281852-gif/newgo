package domain

import (
	"testing"
	"time"
)

func TestTelemetryWindow(t *testing.T) {
	start := testTime()
	points := []TelemetryPoint{{StationID: "a", At: start.Add(time.Minute), Amplitude: 1, Frequency: 2, Noise: .1, Valid: true}, {StationID: "b", At: start.Add(2 * time.Minute), Amplitude: 3, Frequency: 4, Noise: .2, Valid: true}, {StationID: "a", At: start.Add(3 * time.Minute), Amplitude: 2, Frequency: 3, Noise: .1, Valid: true}}
	window := BuildTelemetryWindow(points, start, start.Add(time.Hour))
	if !window.Stable() || window.Coverage() != 1 {
		t.Fatalf("window=%+v", window)
	}
	low, high := window.AmplitudeRange()
	if low != 1 || high != 3 || window.FrequencyMean() != 3 {
		t.Fatalf("range=%v/%v mean=%v", low, high, window.FrequencyMean())
	}
}
func TestTelemetryDeduplication(t *testing.T) {
	at := testTime()
	points := []TelemetryPoint{{StationID: "a", At: at, Amplitude: 1}, {StationID: "a", At: at, Amplitude: 2}, {StationID: "b", At: at, Amplitude: 3}}
	out := DeduplicateTelemetry(points)
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
}
func TestRockLayers(t *testing.T) {
	layers := []RockLayer{{DepthTop: 0, DepthBottom: 2, Formation: "cap", FractureRisk: .2}, {DepthTop: 2, DepthBottom: 4, Formation: "reservoir", FractureRisk: .8}}
	if !LayersCover(layers, 4) || RiskFromLayers(layers) != .8 {
		t.Fatal("layer coverage")
	}
	layer, e := ResolveLayer(layers, 3)
	if e != nil || layer.Formation != "reservoir" {
		t.Fatalf("layer=%+v err=%v", layer, e)
	}
}
func TestMaintenanceTicket(t *testing.T) {
	now := testTime()
	ticket := MaintenanceTicket{ID: "m", StationID: "s", State: "open", Priority: 2}
	next, e := ticket.Assign("operator", now)
	if e != nil || next.State != "assigned" {
		t.Fatal(e)
	}
	next, e = next.Complete("done", now)
	if e != nil || next.State != "completed" {
		t.Fatal(e)
	}
}
func TestMaintenanceOrdering(t *testing.T) {
	now := testTime()
	a, b := now.Add(time.Hour), now.Add(2*time.Hour)
	tickets := []MaintenanceTicket{{ID: "low", State: "open", Priority: 1, DueAt: &a}, {ID: "high", State: "open", Priority: 3, DueAt: &b}}
	out := OrderMaintenance(tickets)
	if out[0].ID != "high" {
		t.Fatalf("out=%+v", out)
	}
	expired := ExpiredTickets(tickets, now.Add(3*time.Hour))
	if len(expired) != 2 {
		t.Fatalf("expired=%d", len(expired))
	}
}
