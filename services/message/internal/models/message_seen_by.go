package models

import (
	"time"

	"github.com/google/uuid"
)

// MessageSeenBy : entrée table message_seen_by (vu par un utilisateur).
type MessageSeenBy struct {
	MessageID   int       `json:"message_id"`
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	SeenAt      time.Time `json:"seen_at"`
}
