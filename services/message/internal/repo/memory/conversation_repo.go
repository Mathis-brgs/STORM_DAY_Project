package memory

import (
	"sync"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
)

type conversationRepo struct {
	mu            sync.RWMutex
	conversations map[int]*models.Conversation
	memberships   map[int]map[uuid.UUID]*models.ConversationMembership
	nextConvID    int
	nextMemberID  int
}

func NewConversationRepo() repo.ConversationRepo {
	return &conversationRepo{
		conversations: make(map[int]*models.Conversation),
		memberships:   make(map[int]map[uuid.UUID]*models.ConversationMembership),
		nextConvID:    1,
		nextMemberID:  1,
	}
}

func (r *conversationRepo) CreateConversation(conversation *models.Conversation) (*models.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	saved := *conversation
	saved.ID = r.nextConvID
	r.nextConvID++
	if saved.CreatedAt.IsZero() {
		saved.CreatedAt = now
	}
	if saved.UpdatedAt.IsZero() {
		saved.UpdatedAt = now
	}

	r.conversations[saved.ID] = &saved
	if _, ok := r.memberships[saved.ID]; !ok {
		r.memberships[saved.ID] = make(map[uuid.UUID]*models.ConversationMembership)
	}

	return cloneConversation(&saved), nil
}

func (r *conversationRepo) GetConversationByID(id int) (*models.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversation, ok := r.conversations[id]
	if !ok || conversation.DeletedAt != nil {
		return nil, repo.ErrConversationNotFound
	}

	return cloneConversation(conversation), nil
}

func (r *conversationRepo) ListConversationsByUser(userID uuid.UUID) ([]*models.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversations := make([]*models.Conversation, 0)
	for conversationID, members := range r.memberships {
		membership, ok := members[userID]
		if !ok || membership.DeletedAt != nil {
			continue
		}

		conversation, exists := r.conversations[conversationID]
		if !exists || conversation.DeletedAt != nil {
			continue
		}
		conversations = append(conversations, cloneConversation(conversation))
	}

	return conversations, nil
}

func (r *conversationRepo) SoftDeleteConversation(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conversation, ok := r.conversations[id]
	if !ok || conversation.DeletedAt != nil {
		return repo.ErrConversationNotFound
	}

	now := time.Now()
	conversation.DeletedAt = &now
	conversation.UpdatedAt = now
	return nil
}

func (r *conversationRepo) CreateMembership(membership *models.ConversationMembership) (*models.ConversationMembership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conversation, ok := r.conversations[membership.ConversationID]
	if !ok || conversation.DeletedAt != nil {
		return nil, repo.ErrConversationNotFound
	}

	if _, ok := r.memberships[membership.ConversationID]; !ok {
		r.memberships[membership.ConversationID] = make(map[uuid.UUID]*models.ConversationMembership)
	}

	if existing, exists := r.memberships[membership.ConversationID][membership.UserID]; exists && existing.DeletedAt == nil {
		return nil, repo.ErrMembershipAlreadyExists
	}

	now := time.Now()
	saved := *membership
	saved.ID = r.nextMemberID
	r.nextMemberID++
	if saved.CreatedAt.IsZero() {
		saved.CreatedAt = now
	}
	saved.DeletedAt = nil

	r.memberships[membership.ConversationID][membership.UserID] = &saved

	return cloneMembership(&saved), nil
}

func (r *conversationRepo) GetMembership(conversationID int, userID uuid.UUID) (*models.ConversationMembership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members, ok := r.memberships[conversationID]
	if !ok {
		return nil, repo.ErrMembershipNotFound
	}
	membership, ok := members[userID]
	if !ok || membership.DeletedAt != nil {
		return nil, repo.ErrMembershipNotFound
	}

	return cloneMembership(membership), nil
}

func (r *conversationRepo) ListMemberships(conversationID int) ([]*models.ConversationMembership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	conversation, ok := r.conversations[conversationID]
	if !ok || conversation.DeletedAt != nil {
		return nil, repo.ErrConversationNotFound
	}

	members := r.memberships[conversationID]
	result := make([]*models.ConversationMembership, 0, len(members))
	for _, membership := range members {
		if membership.DeletedAt != nil {
			continue
		}
		result = append(result, cloneMembership(membership))
	}

	return result, nil
}

func (r *conversationRepo) UpdateMembershipRole(conversationID int, userID uuid.UUID, role models.ConversationRole) (*models.ConversationMembership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	members, ok := r.memberships[conversationID]
	if !ok {
		return nil, repo.ErrMembershipNotFound
	}

	membership, ok := members[userID]
	if !ok || membership.DeletedAt != nil {
		return nil, repo.ErrMembershipNotFound
	}

	membership.Role = role
	return cloneMembership(membership), nil
}

func (r *conversationRepo) SoftDeleteMembership(conversationID int, userID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	members, ok := r.memberships[conversationID]
	if !ok {
		return repo.ErrMembershipNotFound
	}

	membership, ok := members[userID]
	if !ok || membership.DeletedAt != nil {
		return repo.ErrMembershipNotFound
	}

	now := time.Now()
	membership.DeletedAt = &now
	return nil
}

func (r *conversationRepo) SoftDeleteMembershipsByConversation(conversationID int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	members, ok := r.memberships[conversationID]
	if !ok {
		return nil
	}

	now := time.Now()
	for _, membership := range members {
		if membership.DeletedAt == nil {
			membership.DeletedAt = &now
		}
	}
	return nil
}

func (r *conversationRepo) CountOwners(conversationID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	members, ok := r.memberships[conversationID]
	if !ok {
		return 0, nil
	}

	count := 0
	for _, membership := range members {
		if membership.DeletedAt == nil && membership.Role == models.ConversationRoleOwner {
			count++
		}
	}
	return count, nil
}

func cloneConversation(conversation *models.Conversation) *models.Conversation {
	if conversation == nil {
		return nil
	}
	cpy := *conversation
	return &cpy
}

func cloneMembership(membership *models.ConversationMembership) *models.ConversationMembership {
	if membership == nil {
		return nil
	}
	cpy := *membership
	return &cpy
}
