package models

// Constantes pour les actions WebSocket
const (
	WSActionJoin    = "join"
	WSActionMessage = "message"
	WSActionTyping  = "typing"
)

// InputMessage est le payload JSON envoyé par le client sur le WebSocket
type InputMessage struct {
	Action   string `json:"action"`
	Room     string `json:"room"`
	User     string `json:"user"`
	Username string `json:"username,omitempty"`
	Content  string `json:"content"`
	// Attachment fields: either provide base64 payload or an existing mediaId
	AttachmentBase64      string `json:"attachmentBase64,omitempty"`
	AttachmentFilename    string `json:"attachmentFilename,omitempty"`
	AttachmentContentType string `json:"attachmentContentType,omitempty"`
	// Attachment holds the mediaId (returned by media-service) when available
	Attachment            string `json:"attachment,omitempty"`
}
