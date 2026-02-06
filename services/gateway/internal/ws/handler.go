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
	log.Println("Nouvelle connexion socket établie (en attente de Join)")
}
func (h *Handler) OnClose(socket *gws.Conn, err error) {
	if roomName, exist := socket.Session().Load("room"); exist {
		h.hub.Leave(roomName.(string), socket)
	}
}

func (h *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	defer message.Close()

	var msg models.InputMessage
	if err := json.Unmarshal(message.Bytes(), &msg); err != nil {
		log.Printf("Erreur JSON : %v", err)
		return
	}

	switch msg.Action {

	case models.ActionJoin:
		h.hub.Join(msg.Room, socket)
		socket.Session().Store("room", msg.Room)

	case models.ActionMessage:
		h.hub.BroadcastToRoom(msg.Room, message.Bytes())

		_ = h.nats.Publish("NEW_MESSAGE", message.Bytes())

	default:
		log.Printf("Action inconnue : %s", msg.Action)
	}
}
