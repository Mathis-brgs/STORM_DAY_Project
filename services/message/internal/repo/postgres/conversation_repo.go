package postgres

import (
	"database/sql"
	"errors"
	"time"

	models "github.com/Mathis-brgs/storm-project/services/message/internal/models"
	"github.com/Mathis-brgs/storm-project/services/message/internal/repo"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type conversationRepo struct {
	db *sql.DB
}

func NewConversationRepo(db *sql.DB) repo.ConversationRepo {
	return &conversationRepo{db: db}
}

func (r *conversationRepo) CreateConversation(conversation *models.Conversation) (*models.Conversation, error) {
	query := `
		INSERT INTO conversations (name, avatar_url, created_by, created_at, updated_at)
		VALUES ($1, NULLIF($2, ''), $3::uuid, $4, $5)
		RETURNING id, name, COALESCE(avatar_url, ''), COALESCE(created_by::text, ''), created_at, updated_at, deleted_at
	`

	now := time.Now()
	if conversation.CreatedAt.IsZero() {
		conversation.CreatedAt = now
	}
	if conversation.UpdatedAt.IsZero() {
		conversation.UpdatedAt = now
	}

	row := r.db.QueryRow(
		query,
		conversation.Name,
		conversation.AvatarURL,
		nullUUID(conversation.CreatedBy),
		conversation.CreatedAt,
		conversation.UpdatedAt,
	)

	saved, err := scanConversation(row)
	if err != nil {
		return nil, err
	}

	return saved, nil
}

func (r *conversationRepo) GetConversationByID(id int) (*models.Conversation, error) {
	query := `
		SELECT id, name, COALESCE(avatar_url, ''), COALESCE(created_by::text, ''), created_at, updated_at, deleted_at
		FROM conversations
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	conversation, err := scanConversation(r.db.QueryRow(query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrConversationNotFound
		}
		return nil, err
	}

	return conversation, nil
}

func (r *conversationRepo) ListConversationsByUser(userID uuid.UUID) ([]*models.Conversation, error) {
	query := `
		SELECT c.id, c.name, COALESCE(c.avatar_url, ''), COALESCE(c.created_by::text, ''), c.created_at, c.updated_at, c.deleted_at
		FROM conversations c
		INNER JOIN conversations_users cu
		  ON cu.conversation_id = c.id
		WHERE cu.user_id = $1::uuid
		  AND cu.deleted_at IS NULL
		  AND c.deleted_at IS NULL
		ORDER BY c.updated_at DESC, c.id DESC
	`

	rows, err := r.db.Query(query, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	conversations := make([]*models.Conversation, 0)
	for rows.Next() {
		conversation, scanErr := scanConversation(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		conversations = append(conversations, conversation)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *conversationRepo) SoftDeleteConversation(id int) error {
	query := `
		UPDATE conversations
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repo.ErrConversationNotFound
	}

	return nil
}

func (r *conversationRepo) CreateMembership(membership *models.ConversationMembership) (*models.ConversationMembership, error) {
	query := `
		INSERT INTO conversations_users (created_at, user_id, conversation_id, role)
		VALUES ($1, $2::uuid, $3, $4)
		RETURNING id, created_at, deleted_at, user_id, conversation_id, role
	`

	createdAt := membership.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	row := r.db.QueryRow(query, createdAt, membership.UserID.String(), membership.ConversationID, int(membership.Role))
	saved, err := scanMembership(row)
	if err != nil {
		return nil, translateMembershipInsertError(err)
	}
	return saved, nil
}

func (r *conversationRepo) GetMembership(conversationID int, userID uuid.UUID) (*models.ConversationMembership, error) {
	query := `
		SELECT id, created_at, deleted_at, user_id, conversation_id, role
		FROM conversations_users
		WHERE conversation_id = $1
		  AND user_id = $2::uuid
		  AND deleted_at IS NULL
		ORDER BY id DESC
		LIMIT 1
	`

	membership, err := scanMembership(r.db.QueryRow(query, conversationID, userID.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrMembershipNotFound
		}
		return nil, err
	}
	return membership, nil
}

