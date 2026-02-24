package memory

import (
	"errors"
	"sync"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
)

type messageRepo struct {
	mu       sync.RWMutex
	messages []*models.ChatMessage
	counter  int
}

func NewMessageRepo() repo.MessageRepo {
	return &messageRepo{
		messages: make([]*models.ChatMessage, 0),
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

func (r *messageRepo) GetMessagesByGroupId(groupID int) ([]*models.ChatMessage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var messages []*models.ChatMessage
	for _, msg := range r.messages {
		if msg.GroupID == groupID {
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

func (r *messageRepo) DeleteMessageById(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for index, msg := range r.messages {
		if msg.ID == id {
			r.messages = append(r.messages[:index], r.messages[index+1:]...)
			return nil
		}
	}
	return errors.New("message not found")
}
