package auth

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ValidationResult struct {
	IsValid bool     `json:"valid"`
	User    UserInfo `json:"user"`
}

type claims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// ValidateToken validates a JWT token locally without calling the auth service via NATS.
func ValidateToken(token string) (*ValidationResult, error) {
	if token == "" {
		return &ValidationResult{IsValid: false}, nil
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "storm-secret-key"
	}

	parsed := &claims{}
	_, err := jwt.ParseWithClaims(token, parsed, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return &ValidationResult{IsValid: false}, nil
	}

	return &ValidationResult{
		IsValid: true,
		User: UserInfo{
			ID:       parsed.Subject,
			Username: parsed.Username,
		},
	}, nil
}
