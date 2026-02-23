package service

import (
	"errors"
	"log"
	"strings"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
)

const maxMessageContentLength = 10000

type MessageService struct {
	messageRepo repo.MessageRepo
}

func NewMessageService(messageRepo repo.MessageRepo) *MessageService {
	return &MessageService{
		messageRepo: messageRepo,
	}
}

func (s *MessageService) SendMessage(msg *models.ChatMessage) (*models.ChatMessage, error) {
	// Valider: SenderID non vide
	if msg.SenderID == 0 {
		return nil, errors.New("sender ID is empty")
	}

	// Valider: GroupID non vide
	if msg.GroupID == 0 {
		return nil, errors.New("group ID is empty")
	}

	// Valider: Content non vide (trim) et longueur max
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return nil, errors.New("message content is empty")
	}
	if len(content) > maxMessageContentLength {
		return nil, errors.New("message content too long")
	}
	msg.Content = content

	// Sauvegarder le message
	savedMsg, err := s.messageRepo.Save(msg)
	if err != nil {
		log.Printf("[ERROR] Failed to save message: %v", err)
		return nil, err
	}

	return savedMsg, nil
}
