package sqlite

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/repository"
	"path/filepath"
	"testing"
	"time"
)

func openTest(t *testing.T) *Store {
	t.Helper()
	s, e := Open(context.Background(), filepath.Join(t.TempDir(), "geo.db"))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func timestamp() time.Time { return time.Date(2026, 8, 22, 10, 0, 0, 0, time.UTC) }
func seedField(t *testing.T, s *Store) domain.Field {
	t.Helper()
	f := domain.Field{ID: "field-1", Code: "GEO-1", Name: "North", Timezone: "Asia/Shanghai", State: domain.FieldActive, RiskThreshold: 2, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if e := s.InsertField(context.Background(), f); e != nil {
		t.Fatal(e)
	}
	return f
}
func seedWell(t *testing.T, s *Store) domain.Well {
	t.Helper()
	seedField(t, s)
	w := domain.Well{ID: "well-1", FieldID: "field-1", Name: "Well A", APIIdentifier: "API-1", State: domain.WellMonitored, MaxMagnitude: 2, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if e := s.InsertWell(context.Background(), w); e != nil {
		t.Fatal(e)
	}
	return w
}
func TestMigrationCreatesRelations(t *testing.T) {
	s := openTest(t)
	f := seedField(t, s)
	got, e := s.FieldByID(context.Background(), f.ID)
	if e != nil || got.Code != f.Code {
		t.Fatalf("got=%+v err=%v", got, e)
	}
	if _, e = s.FieldByID(context.Background(), "missing"); e != domain.ErrNotFound {
		t.Fatalf("missing=%v", e)
	}
}
func TestOptimisticFieldUpdateRejectsStaleVersion(t *testing.T) {
	s := openTest(t)
	f := seedField(t, s)
	f.Name = "Updated"
	if e := s.UpdateField(context.Background(), f, 99); e == nil {
		t.Fatal("stale update succeeded")
	}
	if e := s.UpdateField(context.Background(), f, 1); e != nil {
		t.Fatal(e)
	}
	got, _ := s.FieldByID(context.Background(), f.ID)
	if got.Name != "Updated" || got.Version != 2 {
		t.Fatalf("got=%+v", got)
	}
}
func TestForeignKeyRejectsUnknownWell(t *testing.T) {
	s := openTest(t)
	w := domain.Well{ID: "w", FieldID: "missing", Name: "x", APIIdentifier: "api", State: domain.WellDrilled, MaxMagnitude: 1, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if e := s.InsertWell(context.Background(), w); e == nil {
		t.Fatal("orphan well inserted")
	}
}
func TestBatchEventsAndPagination(t *testing.T) {
	s := openTest(t)
	seedWell(t, s)
	b := domain.MonitoringBatch{ID: "batch-1", WellID: "well-1", Reference: "RUN-1", State: domain.BatchCollecting, StartsAt: timestamp(), EndsAt: timestamp().Add(time.Hour), Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if e := s.InsertBatch(context.Background(), b); e != nil {
		t.Fatal(e)
	}
	for i := 0; i < 3; i++ {
		e := domain.SeismicEvent{ID: string(rune('a' + i)), BatchID: b.ID, WellID: "well-1", OccurredAt: timestamp().Add(time.Duration(i) * time.Minute), Magnitude: 1, DepthKm: 2, Status: domain.EventUnclassified, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
		if err := s.InsertEvent(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}
	page, e := s.ListEvents(context.Background(), repository.EventFilter{BatchID: b.ID, Page: repository.PageRequest{Page: 1, Size: 2}})
	if e != nil || page.Total != 3 || len(page.Items) != 2 {
		t.Fatalf("page=%+v err=%v", page, e)
	}
	got, _ := s.BatchByID(context.Background(), b.ID)
	if got.EventCount != 3 {
		t.Fatalf("batch=%+v", got)
	}
}
func TestEventUpdateAndListFilter(t *testing.T) {
	s := openTest(t)
	seedWell(t, s)
	b := domain.MonitoringBatch{ID: "b", WellID: "well-1", Reference: "R", State: domain.BatchCollecting, StartsAt: timestamp(), EndsAt: timestamp().Add(time.Hour), Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if e := s.InsertBatch(context.Background(), b); e != nil {
		t.Fatal(e)
	}
	e := domain.SeismicEvent{ID: "e", BatchID: "b", WellID: "well-1", OccurredAt: timestamp(), Magnitude: 1, DepthKm: 2, Status: domain.EventUnclassified, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if err := s.InsertEvent(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	e.Status = domain.EventClassified
	e.Classification = "induced"
	if err := s.UpdateEvent(context.Background(), e, 1); err != nil {
		t.Fatal(err)
	}
	page, err := s.ListEvents(context.Background(), repository.EventFilter{Status: "classified", Page: repository.PageRequest{Page: 1, Size: 5}})
	if err != nil || page.Total != 1 || page.Items[0].Classification != "induced" {
		t.Fatalf("page=%+v err=%v", page, err)
	}
}
func TestRestartKeepsRows(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persistent.db")
	s, e := Open(context.Background(), path)
	if e != nil {
		t.Fatal(e)
	}
	f := domain.Field{ID: "f", Code: "F", Name: "Field", Timezone: "UTC", State: domain.FieldDraft, RiskThreshold: 1, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
	if e = s.InsertField(context.Background(), f); e != nil {
		t.Fatal(e)
	}
	s.Close()
	s, e = Open(context.Background(), path)
	if e != nil {
		t.Fatal(e)
	}
	defer s.Close()
	got, e := s.FieldByID(context.Background(), "f")
	if e != nil || got.Code != "F" {
		t.Fatalf("got=%+v err=%v", got, e)
	}
}
func TestTransactionRollback(t *testing.T) {
	s := openTest(t)
	e := s.WithTx(context.Background(), func(tx repository.Store) error {
		f := domain.Field{ID: "tx-f", Code: "TX", Name: "Tx", Timezone: "UTC", State: domain.FieldDraft, RiskThreshold: 1, Version: 1, CreatedAt: timestamp(), UpdatedAt: timestamp()}
		if err := tx.InsertField(context.Background(), f); err != nil {
			return err
		}
		return domain.ErrConflict
	})
	if e == nil {
		t.Fatal("rollback callback succeeded")
	}
	if _, e = s.FieldByID(context.Background(), "tx-f"); e != domain.ErrNotFound {
		t.Fatalf("row after rollback=%v", e)
	}
}
