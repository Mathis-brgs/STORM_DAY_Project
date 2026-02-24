package models

const (
	ActionJoin    = "join"
	ActionMessage = "message"
	ActionTyping  = "typing"
)

type InputMessage struct {
	Action  string `json:"action"`
	Room    string `json:"room"`
	User    string `json:"user"`
	Content string `json:"content"`
}
