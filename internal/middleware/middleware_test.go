package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequestID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if w.Header().Get("X-Request-ID") == "" {
			t.Fatal("request id missing")
		}
		w.WriteHeader(204)
	}))
	request := httptest.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 204 {
		t.Fatal(response.Code)
	}
}
func TestLimiter(t *testing.T) {
	l := NewLimiter(2, time.Minute)
	now := time.Now()
	if !l.Allow("a", now) || !l.Allow("a", now) || l.Allow("a", now) {
		t.Fatal("limit incorrect")
	}
	if !l.Allow("a", now.Add(2*time.Minute)) {
		t.Fatal("window did not reset")
	}
}
