package service

import (
	"testing"
	"time"
)

func TestMetricsConcurrentSafe(t *testing.T) {
	m := NewMetrics()
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() { m.Observe("request", time.Millisecond); done <- struct{}{} }()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if m.Count("request") != 10 || m.Average("request") != time.Millisecond {
		t.Fatalf("snapshot=%+v", m.Snapshot())
	}
}
