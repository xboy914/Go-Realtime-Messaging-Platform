package realtime

import (
	"context"
	"log/slog"
)

type client struct {
	id          string
	userID      string
	displayName string
	rooms       map[string]struct{}
	send        chan []byte
	done        chan struct{}
}

type publication struct {
	roomID string
	data   []byte
}

type Hub struct {
	logger     *slog.Logger
	register   chan *client
	unregister chan *client
	publish    chan publication
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger: logger, register: make(chan *client), unregister: make(chan *client),
		publish: make(chan publication, 256),
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
			h.logger.Debug("client connected", "client_id", c.id, "user_id", c.userID)
		case c := <-h.unregister:
			if _, exists := clients[c]; exists {
				delete(clients, c)
				close(c.done)
			}
		case message := <-h.publish:
			for c := range clients {
				if _, member := c.rooms[message.roomID]; !member {
					continue
				}
				select {
				case c.send <- message.data:
				default:
					delete(clients, c)
					close(c.done)
					h.logger.Warn("slow client disconnected", "client_id", c.id)
				}
			}
		}
	}
}

func (h *Hub) Publish(ctx context.Context, roomID string, data []byte) error {
	message := publication{roomID: roomID, data: append([]byte(nil), data...)}
	select {
	case h.publish <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
