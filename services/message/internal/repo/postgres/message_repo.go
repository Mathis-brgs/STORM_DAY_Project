package postgres

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
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
		INSERT INTO messages (sender_id, content, conversation_id, attachment, reply_to_id, status, forward_from_id, created_at, updated_at)
		VALUES ($1::uuid, $2, $3, $4, $5, COALESCE(NULLIF($6, ''), 'sent'), $7, $8, $9)
		RETURNING id, created_at
	`

	now := time.Now()
	if msg.CreatedAt.IsZero() {
		msg.CreatedAt = now
	}
	if msg.UpdatedAt.IsZero() {
		msg.UpdatedAt = now
	}
	status := msg.Status
	if status == "" {
		status = "sent"
	}

	var replyToID, forwardFromID interface{}
	if msg.ReplyToID != nil {
		replyToID = *msg.ReplyToID
	}
	if msg.ForwardFromID != nil {
		forwardFromID = *msg.ForwardFromID
	}

	var id int
	var createdAt time.Time
	err := r.db.QueryRow(
		query,
		msg.SenderID.String(), msg.Content, msg.ConversationID, nullString(msg.Attachment),
		replyToID, status, forwardFromID,
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
	saved.Status = status

	return &saved, nil
}

func (r *messageRepo) GetMessageById(id int) (*models.ChatMessage, error) {
	query := `
		SELECT id, sender_id, content, conversation_id, COALESCE(attachment, ''),
		       reply_to_id, COALESCE(NULLIF(TRIM(status), ''), 'sent'), forward_from_id,
		       created_at, updated_at
		FROM messages
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	var msg models.ChatMessage
	var senderIDStr string
	var replyToID, forwardFromID sql.NullInt64
	var status sql.NullString
	err := r.db.QueryRow(query, id).Scan(
		&msg.ID, &senderIDStr, &msg.Content, &msg.ConversationID, &msg.Attachment,
		&replyToID, &status, &forwardFromID,
		&msg.CreatedAt, &msg.UpdatedAt,
	)
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
	if replyToID.Valid {
		ri := int(replyToID.Int64)
		msg.ReplyToID = &ri
	}
	if status.Valid {
		msg.Status = status.String
	} else {
		msg.Status = "sent"
	}
	if forwardFromID.Valid {
		fi := int(forwardFromID.Int64)
		msg.ForwardFromID = &fi
	}

	return &msg, nil
}

