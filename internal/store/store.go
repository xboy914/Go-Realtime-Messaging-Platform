package store

import (
	"context"
	"errors"

	"github.com/xboy914/Go-Realtime-Messaging-Platform/internal/domain"
)

var ErrForbidden = errors.New("room membership required")

type Messages interface {
	RoomIDs(context.Context, string) ([]string, error)
	CreateMessage(context.Context, string, string, string) (domain.Message, error)
	ListMessages(context.Context, string, string, int) ([]domain.Message, error)
}
