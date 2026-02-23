package models

import "time"

// ChatMessage est la structure d'un message de discussion (attachments gérés par media-service)
type ChatMessage struct {
	ID        int        `json:"id"`
	SenderID  int        `json:"sender_id"`
	Content   string     `json:"content"`
	GroupID   int        `json:"group_id"`
	CreatedAt time.Time  `json:"created_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at"`
}
