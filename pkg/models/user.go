package models

import (
	"time"
	"storm-project/pkg/id"	
)

// User définit l'utilisateur qui traverse tout le système
type User struct {
	ID        id.ID `json:"id" gorm:"type:id;default:gen_random_uuid();primaryKey"`
	Email     string    `json:"email" gorm:"uniqueIndex;not null"`
	Username  string    `json:"username" gorm:"uniqueIndex;not null"`
	Password  string    `json:"-"` // Jamais renvoyé en JSON !
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthClaims est ce qu'il y a dans le JWT
type AuthClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}