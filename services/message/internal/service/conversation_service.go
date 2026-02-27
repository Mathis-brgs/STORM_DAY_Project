package service

import (
	"errors"
	"fmt"
	"strings"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
)

const (
	defaultConversationName  = "Untitled conversation"
	maxConversationNameChars = 120
)

var (
	ErrInvalidConversationID = errors.New("conversation ID is empty")
	ErrInvalidUserID         = errors.New("user ID is empty")
	ErrInvalidConversation   = errors.New("conversation payload is invalid")
	ErrInvalidMembershipRole = errors.New("invalid membership role")
	ErrForbidden             = errors.New("forbidden")
	ErrLastOwnerGuard        = errors.New("cannot remove the last owner")
)

type ConversationService struct {
	conversationRepo repo.ConversationRepo
}

func NewConversationService(conversationRepo repo.ConversationRepo) *ConversationService {
	return &ConversationService{
		conversationRepo: conversationRepo,
	}
}

func (s *ConversationService) CreateConversation(ownerID uuid.UUID, name, avatarURL string) (*models.Conversation, error) {
	if ownerID == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	normalizedName := strings.TrimSpace(name)
	if normalizedName == "" {
		normalizedName = defaultConversationName
	}
	if len(normalizedName) > maxConversationNameChars {
		return nil, fmt.Errorf("%w: name too long", ErrInvalidConversation)
	}

	conversation := &models.Conversation{
		Name:      normalizedName,
		AvatarURL: strings.TrimSpace(avatarURL),
		CreatedBy: ownerID,
	}

	savedConversation, err := s.conversationRepo.CreateConversation(conversation)
	if err != nil {
		return nil, err
	}

	ownerMembership := &models.ConversationMembership{
		UserID:         ownerID,
		ConversationID: savedConversation.ID,
		Role:           models.ConversationRoleOwner,
	}
	if _, err := s.conversationRepo.CreateMembership(ownerMembership); err != nil {
		_ = s.conversationRepo.SoftDeleteConversation(savedConversation.ID)
		return nil, err
	}

	return savedConversation, nil
}

func (s *ConversationService) GetConversationByID(conversationID int) (*models.Conversation, error) {
	if conversationID == 0 {
		return nil, ErrInvalidConversationID
	}
	return s.conversationRepo.GetConversationByID(conversationID)
}

func (s *ConversationService) ListConversationsByUser(userID uuid.UUID) ([]*models.Conversation, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	return s.conversationRepo.ListConversationsByUser(userID)
}

func (s *ConversationService) AddMember(actorID uuid.UUID, conversationID int, userID uuid.UUID, role models.ConversationRole) (*models.ConversationMembership, error) {
	if err := validateConversationAndUser(conversationID, actorID); err != nil {
		return nil, err
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if !models.IsValidConversationRole(role) {
		return nil, ErrInvalidMembershipRole
	}

	actorMembership, err := s.requireActorMembership(conversationID, actorID)
	if err != nil {
		return nil, err
	}

	if actorMembership.Role == models.ConversationRoleMember {
		return nil, ErrForbidden
	}
	if actorMembership.Role != models.ConversationRoleOwner && role != models.ConversationRoleMember {
		return nil, ErrForbidden
	}

	if _, err := s.conversationRepo.GetMembership(conversationID, userID); err == nil {
		return nil, repo.ErrMembershipAlreadyExists
	} else if !errors.Is(err, repo.ErrMembershipNotFound) {
		return nil, err
	}

	membership := &models.ConversationMembership{
		UserID:         userID,
		ConversationID: conversationID,
		Role:           role,
	}
	return s.conversationRepo.CreateMembership(membership)
}

func (s *ConversationService) RemoveMember(actorID uuid.UUID, conversationID int, userID uuid.UUID) error {
	if err := validateConversationAndUser(conversationID, actorID); err != nil {
		return err
	}
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}
	if actorID == userID {
		return s.LeaveConversation(actorID, conversationID)
	}

	actorMembership, err := s.requireActorMembership(conversationID, actorID)
	if err != nil {
		return err
	}
	targetMembership, err := s.conversationRepo.GetMembership(conversationID, userID)
	if err != nil {
		return err
	}

	if actorMembership.Role == models.ConversationRoleMember {
		return ErrForbidden
	}
	if actorMembership.Role == models.ConversationRoleAdmin && targetMembership.Role != models.ConversationRoleMember {
		return ErrForbidden
	}
	if targetMembership.Role == models.ConversationRoleOwner {
		if actorMembership.Role != models.ConversationRoleOwner {
			return ErrForbidden
		}
		ownersCount, err := s.conversationRepo.CountOwners(conversationID)
		if err != nil {
			return err
		}
		if ownersCount <= 1 {
			return ErrLastOwnerGuard
		}
	}

	return s.conversationRepo.SoftDeleteMembership(conversationID, userID)
}

