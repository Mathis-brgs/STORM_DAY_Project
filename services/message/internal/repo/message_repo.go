package repo

import (
	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
)

type MessageRepo interface {
	SaveMessage(msg *models.ChatMessage) (*models.ChatMessage, error)
	GetMessageById(id int) (*models.ChatMessage, error)
	GetMessagesByGroupId(groupID int) ([]*models.ChatMessage, error)
	UpdateMessageById(id int, content string) (*models.ChatMessage, error)
	DeleteMessageById(id int) error
}
