package models

import (
	"time"

	"github.com/google/uuid"
)

// ChatMessage : id (PK int), sender_id (UUID), conversation_id (int).
// ReceivedAt est reserve au contexte d'un acteur (ACK), pas un etat global du message.
// ReplyToID, ForwardFromID optionnels. Status: sent | delivered | seen.
type ChatMessage struct {
	ID             int        `json:"id"`
	SenderID       uuid.UUID  `json:"sender_id"`
	ConversationID int        `json:"conversation_id"`
	Content        string     `json:"content"`
	Attachment     string     `json:"attachment,omitempty"`
	ReceivedAt     *time.Time `json:"received_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	ReplyToID     *int          `json:"reply_to_id,omitempty"`
	Status        string        `json:"status"`
	ForwardFromID *int          `json:"forward_from_id,omitempty"`
	ReplyTo       *ReplyToRef   `json:"reply_to,omitempty"`
	SeenBy        []SeenByEntry `json:"seen_by,omitempty"`
}

// ReplyToRef : message référencé pour une réponse (GET /api/messages).
type ReplyToRef struct {
	ID         int    `json:"id"`
	SenderID   string `json:"sender_id,omitempty"`
	SenderName string `json:"sender_name,omitempty"`
	Content    string `json:"content"`
}

// SeenByEntry : un utilisateur ayant vu le message (message_seen_by).
type SeenByEntry struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
	SeenAt      int64  `json:"seen_at,omitempty"`
}
