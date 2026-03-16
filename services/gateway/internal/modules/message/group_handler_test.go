package message

import (
	"bytes"
	"context"
	"gateway/internal/common"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func TestHandler_CreateGroup(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject != subjectGroupCreate {
				t.Fatalf("expected subject %s, got %s", subjectGroupCreate, subject)
			}

			var req apiv1.GroupCreateRequest
			if err := proto.Unmarshal(data, &req); err != nil {
				t.Fatalf("invalid request payload: %v", err)
			}
			if req.GetActorId() == "" {
				t.Fatalf("expected actor_id in proto request")
			}

			resp := &apiv1.GroupCreateResponse{
				Ok: true,
				Data: &apiv1.Group{
					Id:   10,
					Name: "Backend",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	body := `{"actor_id":"a0000001-0000-0000-0000-000000000001","name":"Backend"}`
	req := httptest.NewRequest("POST", "/api/groups", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	handler.CreateGroup(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}

func TestHandler_ListGroups_RequiresActorID(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("GET", "/api/groups", nil)
	w := httptest.NewRecorder()

	handler.ListGroups(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandler_GetGroup_MapsForbidden(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := &apiv1.GroupGetResponse{
				Ok: false,
				Error: &apiv1.Error{
					Code:    "FORBIDDEN",
					Message: "forbidden",
				},
			}
			respBytes, _ := proto.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/api/groups/5?actor_id=a0000001-0000-0000-0000-000000000001", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.GetGroup(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", w.Code)
	}
}

func TestHandler_AddGroupMember_RequiresUserID(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("POST", "/api/groups/5/members?actor_id=a0000001-0000-0000-0000-000000000001", bytes.NewBufferString(`{"role":0}`))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "5")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	handler.AddGroupMember(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
