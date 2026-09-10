package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/realtime"
)

func TestHealth(t *testing.T) {
	hub := realtime.NewHub(slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	response := httptest.NewRecorder()

	NewRouter(slog.Default(), hub).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("security headers are missing")
	}
}

func TestWebSocketBroadcast(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := realtime.NewHub(logger)
	go hub.Run(ctx)

	server := httptest.NewServer(NewRouter(logger, hub))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	first, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer first.CloseNow()
	second, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer second.CloseNow()

	payload := []byte(`{"type":"message.created","room_id":"demo-room","payload":{"text":"hello"}}`)
	if err := first.Write(ctx, websocket.MessageText, payload); err != nil {
		t.Fatal(err)
	}

	readCtx, stopRead := context.WithTimeout(ctx, 2*time.Second)
	defer stopRead()
	_, received, err := second.Read(readCtx)
	if err != nil {
		t.Fatal(err)
	}
	var event map[string]any
	if err := json.Unmarshal(received, &event); err != nil {
		t.Fatal(err)
	}
	if event["type"] != "message.created" || event["room_id"] != "demo-room" {
		t.Fatalf("unexpected event: %s", received)
	}
}
