package nats

import (
	"log"
	"strings"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/service"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const (
	errorCodeBadRequest = "BAD_REQUEST"
	errorCodeInternal   = "INTERNAL"
)

func NewMessageHandler(svc *service.MessageService) *Handler {
	return &Handler{svc: svc}
}

type Handler struct {
	svc *service.MessageService
}

func (h *Handler) handleSendMessage(msg *nats.Msg) {
	var req apiv1.SendMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetGroupId() == 0 {
		h.respondError(msg, errorCodeBadRequest, "group_id required")
		return
	}
	senderID, err := uuid.Parse(req.GetSenderId())
	if err != nil || req.GetSenderId() == "" {
		h.respondError(msg, errorCodeBadRequest, "sender_id required")
		return
	}
	if req.GetContent() == "" {
		h.respondError(msg, errorCodeBadRequest, "content required")
		return
	}

	chatMsg := &models.ChatMessage{
		SenderID:   senderID,
		GroupID:    int(req.GetGroupId()),
		Content:    req.GetContent(),
		Attachment: req.GetAttachment(),
	}

	result, err := h.svc.SendMessage(chatMsg)
	if err != nil {
		code := errorCodeInternal
		if strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "too long") {
			code = errorCodeBadRequest
		}
		h.respondError(msg, code, err.Error())
		return
	}

	resp := &apiv1.SendMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(result),
	}
	data, _ := proto.Marshal(resp)
	if err := msg.Respond(data); err != nil {
		log.Printf("respond: %v", err)
	}
}

func (h *Handler) handleGetMessage(msg *nats.Msg) {
	var req apiv1.GetMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGetMessageError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetId() == 0 {
		h.respondGetMessageError(msg, errorCodeBadRequest, "id required")
		return
	}

	result, err := h.svc.GetMessageById(int(req.GetId()))
	if err != nil {
		h.respondGetMessageError(msg, errorCodeInternal, err.Error())
		return
	}
	if result == nil {
		h.respondGetMessageError(msg, errorCodeBadRequest, "message not found")
		return
	}

	resp := &apiv1.GetMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(result),
	}
	data, _ := proto.Marshal(resp)
	if err := msg.Respond(data); err != nil {
		log.Printf("respond: %v", err)
	}
}

func (h *Handler) handleListMessages(msg *nats.Msg) {
	var req apiv1.ListMessagesRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondListMessagesError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetGroupId() == 0 {
		h.respondListMessagesError(msg, errorCodeBadRequest, "group_id required")
		return
	}

	result, err := h.svc.GetMessagesByGroupId(int(req.GetGroupId()))
	if err != nil {
		h.respondListMessagesError(msg, errorCodeInternal, err.Error())
		return
	}

	resp := &apiv1.ListMessagesResponse{
		Ok:   true,
		Data: chatMessagesToProto(result),
	}
	data, _ := proto.Marshal(resp)
	if err := msg.Respond(data); err != nil {
		log.Printf("respond: %v", err)
	}
}

func (h *Handler) handleUpdateMessage(msg *nats.Msg) {
	var req apiv1.UpdateMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondUpdateMessageError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetId() == 0 {
		h.respondUpdateMessageError(msg, errorCodeBadRequest, "id required")
		return
	}

	if req.GetContent() == "" {
		h.respondUpdateMessageError(msg, errorCodeBadRequest, "content required")
		return
	}

	result, err := h.svc.UpdateMessageById(int(req.GetId()), req.GetContent())
	if err != nil {
		code := errorCodeInternal
		if strings.Contains(err.Error(), "not found") {
			code = errorCodeBadRequest
		} else if strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "too long") {
			code = errorCodeBadRequest
		}
		h.respondUpdateMessageError(msg, code, err.Error())
		return
	}

	resp := &apiv1.UpdateMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(result),
	}
	data, _ := proto.Marshal(resp)
	if err := msg.Respond(data); err != nil {
		log.Printf("respond: %v", err)
	}
}

func (h *Handler) handleDeleteMessage(msg *nats.Msg) {
	var req apiv1.DeleteMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondDeleteMessageError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetId() == 0 {
		h.respondDeleteMessageError(msg, errorCodeBadRequest, "id required")
		return
	}

	err := h.svc.DeleteMessageById(int(req.GetId()))
	if err != nil {
		code := errorCodeInternal
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "empty") {
			code = errorCodeBadRequest
		}
		h.respondDeleteMessageError(msg, code, err.Error())
		return
	}

	resp := &apiv1.DeleteMessageResponse{Ok: true}
	data, _ := proto.Marshal(resp)
	if err := msg.Respond(data); err != nil {
		log.Printf("respond: %v", err)
	}
}

func chatMessagesToProto(messages []*models.ChatMessage) []*apiv1.ChatMessage {
	if messages == nil {
		return nil
	}
	var protoMessages []*apiv1.ChatMessage
	for _, message := range messages {
		protoMessages = append(protoMessages, chatMessageToProto(message))
	}
	return protoMessages
}

func chatMessageToProto(m *models.ChatMessage) *apiv1.ChatMessage {
	if m == nil {
		return nil
	}
	return &apiv1.ChatMessage{
		Id:        int32(m.ID),
		SenderId:  m.SenderID.String(),
		GroupId:   int32(m.GroupID),
		Content:   m.Content,
		Attachment: m.Attachment,
		CreatedAt: m.CreatedAt.Unix(),
		UpdatedAt: m.UpdatedAt.Unix(),
	}
}

func (h *Handler) respondError(msg *nats.Msg, code, text string) {
	resp := &apiv1.SendMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	}
	data, _ := proto.Marshal(resp)
	_ = msg.Respond(data)
}

func (h *Handler) respondGetMessageError(msg *nats.Msg, code, text string) {
	resp := &apiv1.GetMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	}
	data, _ := proto.Marshal(resp)
	_ = msg.Respond(data)
}

func (h *Handler) respondListMessagesError(msg *nats.Msg, code, text string) {
	resp := &apiv1.ListMessagesResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	}
	data, _ := proto.Marshal(resp)
	_ = msg.Respond(data)
}

func (h *Handler) respondUpdateMessageError(msg *nats.Msg, code, text string) {
	resp := &apiv1.UpdateMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	}
	data, _ := proto.Marshal(resp)
	_ = msg.Respond(data)
}

func (h *Handler) respondDeleteMessageError(msg *nats.Msg, code, text string) {
	resp := &apiv1.DeleteMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	}
	data, _ := proto.Marshal(resp)
	_ = msg.Respond(data)
}

func (h *Handler) Listen(nc *nats.Conn) error {
	if _, err := nc.QueueSubscribe("NEW_MESSAGE", "message", h.handleSendMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe("GET_MESSAGE", "message", h.handleGetMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe("LIST_MESSAGES", "message", h.handleListMessages); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe("UPDATE_MESSAGE", "message", h.handleUpdateMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe("DELETE_MESSAGE", "message", h.handleDeleteMessage); err != nil {
		return err
	}
	return nil
}
