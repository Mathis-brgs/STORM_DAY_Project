package memory

import (
	"errors"
	"sync"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
)

type messageRepo struct {
	mu       sync.RWMutex
	messages []*models.ChatMessage
	receipts map[int]map[uuid.UUID]*models.MessageReceipt
	counter  int
}

func NewMessageRepo() repo.MessageRepo {
	return &messageRepo{
		messages: make([]*models.ChatMessage, 0),
		receipts: make(map[int]map[uuid.UUID]*models.MessageReceipt),
		counter:  0,
	}
}

func (r *messageRepo) SaveMessage(msg *models.ChatMessage) (*models.ChatMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	saved := *msg
	r.counter++
	saved.ID = r.counter
	if saved.CreatedAt.IsZero() {
		saved.CreatedAt = time.Now()
	}
	if saved.UpdatedAt.IsZero() {
		saved.UpdatedAt = time.Now()
	}

	r.messages = append(r.messages, &saved)
	return &saved, nil
}

func (r *messageRepo) GetMessageById(id int) (*models.ChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, msg := range r.messages {
		if msg.ID == id {
			return msg, nil
		}
	}

	return nil, errors.New("message not found")
}

func (r *messageRepo) GetMessagesByConversationID(conversationID int) ([]*models.ChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var messages []*models.ChatMessage
	for _, msg := range r.messages {
		if msg.ConversationID == conversationID {
			messages = append(messages, msg)
		}
	}
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	return messages, nil
}

func (r *messageRepo) UpdateMessageById(id int, content string) (*models.ChatMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, msg := range r.messages {
		if msg.ID == id {
			msg.Content = content
			msg.UpdatedAt = time.Now()
			return msg, nil
		}
	}
	return nil, errors.New("message not found")
}

func (r *messageRepo) MarkMessageReceivedByID(id int, userID uuid.UUID, receivedAt time.Time) (*models.MessageReceipt, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, msg := range r.messages {
		if msg.ID == id {
			if r.receipts[id] == nil {
				r.receipts[id] = make(map[uuid.UUID]*models.MessageReceipt)
			}

			if existing, ok := r.receipts[id][userID]; ok {
				existing.UpdatedAt = time.Now()
				return cloneMessageReceipt(existing), nil
			}

			now := time.Now()
			receipt := &models.MessageReceipt{
				MessageID:  id,
				UserID:     userID,
				ReceivedAt: receivedAt,
				CreatedAt:  now,
				UpdatedAt:  now,
			}
			r.receipts[id][userID] = receipt
			return cloneMessageReceipt(receipt), nil
		}
	}
	return nil, errors.New("message not found")
}

func (r *messageRepo) GetMessageReceiptByID(id int, userID uuid.UUID) (*models.MessageReceipt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	usersReceipts, ok := r.receipts[id]
	if !ok {
		return nil, errors.New("message receipt not found")
	}

	receipt, ok := usersReceipts[userID]
	if !ok {
		return nil, errors.New("message receipt not found")
	}

	return cloneMessageReceipt(receipt), nil
}

func (r *messageRepo) DeleteMessageById(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, msg := range r.messages {
		if msg.ID == id {
			r.messages = append(r.messages[:index], r.messages[index+1:]...)
			delete(r.receipts, id)
			return nil
		}
	}
	return errors.New("message not found")
}

func cloneMessageReceipt(receipt *models.MessageReceipt) *models.MessageReceipt {
	if receipt == nil {
		return nil
	}
	cpy := *receipt
	return &cpy
}
