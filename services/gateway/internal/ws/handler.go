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

	wsActiveConnections.Inc()

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
	wsActiveConnections.Dec()
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
		// Ajouter le username pour que le front puisse l'afficher
		if username, ok := socket.Session().Load("username"); ok {
			msg.Username = username.(string)
		}

		// Compat room: conversation:<id> (nouveau) ou group:<id> (legacy).
		conversationID, err := parseConversationRoomID(msg.Room)
		if err != nil {
			log.Printf("Format de room invalide pour un message : %s", msg.Room)
			return
		}

		// Pour un message permanent, on passe par le message-service via Request/Reply
		// If the client included a base64 attachment, upload it first via NATS to media-service
		if msg.AttachmentBase64 != "" {
			uploadReq := struct {
				Filename    string `json:"filename"`
				ContentType string `json:"contentType"`
				Size        int64  `json:"size"`
				DataBase64  string `json:"dataBase64"`
			}{
				Filename:    msg.AttachmentFilename,
				ContentType: msg.AttachmentContentType,
				Size:        int64(len(msg.AttachmentBase64)),
				DataBase64:  msg.AttachmentBase64,
			}

			payload, err := json.Marshal(uploadReq)
			if err != nil {
				log.Printf("failed to marshal media upload request: %v", err)
				return
			}

			reply, err := h.nats.Request("media.upload.requested", payload, 10*time.Second)
			if err != nil {
				log.Printf("media upload request failed: %v", err)
				return
			}

			var mediaResp map[string]any
			if err := json.Unmarshal(reply.Data, &mediaResp); err != nil {
				log.Printf("invalid response from media service: %v", err)
				return
			}

			if errVal, ok := mediaResp["error"]; ok {
				log.Printf("media service error: %v", errVal)
				return
			}

			if id, ok := mediaResp["mediaId"].(string); ok {
				msg.Attachment = id
			} else if id, ok := mediaResp["MediaID"].(string); ok { // fallback
				msg.Attachment = id
			}
			// Clear base64 to avoid broadcasting heavy payloads
			msg.AttachmentBase64 = ""
			msg.AttachmentFilename = ""
			msg.AttachmentContentType = ""
		}

		// Pour un message permanent, on passe par le message-service via Request/Reply
		protoReq := &apiv1.SendMessageRequest{
			GroupId:        int32(conversationID),
			ConversationId: int32(conversationID),
			SenderId:       msg.User,
			Content:        msg.Content,
			Attachment:     msg.Attachment,
		}

		protoData, err := proto.Marshal(protoReq)
		if err != nil {
			log.Printf("Erreur marshal proto : %v", err)
			return
		}

		reply, err := h.nats.Request("NEW_MESSAGE", protoData, 5*time.Second)
		if err != nil {
			log.Printf("Erreur request NEW_MESSAGE : %v", err)
			return
		}

		var resp apiv1.SendMessageResponse
		if err := proto.Unmarshal(reply.Data, &resp); err != nil || !resp.GetOk() {
			log.Printf("Message non sauvegardé : %v", err)
			return
		}

		// Broadcast via NATS seulement si le message a été sauvegardé avec succès
		finalPayload, _ := json.Marshal(msg)
		if err := h.nats.Publish("message.broadcast."+msg.Room, finalPayload); err != nil {
			log.Printf("Erreur publication broadcast NATS : %v", err)
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
