package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage : id (PK int), sender_id (UUID), group_id (int).
type ChatMessage struct {
	ID        int       `json:"id"`
	SenderID  uuid.UUID `json:"sender_id"`
	GroupID   int       `json:"group_id"`
	Content   string    `json:"content"`
	Attachment string   `json:"attachment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
