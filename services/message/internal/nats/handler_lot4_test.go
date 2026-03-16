package nats

import (
	"errors"
	"testing"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo/memory"
	"github.com/Mathis-brgs/storm-project/services/message/internal/service"
	"github.com/google/uuid"
)

var (
	testOwnerID     = uuid.MustParse("a0000001-0000-0000-0000-000000000001")
	testAdminID     = uuid.MustParse("a0000002-0000-0000-0000-000000000002")
	testMemberID    = uuid.MustParse("a0000003-0000-0000-0000-000000000003")
	testMemberTwoID = uuid.MustParse("a0000004-0000-0000-0000-000000000004")
	testExternalID  = uuid.MustParse("a0000005-0000-0000-0000-000000000005")
)

func TestHandlerAuthorizeConversationMember(t *testing.T) {
	handler, conversationID := buildSecurityTestHandler(t)

	if err := handler.authorizeConversationMember(testMemberID, conversationID); err != nil {
		t.Fatalf("member should be authorized, got %v", err)
	}

	err := handler.authorizeConversationMember(testExternalID, conversationID)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("external user should be forbidden, got %v", err)
	}
}

func TestHandlerAuthorizeMessageMutation(t *testing.T) {
	handler, conversationID := buildSecurityTestHandler(t)

	message, err := handler.svc.SendMessage(&models.ChatMessage{
		SenderID:       testMemberID,
		ConversationID: conversationID,
		Content:        "Hello secure world",
	})
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}

	if err := handler.authorizeMessageMutation(testMemberID, message); err != nil {
		t.Fatalf("sender should be authorized to mutate, got %v", err)
	}
	if err := handler.authorizeMessageMutation(testAdminID, message); err != nil {
		t.Fatalf("admin should be authorized to mutate, got %v", err)
	}
	if err := handler.authorizeMessageMutation(testOwnerID, message); err != nil {
		t.Fatalf("owner should be authorized to mutate, got %v", err)
	}

	err = handler.authorizeMessageMutation(testMemberTwoID, message)
	if !errors.Is(err, service.ErrForbidden) {
		t.Fatalf("plain member not sender should be forbidden, got %v", err)
	}
}

func buildSecurityTestHandler(t *testing.T) (*Handler, int) {
	t.Helper()

	messageRepo := memory.NewMessageRepo()
	conversationRepo := memory.NewConversationRepo()
	messageSvc := service.NewMessageService(messageRepo)
	conversationSvc := service.NewConversationService(conversationRepo)

	conversation, err := conversationSvc.CreateConversation(testOwnerID, "Lot4 security", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	if _, err := conversationSvc.AddMember(testOwnerID, conversation.ID, testAdminID, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("AddMember(admin) error = %v", err)
	}
	if _, err := conversationSvc.AddMember(testOwnerID, conversation.ID, testMemberID, models.ConversationRoleMember); err != nil {
		t.Fatalf("AddMember(member) error = %v", err)
	}
	if _, err := conversationSvc.AddMember(testOwnerID, conversation.ID, testMemberTwoID, models.ConversationRoleMember); err != nil {
		t.Fatalf("AddMember(member2) error = %v", err)
	}

	return NewMessageHandler(messageSvc, conversationSvc), conversation.ID
}
