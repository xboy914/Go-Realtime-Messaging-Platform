package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/auth"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/realtime"
)

func NewRouter(logger *slog.Logger, hub *realtime.Hub, manager *auth.Manager) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "version": "0.2.0"})
	})
	mux.Handle("GET /ws", realtime.NewHandler(logger, hub, manager))
	return securityHeaders(mux)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
