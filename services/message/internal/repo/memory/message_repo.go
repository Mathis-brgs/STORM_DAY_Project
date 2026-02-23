package memory

import (
	"sync"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
)

// messageRepo est l'implémentation en mémoire de MessageRepo
type messageRepo struct {
	mu       sync.RWMutex
	messages []*models.ChatMessage
	counter  int
}

// NewMessageRepo crée un nouveau repository de messages en mémoire
func NewMessageRepo() repo.MessageRepo {
	return &messageRepo{
		messages: make([]*models.ChatMessage, 0),
		counter:  0,
	}
}

// Save sauvegarde un message et retourne le message avec son ID généré et CreatedAt
func (r *messageRepo) Save(msg *models.ChatMessage) (*models.ChatMessage, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Créer une copie du message pour éviter les modifications
	saved := *msg

	// Générer un ID auto-incrémenté
	r.counter++
	saved.ID = r.counter

	// Set CreatedAt si vide
	if saved.CreatedAt == (time.Time{}) {
		saved.CreatedAt = time.Now()
	}

	// Sauvegarder
	r.messages = append(r.messages, &saved)

	return &saved, nil
}
