package message

import (
	"encoding/json"
	"gateway/internal/common"
	"gateway/internal/modules/auth"
	"io"
	"net/http"
	"strconv"
	"time"

	"gateway/internal/models"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/proto"
)

const (
	subjectNewMessage    = "NEW_MESSAGE"
	subjectGetMessage    = "GET_MESSAGE"
	subjectListMessages  = "LIST_MESSAGES"
	subjectUpdateMessage = "UPDATE_MESSAGE"
	subjectDeleteMessage = "DELETE_MESSAGE"
	subjectAckMessage    = "ACK_MESSAGE"

	subjectGroupCreate      = "GROUP_CREATE"
	subjectGroupGet         = "GROUP_GET"
	subjectGroupListForUser = "GROUP_LIST_FOR_USER"
	subjectGroupAddMember   = "GROUP_ADD_MEMBER"
	subjectGroupRemove      = "GROUP_REMOVE_MEMBER"
	subjectGroupListMembers = "GROUP_LIST_MEMBERS"
	subjectGroupUpdateRole  = "GROUP_UPDATE_ROLE"
	subjectGroupLeave       = "GROUP_LEAVE"
	subjectGroupDelete      = "GROUP_DELETE"

	requestTimeout = 5 * time.Second
)

const (
	invalidId = "invalid id"
)

type Handler struct {
	nc common.NatsConn
}

func NewHandler(nc common.NatsConn) *Handler {
	return &Handler{nc: nc}
}

// Send gère POST /api/messages - id row = int, sender_id = UUID, conversation_id (ou group_id legacy) = int.
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Valider le token une seule fois et récupérer l'identité de l'utilisateur
	actor := h.actorFromToken(r)
	if actor == nil {
		respondJSON(w, http.StatusUnauthorized, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "UNAUTHORIZED", Message: "invalid or missing token"},
		})
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}
	conversationID := resolveConversationID(req.ConversationID, req.GroupID)
	if conversationID == 0 {
		respondJSON(w, http.StatusBadRequest, models.SendMessageResponse{
			OK: false,
			Error: &models.SendMessageError{
				Code:    "BAD_REQUEST",
				Message: "conversation_id (or legacy group_id) required",
			},
		})
		return
	}

	// Sécurité : on impose l'ID de l'utilisateur authentifié comme sender
	protoReq := &apiv1.SendMessageRequest{
		GroupId:        int32(conversationID),
		SenderId:       actor.ID,
		Content:        req.Content,
		Attachment:     req.Attachment,
		ConversationId: int32(conversationID),
	}
	if req.ReplyToID != nil && *req.ReplyToID > 0 {
		protoReq.ReplyToId = int32(*req.ReplyToID)
	}
	if req.ForwardFromID != nil && *req.ForwardFromID > 0 {
		protoReq.ForwardFromId = int32(*req.ForwardFromID)
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	// Réponse immédiate au client — le publish NATS se fait en arrière-plan
	respondJSON(w, http.StatusAccepted, models.SendMessageResponse{OK: true})

	// Fire-and-forget dans une goroutine pour ne jamais bloquer le handler
	room := "conversation:" + strconv.Itoa(conversationID)
	go func() {
		_ = h.nc.Publish(subjectNewMessage, data)
		broadcast := models.InputMessage{
			Action:   models.WSActionMessage,
			Room:     room,
			User:     actor.ID,
			Username: actor.Username,
			Content:  req.Content,
		}
		if payload, err := json.Marshal(broadcast); err == nil {
			_ = h.nc.Publish("message.broadcast."+room, payload)
		}
	}()
}

// GetById gère GET /api/messages/{id} - id = int (PK row)
func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	protoReq := &apiv1.GetMessageRequest{Id: int32(id)}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGetMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GetMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GetMessageResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		mapped := toSendMessageData(resp.GetData())
		if mapped != nil {
			out.Data = &models.GetMessageData{
				ID:             mapped.ID,
				SenderID:       mapped.SenderID,
				SenderName:     mapped.SenderName,
				SenderUsername: mapped.SenderUsername,
				ConversationID: mapped.ConversationID,
				GroupID:        mapped.GroupID,
				Content:        mapped.Content,
				Attachment:     mapped.Attachment,
				ReceivedAt:     mapped.ReceivedAt,
				CreatedAt:      mapped.CreatedAt,
				UpdatedAt:      mapped.UpdatedAt,
				Status:         mapped.Status,
				ReplyTo:        mapped.ReplyTo,
				SeenBy:         mapped.SeenBy,
			}
			h.enrichSingleMessageData(out.Data)
		}
	}
	if resp.GetError() != nil {
		out.Error = &models.GetMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		if resp.GetError().GetCode() == "BAD_REQUEST" {
			status = http.StatusBadRequest
		} else {
			status = http.StatusNotFound
		}
	}
	respondJSON(w, status, out)
}

