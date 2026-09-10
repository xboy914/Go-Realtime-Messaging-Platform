package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/auth"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/store"
)

const maxMessageBytes = 16 << 10

type inboundEvent struct {
	Type    string `json:"type"`
	RoomID  string `json:"room_id"`
	Payload struct {
		Text string `json:"text"`
	} `json:"payload"`
}

type outboundEvent struct {
	ID     int64  `json:"id"`
	Type   string `json:"type"`
	RoomID string `json:"room_id"`
	Sender struct {
		ID          string `json:"id"`
		DisplayName string `json:"display_name"`
	} `json:"sender"`
	Payload struct {
		Text string `json:"text"`
	} `json:"payload"`
	SentAt time.Time `json:"sent_at"`
}

type Handler struct {
	logger *slog.Logger
	hub    *Hub
	auth   *auth.Manager
	store  store.Messages
	ids    atomic.Uint64
}

func NewHandler(
	logger *slog.Logger, hub *Hub, manager *auth.Manager, messages store.Messages,
) http.Handler {
	return &Handler{logger: logger, hub: hub, auth: manager, store: messages}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawToken, err := auth.ExtractBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeUnauthorized(w)
		return
	}
	claims, err := h.auth.Verify(rawToken)
	if err != nil {
		writeUnauthorized(w)
		return
	}
	roomIDs, err := h.store.RoomIDs(r.Context(), claims.Subject)
	if err != nil {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
		return
	}
	rooms := make(map[string]struct{}, len(roomIDs))
	for _, roomID := range roomIDs {
		rooms[roomID] = struct{}{}
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(maxMessageBytes)
	client := &client{
		id: fmt.Sprintf("client-%d", h.ids.Add(1)), userID: claims.Subject,
		displayName: claims.DisplayName, rooms: rooms,
		send: make(chan []byte, 64), done: make(chan struct{}),
	}
	select {
	case h.hub.register <- client:
	case <-r.Context().Done():
		return
	}
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		h.writeLoop(r.Context(), conn, client)
	}()
	h.readLoop(r.Context(), conn, client)
	select {
	case h.hub.unregister <- client:
	case <-client.done:
	case <-r.Context().Done():
	}
	<-writerDone
}

func (h *Handler) readLoop(ctx context.Context, conn *websocket.Conn, client *client) {
	for {
		messageType, payload, err := conn.Read(ctx)
		if err != nil {
			return
		}
		var incoming inboundEvent
		if messageType != websocket.MessageText || json.Unmarshal(payload, &incoming) != nil ||
			incoming.Type != "message.created" || incoming.RoomID == "" ||
			strings.TrimSpace(incoming.Payload.Text) == "" || len(incoming.Payload.Text) > 4000 {
			_ = conn.Close(websocket.StatusPolicyViolation, "invalid event")
			return
		}
		message, err := h.store.CreateMessage(
			ctx, incoming.RoomID, client.userID, strings.TrimSpace(incoming.Payload.Text),
		)
		if errors.Is(err, store.ErrForbidden) {
			_ = conn.Close(websocket.StatusPolicyViolation, "room membership required")
			return
		}
		if err != nil {
			_ = conn.Close(websocket.StatusInternalError, "message persistence failed")
			return
		}
		event := outboundEvent{
			ID: message.ID, Type: "message.created", RoomID: message.RoomID, SentAt: message.CreatedAt,
		}
		event.Sender.ID, event.Sender.DisplayName = client.userID, client.displayName
		event.Payload.Text = message.Body
		encoded, err := json.Marshal(event)
		if err != nil || h.hub.Publish(ctx, message.RoomID, encoded) != nil {
			return
		}
	}
}

func (h *Handler) writeLoop(ctx context.Context, conn *websocket.Conn, client *client) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.done:
			return
		case message := <-client.send:
			writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := conn.Write(writeCtx, websocket.MessageText, message)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": "valid Bearer token required"})
}
