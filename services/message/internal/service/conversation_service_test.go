package service

import (
	"errors"
	"testing"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo/memory"
	"github.com/google/uuid"
)

var (
	testUserOwner  = uuid.MustParse("a0000001-0000-0000-0000-000000000001")
	testUserAdmin  = uuid.MustParse("a0000002-0000-0000-0000-000000000002")
	testUserMember = uuid.MustParse("a0000003-0000-0000-0000-000000000003")
	testUserOther  = uuid.MustParse("a0000004-0000-0000-0000-000000000004")
)

func TestConversationServiceCreateConversationCreatesOwnerMembership(t *testing.T) {
	svc := NewConversationService(memory.NewConversationRepo())

	conversation, err := svc.CreateConversation(testUserOwner, "  Team Alpha  ", "https://cdn.example.com/team.png")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if conversation.ID == 0 {
		t.Fatalf("expected persisted conversation ID, got %d", conversation.ID)
	}
	if conversation.Name != "Team Alpha" {
		t.Fatalf("expected trimmed name Team Alpha, got %q", conversation.Name)
	}
	if conversation.CreatedBy != testUserOwner {
		t.Fatalf("expected created_by %s, got %s", testUserOwner, conversation.CreatedBy)
	}

	members, err := svc.ListMembers(testUserOwner, conversation.ID)
	if err != nil {
		t.Fatalf("ListMembers() error = %v", err)
	}
	if len(members) != 1 {
		t.Fatalf("expected 1 member, got %d", len(members))
	}
	if members[0].UserID != testUserOwner || members[0].Role != models.ConversationRoleOwner {
		t.Fatalf("expected owner membership for %s, got %+v", testUserOwner, members[0])
	}
}

func TestConversationServiceAddMemberRoleRules(t *testing.T) {
	svc := NewConversationService(memory.NewConversationRepo())

	conversation, err := svc.CreateConversation(testUserOwner, "Rules", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserAdmin, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("owner should be able to add admin, got %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserMember, models.ConversationRoleMember); err != nil {
		t.Fatalf("owner should be able to add member, got %v", err)
	}

	if _, err := svc.AddMember(testUserMember, conversation.ID, testUserOther, models.ConversationRoleMember); !errors.Is(err, ErrForbidden) {
		t.Fatalf("member add should be forbidden, got %v", err)
	}
	if _, err := svc.AddMember(testUserAdmin, conversation.ID, testUserOther, models.ConversationRoleAdmin); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin add admin should be forbidden, got %v", err)
	}
	if _, err := svc.AddMember(testUserAdmin, conversation.ID, testUserOther, models.ConversationRoleMember); err != nil {
		t.Fatalf("admin should add member, got %v", err)
	}
}

func TestConversationServiceLeaveConversationBlocksLastOwner(t *testing.T) {
	svc := NewConversationService(memory.NewConversationRepo())

	conversation, err := svc.CreateConversation(testUserOwner, "Owners", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	if err := svc.LeaveConversation(testUserOwner, conversation.ID); !errors.Is(err, ErrLastOwnerGuard) {
		t.Fatalf("expected last owner guard, got %v", err)
	}

	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserAdmin, models.ConversationRoleOwner); err != nil {
		t.Fatalf("owner should be able to add another owner, got %v", err)
	}
	if err := svc.LeaveConversation(testUserOwner, conversation.ID); err != nil {
		t.Fatalf("leave with two owners should pass, got %v", err)
	}

	ok, err := svc.IsMember(testUserOwner, conversation.ID)
	if err != nil {
		t.Fatalf("IsMember() error = %v", err)
	}
	if ok {
		t.Fatalf("expected owner to be removed from memberships")
	}
}

func TestConversationServiceUpdateMemberRoleRequiresOwner(t *testing.T) {
	svc := NewConversationService(memory.NewConversationRepo())

	conversation, err := svc.CreateConversation(testUserOwner, "Roles", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserAdmin, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("add admin failed: %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserMember, models.ConversationRoleMember); err != nil {
		t.Fatalf("add member failed: %v", err)
	}

	if _, err := svc.UpdateMemberRole(testUserAdmin, conversation.ID, testUserMember, models.ConversationRoleAdmin); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin should not update roles, got %v", err)
	}
	if _, err := svc.UpdateMemberRole(testUserOwner, conversation.ID, testUserOwner, models.ConversationRoleAdmin); !errors.Is(err, ErrLastOwnerGuard) {
		t.Fatalf("expected last owner guard when demoting self, got %v", err)
	}

	if _, err := svc.UpdateMemberRole(testUserOwner, conversation.ID, testUserAdmin, models.ConversationRoleOwner); err != nil {
		t.Fatalf("promote admin to owner failed: %v", err)
	}
	if _, err := svc.UpdateMemberRole(testUserOwner, conversation.ID, testUserOwner, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("demote owner with another owner present failed: %v", err)
	}
}

func TestConversationServiceRemoveMemberRoleRules(t *testing.T) {
	svc := NewConversationService(memory.NewConversationRepo())

	conversation, err := svc.CreateConversation(testUserOwner, "Remove", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserAdmin, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("add admin failed: %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserMember, models.ConversationRoleMember); err != nil {
		t.Fatalf("add member failed: %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserOther, models.ConversationRoleOwner); err != nil {
		t.Fatalf("add second owner failed: %v", err)
	}

	if err := svc.RemoveMember(testUserAdmin, conversation.ID, testUserMember); err != nil {
		t.Fatalf("admin should remove member, got %v", err)
	}
	if err := svc.RemoveMember(testUserAdmin, conversation.ID, testUserOther); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin removing owner should be forbidden, got %v", err)
	}

	if err := svc.RemoveMember(testUserOwner, conversation.ID, testUserOther); err != nil {
		t.Fatalf("owner should remove other owner when at least 2 owners, got %v", err)
	}
	if err := svc.RemoveMember(testUserOwner, conversation.ID, testUserOwner); !errors.Is(err, ErrLastOwnerGuard) {
		t.Fatalf("owner removing self as last owner should fail, got %v", err)
	}
}

func TestConversationServiceDeleteConversationRequiresOwner(t *testing.T) {
	svc := NewConversationService(memory.NewConversationRepo())

	conversation, err := svc.CreateConversation(testUserOwner, "Delete", "")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if _, err := svc.AddMember(testUserOwner, conversation.ID, testUserAdmin, models.ConversationRoleAdmin); err != nil {
		t.Fatalf("add admin failed: %v", err)
	}

	if err := svc.DeleteConversation(testUserAdmin, conversation.ID); !errors.Is(err, ErrForbidden) {
		t.Fatalf("admin delete should be forbidden, got %v", err)
	}
	if err := svc.DeleteConversation(testUserOwner, conversation.ID); err != nil {
		t.Fatalf("owner delete failed: %v", err)
	}

	if _, err := svc.GetConversationByID(conversation.ID); !errors.Is(err, repo.ErrConversationNotFound) {
		t.Fatalf("expected deleted conversation to be not found, got %v", err)
	}
}