func (s *ConversationService) UpdateMemberRole(actorID uuid.UUID, conversationID int, userID uuid.UUID, newRole models.ConversationRole) (*models.ConversationMembership, error) {
	if err := validateConversationAndUser(conversationID, actorID); err != nil {
		return nil, err
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if !models.IsValidConversationRole(newRole) {
		return nil, ErrInvalidMembershipRole
	}

	actorMembership, err := s.requireActorMembership(conversationID, actorID)
	if err != nil {
		return nil, err
	}
	if actorMembership.Role != models.ConversationRoleOwner {
		return nil, ErrForbidden
	}

	targetMembership, err := s.conversationRepo.GetMembership(conversationID, userID)
	if err != nil {
		return nil, err
	}
	if targetMembership.Role == newRole {
		return targetMembership, nil
	}

	if targetMembership.Role == models.ConversationRoleOwner && newRole != models.ConversationRoleOwner {
		ownersCount, err := s.conversationRepo.CountOwners(conversationID)
		if err != nil {
			return nil, err
		}
		if ownersCount <= 1 {
			return nil, ErrLastOwnerGuard
		}
	}

	return s.conversationRepo.UpdateMembershipRole(conversationID, userID, newRole)
}

func (s *ConversationService) LeaveConversation(userID uuid.UUID, conversationID int) error {
	if err := validateConversationAndUser(conversationID, userID); err != nil {
		return err
	}

	membership, err := s.conversationRepo.GetMembership(conversationID, userID)
	if err != nil {
		return err
	}

	if membership.Role == models.ConversationRoleOwner {
		ownersCount, err := s.conversationRepo.CountOwners(conversationID)
		if err != nil {
			return err
		}
		if ownersCount <= 1 {
			return ErrLastOwnerGuard
		}
	}

	return s.conversationRepo.SoftDeleteMembership(conversationID, userID)
}

func (s *ConversationService) DeleteConversation(actorID uuid.UUID, conversationID int) error {
	if err := validateConversationAndUser(conversationID, actorID); err != nil {
		return err
	}

	actorMembership, err := s.requireActorMembership(conversationID, actorID)
	if err != nil {
		return err
	}
	if actorMembership.Role != models.ConversationRoleOwner {
		return ErrForbidden
	}

	if err := s.conversationRepo.SoftDeleteConversation(conversationID); err != nil {
		return err
	}

	return s.conversationRepo.SoftDeleteMembershipsByConversation(conversationID)
}

func (s *ConversationService) ListMembers(actorID uuid.UUID, conversationID int) ([]*models.ConversationMembership, error) {
	if err := validateConversationAndUser(conversationID, actorID); err != nil {
		return nil, err
	}

	if _, err := s.requireActorMembership(conversationID, actorID); err != nil {
		return nil, err
	}

	return s.conversationRepo.ListMemberships(conversationID)
}

func (s *ConversationService) IsMember(userID uuid.UUID, conversationID int) (bool, error) {
	if err := validateConversationAndUser(conversationID, userID); err != nil {
		return false, err
	}

	if _, err := s.conversationRepo.GetMembership(conversationID, userID); err != nil {
		if errors.Is(err, repo.ErrMembershipNotFound) {
			return false, nil
		}
		if errors.Is(err, repo.ErrConversationNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *ConversationService) requireActorMembership(conversationID int, actorID uuid.UUID) (*models.ConversationMembership, error) {
	if _, err := s.conversationRepo.GetConversationByID(conversationID); err != nil {
		return nil, err
	}
	membership, err := s.conversationRepo.GetMembership(conversationID, actorID)
	if err != nil {
		if errors.Is(err, repo.ErrMembershipNotFound) {
			return nil, ErrForbidden
		}
		return nil, err
	}
	return membership, nil
}

func validateConversationAndUser(conversationID int, userID uuid.UUID) error {
	if conversationID == 0 {
		return ErrInvalidConversationID
	}
	if userID == uuid.Nil {
		return ErrInvalidUserID
	}
	return nil
}
