package ws

import (
	"encoding/json"
	"gateway/internal/models"
	"log"

	"github.com/lxzan/gws"
	"github.com/nats-io/nats.go"
)

type Handler struct {
	gws.BuiltinEventHandler
	hub  *Hub
	nats *nats.Conn
}

func NewHandler(hub *Hub, nats *nats.Conn) *Handler {
	return &Handler{hub: hub, nats: nats}
}

func (h *Handler) OnOpen(socket *gws.Conn) {
	userId, _ := socket.Session().Load("userId")
	username, _ := socket.Session().Load("username")
	log.Printf("Nouvelle connexion socket établie : %s (%s)", username, userId)

	// Rejoindre automatiquement une room privée pour l'utilisateur
	if userId != nil {
		h.hub.Join("user:"+userId.(string), socket)
	}
}
func (h *Handler) OnClose(socket *gws.Conn, err error) {
	if roomName, exist := socket.Session().Load("room"); exist {
		h.hub.Leave(roomName.(string), socket)
	}
}

func (h *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer func() {
		err := message.Close()
		if err != nil {
			log.Printf("Erreur fermeture message : %v", err)
		}
	}()

	var msg models.InputMessage
	if err := json.Unmarshal(message.Bytes(), &msg); err != nil {
		log.Printf("Erreur JSON : %v", err)
		return
	}

	userId, _ := socket.Session().Load("userId")

	switch msg.Action {

	case models.ActionJoin:
		h.hub.Join(msg.Room, socket)
		socket.Session().Store("room", msg.Room)

	case models.ActionMessage:
		// Sécurité : on impose l'ID de l'utilisateur authentifié
		if userId != nil {
			msg.User = userId.(string)
		}

		// Re-marshal pour envoyer les données propres (avec l'user ID forcé)
		finalPayload, _ := json.Marshal(msg)

		h.hub.BroadcastToRoom(msg.Room, finalPayload)

		err := h.nats.Publish("message.send", finalPayload)
		if err != nil {
			log.Printf("Erreur publication sur NATS : %v", err)
		}

	default:
		log.Printf("Action inconnue : %s", msg.Action)
	}
}
