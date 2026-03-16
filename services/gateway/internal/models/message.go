package models

// SendMessageRequest accepts the new conversation_id and legacy group_id.
type SendMessageRequest struct {
	ConversationID int    `json:"conversation_id,omitempty"`
	GroupID        int    `json:"group_id,omitempty"` // legacy alias
	SenderID       string `json:"sender_id"`          // UUID
	Content        string `json:"content"`
	Attachment     string `json:"attachment,omitempty"`
}

// SendMessageResponse est la réponse renvoyée par l'API messages
type SendMessageResponse struct {
	OK    bool              `json:"ok"`
	Data  *SendMessageData  `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

// SendMessageData returns conversation_id and keeps group_id for temporary compatibility.
type SendMessageData struct {
	ID             int    `json:"id"`
	SenderID       string `json:"sender_id"` // UUID
	ConversationID int    `json:"conversation_id"`
	GroupID        int    `json:"group_id,omitempty"` // legacy alias
	Content        string `json:"content"`
	Attachment     string `json:"attachment,omitempty"`
	ReceivedAt     int64  `json:"received_at,omitempty"` // actor-scoped receipt when available
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
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

// GetMessageData : id (int), sender_id (UUID), conversation_id (int).
type GetMessageData struct {
	ID             int    `json:"id"`
	SenderID       string `json:"sender_id"`
	ConversationID int    `json:"conversation_id"`
	GroupID        int    `json:"group_id,omitempty"` // legacy alias
	Content        string `json:"content"`
	Attachment     string `json:"attachment,omitempty"`
	ReceivedAt     int64  `json:"received_at,omitempty"` // actor-scoped receipt when available
	CreatedAt      int64  `json:"created_at"`
	UpdatedAt      int64  `json:"updated_at"`
}

// ListMessagesResponse est la réponse de GET /api/messages
type ListMessagesResponse struct {
	OK         bool              `json:"ok"`
	Data       []SendMessageData `json:"data,omitempty"`
	NextCursor string            `json:"next_cursor,omitempty"`
	Error      *SendMessageError `json:"error,omitempty"`
}

// UpdateMessageRequest est le payload de PUT /api/messages/{id}
type UpdateMessageRequest struct {
	Content string `json:"content"`
	Message string `json:"message"`
	ActorID string `json:"actor_id,omitempty"` // UUID
}

// UpdateMessageResponse est la réponse de PUT /api/messages/{id}
type UpdateMessageResponse struct {
	OK    bool              `json:"ok"`
	Data  *SendMessageData  `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

// AckMessageRequest est le payload de POST /api/messages/{id}/receipt
type AckMessageRequest struct {
	ActorID    string `json:"actor_id,omitempty"`    // UUID
	ReceivedAt int64  `json:"received_at,omitempty"` // Unix timestamp (optionnel)
}

// AckMessageResponse est la réponse de POST /api/messages/{id}/receipt
type AckMessageResponse struct {
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

type Group struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
	CreatedBy string `json:"created_by,omitempty"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type GroupMember struct {
	ID             int    `json:"id"`
	ConversationID int    `json:"conversation_id"`
	GroupID        int    `json:"group_id,omitempty"` // legacy alias
	UserID         string `json:"user_id"`
	Role           int    `json:"role"`
	CreatedAt      int64  `json:"created_at"`
}

type CreateGroupRequest struct {
	ActorID   string `json:"actor_id,omitempty"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type GroupResponse struct {
	OK    bool              `json:"ok"`
	Data  *Group            `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

type GroupsResponse struct {
	OK    bool              `json:"ok"`
	Data  []Group           `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

type AddGroupMemberRequest struct {
	ActorID string `json:"actor_id,omitempty"`
	UserID  string `json:"user_id"`
	Role    int    `json:"role"`
}

type UpdateGroupMemberRoleRequest struct {
	ActorID string `json:"actor_id,omitempty"`
	Role    int    `json:"role"`
}

type GroupMemberResponse struct {
	OK    bool              `json:"ok"`
	Data  *GroupMember      `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}

type GroupMembersResponse struct {
	OK    bool              `json:"ok"`
	Data  []GroupMember     `json:"data,omitempty"`
	Error *SendMessageError `json:"error,omitempty"`
}
