package models

// EventType définit le type d'événement
type EventType string

const (
	EventNewMessage EventType = "NEW_MESSAGE"
	EventUserTyping EventType = "USER_TYPING"
	EventUserOnline EventType = "USER_ONLINE"
)

// EventMessage est le payload envoyé dans NATS
type EventMessage struct {
	Type      EventType `json:"type"`
	Payload   any       `json:"payload"` // 
	Timestamp int64     `json:"timestamp"`
}

// ChatMessage est la structure d'un message de discussion
type ChatMessage struct {
	ID             string `json:"id"`
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	Content        string `json:"content"`
	CreatedAt      string `json:"created_at"`
}