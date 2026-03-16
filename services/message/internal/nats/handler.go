package nats

import (
	"errors"
	"log"
	"strings"
	"time"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/Mathis-brgs/storm-project/services/message/internal/service"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const (
	errorCodeBadRequest = "BAD_REQUEST"
	errorCodeNotFound   = "NOT_FOUND"
	errorCodeForbidden  = "FORBIDDEN"
	errorCodeConflict   = "CONFLICT"
	errorCodeInternal   = "INTERNAL"
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
)

func NewMessageHandler(svc *service.MessageService, conversationSvc *service.ConversationService) *Handler {
	return &Handler{
		svc:             svc,
		conversationSvc: conversationSvc,
	}
}

type Handler struct {
	svc             *service.MessageService
	conversationSvc *service.ConversationService
}

func (h *Handler) handleSendMessage(msg *nats.Msg) {
	var req apiv1.SendMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondSendMessageError(msg, errorCodeBadRequest, "invalid request format")
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondSendMessageError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}
	senderID, err := parseUUID("sender_id", req.GetSenderId())
	if err != nil {
		h.respondSendMessageError(msg, errorCodeBadRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.GetContent()) == "" {
		h.respondSendMessageError(msg, errorCodeBadRequest, "content required")
		return
	}
	if err := h.authorizeConversationMember(senderID, conversationID); err != nil {
		code := mapConversationError(err)
		h.respondSendMessageError(msg, code, err.Error())
		return
	}

	chatMsg := &models.ChatMessage{
		SenderID:       senderID,
		ConversationID: conversationID,
		Content:        req.GetContent(),
		Attachment:     req.GetAttachment(),
	}

	result, err := h.svc.SendMessage(chatMsg)
	if err != nil {
		code := mapMessageError(err)
		h.respondSendMessageError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.SendMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(result),
	})
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
		code := mapMessageError(err)
		h.respondGetMessageError(msg, code, err.Error())
		return
	}
	if result == nil {
		h.respondGetMessageError(msg, errorCodeNotFound, "message not found")
		return
	}

	h.respondProto(msg, &apiv1.GetMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(result),
	})
}

func (h *Handler) handleListMessages(msg *nats.Msg) {
	var req apiv1.ListMessagesRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondListMessagesError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondListMessagesError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}
	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondListMessagesError(msg, errorCodeBadRequest, err.Error())
		return
	}
	if err := h.authorizeConversationMember(actorID, conversationID); err != nil {
		code := mapConversationError(err)
		h.respondListMessagesError(msg, code, err.Error())
		return
	}

	result, err := h.svc.GetMessagesByConversationID(conversationID)
	if err != nil {
		code := mapMessageError(err)
		h.respondListMessagesError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.ListMessagesResponse{
		Ok:   true,
		Data: chatMessagesToProto(result),
	})
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
	if strings.TrimSpace(req.GetContent()) == "" {
		h.respondUpdateMessageError(msg, errorCodeBadRequest, "content required")
		return
	}
	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondUpdateMessageError(msg, errorCodeBadRequest, err.Error())
		return
	}

	existingMessage, err := h.svc.GetMessageById(int(req.GetId()))
	if err != nil {
		code := mapMessageError(err)
		h.respondUpdateMessageError(msg, code, err.Error())
		return
	}
	if existingMessage == nil {
		h.respondUpdateMessageError(msg, errorCodeNotFound, "message not found")
		return
	}
	if err := h.authorizeMessageMutation(actorID, existingMessage); err != nil {
		code := mapConversationError(err)
		h.respondUpdateMessageError(msg, code, err.Error())
		return
	}

	result, err := h.svc.UpdateMessageById(int(req.GetId()), req.GetContent())
	if err != nil {
		code := mapMessageError(err)
		h.respondUpdateMessageError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.UpdateMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(result),
	})
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
	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondDeleteMessageError(msg, errorCodeBadRequest, err.Error())
		return
	}

	existingMessage, err := h.svc.GetMessageById(int(req.GetId()))
	if err != nil {
		code := mapMessageError(err)
		h.respondDeleteMessageError(msg, code, err.Error())
		return
	}
	if existingMessage == nil {
		h.respondDeleteMessageError(msg, errorCodeNotFound, "message not found")
		return
	}
	if err := h.authorizeMessageMutation(actorID, existingMessage); err != nil {
		code := mapConversationError(err)
		h.respondDeleteMessageError(msg, code, err.Error())
		return
	}

	if err := h.svc.DeleteMessageById(int(req.GetId())); err != nil {
		code := mapMessageError(err)
		h.respondDeleteMessageError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.DeleteMessageResponse{Ok: true})
}

