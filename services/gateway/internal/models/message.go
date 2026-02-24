package models

// SendMessageRequest est le payload JSON pour POST /api/messages
type SendMessageRequest struct {
	GroupID  int32  `json:"group_id"`
	SenderID  int32  `json:"sender_id"`
	Content  string `json:"content"`
}

// SendMessageResponse est la réponse JSON de POST /api/messages
type SendMessageResponse struct {
	OK    bool               `json:"ok"`
	Data  *SendMessageData   `json:"data,omitempty"`
	Error *SendMessageError  `json:"error,omitempty"`
}

// SendMessageData représente un message dans les réponses
type SendMessageData struct {
	ID        int32  `json:"id"`
	SenderID  int32  `json:"sender_id"`
	GroupID   int32  `json:"group_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// SendMessageError représente une erreur dans les réponses message
type SendMessageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetMessageResponse est la réponse JSON de GET /api/messages/{id}
type GetMessageResponse struct {
	OK    bool              `json:"ok"`
	Data  *GetMessageData   `json:"data,omitempty"`
	Error *GetMessageError  `json:"error,omitempty"`
}

// GetMessageData représente un message (alias pour réutilisation)
type GetMessageData = SendMessageData

// GetMessageError représente une erreur pour Get
type GetMessageError = SendMessageError

// ListMessagesResponse est la réponse JSON de GET /api/messages?group_id=...
type ListMessagesResponse struct {
	OK         bool               `json:"ok"`
	Data       []SendMessageData  `json:"data,omitempty"`
	NextCursor string             `json:"next_cursor,omitempty"`
	Error      *SendMessageError  `json:"error,omitempty"`
}
