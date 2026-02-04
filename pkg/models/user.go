package models

import (
	"time"

	"github.com/google/uuid"
)

// User définit l'utilisateur qui traverse tout le système
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Username     string    `json:"username" gorm:"not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"column:password_hash;not null"`
	AvatarURL    string    `json:"avatar_url" gorm:"type:text"`
}

// NewUser crée un User avec un UUIDv7
func NewUser(username, email, passwordHash string) User {
	id, _ := uuid.NewV7()
	return User{
		ID:           id,
		Username:     username,
		Email:        email,
		PasswordHash: passwordHash,
	}
}

// Jwt représente un token JWT stocké en base
type Jwt struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	Token       string    `json:"jwt" gorm:"column:jwt;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"type:timestamptz"`
	ExpiratedAt time.Time `json:"expirated_at" gorm:"type:timestamptz"`
}

// AuthClaims est ce qu'il y a dans le JWT
type AuthClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}
