package repo

import models "github.com/Mathis-brgs/storm-project/services/message/internal/models"

type MessageRepo interface {
	Save(msg *models.ChatMessage) (*models.ChatMessage, error)
}
