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
	seenBy   map[int][]*models.MessageSeenBy
	counter  int
}

func NewMessageRepo() repo.MessageRepo {
	return &messageRepo{
		messages: make([]*models.ChatMessage, 0),
		receipts: make(map[int]map[uuid.UUID]*models.MessageReceipt),
		seenBy:   make(map[int][]*models.MessageSeenBy),
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
	if saved.Status == "" {
		saved.Status = "sent"
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
	for _, m := range messages {
		if m.ReplyToID != nil {
			if replyMsg := r.findByIDLocked(*m.ReplyToID); replyMsg != nil {
				m.ReplyTo = &models.ReplyToRef{
					ID:       replyMsg.ID,
					SenderID: replyMsg.SenderID.String(),
					Content:  replyMsg.Content,
				}
			}
		}
		if list := r.seenBy[m.ID]; len(list) > 0 {
			m.SeenBy = make([]models.SeenByEntry, 0, len(list))
			for _, e := range list {
				m.SeenBy = append(m.SeenBy, models.SeenByEntry{
					UserID:      e.UserID.String(),
					DisplayName: e.DisplayName,
					SeenAt:      e.SeenAt.Unix(),
				})
			}
		}
	}
	return messages, nil
}

func (r *messageRepo) findByIDLocked(id int) *models.ChatMessage {
	for _, m := range r.messages {
		if m.ID == id {
			return m
		}
	}
	return nil
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

func (r *messageRepo) SetMessageStatus(id int, status string) error {
	if status != "sent" && status != "delivered" && status != "seen" {
		return errors.New("invalid status")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, msg := range r.messages {
		if msg.ID == id {
			msg.Status = status
			return nil
		}
	}
	return errors.New("message not found")
}

func (r *messageRepo) MarkMessageSeenBy(id int, userID uuid.UUID, displayName string) (*models.MessageSeenBy, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, msg := range r.messages {
		if msg.ID == id {
			if r.seenBy[id] == nil {
				r.seenBy[id] = make([]*models.MessageSeenBy, 0)
			}
			for _, e := range r.seenBy[id] {
				if e.UserID == userID {
					e.DisplayName = displayName
					e.SeenAt = time.Now()
					return &models.MessageSeenBy{MessageID: e.MessageID, UserID: e.UserID, DisplayName: e.DisplayName, SeenAt: e.SeenAt}, nil
				}
			}
			now := time.Now()
			e := &models.MessageSeenBy{MessageID: id, UserID: userID, DisplayName: displayName, SeenAt: now}
			r.seenBy[id] = append(r.seenBy[id], e)
			return &models.MessageSeenBy{MessageID: e.MessageID, UserID: e.UserID, DisplayName: e.DisplayName, SeenAt: e.SeenAt}, nil
		}
	}
	return nil, errors.New("message not found")
}

func (r *messageRepo) GetSeenByForMessage(id int) ([]*models.MessageSeenBy, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := r.seenBy[id]
	if len(list) == 0 {
		return nil, nil
	}
	out := make([]*models.MessageSeenBy, len(list))
	for i, e := range list {
		cpy := *e
		out[i] = &cpy
	}
	return out, nil
}

func (r *messageRepo) DeleteMessageById(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, msg := range r.messages {
		if msg.ID == id {
			r.messages = append(r.messages[:index], r.messages[index+1:]...)
			delete(r.receipts, id)
			delete(r.seenBy, id)
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