func (r *conversationRepo) ListMemberships(conversationID int) ([]*models.ConversationMembership, error) {
	if _, err := r.GetConversationByID(conversationID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, created_at, deleted_at, user_id, conversation_id, role
		FROM conversations_users
		WHERE conversation_id = $1
		  AND deleted_at IS NULL
		ORDER BY role DESC, id ASC
	`

	rows, err := r.db.Query(query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberships := make([]*models.ConversationMembership, 0)
	for rows.Next() {
		membership, scanErr := scanMembership(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		memberships = append(memberships, membership)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return memberships, nil
}

func (r *conversationRepo) UpdateMembershipRole(conversationID int, userID uuid.UUID, role models.ConversationRole) (*models.ConversationMembership, error) {
	query := `
		UPDATE conversations_users
		SET role = $1
		WHERE conversation_id = $2
		  AND user_id = $3::uuid
		  AND deleted_at IS NULL
		RETURNING id, created_at, deleted_at, user_id, conversation_id, role
	`

	updated, err := scanMembership(r.db.QueryRow(query, int(role), conversationID, userID.String()))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repo.ErrMembershipNotFound
		}
		return nil, err
	}
	return updated, nil
}

func (r *conversationRepo) SoftDeleteMembership(conversationID int, userID uuid.UUID) error {
	query := `
		UPDATE conversations_users
		SET deleted_at = NOW()
		WHERE conversation_id = $1
		  AND user_id = $2::uuid
		  AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, conversationID, userID.String())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return repo.ErrMembershipNotFound
	}

	return nil
}

func (r *conversationRepo) SoftDeleteMembershipsByConversation(conversationID int) error {
	query := `
		UPDATE conversations_users
		SET deleted_at = NOW()
		WHERE conversation_id = $1
		  AND deleted_at IS NULL
	`

	_, err := r.db.Exec(query, conversationID)
	return err
}

func (r *conversationRepo) CountOwners(conversationID int) (int, error) {
	query := `
		SELECT COUNT(1)
		FROM conversations_users
		WHERE conversation_id = $1
		  AND role = $2
		  AND deleted_at IS NULL
	`

	var count int
	if err := r.db.QueryRow(query, conversationID, int(models.ConversationRoleOwner)).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanConversation(row scanner) (*models.Conversation, error) {
	var (
		conversation models.Conversation
		createdByStr string
		deletedAt    sql.NullTime
	)

	if err := row.Scan(
		&conversation.ID,
		&conversation.Name,
		&conversation.AvatarURL,
		&createdByStr,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
		&deletedAt,
	); err != nil {
		return nil, err
	}

	if createdByStr != "" {
		parsed, err := uuid.Parse(createdByStr)
		if err != nil {
			return nil, err
		}
		conversation.CreatedBy = parsed
	}
	if deletedAt.Valid {
		conversation.DeletedAt = &deletedAt.Time
	}

	return &conversation, nil
}

func scanMembership(row scanner) (*models.ConversationMembership, error) {
	var (
		membership models.ConversationMembership
		role       int
		userIDStr  string
		deletedAt  sql.NullTime
	)

	if err := row.Scan(
		&membership.ID,
		&membership.CreatedAt,
		&deletedAt,
		&userIDStr,
		&membership.ConversationID,
		&role,
	); err != nil {
		return nil, err
	}

	parsed, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}
	membership.UserID = parsed
	membership.Role = models.ConversationRole(role)
	if deletedAt.Valid {
		membership.DeletedAt = &deletedAt.Time
	}

	return &membership, nil
}

func nullUUID(id uuid.UUID) any {
	if id == uuid.Nil {
		return nil
	}
	return id.String()
}

func translateMembershipInsertError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return repo.ErrMembershipAlreadyExists
		case "23503":
			return repo.ErrConversationNotFound
		}
	}
	return err
}
