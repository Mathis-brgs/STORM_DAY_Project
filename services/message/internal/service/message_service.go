package service

import (
	"errors"
	"log"
	"strings"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
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
	if msg.ConversationID == 0 {
		return nil, errors.New("conversation ID is empty")
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

func (s *MessageService) GetMessagesByConversationID(conversationID int) ([]*models.ChatMessage, error) {
	return s.messageRepo.GetMessagesByConversationID(conversationID)
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

func (s *MessageService) MarkMessageReceivedByID(id int, userID uuid.UUID, receivedAt time.Time) (*models.MessageReceipt, error) {
	if id == 0 {
		return nil, errors.New("id is empty")
	}
	if userID == uuid.Nil {
		return nil, errors.New("user ID is empty")
	}
	if receivedAt.IsZero() {
		receivedAt = time.Now()
	}

	receipt, err := s.messageRepo.MarkMessageReceivedByID(id, userID, receivedAt)
	if err != nil {
		return nil, err
	}
	return receipt, nil
}

func (s *MessageService) GetMessageReceiptByID(id int, userID uuid.UUID) (*models.MessageReceipt, error) {
	if id == 0 {
		return nil, errors.New("id is empty")
	}
	if userID == uuid.Nil {
		return nil, errors.New("user ID is empty")
	}
	return s.messageRepo.GetMessageReceiptByID(id, userID)
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
