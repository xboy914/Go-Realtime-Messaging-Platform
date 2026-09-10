package domain

import "time"

type Message struct {
	ID          int64     `json:"id"`
	RoomID      string    `json:"room_id"`
	SenderID    string    `json:"sender_id"`
	DisplayName string    `json:"display_name"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
}
