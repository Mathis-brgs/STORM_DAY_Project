package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageReceipt stocke l'accuse de reception d'un utilisateur pour un message.
type MessageReceipt struct {
	MessageID  int       `json:"message_id"`
	UserID     uuid.UUID `json:"user_id"`
	ReceivedAt time.Time `json:"received_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
