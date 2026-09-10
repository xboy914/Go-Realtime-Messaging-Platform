package realtime

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

const maxMessageBytes = 16 << 10

type inboundEvent struct {
	Type    string          `json:"type"`
	RoomID  string          `json:"room_id"`
	Payload json.RawMessage `json:"payload"`
}

type outboundEvent struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	RoomID  string          `json:"room_id"`
	Payload json.RawMessage `json:"payload"`
	SentAt  time.Time       `json:"sent_at"`
}

type Handler struct {
	logger *slog.Logger
	hub    *Hub
	ids    atomic.Uint64
}

func NewHandler(logger *slog.Logger, hub *Hub) http.Handler {
	return &Handler{logger: logger, hub: hub}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
	})
	if err != nil {
		h.logger.Warn("websocket upgrade rejected", "error", err)
		return
	}
	defer conn.CloseNow()
	conn.SetReadLimit(maxMessageBytes)

	client := &client{
		id:   fmt.Sprintf("client-%d", h.ids.Add(1)),
		send: make(chan []byte, 64),
		done: make(chan struct{}),
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

	h.readLoop(r.Context(), conn)
	select {
	case h.hub.unregister <- client:
	case <-client.done:
	case <-r.Context().Done():
	}
	<-writerDone
}

func (h *Handler) readLoop(ctx context.Context, conn *websocket.Conn) {
	for {
		messageType, payload, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if messageType != websocket.MessageText {
			_ = conn.Close(websocket.StatusUnsupportedData, "text messages only")
			return
		}

		var incoming inboundEvent
		if err := json.Unmarshal(payload, &incoming); err != nil ||
			incoming.Type != "message.created" ||
			incoming.RoomID == "" ||
			len(incoming.Payload) == 0 {
			_ = conn.Close(websocket.StatusPolicyViolation, "invalid event")
			return
		}

		event := outboundEvent{
			ID:      fmt.Sprintf("event-%d", h.ids.Add(1)),
			Type:    incoming.Type,
			RoomID:  incoming.RoomID,
			Payload: incoming.Payload,
			SentAt:  time.Now().UTC(),
		}
		encoded, err := json.Marshal(event)
		if err != nil {
			return
		}
		if err := h.hub.Publish(ctx, encoded); err != nil {
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
