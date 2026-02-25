package postgres

import (
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
)

type messageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) repo.MessageRepo {
	return &messageRepo{db: db}
}

func (r *messageRepo) SaveMessage(msg *models.ChatMessage) (*models.ChatMessage, error) {
	query := `
		INSERT INTO messages (sender_id, content, group_id, attachment, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = now
	}

	var id int
	var createdAt time.Time
	err := r.db.QueryRow(
		query,
		msg.SenderID.String(), msg.Content, msg.GroupID, nullString(msg.Attachment),
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

func (r *messageRepo) GetMessageById(id int) (*models.ChatMessage, error) {
	query := `
		SELECT id, sender_id, content, group_id, COALESCE(attachment, ''), created_at, updated_at
		FROM messages
		WHERE id = $1
	`

	var msg models.ChatMessage
	var senderIDStr string
	err := r.db.QueryRow(query, id).Scan(&msg.ID, &senderIDStr, &msg.Content, &msg.GroupID, &msg.Attachment, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("message not found")
		}
		return nil, err
	}
	if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *messageRepo) GetMessagesByGroupId(groupID int) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, sender_id, content, group_id, COALESCE(attachment, ''), created_at, updated_at
		FROM messages
		WHERE group_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`
	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		var senderIDStr string
		if err := rows.Scan(&msg.ID, &senderIDStr, &msg.Content, &msg.GroupID, &msg.Attachment, &msg.CreatedAt, &msg.UpdatedAt); err != nil {
			return nil, err
		}
		if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return messages, nil
}

func (r *messageRepo) UpdateMessageById(id int, content string) (*models.ChatMessage, error) {
	query := `
		UPDATE messages
		SET content = $1, updated_at = $2
		WHERE id = $3
		RETURNING id, sender_id, group_id, content, COALESCE(attachment, ''), created_at, updated_at
	`

	var msg models.ChatMessage
	var senderIDStr string
	err := r.db.QueryRow(query, content, time.Now(), id).Scan(&msg.ID, &senderIDStr, &msg.GroupID, &msg.Content, &msg.Attachment, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("message not found")
		}
		return nil, err
	}
	if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *messageRepo) DeleteMessageById(id int) error {
	query := `DELETE FROM messages WHERE id = $1`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("message not found")
	}
	log.Printf("Message deleted: %d", id)
	return nil
}

func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
