package clock

import (
	"testing"
	"time"
)

func TestFixedClock(t *testing.T) {
	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := (Fixed{Value: want}).Now(); !got.Equal(want) {
		t.Fatal(got)
	}
}
func TestOffsetClock(t *testing.T) {
	base := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	got := (Offset{Base: Fixed{Value: base}, Delta: time.Hour}).Now()
	if !got.Equal(base.Add(time.Hour)) {
		t.Fatal(got)
	}
}
