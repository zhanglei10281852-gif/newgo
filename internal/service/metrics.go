package service

import (
	"sync"
	"time"
)

type Metrics struct {
	mu        sync.RWMutex
	requests  map[string]int64
	durations map[string]time.Duration
}

func NewMetrics() *Metrics {
	return &Metrics{requests: map[string]int64{}, durations: map[string]time.Duration{}}
}
func (m *Metrics) Observe(name string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests[name]++
	m.durations[name] += duration
}
func (m *Metrics) Count(name string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.requests[name]
}
func (m *Metrics) Average(name string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := m.requests[name]
	if count == 0 {
		return 0
	}
	return m.durations[name] / time.Duration(count)
}
func (m *Metrics) Snapshot() map[string]struct {
	Count   int64
	Average time.Duration
} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := map[string]struct {
		Count   int64
		Average time.Duration
	}{}
	for name, count := range m.requests {
		out[name] = struct {
			Count   int64
			Average time.Duration
		}{count, m.durations[name] / time.Duration(count)}
	}
	return out
}
