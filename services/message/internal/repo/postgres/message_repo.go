package postgres

import (
	"database/sql"
	"errors"
	"log"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
)

type messageRepo struct {
	db *sql.DB
}

func NewMessageRepo(db *sql.DB) repo.MessageRepo {
	return &messageRepo{db: db}
}

func (r *messageRepo) SaveMessage(msg *models.ChatMessage) (*models.ChatMessage, error) {
	query := `
		INSERT INTO messages (sender_id, content, conversation_id, attachment, created_at, updated_at)
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
		msg.SenderID.String(), msg.Content, msg.ConversationID, nullString(msg.Attachment),
		msg.CreatedAt, msg.UpdatedAt,
	).Scan(&id, &createdAt)
	if err != nil {
		return nil, err
	}

	saved := *msg
	saved.ID = id
	saved.CreatedAt = createdAt
	saved.UpdatedAt = msg.UpdatedAt
	saved.ReceivedAt = nil

	return &saved, nil
}

func (r *messageRepo) GetMessageById(id int) (*models.ChatMessage, error) {
	query := `
		SELECT id, sender_id, content, conversation_id, COALESCE(attachment, ''), created_at, updated_at
		FROM messages
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	var msg models.ChatMessage
	var senderIDStr string
	err := r.db.QueryRow(query, id).Scan(&msg.ID, &senderIDStr, &msg.Content, &msg.ConversationID, &msg.Attachment, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("message not found")
		}
		return nil, err
	}
	if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
		return nil, err
	}
	msg.ReceivedAt = nil

	return &msg, nil
}

func (r *messageRepo) GetMessagesByConversationID(conversationID int) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, sender_id, content, conversation_id, COALESCE(attachment, ''), created_at, updated_at
		FROM messages
		WHERE conversation_id = $1
		  AND deleted_at IS NULL
		ORDER BY created_at DESC, id DESC
		LIMIT 100
	`
	rows, err := r.db.Query(query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		var senderIDStr string
		if err := rows.Scan(&msg.ID, &senderIDStr, &msg.Content, &msg.ConversationID, &msg.Attachment, &msg.CreatedAt, &msg.UpdatedAt); err != nil {
			return nil, err
		}
		if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
			return nil, err
		}
		msg.ReceivedAt = nil
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
		  AND deleted_at IS NULL
		RETURNING id, sender_id, conversation_id, content, COALESCE(attachment, ''), created_at, updated_at
	`

	var msg models.ChatMessage
	var senderIDStr string
	err := r.db.QueryRow(query, content, time.Now(), id).Scan(&msg.ID, &senderIDStr, &msg.ConversationID, &msg.Content, &msg.Attachment, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("message not found")
		}
		return nil, err
	}
	if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
		return nil, err
	}
	msg.ReceivedAt = nil

	return &msg, nil
}

func (r *messageRepo) MarkMessageReceivedByID(id int, userID uuid.UUID, receivedAt time.Time) (*models.MessageReceipt, error) {
	query := `
		WITH target AS (
			SELECT id
			FROM messages
			WHERE id = $1
			  AND deleted_at IS NULL
		)
		INSERT INTO message_receipts (message_id, user_id, received_at, created_at, updated_at)
		SELECT target.id, $2::uuid, $3, NOW(), NOW()
		FROM target
		ON CONFLICT (message_id, user_id)
		DO UPDATE SET updated_at = NOW()
		RETURNING message_id, user_id, received_at, created_at, updated_at
	`

	var (
		receipt   models.MessageReceipt
		userIDStr string
	)
	err := r.db.QueryRow(query, id, userID.String(), receivedAt).Scan(
		&receipt.MessageID,
		&userIDStr,
		&receipt.ReceivedAt,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("message not found")
		}
		return nil, err
	}
	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	receipt.UserID = parsedUserID

	return &receipt, nil
}

func (r *messageRepo) GetMessageReceiptByID(id int, userID uuid.UUID) (*models.MessageReceipt, error) {
	query := `
		SELECT message_id, user_id, received_at, created_at, updated_at
		FROM message_receipts
		WHERE message_id = $1
		  AND user_id = $2::uuid
	`

	var (
		receipt   models.MessageReceipt
		userIDStr string
	)
	err := r.db.QueryRow(query, id, userID.String()).Scan(
		&receipt.MessageID,
		&userIDStr,
		&receipt.ReceivedAt,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("message receipt not found")
		}
		return nil, err
	}
	parsedUserID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	receipt.UserID = parsedUserID

	return &receipt, nil
}

func (r *messageRepo) DeleteMessageById(id int) error {
	query := `
		UPDATE messages
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1
		  AND deleted_at IS NULL
	`
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
