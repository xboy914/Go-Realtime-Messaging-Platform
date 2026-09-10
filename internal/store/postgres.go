package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/domain"
)

type Postgres struct {
	db *sql.DB
}

func NewPostgres(db *sql.DB) *Postgres {
	return &Postgres{db: db}
}

func (p *Postgres) RoomIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := p.db.QueryContext(ctx,
		`SELECT room_id FROM room_members WHERE user_id = $1 ORDER BY room_id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list memberships: %w", err)
	}
	defer rows.Close()
	var roomIDs []string
	for rows.Next() {
		var roomID string
		if err := rows.Scan(&roomID); err != nil {
			return nil, fmt.Errorf("scan membership: %w", err)
		}
		roomIDs = append(roomIDs, roomID)
	}
	return roomIDs, rows.Err()
}

func (p *Postgres) CreateMessage(
	ctx context.Context, roomID, userID, body string,
) (domain.Message, error) {
	var message domain.Message
	err := p.db.QueryRowContext(ctx, `
		INSERT INTO messages (room_id, sender_id, body)
		SELECT $1, $2, $3
		WHERE EXISTS (
			SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2
		)
		RETURNING id, room_id, sender_id, body, created_at
	`, roomID, userID, body).Scan(
		&message.ID, &message.RoomID, &message.SenderID, &message.Body, &message.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Message{}, ErrForbidden
	}
	if err != nil {
		return domain.Message{}, fmt.Errorf("create message: %w", err)
	}
	return message, nil
}

func (p *Postgres) ListMessages(
	ctx context.Context, roomID, userID string, limit int,
) ([]domain.Message, error) {
	var member bool
	if err := p.db.QueryRowContext(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM room_members WHERE room_id = $1 AND user_id = $2
		)`, roomID, userID).Scan(&member); err != nil {
		return nil, fmt.Errorf("check membership: %w", err)
	}
	if !member {
		return nil, ErrForbidden
	}
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, room_id, sender_id, display_name, body, created_at
		FROM (
			SELECT m.id, m.room_id, m.sender_id, u.display_name, m.body, m.created_at
			FROM messages m
			JOIN users u ON u.id = m.sender_id
			WHERE m.room_id = $1
			ORDER BY m.created_at DESC, m.id DESC
			LIMIT $2
		) recent
		ORDER BY created_at, id
	`, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()
	messages := make([]domain.Message, 0, limit)
	for rows.Next() {
		var message domain.Message
		if err := rows.Scan(
			&message.ID, &message.RoomID, &message.SenderID, &message.DisplayName,
			&message.Body, &message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}