func (h *Handler) handleAckMessage(msg *nats.Msg) {
	var req apiv1.AckMessageRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondAckMessageError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	if req.GetId() == 0 {
		h.respondAckMessageError(msg, errorCodeBadRequest, "id required")
		return
	}
	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondAckMessageError(msg, errorCodeBadRequest, err.Error())
		return
	}

	existingMessage, err := h.svc.GetMessageById(int(req.GetId()))
	if err != nil {
		code := mapMessageError(err)
		h.respondAckMessageError(msg, code, err.Error())
		return
	}
	if existingMessage == nil {
		h.respondAckMessageError(msg, errorCodeNotFound, "message not found")
		return
	}
	if err := h.authorizeConversationMember(actorID, existingMessage.ConversationID); err != nil {
		code := mapConversationError(err)
		h.respondAckMessageError(msg, code, err.Error())
		return
	}

	receivedAt := time.Now()
	if req.GetReceivedAt() > 0 {
		receivedAt = time.Unix(req.GetReceivedAt(), 0).UTC()
	}

	receipt, err := h.svc.MarkMessageReceivedByID(int(req.GetId()), actorID, receivedAt)
	if err != nil {
		code := mapMessageError(err)
		h.respondAckMessageError(msg, code, err.Error())
		return
	}

	responseMessage := *existingMessage
	responseMessage.ReceivedAt = &receipt.ReceivedAt

	h.respondProto(msg, &apiv1.AckMessageResponse{
		Ok:   true,
		Data: chatMessageToProto(&responseMessage),
	})
}

func (h *Handler) handleGroupCreate(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupCreateError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupCreateRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupCreateError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupCreateError(msg, errorCodeBadRequest, err.Error())
		return
	}

	conversation, err := h.conversationSvc.CreateConversation(actorID, req.GetName(), req.GetAvatarUrl())
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupCreateError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupCreateResponse{
		Ok:   true,
		Data: conversationToProto(conversation),
	})
}

func (h *Handler) handleGroupGet(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupGetError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupGetRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupGetError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupGetError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupGetError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	conversation, err := h.conversationSvc.GetConversationByID(conversationID)
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupGetError(msg, code, err.Error())
		return
	}
	isMember, err := h.conversationSvc.IsMember(actorID, conversationID)
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupGetError(msg, code, err.Error())
		return
	}
	if !isMember {
		h.respondGroupGetError(msg, errorCodeForbidden, service.ErrForbidden.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupGetResponse{
		Ok:   true,
		Data: conversationToProto(conversation),
	})
}

func (h *Handler) handleGroupListForUser(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupListForUserError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupListForUserRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupListForUserError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		h.respondGroupListForUserError(msg, errorCodeBadRequest, err.Error())
		return
	}

	conversations, err := h.conversationSvc.ListConversationsByUser(userID)
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupListForUserError(msg, code, err.Error())
		return
	}

	data := make([]*apiv1.Group, 0, len(conversations))
	for _, conversation := range conversations {
		data = append(data, conversationToProto(conversation))
	}

	h.respondProto(msg, &apiv1.GroupListForUserResponse{
		Ok:   true,
		Data: data,
	})
}

func (h *Handler) handleGroupAddMember(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupAddMemberError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupAddMemberRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupAddMemberError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupAddMemberError(msg, errorCodeBadRequest, err.Error())
		return
	}
	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		h.respondGroupAddMemberError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupAddMemberError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	membership, err := h.conversationSvc.AddMember(
		actorID,
		conversationID,
		userID,
		models.ConversationRole(req.GetRole()),
	)
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupAddMemberError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupAddMemberResponse{
		Ok:   true,
		Data: membershipToProto(membership),
	})
}

func (h *Handler) handleGroupRemoveMember(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupRemoveMemberError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupRemoveMemberRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupRemoveMemberError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupRemoveMemberError(msg, errorCodeBadRequest, err.Error())
		return
	}
	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		h.respondGroupRemoveMemberError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupRemoveMemberError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	if err := h.conversationSvc.RemoveMember(actorID, conversationID, userID); err != nil {
		code := mapConversationError(err)
		h.respondGroupRemoveMemberError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupRemoveMemberResponse{Ok: true})
}

func (h *Handler) handleGroupListMembers(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupListMembersError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupListMembersRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupListMembersError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupListMembersError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupListMembersError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	memberships, err := h.conversationSvc.ListMembers(actorID, conversationID)
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupListMembersError(msg, code, err.Error())
		return
	}

	data := make([]*apiv1.GroupMember, 0, len(memberships))
	for _, membership := range memberships {
		data = append(data, membershipToProto(membership))
	}

	h.respondProto(msg, &apiv1.GroupListMembersResponse{
		Ok:   true,
		Data: data,
	})
}

