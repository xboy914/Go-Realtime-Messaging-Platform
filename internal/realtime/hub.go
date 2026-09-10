package realtime

import (
	"context"
	"log/slog"
)

type client struct {
	id   string
	send chan []byte
	done chan struct{}
}

type Hub struct {
	logger     *slog.Logger
	register   chan *client
	unregister chan *client
	publish    chan []byte
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:     logger,
		register:   make(chan *client),
		unregister: make(chan *client),
		publish:    make(chan []byte, 256),
	}
}

func (h *Hub) Run(ctx context.Context) {
	clients := make(map[*client]struct{})
	for {
		select {
		case <-ctx.Done():
			for c := range clients {
				close(c.done)
			}
			return
		case c := <-h.register:
			clients[c] = struct{}{}
			h.logger.Debug("client connected", "client_id", c.id, "connections", len(clients))
		case c := <-h.unregister:
			if _, exists := clients[c]; exists {
				delete(clients, c)
				close(c.done)
				h.logger.Debug("client disconnected", "client_id", c.id, "connections", len(clients))
			}
		case message := <-h.publish:
			for c := range clients {
				select {
				case c.send <- message:
				default:
					delete(clients, c)
					close(c.done)
					h.logger.Warn("slow client disconnected", "client_id", c.id)
				}
			}
		}
	}
}

func (h *Hub) Publish(ctx context.Context, message []byte) error {
	snapshot := append([]byte(nil), message...)
	select {
	case h.publish <- snapshot:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
