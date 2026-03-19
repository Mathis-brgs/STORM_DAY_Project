package models

// Constantes pour les actions WebSocket
const (
	WSActionJoin      = "join"
	WSActionMessage   = "message"
	WSActionTyping    = "typing"
	WSActionDelivered = "delivered"
	WSActionSeen      = "seen"
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
	Attachment            string `json:"attachment,omitempty"`
	// ID / message_id : PK ligne `messages` (même valeur que l’API REST) pour le front (édition, delivered, etc.).
	ID        int    `json:"id,omitempty"`
	MessageID string `json:"message_id,omitempty"`
	// Réponse à un message : même forme que GET /api/messages pour afficher la citation sans attendre le resync.
	ReplyToID *int          `json:"reply_to_id,omitempty"`
	ReplyTo   *ReplyToData  `json:"reply_to,omitempty"`
}
