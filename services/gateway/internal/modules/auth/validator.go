package auth

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ValidationResult struct {
	Valid bool     `json:"valid"`
	User  UserInfo `json:"user"`
}

// ValidateToken sends a request to the Auth service via NATS to validate a JWT token.
func ValidateToken(nc *nats.Conn, token string) (*ValidationResult, error) {
	if token == "" {
		return &ValidationResult{Valid: false}, nil
	}

	// Wrap for NestJS Microservice
	reqID := time.Now().String()
	request := struct {
		Pattern string            `json:"pattern"`
		Data    map[string]string `json:"data"`
		ID      string            `json:"id"`
	}{
		Pattern: "auth.validate",
		Data:    map[string]string{"token": token},
		ID:      reqID,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	msg, err := nc.Request("auth.validate", payload, 2*time.Second)
	if err != nil {
		return nil, fmt.Errorf("nats request failed: %w", err)
	}

	var wrapper struct {
		Response ValidationResult `json:"response"`
	}
	if err := json.Unmarshal(msg.Data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal validation result: %w", err)
	}

	return &wrapper.Response, nil
}