func (r *messageRepo) GetMessagesByConversationID(conversationID int) ([]*models.ChatMessage, error) {
	query := `
		SELECT m.id, m.sender_id, m.content, m.conversation_id, COALESCE(m.attachment, ''),
		       m.reply_to_id, COALESCE(m.status, 'sent'), m.forward_from_id,
		       m.created_at, m.updated_at,
		       r.id AS reply_id, r.sender_id AS reply_sender_id, r.content AS reply_content
		FROM messages m
		LEFT JOIN messages r ON r.id = m.reply_to_id AND r.deleted_at IS NULL
		WHERE m.conversation_id = $1
		  AND m.deleted_at IS NULL
		ORDER BY m.created_at DESC, m.id DESC
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
		var replyToID, forwardFromID sql.NullInt64
		var status sql.NullString
		var replyID sql.NullInt64
		var replySenderID, replyContent sql.NullString
		if err := rows.Scan(
			&msg.ID, &senderIDStr, &msg.Content, &msg.ConversationID, &msg.Attachment,
			&replyToID, &status, &forwardFromID,
			&msg.CreatedAt, &msg.UpdatedAt,
			&replyID, &replySenderID, &replyContent,
		); err != nil {
			return nil, err
		}
		if msg.SenderID, err = uuid.Parse(senderIDStr); err != nil {
			return nil, err
		}
		msg.ReceivedAt = nil
		if replyToID.Valid {
			ri := int(replyToID.Int64)
			msg.ReplyToID = &ri
		}
		if status.Valid {
			msg.Status = status.String
		} else {
			msg.Status = "sent"
		}
		if forwardFromID.Valid {
			fi := int(forwardFromID.Int64)
			msg.ForwardFromID = &fi
		}
		if replyID.Valid && replySenderID.Valid {
			msg.ReplyTo = &models.ReplyToRef{
				ID:       int(replyID.Int64),
				SenderID: replySenderID.String,
				Content:  replyContent.String,
			}
		}
		messages = append(messages, &msg)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(messages) > 0 {
		seenByMap, err := r.getSeenByForMessageIDs(messages)
		if err != nil {
			return nil, err
		}
		for _, m := range messages {
			m.SeenBy = seenByMap[m.ID]
		}
	}

	return messages, nil
}

func (r *messageRepo) UpdateMessageById(id int, content string) (*models.ChatMessage, error) {
	query := `
		UPDATE messages
		SET content = $1, updated_at = $2
		WHERE id = $3
		  AND deleted_at IS NULL
		RETURNING id, sender_id, conversation_id, content, COALESCE(attachment, ''),
		          reply_to_id, COALESCE(NULLIF(TRIM(status), ''), 'sent'), forward_from_id,
		          created_at, updated_at
	`

	var msg models.ChatMessage
	var senderIDStr string
	var replyToID, forwardFromID sql.NullInt64
	var status sql.NullString
	err := r.db.QueryRow(query, content, time.Now(), id).Scan(
		&msg.ID, &senderIDStr, &msg.ConversationID, &msg.Content, &msg.Attachment,
		&replyToID, &status, &forwardFromID,
		&msg.CreatedAt, &msg.UpdatedAt,
	)
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
	if replyToID.Valid {
		ri := int(replyToID.Int64)
		msg.ReplyToID = &ri
	}
	if status.Valid {
		msg.Status = status.String
	} else {
		msg.Status = "sent"
	}
	if forwardFromID.Valid {
		fi := int(forwardFromID.Int64)
		msg.ForwardFromID = &fi
	}

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

func (r *messageRepo) SetMessageStatus(id int, status string) error {
	if status != "sent" && status != "delivered" && status != "seen" {
		return errors.New("invalid status")
	}
	query := `UPDATE messages SET status = $1, updated_at = NOW() WHERE id = $2 AND deleted_at IS NULL`
	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("message not found")
	}
	return nil
}

func (r *messageRepo) MarkMessageSeenBy(id int, userID uuid.UUID, displayName string) (*models.MessageSeenBy, error) {
	query := `
		INSERT INTO message_seen_by (message_id, user_id, display_name, seen_at)
		VALUES ($1, $2::uuid, $3, NOW())
		ON CONFLICT (message_id, user_id)
		DO UPDATE SET display_name = EXCLUDED.display_name, seen_at = NOW()
		RETURNING message_id, user_id, display_name, seen_at
	`
	var out models.MessageSeenBy
	var userIDStr string
	err := r.db.QueryRow(query, id, userID.String(), displayName).Scan(
		&out.MessageID, &userIDStr, &out.DisplayName, &out.SeenAt,
	)
	if err != nil {
		return nil, err
	}
	if out.UserID, err = uuid.Parse(userIDStr); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *messageRepo) GetSeenByForMessage(id int) ([]*models.MessageSeenBy, error) {
	query := `
		SELECT message_id, user_id, display_name, seen_at
		FROM message_seen_by
		WHERE message_id = $1
		ORDER BY seen_at ASC
	`
	rows, err := r.db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*models.MessageSeenBy
	for rows.Next() {
		var e models.MessageSeenBy
		var userIDStr string
		if err := rows.Scan(&e.MessageID, &userIDStr, &e.DisplayName, &e.SeenAt); err != nil {
			return nil, err
		}
		if e.UserID, err = uuid.Parse(userIDStr); err != nil {
			return nil, err
		}
		list = append(list, &e)
	}
	return list, rows.Err()
}

func (r *messageRepo) getSeenByForMessageIDs(messages []*models.ChatMessage) (map[int][]models.SeenByEntry, error) {
	if len(messages) == 0 {
		return nil, nil
	}
	ids := make([]int, 0, len(messages))
	for _, m := range messages {
		ids = append(ids, m.ID)
	}
	placeholders := ""
	for i := range ids {
		if i > 0 {
			placeholders += ","
		}
		placeholders += "$" + strconv.Itoa(i+1)
	}
	query := `
		SELECT message_id, user_id, display_name, seen_at
		FROM message_seen_by
		WHERE message_id IN (` + placeholders + `)
		ORDER BY message_id, seen_at ASC
	`
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int][]models.SeenByEntry)
	for rows.Next() {
		var mid int
		var userIDStr, displayName string
		var seenAt time.Time
		if err := rows.Scan(&mid, &userIDStr, &displayName, &seenAt); err != nil {
			return nil, err
		}
		out[mid] = append(out[mid], models.SeenByEntry{
			UserID:      userIDStr,
			DisplayName: displayName,
			SeenAt:      seenAt.Unix(),
		})
	}
	return out, rows.Err()
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
