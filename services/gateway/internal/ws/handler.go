package ws

import (
	"encoding/json"
	"gateway/internal/models"
	"log"
	"strconv"
	"strings"
	"time"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/lxzan/gws"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	gws.BuiltinEventHandler
	hub               *Hub
	nats              NatsConn
	HeartbeatInterval time.Duration
}

func NewHandler(hub *Hub, nats NatsConn) *Handler {
	return &Handler{
		hub:               hub,
		nats:              nats,
		HeartbeatInterval: 30 * time.Second,
	}
}

func (h *Handler) OnOpen(socket *gws.Conn) {
	h.onOpen(socket)
}

func (h *Handler) onOpen(socket Socket) {
	userId, _ := socket.Session().Load("userId")
	username, _ := socket.Session().Load("username")
	log.Printf("Nouvelle connexion socket établie : %s (%s)", username, userId)

	// Rejoindre automatiquement une room privée pour l'utilisateur
	if userId != nil {
		h.hub.Join("user:"+userId.(string), socket)
	}

	// Démarrer le heartbeat (Ping à intervalles configurables)
	go func() {
		ticker := time.NewTicker(h.HeartbeatInterval)
		defer ticker.Stop()
		for range ticker.C {
			if err := socket.WritePing(nil); err != nil {
				return
			}
		}
	}()
}

func (h *Handler) OnPing(socket *gws.Conn, payload []byte) {
	h.onPing(socket, payload)
}

func (h *Handler) onPing(socket Socket, payload []byte) {
	_ = socket.WritePong(payload)
}

func (h *Handler) OnClose(socket *gws.Conn, err error) {
	h.onClose(socket, err)
}

func (h *Handler) onClose(socket Socket, err error) {
	if roomName, exist := socket.Session().Load("room"); exist {
		h.hub.Leave(roomName.(string), socket)
	}
}

func (h *Handler) OnMessage(socket *gws.Conn, message *gws.Message) {
	h.onMessage(socket, message)
}

func (h *Handler) onMessage(socket Socket, message WSMessage) {
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
		userId, _ := socket.Session().Load("userId")
		// Sécurité : on impose l'ID de l'utilisateur authentifié
		if userId != nil {
			msg.User = userId.(string)
		}

		// On extrait l'ID du groupe de la room (ex: "group:123")
		roomParts := strings.Split(msg.Room, ":")
		if len(roomParts) < 2 || roomParts[0] != "group" {
			log.Printf("Format de room invalide pour un message : %s", msg.Room)
			return
		}

		groupID, err := strconv.Atoi(roomParts[1])
		if err != nil {
			log.Printf("ID de groupe invalide dans la room %s : %v", msg.Room, err)
			return
		}

		// Pour un message permanent, on passe par le message-service
		protoReq := &apiv1.SendMessageRequest{
			GroupId:  int32(groupID),
			SenderId: msg.User,
			Content:  msg.Content,
		}

		protoData, err := proto.Marshal(protoReq)
		if err != nil {
			log.Printf("Erreur marshal proto : %v", err)
			return
		}

		err = h.nats.Publish("NEW_MESSAGE", protoData)
		if err != nil {
			log.Printf("Erreur publication sur NATS (NEW_MESSAGE) : %v", err)
		}

	case models.WSActionTyping:
		userId, _ := socket.Session().Load("userId")
		if userId != nil {
			msg.User = userId.(string)
		}
		finalPayload, _ := json.Marshal(msg)
		_ = h.nats.Publish("message.broadcast."+msg.Room, finalPayload)

	default:
		log.Printf("Action inconnue : %s", msg.Action)
	}
}
