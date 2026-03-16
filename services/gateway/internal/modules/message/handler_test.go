package message

import (
	"bytes"
	"context"
	"encoding/json"
	"gateway/internal/common"
	"gateway/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func TestHandler_Send(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.SendMessageResponse{
				Ok: true,
				Data: &apiv1.ChatMessage{
					Id:      1,
					Content: "hello",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	body := `{"group_id": 123, "sender_id": "user-123", "content": "hello"}`
	req := httptest.NewRequest("POST", "/api/messages", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handler.Send(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Send_WithConversationID(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.SendMessageResponse{
				Ok: true,
				Data: &apiv1.ChatMessage{
					Id:      1,
					Content: "hello",
					GroupId: 321,
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	body := `{"conversation_id": 321, "sender_id": "user-123", "content": "hello"}`
	req := httptest.NewRequest("POST", "/api/messages", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handler.Send(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_GetById(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.GetMessageResponse{
				Ok: true,
				Data: &apiv1.ChatMessage{
					Id:      1,
					Content: "hello",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.GetById(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Send_JSONError(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("POST", "/api/messages", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	handler.Send(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_Send_NATSError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return nil, nats.ErrTimeout
		},
	}
	handler := NewHandler(mockNc)
	body := `{"group_id": 123, "sender_id": "user-123", "content": "hello"}`
	req := httptest.NewRequest("POST", "/api/messages", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	handler.Send(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status BadGateway, got %d", w.Code)
	}
}

func TestHandler_GetById_InvalidId(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("GET", "/api/messages/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetById(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_List(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.ListMessagesResponse{
				Ok: true,
				Data: []*apiv1.ChatMessage{
					{Id: 1, Content: "hi"},
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages?group_id=123&actor_id=a0000001-0000-0000-0000-000000000001", nil)
	w := httptest.NewRecorder()
	handler.GetByGroupId(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_List_WithConversationID(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.ListMessagesResponse{
				Ok: true,
				Data: []*apiv1.ChatMessage{
					{Id: 1, Content: "hi", GroupId: 123},
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages?conversation_id=123&actor_id=a0000001-0000-0000-0000-000000000001", nil)
	w := httptest.NewRecorder()
	handler.GetByGroupId(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_List_NoGroupId(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("GET", "/api/messages", nil)
	w := httptest.NewRecorder()
	handler.GetByGroupId(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_Update(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.UpdateMessageResponse{
				Ok:   true,
				Data: &apiv1.ChatMessage{Id: 1, Content: "updated"},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	body := `{"content": "updated"}`
	req := httptest.NewRequest("PUT", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Update_NoContent(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	body := `{"content": ""}`
	req := httptest.NewRequest("PUT", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_Delete(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.DeleteMessageResponse{Ok: true}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("DELETE", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Delete(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Send_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("invalid proto bytes")}, nil
		},
	}
	handler := NewHandler(mockNc)
	body := `{"group_id": 123, "sender_id": "user-123", "content": "hello"}`
	req := httptest.NewRequest("POST", "/api/messages", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	handler.Send(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status BadGateway, got %d", w.Code)
	}
}

func TestHandler_Send_BusinessError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.SendMessageResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "BAD_REQUEST",
					Message: "invalid content",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	body := `{"group_id": 123, "sender_id": "user-123", "content": "bad"}`
	req := httptest.NewRequest("POST", "/api/messages", bytes.NewBufferString(body))
	w := httptest.NewRecorder()
	handler.Send(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_GetById_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("invalid proto bytes")}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetById(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status BadGateway, got %d", w.Code)
	}
}

func TestHandler_GetById_NotFound(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.GetMessageResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "NOT_FOUND",
					Message: "msg not found",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetById(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status NotFound, got %d", w.Code)
	}
}

func TestHandler_List_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("invalid proto bytes")}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages?group_id=123&actor_id=a0000001-0000-0000-0000-000000000001", nil)
	w := httptest.NewRecorder()
	handler.GetByGroupId(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status BadGateway, got %d", w.Code)
	}
}

func TestHandler_List_BusinessError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.ListMessagesResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "OTHER",
					Message: "err",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages?group_id=123&actor_id=a0000001-0000-0000-0000-000000000001", nil)
	w := httptest.NewRecorder()
	handler.GetByGroupId(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("Expected status UnprocessableEntity, got %d", w.Code)
	}
}

func TestHandler_Update_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("invalid proto bytes")}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("PUT", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(`{"content":"ok"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status BadGateway, got %d", w.Code)
	}
}

func TestHandler_Update_FallbackMessage(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.UpdateMessageResponse{Ok: true}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	body := `{"message": "using fallback"}`
	req := httptest.NewRequest("PUT", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Delete_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("invalid proto bytes")}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("DELETE", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Delete(w, req)
	if w.Code != http.StatusBadGateway {
		t.Errorf("Expected status BadGateway, got %d", w.Code)
	}
}

func TestHandler_Send_WrongMethod(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("GET", "/api/messages", nil)
	w := httptest.NewRecorder()
	handler.Send(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405, got %d", w.Code)
	}
}

func TestHandler_GetById_BusinessError_BadRequest(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.GetMessageResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "BAD_REQUEST",
					Message: "err",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/messages/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.GetById(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestHandler_AckReceipt(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject != subjectAckMessage {
				t.Fatalf("expected subject %s, got %s", subjectAckMessage, subject)
			}

			var req apiv1.AckMessageRequest
			if err := proto.Unmarshal(data, &req); err != nil {
				t.Fatalf("invalid request payload: %v", err)
			}
			if req.GetActorId() == "" {
				t.Fatalf("expected actor_id in request payload")
			}

			resp := &apiv1.AckMessageResponse{
				Ok: true,
				Data: &apiv1.ChatMessage{
					Id:             1,
					ConversationId: 123,
					GroupId:        123,
					Content:        "hello",
					ReceivedAt:     1710000000,
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)

	req := httptest.NewRequest("POST", "/api/messages/1/receipt?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(`{}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.AckReceipt(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status OK, got %d", w.Code)
	}

	var payload models.AckMessageResponse
	if err := json.NewDecoder(w.Body).Decode(&payload); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if payload.Data == nil || payload.Data.ReceivedAt == 0 {
		t.Fatalf("expected received_at in response, got %+v", payload.Data)
	}
}

func TestHandler_AckReceipt_RequiresActorID(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})

	req := httptest.NewRequest("POST", "/api/messages/1/receipt", bytes.NewBufferString(`{}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.AckReceipt(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_Update_BusinessError_BadRequest(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.UpdateMessageResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "BAD_REQUEST",
					Message: "err",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("PUT", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(`{"content":"ok"}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}

func TestHandler_Delete_BusinessError_BadRequest(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.DeleteMessageResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "BAD_REQUEST",
					Message: "err",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("DELETE", "/api/messages/1?actor_id=a0000001-0000-0000-0000-000000000001", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Delete(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}
}
