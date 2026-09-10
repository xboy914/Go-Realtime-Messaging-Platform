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
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/auth"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/domain"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/realtime"
)

const testSecret = "test-secret-with-at-least-thirty-two-bytes"

type memoryMessages struct{}

func (memoryMessages) RoomIDs(context.Context, string) ([]string, error) {
	return []string{"demo-room"}, nil
}

func (memoryMessages) CreateMessage(
	_ context.Context, roomID, userID, body string,
) (domain.Message, error) {
	return domain.Message{
		ID: 1, RoomID: roomID, SenderID: userID, Body: body, CreatedAt: time.Now().UTC(),
	}, nil
}

func (memoryMessages) ListMessages(
	context.Context, string, string, int,
) ([]domain.Message, error) {
	return []domain.Message{}, nil
}

func testServer(t *testing.T) (*httptest.Server, *auth.Manager, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	hub := realtime.NewHub(logger)
	go hub.Run(ctx)
	manager, err := auth.NewManager(testSecret, "messaging-test", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(NewRouter(logger, hub, manager, memoryMessages{})), manager, cancel
}

func TestHealth(t *testing.T) {
	server, _, cancel := testServer(t)
	defer cancel()
	defer server.Close()
	response, err := http.Get(server.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.StatusCode)
	}
	if response.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("security headers are missing")
	}
}

func TestWebSocketRequiresJWT(t *testing.T) {
	server, _, cancel := testServer(t)
	defer cancel()
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	connection, response, err := websocket.Dial(context.Background(), wsURL, nil)
	if connection != nil {
		connection.CloseNow()
	}
	if err == nil || response == nil || response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized handshake, response=%v err=%v", response, err)
	}
	response.Body.Close()
}

func TestAuthenticatedBroadcastUsesTokenIdentity(t *testing.T) {
	server, manager, cancel := testServer(t)
	defer cancel()
	defer server.Close()
	ctx := context.Background()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	aliceToken, err := manager.Issue("user-alice", "Alice")
	if err != nil {
		t.Fatal(err)
	}
	bobToken, err := manager.Issue("user-bob", "Bob")
	if err != nil {
		t.Fatal(err)
	}
	dial := func(token string) *websocket.Conn {
		headers := http.Header{"Authorization": []string{"Bearer " + token}}
		connection, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{HTTPHeader: headers})
		if err != nil {
			t.Fatal(err)
		}
		return connection
	}
	alice := dial(aliceToken)
	defer alice.CloseNow()
	bob := dial(bobToken)
	defer bob.CloseNow()

	payload := []byte(`{"type":"message.created","room_id":"demo-room","payload":{"text":"hello"}}`)
	if err := alice.Write(ctx, websocket.MessageText, payload); err != nil {
		t.Fatal(err)
	}
	readCtx, stopRead := context.WithTimeout(ctx, 2*time.Second)
	defer stopRead()
	_, received, err := bob.Read(readCtx)
	if err != nil {
		t.Fatal(err)
	}
	var event struct {
		Sender struct {
			ID string `json:"id"`
		} `json:"sender"`
	}
	if err := json.Unmarshal(received, &event); err != nil {
		t.Fatal(err)
	}
	if event.Sender.ID != "user-alice" {
		t.Fatalf("sender must come from JWT claims: %s", received)
	}
}
