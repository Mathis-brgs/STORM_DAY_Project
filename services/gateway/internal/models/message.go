package models

// SendMessageRequest : group_id (int), sender_id (UUID).
type SendMessageRequest struct {
	GroupID     int    `json:"group_id"`
	SenderID    string `json:"sender_id"` // UUID
	Content     string `json:"content"`
	Attachment  string `json:"attachment,omitempty"`
}

// SendMessageResponse est la réponse renvoyée par l'API messages
type SendMessageResponse struct {
	OK    bool              `json:"ok"`
	Data  *SendMessageData  `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

// SendMessageData : id (PK int), sender_id (UUID), group_id (int).
type SendMessageData struct {
	ID         int    `json:"id"`
	SenderID   string `json:"sender_id"`  // UUID
	GroupID    int    `json:"group_id"`
	Content    string `json:"content"`
	Attachment string `json:"attachment,omitempty"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// SendMessageError représente une erreur dans la réponse message
type SendMessageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetMessageError représente une erreur dans la réponse message
type GetMessageError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GetMessageRequest : id (PK row, int).
type GetMessageRequest struct {
	ID int `json:"id"`
}

// GetMessageResponse est la réponse renvoyée par l'API messages
type GetMessageResponse struct {
	OK    bool             `json:"ok"`
	Data  *GetMessageData  `json:"data,omitempty"`
	Error *GetMessageError `json:"error,omitempty"`
}

// GetMessageData : id (int), sender_id (UUID), group_id (int).
type GetMessageData struct {
	ID         int    `json:"id"`
	SenderID   string `json:"sender_id"`
	GroupID    int    `json:"group_id"`
	Content    string `json:"content"`
	Attachment string `json:"attachment,omitempty"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

// ListMessagesResponse est la réponse de GET /api/messages
type ListMessagesResponse struct {
	OK         bool              `json:"ok"`
	Data       []SendMessageData  `json:"data,omitempty"`
	NextCursor string            `json:"next_cursor,omitempty"`
	Error      *SendMessageError  `json:"error,omitempty"`
}

// UpdateMessageRequest est le payload de PUT /api/messages/{id}
type UpdateMessageRequest struct {
	Content  string `json:"content"`
	Message  string `json:"message"`
}

// UpdateMessageResponse est la réponse de PUT /api/messages/{id}
type UpdateMessageResponse struct {
	OK    bool              `json:"ok"`
	Data  *SendMessageData  `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

// DeleteMessageRequest : id (int).
type DeleteMessageRequest struct {
	ID int `json:"id"`
}

// DeleteMessageResponse est la réponse de DELETE /api/messages/{id}
type DeleteMessageResponse struct {
	OK    bool              `json:"ok"`
	Error *SendMessageError `json:"error,omitempty"`
}