func (h *Handler) handleGroupUpdateRole(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupUpdateRoleError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupUpdateRoleRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupUpdateRoleError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupUpdateRoleError(msg, errorCodeBadRequest, err.Error())
		return
	}
	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		h.respondGroupUpdateRoleError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupUpdateRoleError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	membership, err := h.conversationSvc.UpdateMemberRole(
		actorID,
		conversationID,
		userID,
		models.ConversationRole(req.GetRole()),
	)
	if err != nil {
		code := mapConversationError(err)
		h.respondGroupUpdateRoleError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupUpdateRoleResponse{
		Ok:   true,
		Data: membershipToProto(membership),
	})
}

func (h *Handler) handleGroupLeave(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupLeaveError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupLeaveRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupLeaveError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	userID, err := parseUUID("user_id", req.GetUserId())
	if err != nil {
		h.respondGroupLeaveError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupLeaveError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	if err := h.conversationSvc.LeaveConversation(userID, conversationID); err != nil {
		code := mapConversationError(err)
		h.respondGroupLeaveError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupLeaveResponse{Ok: true})
}

func (h *Handler) handleGroupDelete(msg *nats.Msg) {
	if h.conversationSvc == nil {
		h.respondGroupDeleteError(msg, errorCodeInternal, "conversation service unavailable")
		return
	}

	var req apiv1.GroupDeleteRequest
	if err := proto.Unmarshal(msg.Data, &req); err != nil {
		h.respondGroupDeleteError(msg, errorCodeBadRequest, "invalid request format")
		return
	}

	actorID, err := parseUUID("actor_id", req.GetActorId())
	if err != nil {
		h.respondGroupDeleteError(msg, errorCodeBadRequest, err.Error())
		return
	}
	conversationID := resolveConversationID(req.GetConversationId(), req.GetGroupId())
	if conversationID == 0 {
		h.respondGroupDeleteError(msg, errorCodeBadRequest, "conversation_id required")
		return
	}

	if err := h.conversationSvc.DeleteConversation(actorID, conversationID); err != nil {
		code := mapConversationError(err)
		h.respondGroupDeleteError(msg, code, err.Error())
		return
	}

	h.respondProto(msg, &apiv1.GroupDeleteResponse{Ok: true})
}

func (h *Handler) Listen(nc *nats.Conn) error {
	if _, err := nc.QueueSubscribe(subjectNewMessage, "message", h.handleSendMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGetMessage, "message", h.handleGetMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectListMessages, "message", h.handleListMessages); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectUpdateMessage, "message", h.handleUpdateMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectDeleteMessage, "message", h.handleDeleteMessage); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectAckMessage, "message", h.handleAckMessage); err != nil {
		return err
	}

	if _, err := nc.QueueSubscribe(subjectGroupCreate, "message", h.handleGroupCreate); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupGet, "message", h.handleGroupGet); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupListForUser, "message", h.handleGroupListForUser); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupAddMember, "message", h.handleGroupAddMember); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupRemove, "message", h.handleGroupRemoveMember); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupListMembers, "message", h.handleGroupListMembers); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupUpdateRole, "message", h.handleGroupUpdateRole); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupLeave, "message", h.handleGroupLeave); err != nil {
		return err
	}
	if _, err := nc.QueueSubscribe(subjectGroupDelete, "message", h.handleGroupDelete); err != nil {
		return err
	}

	return nil
}

func chatMessagesToProto(messages []*models.ChatMessage) []*apiv1.ChatMessage {
	if messages == nil {
		return nil
	}
	protoMessages := make([]*apiv1.ChatMessage, 0, len(messages))
	for _, message := range messages {
		protoMessages = append(protoMessages, chatMessageToProto(message))
	}
	return protoMessages
}

func chatMessageToProto(m *models.ChatMessage) *apiv1.ChatMessage {
	if m == nil {
		return nil
	}
	conversationID := int32(m.ConversationID)
	receivedAt := int64(0)
	if m.ReceivedAt != nil {
		receivedAt = m.ReceivedAt.Unix()
	}
	return &apiv1.ChatMessage{
		Id:             int32(m.ID),
		SenderId:       m.SenderID.String(),
		GroupId:        conversationID,
		Content:        m.Content,
		Attachment:     m.Attachment,
		CreatedAt:      m.CreatedAt.Unix(),
		UpdatedAt:      m.UpdatedAt.Unix(),
		ConversationId: conversationID,
		ReceivedAt:     receivedAt,
	}
}

