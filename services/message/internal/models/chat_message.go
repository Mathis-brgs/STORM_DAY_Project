package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage : id (PK int), sender_id (UUID), conversation_id (int).
// ReceivedAt est reserve au contexte d'un acteur (ACK), pas un etat global du message.
type ChatMessage struct {
	ID             int        `json:"id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	ConversationID int        `json:"conversation_id"`
	Content        string     `json:"content"`
	Attachment     string     `json:"attachment,omitempty"`
	ReceivedAt     *time.Time `json:"received_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
