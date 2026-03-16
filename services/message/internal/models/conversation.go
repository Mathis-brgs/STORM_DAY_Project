package models

import (
	"time"

	"github.com/google/uuid"
)

type ConversationRole int

const (
	ConversationRoleMember ConversationRole = 0
	ConversationRoleAdmin  ConversationRole = 1
	ConversationRoleOwner  ConversationRole = 2
)

func IsValidConversationRole(role ConversationRole) bool {
	return role == ConversationRoleMember || role == ConversationRoleAdmin || role == ConversationRoleOwner
}

type Conversation struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	CreatedBy uuid.UUID  `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type ConversationMembership struct {
	ID             int              `json:"id"`
	UserID         uuid.UUID        `json:"user_id"`
	ConversationID int              `json:"conversation_id"`
	Role           ConversationRole `json:"role"`
	CreatedAt      time.Time        `json:"created_at"`
	DeletedAt      *time.Time       `json:"deleted_at,omitempty"`
}
