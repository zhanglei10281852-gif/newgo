package middleware

import (
	"github.com/zhanglei10281852-gif/newgo/internal/identity"
	"github.com/zhanglei10281852-gif/newgo/internal/requestmeta"
	"log/slog"
	"net/http"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = identity.New("req")
		}
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(requestmeta.WithID(r.Context(), id)))
	})
}
func Recovery(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				logger.Error("panic recovered", "value", v)
				http.Error(w, `{"error":"internal"}`, 500)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
func Logging(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request", "method", r.Method, "path", r.URL.Path, "request_id", requestmeta.ID(r.Context()))
		next.ServeHTTP(w, r)
	})
}
