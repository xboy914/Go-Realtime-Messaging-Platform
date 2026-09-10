package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/database"
)

func TestPostgresMessagingLifecycle(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL is required for PostgreSQL integration test")
	}
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := database.Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "DELETE FROM messages"); err != nil {
		t.Fatal(err)
	}
	repository := NewPostgres(db)
	message, err := repository.CreateMessage(ctx, "demo-room", "demo-alice", "persisted hello")
	if err != nil {
		t.Fatal(err)
	}
	history, err := repository.ListMessages(ctx, "demo-room", "demo-alice", 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || history[0].ID != message.ID {
		t.Fatalf("unexpected history: %#v", history)
	}
	if _, err := repository.CreateMessage(
		ctx, "demo-room", "demo-outsider", "forbidden",
	); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected membership failure, got %v", err)
	}
}
