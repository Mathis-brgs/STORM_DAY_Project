package memory

import (
	"errors"
	"testing"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
)

var (
	repoOwnerID     = uuid.MustParse("b1000001-0000-0000-0000-000000000001")
	repoAdminID     = uuid.MustParse("b1000002-0000-0000-0000-000000000002")
	repoMemberID    = uuid.MustParse("b1000003-0000-0000-0000-000000000003")
	repoSecondOwner = uuid.MustParse("b1000004-0000-0000-0000-000000000004")
	repoUnknownUser = uuid.MustParse("b1000005-0000-0000-0000-000000000005")
)

func TestConversationRepoLifecycle(t *testing.T) {
	r := NewConversationRepo()

	conversation, err := r.CreateConversation(&models.Conversation{
		Name:      "Backend",
		AvatarURL: "https://cdn.example.com/backend.png",
		CreatedBy: repoOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if conversation.ID == 0 {
		t.Fatalf("expected persisted conversation ID, got %d", conversation.ID)
	}

	if _, err := r.GetConversationByID(conversation.ID); err != nil {
		t.Fatalf("GetConversationByID() error = %v", err)
	}

	ownerMembership, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoOwnerID,
		Role:           models.ConversationRoleOwner,
	})
	if err != nil {
		t.Fatalf("CreateMembership(owner) error = %v", err)
	}
	if ownerMembership.ID == 0 {
		t.Fatalf("expected persisted owner membership ID, got %d", ownerMembership.ID)
	}

	if _, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoOwnerID,
		Role:           models.ConversationRoleOwner,
	}); !errors.Is(err, repo.ErrMembershipAlreadyExists) {
		t.Fatalf("expected ErrMembershipAlreadyExists, got %v", err)
	}

	if _, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoAdminID,
		Role:           models.ConversationRoleAdmin,
	}); err != nil {
		t.Fatalf("CreateMembership(admin) error = %v", err)
	}
	memberMembership, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoMemberID,
		Role:           models.ConversationRoleMember,
	})
	if err != nil {
		t.Fatalf("CreateMembership(member) error = %v", err)
	}

	conversationsForOwner, err := r.ListConversationsByUser(repoOwnerID)
	if err != nil {
		t.Fatalf("ListConversationsByUser(owner) error = %v", err)
	}
	if len(conversationsForOwner) != 1 {
		t.Fatalf("expected 1 conversation for owner, got %d", len(conversationsForOwner))
	}
	if len(conversationsForOwner) == 1 && conversationsForOwner[0].ID != conversation.ID {
		t.Fatalf("expected conversation ID %d, got %d", conversation.ID, conversationsForOwner[0].ID)
	}

	updatedOwner, err := r.UpdateMembershipRole(conversation.ID, repoOwnerID, models.ConversationRoleAdmin)
	if err != nil {
		t.Fatalf("UpdateMembershipRole(owner->admin) error = %v", err)
	}
	if updatedOwner.Role != models.ConversationRoleAdmin {
		t.Fatalf("expected updated role admin, got %d", updatedOwner.Role)
	}

	if _, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoSecondOwner,
		Role:           models.ConversationRoleOwner,
	}); err != nil {
		t.Fatalf("CreateMembership(second owner) error = %v", err)
	}
	owners, err := r.CountOwners(conversation.ID)
	if err != nil {
		t.Fatalf("CountOwners() error = %v", err)
	}
	if owners != 1 {
		t.Fatalf("expected 1 owner after owner demotion, got %d", owners)
	}

	if err := r.SoftDeleteMembership(conversation.ID, repoMemberID); err != nil {
		t.Fatalf("SoftDeleteMembership(member) error = %v", err)
	}
	if _, err := r.GetMembership(conversation.ID, repoMemberID); !errors.Is(err, repo.ErrMembershipNotFound) {
		t.Fatalf("expected deleted member to be not found, got %v", err)
	}

	recreatedMember, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoMemberID,
		Role:           models.ConversationRoleMember,
	})
	if err != nil {
		t.Fatalf("CreateMembership(member recreate) error = %v", err)
	}
	if recreatedMember.ID == memberMembership.ID {
		t.Fatalf("expected recreated membership to have a new ID")
	}

	if err := r.SoftDeleteConversation(conversation.ID); err != nil {
		t.Fatalf("SoftDeleteConversation() error = %v", err)
	}
	if _, err := r.GetConversationByID(conversation.ID); !errors.Is(err, repo.ErrConversationNotFound) {
		t.Fatalf("expected deleted conversation to be not found, got %v", err)
	}
	if _, err := r.ListMemberships(conversation.ID); !errors.Is(err, repo.ErrConversationNotFound) {
		t.Fatalf("expected ListMemberships on deleted conversation to fail, got %v", err)
	}

	if err := r.SoftDeleteMembership(conversation.ID, repoUnknownUser); !errors.Is(err, repo.ErrMembershipNotFound) {
		t.Fatalf("expected SoftDeleteMembership unknown user to return ErrMembershipNotFound, got %v", err)
	}
}

func TestConversationRepoSoftDeleteMembershipsByConversation(t *testing.T) {
	r := NewConversationRepo()

	conversation, err := r.CreateConversation(&models.Conversation{
		Name:      "Infra",
		CreatedBy: repoOwnerID,
	})
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}

	if _, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoOwnerID,
		Role:           models.ConversationRoleOwner,
	}); err != nil {
		t.Fatalf("CreateMembership(owner) error = %v", err)
	}
	if _, err := r.CreateMembership(&models.ConversationMembership{
		ConversationID: conversation.ID,
		UserID:         repoAdminID,
		Role:           models.ConversationRoleAdmin,
	}); err != nil {
		t.Fatalf("CreateMembership(admin) error = %v", err)
	}

	if err := r.SoftDeleteMembershipsByConversation(conversation.ID); err != nil {
		t.Fatalf("SoftDeleteMembershipsByConversation() error = %v", err)
	}

	members, err := r.ListMemberships(conversation.ID)
	if err != nil {
		t.Fatalf("ListMemberships() error = %v", err)
	}
	if len(members) != 0 {
		t.Fatalf("expected all memberships to be soft deleted, got %d active memberships", len(members))
	}

	owners, err := r.CountOwners(conversation.ID)
	if err != nil {
		t.Fatalf("CountOwners() error = %v", err)
	}
	if owners != 0 {
		t.Fatalf("expected 0 owners after bulk soft delete, got %d", owners)
	}
}
