package httpapi

import (
	"context"
	"github.com/zhanglei10281852-gif/newgo/internal/clock"
	"github.com/zhanglei10281852-gif/newgo/internal/service"
	"github.com/zhanglei10281852-gif/newgo/internal/storage/sqlite"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func testHandler(t *testing.T) http.Handler {
	store, e := sqlite.Open(context.Background(), ":memory:")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { store.Close() })
	services := service.New(store, clock.Fixed{Value: time.Now()}, time.Hour, time.Hour)
	return (Server{Services: services, Logger: slog.Default()}).Handler()
}
func TestHealthAndReadiness(t *testing.T) {
	handler := testHandler(t)
	for _, path := range []string{"/healthz", "/readyz"} {
		request := httptest.NewRequest("GET", path, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatalf("%s=%d body=%s", path, response.Code, response.Body.String())
		}
		if response.Header().Get("X-Request-ID") == "" {
			t.Fatal("request id missing")
		}
	}
}
func TestLoginRejectsMalformedJSON(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest("POST", "/api/v1/auth/login", http.NoBody)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
}
func TestLogoutRequiresToken(t *testing.T) {
	handler := testHandler(t)
	request := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != 401 {
		t.Fatalf("code=%d body=%s", response.Code, response.Body.String())
	}
}
