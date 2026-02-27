package repo

import (
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/google/uuid"
)

type MessageRepo interface {
	SaveMessage(msg *models.ChatMessage) (*models.ChatMessage, error)
	GetMessageById(id int) (*models.ChatMessage, error)
	GetMessagesByConversationID(conversationID int) ([]*models.ChatMessage, error)
	MarkMessageReceivedByID(id int, userID uuid.UUID, receivedAt time.Time) (*models.MessageReceipt, error)
	GetMessageReceiptByID(id int, userID uuid.UUID) (*models.MessageReceipt, error)
	UpdateMessageById(id int, content string) (*models.ChatMessage, error)
	DeleteMessageById(id int) error
}
