package postgres

import (
	"database/sql"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
)

type messageRepo struct {
	db *sql.DB
}

// NewMessageRepo crée un MessageRepo PostgreSQL
func NewMessageRepo(db *sql.DB) repo.MessageRepo {
	return &messageRepo{db: db}
}

func (r *messageRepo) Save(msg *models.ChatMessage) (*models.ChatMessage, error) {
	query := `
		INSERT INTO messages (sender_id, content, group_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	now := time.Now()
	if msg.CreatedAt == (time.Time{}) {
		msg.CreatedAt = now
	}
	if msg.UpdatedAt == (time.Time{}) {
		msg.UpdatedAt = now
	}

	var id int
	var createdAt time.Time
	err := r.db.QueryRow(
		query,
		msg.SenderID, msg.Content, msg.GroupID,
		msg.CreatedAt, msg.UpdatedAt,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	saved := *msg
	saved.ID = id
	saved.CreatedAt = createdAt
	saved.UpdatedAt = msg.UpdatedAt

	return &saved, nil
}
