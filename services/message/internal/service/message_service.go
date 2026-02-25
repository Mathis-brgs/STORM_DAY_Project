package service

import (
	"errors"
	"log"
	"strings"

	"github.com/google/uuid"
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
	if msg.SenderID == uuid.Nil {
		return nil, errors.New("sender ID is empty")
	}
	if msg.GroupID == 0 {
		return nil, errors.New("group ID is empty")
	}

	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return nil, errors.New("message content is empty")
	}
	if len(content) > maxMessageContentLength {
		return nil, errors.New("message content too long")
	}
	msg.Content = content

	savedMsg, err := s.messageRepo.SaveMessage(msg)
	if err != nil {
		log.Printf("[ERROR] Failed to save message: %v", err)
		return nil, err
	}

	return savedMsg, nil
}

func (s *MessageService) GetMessageById(id int) (*models.ChatMessage, error) {
	return s.messageRepo.GetMessageById(id)
}

func (s *MessageService) GetMessagesByGroupId(groupID int) ([]*models.ChatMessage, error) {
	return s.messageRepo.GetMessagesByGroupId(groupID)
}

func (s *MessageService) UpdateMessageById(id int, content string) (*models.ChatMessage, error) {
	if id == 0 {
		return nil, errors.New("id is empty")
	}

	content = strings.TrimSpace(content)
	if content == "" {
		return nil, errors.New("message content is empty")
	}
	if len(content) > maxMessageContentLength {
		return nil, errors.New("message content too long")
	}

	updatedMsg, err := s.messageRepo.UpdateMessageById(id, content)
	if err != nil {
		return nil, err
	}
	return updatedMsg, nil
}

func (s *MessageService) DeleteMessageById(id int) error {
	if id == 0 {
		return errors.New("id is empty")
	}

	err := s.messageRepo.DeleteMessageById(id)
	if err != nil {
		log.Printf("[ERROR] Failed to delete message: %v", err)
		return err
	}
	log.Printf("Message deleted: %d", id)
	return nil
}
