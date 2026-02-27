package ws

import (
	"encoding/json"
	"errors"
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
		if isConversationRoom(msg.Room) && !h.canJoinConversationRoom(socket, msg.Room) {
			log.Printf("Acces refuse a la room %s", msg.Room)
			return
		}
		h.hub.Join(msg.Room, socket)
		socket.Session().Store("room", msg.Room)

	case models.WSActionMessage:
		userId, _ := socket.Session().Load("userId")
		// Sécurité : on impose l'ID de l'utilisateur authentifié
		if userId != nil {
			msg.User = userId.(string)
		}

		// Compat room: conversation:<id> (nouveau) ou group:<id> (legacy).
		conversationID, err := parseConversationRoomID(msg.Room)
		if err != nil {
			log.Printf("Format de room invalide pour un message : %s", msg.Room)
			return
		}

		// Pour un message permanent, on passe par le message-service
		protoReq := &apiv1.SendMessageRequest{
			GroupId:        int32(conversationID),
			ConversationId: int32(conversationID),
			SenderId:       msg.User,
			Content:        msg.Content,
		}

		protoData, err := proto.Marshal(protoReq)
		if err != nil {
			log.Printf("Erreur marshal proto : %v", err)
			return
		}

		if err := h.nats.Publish("NEW_MESSAGE", protoData); err != nil {
			log.Printf("Erreur publication sur NATS (NEW_MESSAGE) : %v", err)
		}

		// Echo local pour que l'emetteur (et les clients de la room locale) reçoive le message instantanement.
		finalPayload, _ := json.Marshal(msg)
		h.hub.BroadcastToRoom(msg.Room, finalPayload)

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

func parseConversationRoomID(room string) (int, error) {
	roomParts := strings.Split(room, ":")
	if len(roomParts) < 2 {
		return 0, errors.New("invalid room format")
	}
	if roomParts[0] != "group" && roomParts[0] != "conversation" {
		return 0, errors.New("unsupported room prefix")
	}
	conversationID, err := strconv.Atoi(roomParts[1])
	if err != nil {
		return 0, err
	}
	if conversationID <= 0 {
		return 0, errors.New("invalid conversation id")
	}
	return conversationID, nil
}

func isConversationRoom(room string) bool {
	return strings.HasPrefix(room, "group:") || strings.HasPrefix(room, "conversation:")
}

func (h *Handler) canJoinConversationRoom(socket Socket, room string) bool {
	userIDRaw, ok := socket.Session().Load("userId")
	if !ok {
		return false
	}
	userID, ok := userIDRaw.(string)
	if !ok || strings.TrimSpace(userID) == "" {
		return false
	}

	conversationID, err := parseConversationRoomID(room)
	if err != nil {
		return false
	}

	protoReq := &apiv1.GroupGetRequest{
		ActorId:        userID,
		ConversationId: int32(conversationID),
		GroupId:        int32(conversationID),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		log.Printf("Erreur marshal GROUP_GET: %v", err)
		return false
	}

	reply, err := h.nats.Request("GROUP_GET", data, 3*time.Second)
	if err != nil {
		log.Printf("Erreur request GROUP_GET: %v", err)
		return false
	}

	var resp apiv1.GroupGetResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		log.Printf("Erreur unmarshal GROUP_GET: %v", err)
		return false
	}

	return resp.GetOk()
}
