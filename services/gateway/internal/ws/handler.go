package ws

import (
	"encoding/json"
	"gateway/internal/models"
	"log"
	"time"

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

	// Démarrer le heartbeat (Ping toutes les 30s)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := socket.WritePing(nil); err != nil {
				return
			}
		}
	}()
}

func (h *Handler) OnPing(socket *gws.Conn, payload []byte) {
	_ = socket.WritePong(payload)
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

	switch msg.Action {

	case models.WSActionJoin:
		h.hub.Join(msg.Room, socket)
		socket.Session().Store("room", msg.Room)

	case models.WSActionMessage:
		h.hub.BroadcastToRoom(msg.Room, message.Bytes())

		err := h.nats.Publish("NEW_MESSAGE", message.Bytes())
		if err != nil {
			log.Printf("Erreur publication sur NATS : %v", err)
		}

	default:
		log.Printf("Action inconnue : %s", msg.Action)
	}
}
