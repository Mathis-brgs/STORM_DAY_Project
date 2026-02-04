package models

// On définit des constantes pour éviter les fautes de frappe dans le code
const (
	ActionJoin    = "join"
	ActionMessage = "message"
)

// InputMessage représente ce que le client envoie au serveur
type InputMessage struct {
	Action  string `json:"action"`  // Ex: "join" ou "message"
	Room    string `json:"room"`    // Ex: "salon-1"
	User    string `json:"user"`    // Ex: "Alice"
	Content string `json:"content"` // Ex: "Salut tout le monde"
}