func conversationToProto(c *models.Conversation) *apiv1.Group {
	if c == nil {
		return nil
	}
	createdBy := ""
	if c.CreatedBy != uuid.Nil {
		createdBy = c.CreatedBy.String()
	}
	return &apiv1.Group{
		Id:        int32(c.ID),
		Name:      c.Name,
		AvatarUrl: c.AvatarURL,
		CreatedBy: createdBy,
		CreatedAt: c.CreatedAt.Unix(),
		UpdatedAt: c.UpdatedAt.Unix(),
	}
}

func membershipToProto(m *models.ConversationMembership) *apiv1.GroupMember {
	if m == nil {
		return nil
	}
	conversationID := int32(m.ConversationID)
	return &apiv1.GroupMember{
		Id:             int32(m.ID),
		ConversationId: conversationID,
		GroupId:        conversationID,
		UserId:         m.UserID.String(),
		Role:           int32(m.Role),
		CreatedAt:      m.CreatedAt.Unix(),
	}
}

func resolveConversationID(conversationID, legacyGroupID int32) int {
	if conversationID > 0 {
		return int(conversationID)
	}
	if legacyGroupID > 0 {
		return int(legacyGroupID)
	}
	return 0
}

func parseUUID(fieldName, value string) (uuid.UUID, error) {
	if strings.TrimSpace(value) == "" {
		return uuid.Nil, errors.New(fieldName + " required")
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, errors.New(fieldName + " must be a valid UUID")
	}
	return id, nil
}

func (h *Handler) authorizeConversationMember(userID uuid.UUID, conversationID int) error {
	if h.conversationSvc == nil {
		return errors.New("conversation service unavailable")
	}

	isMember, err := h.conversationSvc.IsMember(userID, conversationID)
	if err != nil {
		return err
	}
	if !isMember {
		return service.ErrForbidden
	}

	return nil
}

func (h *Handler) authorizeMessageMutation(actorID uuid.UUID, message *models.ChatMessage) error {
	if message == nil {
		return errors.New("message not found")
	}
	if err := h.authorizeConversationMember(actorID, message.ConversationID); err != nil {
		return err
	}

	if actorID == message.SenderID {
		return nil
	}

	memberships, err := h.conversationSvc.ListMembers(actorID, message.ConversationID)
	if err != nil {
		return err
	}

	for _, membership := range memberships {
		if membership.UserID != actorID {
			continue
		}
		if membership.Role == models.ConversationRoleAdmin || membership.Role == models.ConversationRoleOwner {
			return nil
		}
		return service.ErrForbidden
	}

	return service.ErrForbidden
}

func mapMessageError(err error) string {
	if err == nil {
		return errorCodeInternal
	}
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "not found"):
		return errorCodeNotFound
	case strings.Contains(text, "empty"),
		strings.Contains(text, "too long"),
		strings.Contains(text, "invalid"):
		return errorCodeBadRequest
	default:
		return errorCodeInternal
	}
}

func mapConversationError(err error) string {
	if err == nil {
		return errorCodeInternal
	}
	switch {
	case errors.Is(err, service.ErrInvalidConversationID),
		errors.Is(err, service.ErrInvalidUserID),
		errors.Is(err, service.ErrInvalidConversation),
		errors.Is(err, service.ErrInvalidMembershipRole):
		return errorCodeBadRequest
	case errors.Is(err, service.ErrForbidden):
		return errorCodeForbidden
	case errors.Is(err, repo.ErrMembershipAlreadyExists),
		errors.Is(err, service.ErrLastOwnerGuard):
		return errorCodeConflict
	case errors.Is(err, repo.ErrConversationNotFound),
		errors.Is(err, repo.ErrMembershipNotFound):
		return errorCodeNotFound
	default:
		return errorCodeInternal
	}
}

func (h *Handler) respondProto(msg *nats.Msg, response proto.Message) {
	data, err := proto.Marshal(response)
	if err != nil {
		log.Printf("marshal response: %v", err)
		return
	}
	if err := msg.Respond(data); err != nil {
		log.Printf("respond: %v", err)
	}
}

func (h *Handler) respondSendMessageError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.SendMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGetMessageError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GetMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondListMessagesError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.ListMessagesResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondUpdateMessageError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.UpdateMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondDeleteMessageError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.DeleteMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondAckMessageError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.AckMessageResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupCreateError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupCreateResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupGetError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupGetResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupListForUserError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupListForUserResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupAddMemberError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupAddMemberResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupRemoveMemberError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupRemoveMemberResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupListMembersError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupListMembersResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupUpdateRoleError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupUpdateRoleResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupLeaveError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupLeaveResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}

func (h *Handler) respondGroupDeleteError(msg *nats.Msg, code, text string) {
	h.respondProto(msg, &apiv1.GroupDeleteResponse{
		Ok: false,
		Error: &apiv1.Error{
			Code:    code,
			Message: text,
		},
	})
}
