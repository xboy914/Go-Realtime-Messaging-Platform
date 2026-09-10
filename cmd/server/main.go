package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/auth"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/database"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/httpapi"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/realtime"
	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	manager, err := auth.NewManager(os.Getenv("JWT_SECRET"), "go-realtime-messaging", 15*time.Minute)
	if err != nil {
		logger.Error("invalid authentication configuration", "error", err)
		return
	}
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("database configuration failed", "error", err)
		return
	}
	defer db.Close()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := db.PingContext(ctx); err != nil {
		logger.Error("database unavailable", "error", err)
		return
	}
	if err := database.Migrate(ctx, db); err != nil {
		logger.Error("migration failed", "error", err)
		return
	}
	messages := store.NewPostgres(db)
	hub := realtime.NewHub(logger)
	go hub.Run(ctx)
	server := &http.Server{
		Addr: env("HTTP_ADDR", ":8080"), Handler: httpapi.NewRouter(logger, hub, manager, messages),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second,
		WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second,
	}
	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			stop()
		}
	}()
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
