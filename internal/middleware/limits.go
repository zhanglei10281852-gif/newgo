package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Limiter struct {
	mu     sync.Mutex
	limit  int
	window time.Duration
	seen   map[string][]time.Time
}

func NewLimiter(limit int, window time.Duration) *Limiter {
	if limit < 1 {
		limit = 1
	}
	return &Limiter{limit: limit, window: window, seen: map[string][]time.Time{}}
}
func (l *Limiter) Allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-l.window)
	items := l.seen[key][:0]
	for _, at := range l.seen[key] {
		if at.After(cutoff) {
			items = append(items, at)
		}
	}
	if len(items) >= l.limit {
		l.seen[key] = items
		return false
	}
	l.seen[key] = append(items, now)
	return true
}
func RateLimit(l *Limiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(r.RemoteAddr, time.Now()) {
			w.Header().Set("Retry-After", strconv.Itoa(int(l.window.Seconds())))
			http.Error(w, "rate limit", 429)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func ClientKey(r *http.Request) string {
	if value := r.Header.Get("X-Client-ID"); value != "" {
		return value
	}
	return r.RemoteAddr
}
