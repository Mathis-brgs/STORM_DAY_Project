package nats

import (
	"errors"
	"strings"
	"testing"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo/memory"
	"github.com/Mathis-brgs/storm-project/services/message/internal/service"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

var (
	lot6OwnerID    = uuid.MustParse("b2000001-0000-0000-0000-000000000001")
	lot6AdminID    = uuid.MustParse("b2000002-0000-0000-0000-000000000002")
	lot6MemberID   = uuid.MustParse("b2000003-0000-0000-0000-000000000003")
	lot6Member2ID  = uuid.MustParse("b2000004-0000-0000-0000-000000000004")
	lot6ExternalID = uuid.MustParse("b2000005-0000-0000-0000-000000000005")
)

type lot6Fixture struct {
	handler         *Handler
	messageSvc      *service.MessageService
	conversationSvc *service.ConversationService
	conversationID  int
}

func TestHandlerLot6GroupMembershipPermissions(t *testing.T) {
	fix := newLot6Fixture(t)

	dispatchNATSHandler(
		t,
		&apiv1.GroupAddMemberRequest{
			ActorId:        lot6MemberID.String(),
			ConversationId: int32(fix.conversationID),
			UserId:         lot6ExternalID.String(),
			Role:           int32(models.ConversationRoleMember),
		},
		fix.handler.handleGroupAddMember,
	)

	isMember, err := fix.conversationSvc.IsMember(lot6ExternalID, fix.conversationID)
	if err != nil {
		t.Fatalf("IsMember(external) error = %v", err)
	}
	if isMember {
		t.Fatalf("member actor should not be able to add a new member")
	}

	dispatchNATSHandler(
		t,
		&apiv1.GroupAddMemberRequest{
			ActorId:        lot6OwnerID.String(),
			ConversationId: int32(fix.conversationID),
			UserId:         lot6ExternalID.String(),
			Role:           int32(models.ConversationRoleMember),
		},
		fix.handler.handleGroupAddMember,
	)

	isMember, err = fix.conversationSvc.IsMember(lot6ExternalID, fix.conversationID)
	if err != nil {
		t.Fatalf("IsMember(external) error = %v", err)
	}
	if !isMember {
		t.Fatalf("owner actor should be able to add a new member")
	}
}

func TestHandlerLot6GroupRoleAndDeletePermissions(t *testing.T) {
	fix := newLot6Fixture(t)

	dispatchNATSHandler(
		t,
		&apiv1.GroupUpdateRoleRequest{
			ActorId:        lot6AdminID.String(),
			ConversationId: int32(fix.conversationID),
			UserId:         lot6MemberID.String(),
			Role:           int32(models.ConversationRoleAdmin),
		},
		fix.handler.handleGroupUpdateRole,
	)

	if gotRole := memberRoleForUser(t, fix.conversationSvc, fix.conversationID, lot6MemberID); gotRole != models.ConversationRoleMember {
		t.Fatalf("admin should not be able to promote roles, got role=%d", gotRole)
	}

	dispatchNATSHandler(
		t,
		&apiv1.GroupUpdateRoleRequest{
			ActorId:        lot6OwnerID.String(),
			ConversationId: int32(fix.conversationID),
			UserId:         lot6MemberID.String(),
			Role:           int32(models.ConversationRoleAdmin),
		},
		fix.handler.handleGroupUpdateRole,
	)

	if gotRole := memberRoleForUser(t, fix.conversationSvc, fix.conversationID, lot6MemberID); gotRole != models.ConversationRoleAdmin {
		t.Fatalf("owner should be able to update roles, got role=%d", gotRole)
	}

	dispatchNATSHandler(
		t,
		&apiv1.GroupRemoveMemberRequest{
			ActorId:        lot6AdminID.String(),
			ConversationId: int32(fix.conversationID),
			UserId:         lot6OwnerID.String(),
		},
		fix.handler.handleGroupRemoveMember,
	)

	ownerStillMember, err := fix.conversationSvc.IsMember(lot6OwnerID, fix.conversationID)
	if err != nil {
		t.Fatalf("IsMember(owner) error = %v", err)
	}
	if !ownerStillMember {
		t.Fatalf("admin should not be able to remove owner")
	}

	dispatchNATSHandler(
		t,
		&apiv1.GroupDeleteRequest{
			ActorId:        lot6AdminID.String(),
			ConversationId: int32(fix.conversationID),
		},
		fix.handler.handleGroupDelete,
	)
	if _, err := fix.conversationSvc.GetConversationByID(fix.conversationID); err != nil {
		t.Fatalf("admin delete should not delete conversation, got %v", err)
	}

	dispatchNATSHandler(
		t,
		&apiv1.GroupDeleteRequest{
			ActorId:        lot6OwnerID.String(),
			ConversationId: int32(fix.conversationID),
		},
		fix.handler.handleGroupDelete,
	)
	if _, err := fix.conversationSvc.GetConversationByID(fix.conversationID); !errors.Is(err, repo.ErrConversationNotFound) {
		t.Fatalf("owner delete should remove conversation, got %v", err)
	}
}

func TestHandlerLot6MessageCRUDMembershipGuards(t *testing.T) {
	fix := newLot6Fixture(t)

	dispatchNATSHandler(
		t,
		&apiv1.SendMessageRequest{
			ConversationId: int32(fix.conversationID),
			SenderId:       lot6MemberID.String(),
			Content:        "initial message",
		},
		fix.handler.handleSendMessage,
	)

	messages, err := fix.messageSvc.GetMessagesByConversationID(fix.conversationID)
	if err != nil {
		t.Fatalf("GetMessagesByConversationID() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message after valid sender publish, got %d", len(messages))
	}
	messageID := messages[0].ID

	dispatchNATSHandler(
		t,
		&apiv1.SendMessageRequest{
			ConversationId: int32(fix.conversationID),
			SenderId:       lot6ExternalID.String(),
			Content:        "forbidden message",
		},
		fix.handler.handleSendMessage,
	)

	messages, err = fix.messageSvc.GetMessagesByConversationID(fix.conversationID)
	if err != nil {
		t.Fatalf("GetMessagesByConversationID() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("forbidden sender should not persist new message, got %d", len(messages))
	}

	dispatchNATSHandler(
		t,
		&apiv1.UpdateMessageRequest{
			Id:      int32(messageID),
			ActorId: lot6ExternalID.String(),
			Content: "hacked content",
		},
		fix.handler.handleUpdateMessage,
	)

	message, err := fix.messageSvc.GetMessageById(messageID)
	if err != nil {
		t.Fatalf("GetMessageById() error = %v", err)
	}
	if message.Content != "initial message" {
		t.Fatalf("forbidden actor should not update message content, got %q", message.Content)
	}

	dispatchNATSHandler(
		t,
		&apiv1.UpdateMessageRequest{
			Id:      int32(messageID),
			ActorId: lot6AdminID.String(),
			Content: "moderated content",
		},
		fix.handler.handleUpdateMessage,
	)

	message, err = fix.messageSvc.GetMessageById(messageID)
	if err != nil {
		t.Fatalf("GetMessageById() error = %v", err)
	}
	if message.Content != "moderated content" {
		t.Fatalf("admin should be able to update message content, got %q", message.Content)
	}

	dispatchNATSHandler(
		t,
		&apiv1.DeleteMessageRequest{
			Id:      int32(messageID),
			ActorId: lot6ExternalID.String(),
		},
		fix.handler.handleDeleteMessage,
	)

	if _, err := fix.messageSvc.GetMessageById(messageID); err != nil {
		t.Fatalf("forbidden actor should not delete message, got %v", err)
	}

	dispatchNATSHandler(
		t,
		&apiv1.DeleteMessageRequest{
			Id:      int32(messageID),
			ActorId: lot6OwnerID.String(),
		},
		fix.handler.handleDeleteMessage,
	)

	if _, err := fix.messageSvc.GetMessageById(messageID); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("owner should be able to delete message, got %v", err)
	}
}

func TestHandlerLot6AckMessageReceipt(t *testing.T) {
	fix := newLot6Fixture(t)

	created, err := fix.messageSvc.SendMessage(&models.ChatMessage{
		SenderID:       lot6MemberID,
		ConversationID: fix.conversationID,
		Content:        "receipt target",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	dispatchNATSHandler(
		t,
		&apiv1.AckMessageRequest{
			Id:      int32(created.ID),
			ActorId: lot6ExternalID.String(),
		},
		fix.handler.handleAckMessage,
	)

	if _, err := fix.messageSvc.GetMessageReceiptByID(created.ID, lot6ExternalID); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("external actor should not be able to create a receipt, got %v", err)
	}

	dispatchNATSHandler(
		t,
		&apiv1.AckMessageRequest{
			Id:         int32(created.ID),
			ActorId:    lot6AdminID.String(),
			ReceivedAt: 1710000000,
		},
		fix.handler.handleAckMessage,
	)

	adminReceipt, err := fix.messageSvc.GetMessageReceiptByID(created.ID, lot6AdminID)
	if err != nil {
		t.Fatalf("GetMessageReceiptByID(admin) error = %v", err)
	}
	if got := adminReceipt.ReceivedAt.Unix(); got != 1710000000 {
		t.Fatalf("expected received_at=1710000000, got %d", got)
	}

	dispatchNATSHandler(
		t,
		&apiv1.AckMessageRequest{
			Id:         int32(created.ID),
			ActorId:    lot6AdminID.String(),
			ReceivedAt: 1710000100,
		},
		fix.handler.handleAckMessage,
	)

	adminReceipt, err = fix.messageSvc.GetMessageReceiptByID(created.ID, lot6AdminID)
	if err != nil {
		t.Fatalf("GetMessageReceiptByID(admin) error = %v", err)
	}
	if got := adminReceipt.ReceivedAt.Unix(); got != 1710000000 {
		t.Fatalf("received_at should remain first ack timestamp per user, got %d", got)
	}

	dispatchNATSHandler(
		t,
		&apiv1.AckMessageRequest{
			Id:         int32(created.ID),
			ActorId:    lot6OwnerID.String(),
			ReceivedAt: 1710000100,
		},
		fix.handler.handleAckMessage,
	)

	ownerReceipt, err := fix.messageSvc.GetMessageReceiptByID(created.ID, lot6OwnerID)
	if err != nil {
		t.Fatalf("GetMessageReceiptByID(owner) error = %v", err)
	}
	if got := ownerReceipt.ReceivedAt.Unix(); got != 1710000100 {
		t.Fatalf("expected owner receipt to be independent, got %d", got)
	}
}

func TestMapConversationErrorLot6(t *testing.T) {
	if got := mapConversationError(service.ErrForbidden); got != errorCodeForbidden {
		t.Fatalf("mapConversationError(forbidden) expected %s, got %s", errorCodeForbidden, got)
	}
	if got := mapConversationError(service.ErrLastOwnerGuard); got != errorCodeConflict {
		t.Fatalf("mapConversationError(last owner guard) expected %s, got %s", errorCodeConflict, got)
	}
	if got := mapConversationError(repo.ErrConversationNotFound); got != errorCodeNotFound {
		t.Fatalf("mapConversationError(conversation not found) expected %s, got %s", errorCodeNotFound, got)
	}
}

func newLot6Fixture(t *testing.T) *lot6Fixture {
	t.Helper()

	messageRepo := memory.NewMessageRepo()
	conversationRepo := memory.NewConversationRepo()
	messageSvc := service.NewMessageService(messageRepo)
	conversationSvc := service.NewConversationService(conversationRepo)
	handler := NewMessageHandler(messageSvc, conversationSvc)

	conversation, err := conversationSvc.CreateConversation(lot6OwnerID, "Lot6 validation", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if _, err := conversationSvc.AddMember(lot6OwnerID, conversation.ID, lot6AdminID, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("AddMember(admin) error = %v", err)
	}
	if _, err := conversationSvc.AddMember(lot6OwnerID, conversation.ID, lot6MemberID, models.ConversationRoleMember); err != nil {
		t.Fatalf("AddMember(member1) error = %v", err)
	}
	if _, err := conversationSvc.AddMember(lot6OwnerID, conversation.ID, lot6Member2ID, models.ConversationRoleMember); err != nil {
		t.Fatalf("AddMember(member2) error = %v", err)
	}

	return &lot6Fixture{
		handler:         handler,
		messageSvc:      messageSvc,
		conversationSvc: conversationSvc,
		conversationID:  conversation.ID,
	}
}

func memberRoleForUser(t *testing.T, conversationSvc *service.ConversationService, conversationID int, userID uuid.UUID) models.ConversationRole {
	t.Helper()

	members, err := conversationSvc.ListMembers(lot6OwnerID, conversationID)
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}
	for _, member := range members {
		if member.UserID == userID {
			return member.Role
		}
	}

	t.Fatalf("member %s not found in conversation %d", userID, conversationID)
	return models.ConversationRoleMember
}

func dispatchNATSHandler(t *testing.T, request proto.Message, handlerFunc func(*nats.Msg)) {
	t.Helper()

	data, err := proto.Marshal(request)
	if err != nil {
		t.Fatalf("proto.Marshal() error = %v", err)
	}

	handlerFunc(&nats.Msg{Data: data})
}
