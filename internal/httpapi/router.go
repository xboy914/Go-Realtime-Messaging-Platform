package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/auth"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/realtime"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/store"
)

func NewRouter(
	logger *slog.Logger, hub *realtime.Hub, manager *auth.Manager, messages store.Messages,
) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "0.3.0"})
	})
	mux.Handle("GET /ws", realtime.NewHandler(logger, hub, manager, messages))
	mux.HandleFunc("GET /rooms/{roomID}/messages", func(w http.ResponseWriter, r *http.Request) {
		rawToken, err := auth.ExtractBearer(r.Header.Get("Authorization"))
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		claims, err := manager.Verify(rawToken)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
			return
		}
		history, err := messages.ListMessages(r.Context(), r.PathValue("roomID"), claims.Subject, 50)
		if errors.Is(err, store.ErrForbidden) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "membership required"})
			return
		}
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "history unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"messages": history})
	})
	return securityHeaders(mux)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
