package httpapi

import (
	"encoding/json"
	"errors"
	"github.com/zhanglei10281852-gif/newgo/internal/domain"
	"github.com/zhanglei10281852-gif/newgo/internal/middleware"
	"github.com/zhanglei10281852-gif/newgo/internal/requestmeta"
	"github.com/zhanglei10281852-gif/newgo/internal/service"
	"log/slog"
	"net/http"
	"strings"
)

type Server struct {
	Services *service.Services
	Logger   *slog.Logger
}

func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	Protected{Services: s.Services}.Register(mux)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { write(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.Services.Ready(r.Context()); err != nil {
			write(w, 503, errorBody(err, r))
			return
		}
		write(w, 200, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("POST /api/v1/auth/login", s.login)
	mux.HandleFunc("POST /api/v1/auth/logout", s.logout)
	return middleware.RequestID(middleware.Recovery(s.Logger, middleware.Logging(s.Logger, mux)))
}
func (s Server) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if json.NewDecoder(r.Body).Decode(&in) != nil {
		write(w, 400, errorBody(domain.ErrInvalid, r))
		return
	}
	sess, _, err := s.Services.Auth.Login(r.Context(), in.Email, in.Password)
	if err != nil {
		write(w, 401, errorBody(err, r))
		return
	}
	write(w, 200, map[string]string{"token": sess.ID, "expires_at": sess.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")})
}
func (s Server) logout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token == "" {
		write(w, 401, errorBody(domain.ErrForbidden, r))
		return
	}
	if err := s.Services.Auth.Logout(r.Context(), token); err != nil {
		write(w, 404, errorBody(err, r))
		return
	}
	write(w, 204, nil)
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func errorBody(err error, r *http.Request) map[string]any {
	code := "internal"
	status := 500
	if errors.Is(err, domain.ErrInvalid) {
		code = "invalid_input"
		status = 400
	}
	if errors.Is(err, domain.ErrForbidden) {
		code = "forbidden"
		status = 403
	}
	if errors.Is(err, domain.ErrConflict) {
		code = "conflict"
		status = 409
	}
	if errors.Is(err, domain.ErrNotFound) {
		code = "not_found"
		status = 404
	}
	return map[string]any{"error": map[string]any{"code": code, "message": err.Error(), "request_id": requestmeta.ID(r.Context()), "status": status}}
}
