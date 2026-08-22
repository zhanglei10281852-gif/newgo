package service

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/storage/sqlite"
	"testing"
	"time"
)

func TestWorkflowCreatesAndClassifiesEvent(t *testing.T) {
	store, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	now := time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC)
	ops := Operations{Store: store, Clock: clock.Fixed{Value: now}}
	field, e := ops.CreateField(context.Background(), string(domain.RoleFieldEngineer), domain.Field{Code: "G", Name: "Field", Timezone: "UTC", RiskThreshold: 2})
	if e != nil {
		t.Fatal(e)
	}
	well, e := ops.CreateWell(context.Background(), string(domain.RoleFieldEngineer), domain.Well{FieldID: field.ID, Name: "Well", APIIdentifier: "API", MaxMagnitude: 2})
	if e != nil {
		t.Fatal(e)
	}
	batch := Workflow{Store: store, Clock: func() time.Time { return now }}
	b, e := batch.OpenBatch(context.Background(), domain.MonitoringBatch{WellID: well.ID, Reference: "R", StartsAt: now, EndsAt: now.Add(time.Hour)})
	if e != nil {
		t.Fatal(e)
	}
	event, e := batch.RecordEvent(context.Background(), domain.SeismicEvent{BatchID: b.ID, WellID: well.ID, OccurredAt: now.Add(time.Minute), Magnitude: 1, DepthKm: 2})
	if e != nil {
		t.Fatal(e)
	}
	event, e = batch.ClassifyEvent(context.Background(), event.ID, "induced", "")
	if e != nil || event.Status != domain.EventClassified {
		t.Fatalf("event=%+v err=%v", event, e)
	}
}
func TestWorkflowRejectsPermitForUnmonitoredWell(t *testing.T) {
	store, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	defer store.Close()
	now := time.Now().UTC()
	ops := Operations{Store: store, Clock: clock.Fixed{Value: now}}
	field, _ := ops.CreateField(context.Background(), "field_engineer", domain.Field{Code: "G", Name: "Field", Timezone: "UTC", RiskThreshold: 2})
	well, _ := ops.CreateWell(context.Background(), "field_engineer", domain.Well{FieldID: field.ID, Name: "Well", APIIdentifier: "API", MaxMagnitude: 2})
	workflow := Workflow{Store: store, Clock: func() time.Time { return now }, PermitTTL: time.Hour}
	if _, e := workflow.SubmitPermit(context.Background(), domain.StimulationPermit{WellID: well.ID, BatchID: "b", RequesterID: "u", ReviewerID: "r", EquipmentSlot: "slot"}); e == nil {
		t.Fatal("permit accepted for drilled well")
	}
}
