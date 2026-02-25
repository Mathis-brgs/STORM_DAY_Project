package models

// Constantes pour les actions WebSocket
const (
	WSActionJoin    = "join"
	WSActionMessage = "message"
	WSActionTyping  = "typing"
)

// InputMessage est le payload JSON envoyé par le client sur le WebSocket
type InputMessage struct {
	Action  string `json:"action"`
	Room    string `json:"room"`
	User    string `json:"user"`
	Content string `json:"content"`
}
