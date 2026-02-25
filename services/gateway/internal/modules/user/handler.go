package user

import (
	"encoding/json"
	"gateway/internal/common"
	"gateway/internal/modules/auth"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	nc common.NatsConn
}

func NewHandler(nc common.NatsConn) *Handler {
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
	payload, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg, err := h.nc.Request("user.get", payload, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var wrapper struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(msg.Data, &wrapper); err == nil && len(wrapper.Response) > 0 {
		_, _ = w.Write(wrapper.Response)
	} else {
		_, _ = w.Write(msg.Data)
	}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, `{"error":"missing query param q"}`, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}
	valResult, err := auth.ValidateToken(h.nc, token)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if !valResult.IsValid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	request := struct {
		Pattern string      `json:"pattern"`
		Data    interface{} `json:"data"`
		ID      string      `json:"id"`
	}{
		Pattern: "user.search",
		Data:    map[string]string{"query": q},
		ID:      time.Now().String(),
	}
	payload, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	msg, err := h.nc.Request("user.search", payload, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var wrapper struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(msg.Data, &wrapper); err == nil && len(wrapper.Response) > 0 {
		_, _ = w.Write(wrapper.Response)
	} else {
		_, _ = w.Write(msg.Data)
	}
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
	valResult, err := auth.ValidateToken(h.nc, token)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if !valResult.IsValid {
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
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	resp, err := h.nc.Request("user.update", payloadBytes, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var wrapper struct {
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal(resp.Data, &wrapper); err == nil && len(wrapper.Response) > 0 {
		_, _ = w.Write(wrapper.Response)
	} else {
		_, _ = w.Write(resp.Data)
	}
}
