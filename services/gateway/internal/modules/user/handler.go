package user

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

type Handler struct {
	nc *nats.Conn
}

func NewHandler(nc *nats.Conn) *Handler {
	return &Handler{nc: nc}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Wrap for NestJS Microservice
	reqID := time.Now().String()
	request := struct {
		Pattern string      `json:"pattern"`
		Data    interface{} `json:"data"`
		ID      string      `json:"id"`
	}{
		Pattern: "user.get",
		Data:    map[string]string{"id": id},
		ID:      reqID,
	}
	payload, _ := json.Marshal(request)

	msg, err := h.nc.Request("user.get", payload, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(msg.Data)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 1. Validate Token to get User ID
	reqID := time.Now().String()
	valRequest := struct {
		Pattern string      `json:"pattern"`
		Data    interface{} `json:"data"`
		ID      string      `json:"id"`
	}{
		Pattern: "auth.validate",
		Data:    map[string]string{"token": token},
		ID:      reqID,
	}
	valPayload, _ := json.Marshal(valRequest)

	msg, err := h.nc.Request("auth.validate", valPayload, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	var valResult struct {
		Valid bool `json:"valid"`
		User  struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(msg.Data, &valResult); err != nil || !valResult.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user is updating themselves
	if valResult.User.ID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// 2. Prepare Update payload
	var dto interface{}
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	updatePayload := map[string]interface{}{
		"id":     id,
		"userId": valResult.User.ID,
		"dto":    dto,
	}

	reqID2 := time.Now().String()
	request := struct {
		Pattern string      `json:"pattern"`
		Data    interface{} `json:"data"`
		ID      string      `json:"id"`
	}{
		Pattern: "user.update",
		Data:    updatePayload,
		ID:      reqID2,
	}
	payloadBytes, _ := json.Marshal(request)

	resp, err := h.nc.Request("user.update", payloadBytes, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(resp.Data)
}
