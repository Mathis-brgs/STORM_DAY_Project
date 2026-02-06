package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

type Handler struct {
	nc *nats.Conn
}

func NewHandler(nc *nats.Conn) *Handler {
	return &Handler{nc: nc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	proxyRequest(h.nc, "auth.register", w, r)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	proxyRequest(h.nc, "auth.login", w, r)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	proxyRequest(h.nc, "auth.refresh", w, r)
}

// Logout requires extracting the user ID from the token (validated via NATS first)
// For simplicity, we assume the token is in the header and we ask 'auth.validate' first?
// Or we just forward the token if the user service logic changed?
// Plan said: "Gateway ... sends a NATS request".
// My previous thought: "Gateway ... Call auth.validate ... then auth.logout"
// Let's implement that.

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Strip "Bearer " if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// 1. Validate Token to get User ID
	valPayload, err := json.Marshal(map[string]string{"token": token})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
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

	// 2. Call Logout with User ID
	logoutPayload, err := json.Marshal(map[string]string{"userId": valResult.User.ID})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	resp, err := h.nc.Request("auth.logout", logoutPayload, 2*time.Second)
	if err != nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(resp.Data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func proxyRequest(nc *nats.Conn, subject string, w http.ResponseWriter, r *http.Request) {
	var body interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Wrap for NestJS Microservice (needs "pattern", "data", "id")
	// "id" is required for it to be treated as a Request-Response, otherwise it's an Event.
	reqID := time.Now().String() // Simple unique ID
	request := struct {
		Pattern string      `json:"pattern"`
		Data    interface{} `json:"data"`
		ID      string      `json:"id"`
	}{
		Pattern: subject,
		Data:    body,
		ID:      reqID,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("[Gateway] Sending NATS request to %s with payload: %s", subject, string(payload))
	msg, err := nc.Request(subject, payload, 2*time.Second)
	if err != nil {
		log.Printf("[Gateway] NATS Error for %s: %v", subject, err)
		http.Error(w, "Service unavailable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	log.Printf("[Gateway] Received NATS response from %s: %s", subject, string(msg.Data))

	// Assuming User Service returns JSON error if something went wrong
	// We just pipe the response for now
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(msg.Data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
