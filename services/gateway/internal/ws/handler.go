package ws

import (
	"encoding/json"
	"gateway/internal/models"
	"log"

	"github.com/lxzan/gws"
)

type Handler struct {
	gws.BuiltinEventHandler
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// OnOpen : On ne fait plus rien ici.
// La connexion technique est établie, mais l'utilisateur n'est pas encore dans une room.
func (h *Handler) OnOpen(socket *gws.Conn) {
	log.Println("Nouvelle connexion socket établie (en attente de Join)")
}

// OnClose : Le nettoyage
func (h *Handler) OnClose(socket *gws.Conn, err error) {
	// On ouvre le "sac à dos" du socket pour retrouver le nom de la room
	if roomName, exist := socket.Session().Load("room"); exist {
		// Si on trouve une room, on le supprime du Hub
		h.hub.Leave(roomName.(string), socket)
	}
}

func (h *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	// 1. On décode le JSON
	var msg models.InputMessage
	if err := json.Unmarshal(message.Bytes(), &msg); err != nil {
		log.Printf("Erreur JSON : %v", err)
		return
	}

	// 2. Aiguillage selon l'action
	switch msg.Action {

	case models.ActionJoin:
		// L'utilisateur veut rejoindre une room
		h.hub.Join(msg.Room, socket)

		// IMPORTANT : On note le nom de la room dans le "sac à dos" du socket
		// Ça nous servira pour le OnClose plus tard
		socket.Session().Store("room", msg.Room)

	case models.ActionMessage:
		// L'utilisateur veut parler dans sa room
		// On renvoie le JSON brut tel quel (ou on pourrait le modifier)
		h.hub.BroadcastToRoom(msg.Room, message.Bytes())

	default:
		log.Printf("Action inconnue : %s", msg.Action)
	}
}
