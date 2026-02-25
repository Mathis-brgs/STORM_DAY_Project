package user

import (
	"bytes"
	"context"
	"encoding/json"
	"gateway/internal/common"
	"gateway/internal/modules/auth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nats-io/nats.go"
)

func TestHandler_Get(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := map[string]interface{}{
				"response": map[string]string{"id": "user-123", "username": "testuser"},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)

	// Use chi context for URL param
	req := httptest.NewRequest("GET", "/users/user-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}

	var res map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &res)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	if res["id"] != "user-123" {
		t.Errorf("Expected id user-123, got %v", res["id"])
	}
}

func TestHandler_Update(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				type respWrapper struct {
					Response auth.ValidationResult `json:"response"`
				}
				resp := respWrapper{
					Response: auth.ValidationResult{
						IsValid: true,
						User:    auth.UserInfo{ID: "user-123"},
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			resp := map[string]interface{}{
				"response": map[string]string{"id": "user-123", "status": "updated"},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	body := map[string]string{"username": "newname"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "token")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Update_Forbidden(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			type respWrapper struct {
				Response auth.ValidationResult `json:"response"`
			}
			resp := respWrapper{
				Response: auth.ValidationResult{
					IsValid: true,
					User:    auth.UserInfo{ID: "user-123"},
				},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	req := httptest.NewRequest("PUT", "/users/other-user", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Authorization", "token")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "other-user")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status Forbidden, got %d", w.Code)
	}
}

func TestHandler_Get_NATSError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return nil, nats.ErrTimeout
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/users/user-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Get(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %d", w.Code)
	}
}

func TestHandler_Update_JSONError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			type respWrapper struct {
				Response auth.ValidationResult `json:"response"`
			}
			resp := respWrapper{
				Response: auth.ValidationResult{
					IsValid: true,
					User:    auth.UserInfo{ID: "user-123"},
				},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBufferString("invalid json"))
	req.Header.Set("Authorization", "token")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "user-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_Update_NoToken(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("PUT", "/users/user-123", nil)
	w := httptest.NewRecorder()
	handler.Update(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %d", w.Code)
	}
}

func TestHandler_Update_ValidationError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return nil, nats.ErrNoResponders
		},
	}
	handler := NewHandler(mockNc)
	r := chi.NewRouter()
	r.Put("/users/{id}", handler.Update)

	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %d", w.Code)
	}
}

func TestHandler_Update_ValidationInvalid(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			type respWrapper struct {
				Response auth.ValidationResult `json:"response"`
			}
			resp := respWrapper{
				Response: auth.ValidationResult{IsValid: false},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	r := chi.NewRouter()
	r.Put("/users/{id}", handler.Update)

	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %d", w.Code)
	}
}

func TestHandler_Update_NATSError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				type respWrapper struct {
					Response auth.ValidationResult `json:"response"`
				}
				resp := respWrapper{
					Response: auth.ValidationResult{
						IsValid: true,
						User:    auth.UserInfo{ID: "user-123"},
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return nil, nats.ErrTimeout
		},
	}
	handler := NewHandler(mockNc)
	r := chi.NewRouter()
	r.Put("/users/{id}", handler.Update)

	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %d", w.Code)
	}
}

func TestHandler_Update_Bearer(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				type respWrapper struct {
					Response auth.ValidationResult `json:"response"`
				}
				resp := respWrapper{
					Response: auth.ValidationResult{
						IsValid: true,
						User:    auth.UserInfo{ID: "user-123"},
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{Data: []byte(`{"response":{"id":"user-123"}}`)}, nil
		},
	}
	handler := NewHandler(mockNc)
	r := chi.NewRouter()
	r.Put("/users/{id}", handler.Update)
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBufferString(`{"foo":"bar"}`))
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Get_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("direct data")}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("GET", "/users/123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()
	handler.Get(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Update_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				type respWrapper struct {
					Response auth.ValidationResult `json:"response"`
				}
				resp := respWrapper{
					Response: auth.ValidationResult{
						IsValid: true,
						User:    auth.UserInfo{ID: "user-123"},
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{Data: []byte("direct data")}, nil
		},
	}
	handler := NewHandler(mockNc)
	r := chi.NewRouter()
	r.Put("/users/{id}", handler.Update)
	req := httptest.NewRequest("PUT", "/users/user-123", bytes.NewBufferString(`{"foo":"bar"}`))
	req.Header.Set("Authorization", "token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}
