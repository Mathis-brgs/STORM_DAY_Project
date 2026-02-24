package nats

import (
	"log"
	"strings"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/service"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/nats.go"
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

func (h *Handler) handleMessage(msg *nats.Msg) {
	var req apiv1.SendMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetGroupId() == 0 {
		h.respondError(msg, errorCodeBadRequest, "group_id required")
		return
	}
	if req.GetSenderId() == 0 {
		h.respondError(msg, errorCodeBadRequest, "sender_id required")
		return
	}
	if req.GetContent() == "" {
		h.respondError(msg, errorCodeBadRequest, "content required")
		return
	}

	chatMsg := &models.ChatMessage{
		SenderID: int(req.GetSenderId()),
		Content:  req.GetContent(),
		GroupID:  int(req.GetGroupId()),
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

func chatMessageToProto(m *models.ChatMessage) *apiv1.ChatMessage {
	if m == nil {
		return nil
	}
	return &apiv1.ChatMessage{
		Id:        int32(m.ID),
		SenderId:  int32(m.SenderID),
		GroupId:   int32(m.GroupID),
		Content:   m.Content,
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

func (h *Handler) Listen(nc *nats.Conn) error {
	_, err := nc.QueueSubscribe("NEW_MESSAGE", "message-service", h.handleMessage)
	return err
}
