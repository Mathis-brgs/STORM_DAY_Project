package message

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"gateway/internal/models"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/protobuf/proto"
	"github.com/nats-io/nats.go"
)

const (
	subjectNewMessage    = "NEW_MESSAGE"
	subjectGetMessage    = "GET_MESSAGE"
	subjectListMessages  = "LIST_MESSAGES"
	subjectUpdateMessage = "UPDATE_MESSAGE"
	subjectDeleteMessage = "DELETE_MESSAGE"
	requestTimeout       = 5 * time.Second
)

const (
	invalidId = "invalid id"
)

type Handler struct {
	nc *nats.Conn
}

func NewHandler(nc *nats.Conn) *Handler {
	return &Handler{nc: nc}
}

// Send gère POST /api/messages - id row = int, sender_id = UUID, group_id = int
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
		GroupId:    int32(req.GroupID),
		SenderId:   req.SenderID,
		Content:    req.Content,
		Attachment: req.Attachment,
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

	var resp apiv1.SendMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
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
			ID:         int(d.GetId()),
			SenderID:   d.GetSenderId(),
			GroupID:    int(d.GetGroupId()),
			Content:    d.GetContent(),
			Attachment: d.GetAttachment(),
			CreatedAt:  d.GetCreatedAt(),
			UpdatedAt:  d.GetUpdatedAt(),
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

// GetById gère GET /api/messages/{id} - id = int (PK row)
func (h *Handler) GetById(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	protoReq := &apiv1.GetMessageRequest{Id: int32(id)}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectGetMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.GetMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.GetMessageResponse{
			OK: false, Error: &models.GetMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.GetMessageResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		d := resp.GetData()
		out.Data = &models.GetMessageData{
			ID:         int(d.GetId()),
			SenderID:   d.GetSenderId(),
			GroupID:    int(d.GetGroupId()),
			Content:    d.GetContent(),
			Attachment: d.GetAttachment(),
			CreatedAt:  d.GetCreatedAt(),
			UpdatedAt:  d.GetUpdatedAt(),
		}
	}
	if resp.GetError() != nil {
		out.Error = &models.GetMessageError{
			Code:    resp.GetError().GetCode(),
			Message: resp.GetError().GetMessage(),
		}
	}

	status := http.StatusOK
	if !resp.GetOk() && resp.GetError() != nil {
		if resp.GetError().GetCode() == "BAD_REQUEST" {
			status = http.StatusBadRequest
		} else {
			status = http.StatusNotFound
		}
	}
	respondJSON(w, status, out)
}

// List gère GET /api/messages?group_id= - group_id = int
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	groupIDStr := r.URL.Query().Get("group_id")
	groupID, err := strconv.ParseInt(groupIDStr, 10, 32)
	if err != nil || groupID <= 0 {
		respondJSON(w, http.StatusBadRequest, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "group_id required"},
		})
		return
	}

	protoReq := &apiv1.ListMessagesRequest{
		GroupId: int32(groupID),
		Limit:   100,
		Cursor:  r.URL.Query().Get("cursor"),
	}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectListMessages, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.ListMessagesResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.ListMessagesResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.ListMessagesResponse{OK: resp.GetOk(), NextCursor: resp.GetNextCursor()}
	for _, d := range resp.GetData() {
		out.Data = append(out.Data, models.SendMessageData{
			ID:         int(d.GetId()),
			SenderID:   d.GetSenderId(),
			GroupID:    int(d.GetGroupId()),
			Content:    d.GetContent(),
			Attachment: d.GetAttachment(),
			CreatedAt:  d.GetCreatedAt(),
			UpdatedAt:  d.GetUpdatedAt(),
		})
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

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	var body models.UpdateMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "invalid JSON"},
		})
		return
	}
	content := body.Content
	if content == "" {
		content = body.Message
	}
	if content == "" {
		respondJSON(w, http.StatusBadRequest, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: "content required"},
		})
		return
	}

	protoReq := &apiv1.UpdateMessageRequest{Id: int32(id), Content: content}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectUpdateMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.UpdateMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.UpdateMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.UpdateMessageResponse{OK: resp.GetOk()}
	if resp.GetData() != nil {
		d := resp.GetData()
		out.Data = &models.SendMessageData{
			ID:         int(d.GetId()),
			SenderID:   d.GetSenderId(),
			GroupID:    int(d.GetGroupId()),
			Content:    d.GetContent(),
			Attachment: d.GetAttachment(),
			CreatedAt:  d.GetCreatedAt(),
			UpdatedAt:  d.GetUpdatedAt(),
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

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil || id <= 0 {
		respondJSON(w, http.StatusBadRequest, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "BAD_REQUEST", Message: invalidId},
		})
		return
	}

	protoReq := &apiv1.DeleteMessageRequest{Id: int32(id)}
	data, err := proto.Marshal(protoReq)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "INTERNAL", Message: err.Error()},
		})
		return
	}

	reply, err := h.nc.Request(subjectDeleteMessage, data, requestTimeout)
	if err != nil {
		respondJSON(w, http.StatusBadGateway, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
		})
		return
	}

	var resp apiv1.DeleteMessageResponse
	if err := proto.Unmarshal(reply.Data, &resp); err != nil {
		respondJSON(w, http.StatusBadGateway, models.DeleteMessageResponse{
			OK: false, Error: &models.SendMessageError{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
		})
		return
	}

	out := models.DeleteMessageResponse{OK: resp.GetOk()}
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
			status = http.StatusNotFound
		}
	}
	respondJSON(w, status, out)
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