// GetByGroupId gère GET /api/messages?conversation_id= (ou ?group_id= legacy).
func (h *Handler) GetByGroupId(w http.ResponseWriter, r *http.Request) {
	conversationID, ok := queryConversationID(r)
	if !ok {
		respondJSON(w, http.StatusBadRequest, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "conversation_id (or legacy group_id) required"},
		})
		return
	}
	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.ListMessagesRequest{
		GroupId:        int32(conversationID),
		Limit:          100,
		Cursor:         r.URL.Query().Get("cursor"),
		ConversationId: int32(conversationID),
		ActorId:        actorID,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectListMessages, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.ListMessagesResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.ListMessagesResponse{
		OK:         resp.GetOk(),
		NextCursor: resp.GetNextCursor(),
		Data:       []models.SendMessageData{},
	}
	for _, d := range resp.GetData() {
		if mapped := toSendMessageData(d); mapped != nil {
			out.Data = append(out.Data, *mapped)
		}
	}
	h.enrichMessageListSenderNames(&out.Data)
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	var body models.UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}
	content := body.Content
	if content == "" {
		content = body.Message
	}
	if content == "" {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "content required"},
		})
		return
	}
	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.UpdateMessageRequest{
		Id:      int32(id),
		Content: content,
		ActorId: actorID,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectUpdateMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.UpdateMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.UpdateMessageResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		out.Data = toSendMessageData(resp.GetData())
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}

	// Contrat front : après PATCH réussi, broadcaster message_updated (aliases: message_edited, message_edit, updated).
	if resp.GetOk() && resp.GetData() != nil {
		convID := resp.GetData().GetConversationId()
		if convID == 0 {
			convID = resp.GetData().GetGroupId()
		}
		if convID > 0 {
			room := "conversation:" + strconv.Itoa(int(convID))
			payload, _ := json.Marshal(map[string]interface{}{
				"action":     "message_updated",
				"room":       room,
				"message_id": strconv.Itoa(int(resp.GetData().GetId())),
				"content":    resp.GetData().GetContent(),
			})
			_ = h.nc.Publish("message.broadcast."+room, payload)
		}
	}

	respondJSON(w, status, out)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}
	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.DeleteMessageRequest{
		Id:      int32(id),
		ActorId: actorID,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectDeleteMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.DeleteMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.DeleteMessageResponse{OK: resp.GetOk()}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusNotFound)
	}
	respondJSON(w, status, out)
}

func (h *Handler) AckReceipt(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.AckMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	var body models.AckMessageRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil && err != io.EOF {
			respondJSON(w, http.StatusBadRequest, models.AckMessageResponse{
				OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
			})
			return
		}
	}

	actorID := h.actorIDFromToken(r)
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, models.AckMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "actor_id (or user_id / X-User-ID) required"},
		})
		return
	}

	protoReq := &apiv1.AckMessageRequest{
		Id:         int32(id),
		ActorId:    actorID,
		ReceivedAt: body.ReceivedAt,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.AckMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectAckMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.AckMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.AckMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.AckMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.AckMessageResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		out.Data = toSendMessageData(resp.GetData())
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		status = statusFromServiceCode(resp.GetError().GetCode(), http.StatusUnprocessableEntity)
	}
	respondJSON(w, status, out)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func queryConversationID(r *http.Request) (int, bool) {
	if raw := r.URL.Query().Get("conversation_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 32)
		return int(id), err == nil && id > 0
	}
	if raw := r.URL.Query().Get("group_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 32)
		return int(id), err == nil && id > 0
	}
	return 0, false
}

func resolveConversationID(conversationID, legacyGroupID int) int {
	if conversationID > 0 {
		return conversationID
	}
	if legacyGroupID > 0 {
		return legacyGroupID
	}
	return 0
}

func (h *Handler) actorFromToken(r *http.Request) *auth.UserInfo {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	if token == "" {
		return nil
	}
	result, err := auth.ValidateToken(token)
	if err != nil || !result.IsValid {
		return nil
	}
	return &result.User
}

func (h *Handler) actorIDFromToken(r *http.Request) string {
	actor := h.actorFromToken(r)
	if actor == nil {
		return ""
	}
	return actor.ID
}


func statusFromServiceCode(code string, fallback int) int {
	switch code {
	case "BAD_REQUEST":
		return http.StatusBadRequest
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	default:
		return fallback
	}
}

func toSendMessageData(d *apiv1.ChatMessage) *models.SendMessageData {
	if d == nil {
		return nil
	}
	conversationID := int(d.GetConversationId())
	if conversationID == 0 {
		conversationID = int(d.GetGroupId())
	}
	out := &models.SendMessageData{
		ID:             int(d.GetId()),
		SenderID:       d.GetSenderId(),
		ConversationID: conversationID,
		GroupID:        conversationID,
		Content:        d.GetContent(),
		Attachment:     d.GetAttachment(),
		ReceivedAt:     d.GetReceivedAt(),
		CreatedAt:      d.GetCreatedAt(),
		UpdatedAt:      d.GetUpdatedAt(),
		Status:         d.GetStatus(),
	}
	if d.GetReplyTo() != nil {
		out.ReplyTo = &models.ReplyToData{
			ID:       int(d.GetReplyTo().GetId()),
			SenderID: d.GetReplyTo().GetSenderId(),
			Content:  d.GetReplyTo().GetContent(),
		}
	}
	for _, e := range d.GetSeenBy() {
		out.SeenBy = append(out.SeenBy, models.SeenByEntry{
			UserID:      e.GetUserId(),
			DisplayName: e.GetDisplayName(),
		})
	}
	return out
}

func (h *Handler) enrichMessageListSenderNames(data *[]models.SendMessageData) {
	if data == nil {
		return
	}
	for i := range *data {
		m := &(*data)[i]
		m.SenderName = h.fetchUsername(m.SenderID)
		m.SenderUsername = m.SenderName
		if m.ReplyTo != nil && m.ReplyTo.SenderID != "" {
			m.ReplyTo.SenderName = h.fetchUsername(m.ReplyTo.SenderID)
		}
	}
}

func (h *Handler) enrichSingleMessageData(d *models.GetMessageData) {
	if d == nil {
		return
	}
	d.SenderName = h.fetchUsername(d.SenderID)
	d.SenderUsername = d.SenderName
	if d.ReplyTo != nil && d.ReplyTo.SenderID != "" {
		d.ReplyTo.SenderName = h.fetchUsername(d.ReplyTo.SenderID)
	}
}
