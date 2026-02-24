package message

import (
	"encoding/json"
	"net/http"
	"time"

	"gateway/internal/models"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

const (
	subjectNewMessage = "NEW_MESSAGE"
	requestTimeout    = 5 * time.Second
)

type Handler struct {
	nc *nats.Conn
}

func NewHandler(nc *nats.Conn) *Handler {
	return &Handler{nc: nc}
}

// Send gère POST /api/messages - envoie un message via le message-service
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}

	protoReq := &apiv1.SendMessageRequest{
		GroupId:  req.GroupID,
		SenderId: req.SenderID,
		Content:  req.Content,
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectNewMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	resp := &apiv1.SendMessageResponse{}
	if err := proto.Unmarshal(reply.Data, resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.SendMessageResponse{
			OK:    false,
			Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.SendMessageResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		d := resp.GetData()
		out.Data = &models.SendMessageData{
			ID:        d.GetId(),
			SenderID:  d.GetSenderId(),
			GroupID:   d.GetGroupId(),
			Content:   d.GetContent(),
			CreatedAt: d.GetCreatedAt(),
			UpdatedAt: d.GetUpdatedAt(),
		}
	}
	if resp.GetError() != nil {
		out.Error = &models.SendMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		if resp.GetError().GetCode() == "BAD_REQUEST" {
			status = http.StatusBadRequest
		} else {
			status = http.StatusUnprocessableEntity
		}
	}
	respondJSON(w, status, out)
}

// GetById gère GET /api/messages/{id} - non implémenté côté message-service pour l’instant
func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "id")
	respondJSON(w, http.StatusNotImplemented, models.GetMessageResponse{
		OK: false, Error: &models.GetMessageError{Code: "NOT_IMPLEMENTED", Message: "GET /api/messages/{id} not implemented yet"},
	})
}

// List gère GET /api/messages - non implémenté côté message-service pour l’instant
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusNotImplemented, models.ListMessagesResponse{
		OK: false, Error: &models.SendMessageError{Code: "NOT_IMPLEMENTED", Message: "GET /api/messages?group_id=... not implemented yet"},
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
