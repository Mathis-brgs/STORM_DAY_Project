package api

import (
	"encoding/json"
	"net/http"
	"time"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"google.golang.org/protobuf/proto"
	"github.com/nats-io/nats.go"
)

const subjectNewMessage = "NEW_MESSAGE"
const requestTimeout = 5 * time.Second

// SendMessageRequestJSON est le payload attendu par POST /api/messages
type SendMessageRequestJSON struct {
	GroupID  int32  `json:"group_id"`
	SenderID int32  `json:"sender_id"`
	Content  string `json:"content"`
}

// SendMessageResponseJSON est la réponse renvoyée
type SendMessageResponseJSON struct {
	OK    bool                  `json:"ok"`
	Data  *SendMessageDataJSON  `json:"data,omitempty"`
	Error *SendMessageErrorJSON `json:"error,omitempty"`
}

type SendMessageDataJSON struct {
	ID        int32  `json:"id"`
	SenderID  int32  `json:"sender_id"`
	GroupID   int32  `json:"group_id"`
	Content   string `json:"content"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type SendMessageErrorJSON struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMessagesHandler crée un handler HTTP pour envoyer des messages via le message-service
func NewMessagesHandler(nc *nats.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SendMessageRequestJSON
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondJSON(w, http.StatusBadRequest, SendMessageResponseJSON{
				OK:    false,
				Error: &SendMessageErrorJSON{Code: "BAD_REQUEST", Message: "invalid JSON"},
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
			respondJSON(w, http.StatusInternalServerError, SendMessageResponseJSON{
				OK:    false,
				Error: &SendMessageErrorJSON{Code: "INTERNAL", Message: err.Error()},
			})
			return
		}

		reply, err := nc.Request(subjectNewMessage, data, requestTimeout)
		if err != nil {
			respondJSON(w, http.StatusBadGateway, SendMessageResponseJSON{
				OK:    false,
				Error: &SendMessageErrorJSON{Code: "GATEWAY_ERROR", Message: "message-service unreachable: " + err.Error()},
			})
			return
		}

		var resp apiv1.SendMessageResponse
		if err := proto.Unmarshal(reply.Data, &resp); err != nil {
			respondJSON(w, http.StatusBadGateway, SendMessageResponseJSON{
				OK:    false,
				Error: &SendMessageErrorJSON{Code: "GATEWAY_ERROR", Message: "invalid response from message-service"},
			})
			return
		}

		out := SendMessageResponseJSON{OK: resp.GetOk()}
		if resp.GetData() != nil {
			d := resp.GetData()
			out.Data = &SendMessageDataJSON{
				ID:        d.GetId(),
				SenderID:  d.GetSenderId(),
				GroupID:   d.GetGroupId(),
				Content:   d.GetContent(),
				CreatedAt: d.GetCreatedAt(),
				UpdatedAt: d.GetUpdatedAt(),
			}
		}
		if resp.GetError() != nil {
			out.Error = &SendMessageErrorJSON{
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
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
