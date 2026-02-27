package repo

import (
	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/google/uuid"
)

type ConversationRepo interface {
	CreateConversation(conversation *models.Conversation) (*models.Conversation, error)
	GetConversationByID(id int) (*models.Conversation, error)
	ListConversationsByUser(userID uuid.UUID) ([]*models.Conversation, error)
	SoftDeleteConversation(id int) error

	CreateMembership(membership *models.ConversationMembership) (*models.ConversationMembership, error)
	GetMembership(conversationID int, userID uuid.UUID) (*models.ConversationMembership, error)
	ListMemberships(conversationID int) ([]*models.ConversationMembership, error)
	UpdateMembershipRole(conversationID int, userID uuid.UUID, role models.ConversationRole) (*models.ConversationMembership, error)
	SoftDeleteMembership(conversationID int, userID uuid.UUID) error
	SoftDeleteMembershipsByConversation(conversationID int) error
	CountOwners(conversationID int) (int, error)
}
