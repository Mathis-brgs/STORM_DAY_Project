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
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
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

		// On extrait l'ID du groupe de la room (ex: "group:123")
		// On suppose que le format est stable
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

		senderID, _ := strconv.Atoi(msg.User)

		// Pour un message permanent, on passe par le message-service
		// Le message-service se chargera de diffuser le message (broadcast) via NATS
		// une fois sauvegardé, ce qui évitera les doublons.
		protoReq := &apiv1.SendMessageRequest{
			GroupId:  int32(groupID),
			SenderId: int32(senderID),
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

	case models.ActionTyping:
		if userId != nil {
			msg.User = userId.(string)
		}
		finalPayload, _ := json.Marshal(msg)

		// On ne diffuse PLUS en local (hub.BroadcastToRoom) ici.
		// On utilise exclusivement NATS pour la diffusion (fan-out).
		// Le Hub écoute "message.broadcast.>" et redistribuera à tout le monde.
		_ = h.nats.Publish("message.broadcast."+msg.Room, finalPayload)

	default:
		log.Printf("Action inconnue : %s", msg.Action)
	}
}
