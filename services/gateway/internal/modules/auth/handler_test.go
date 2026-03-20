package auth

import (
	"bytes"
	"encoding/json"
	"gateway/internal/common"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nats-io/nats.go"
)

// makeTestToken génère un JWT valide signé avec la clé de test.
func makeTestToken(userID, username string, expired bool) string {
	exp := time.Now().Add(15 * time.Minute)
	if expired {
		exp = time.Now().Add(-1 * time.Minute)
	}
	c := &claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	t, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte("storm-secret-key"))
	return t
}

func TestHandler_Register(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject != "auth.register" {
				t.Errorf("Expected subject auth.register, got %s", subject)
			}
			resp := map[string]interface{}{
				"response": map[string]string{"id": "user-123", "username": "testuser"},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	body := map[string]string{"username": "testuser", "password": "password"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created, got %d", w.Code)
	}

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Error unmarshalling response: %v", err)
	}
	if resp["username"] != "testuser" {
		t.Errorf("Expected username testuser, got %v", resp["username"])
	}
}

func TestHandler_Login(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := map[string]interface{}{
				"response": map[string]string{"access_token": "token123"},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	body := map[string]string{"username": "testuser", "password": "password"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Logout(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := map[string]interface{}{
				"response": map[string]bool{"success": true},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}

	handler := NewHandler(mockNc)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+makeTestToken("user-123", "testuser", false))
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Logout_Bearer(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte(`{"response":{"success":true}}`)}, nil
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+makeTestToken("user-123", "testuser", false))
	w := httptest.NewRecorder()
	handler.Logout(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestHandler_Register_JSONError(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBufferString("invalid json"))
	w := httptest.NewRecorder()
	handler.Register(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %d", w.Code)
	}
}

func TestHandler_Register_NATSError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return nil, nats.ErrTimeout
		},
	}
	handler := NewHandler(mockNc)
	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()
	handler.Register(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %d", w.Code)
	}
}

func TestHandler_Register_UnwrappedResponse(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte(`{"direct": "data"}`)}, nil
		},
	}
	handler := NewHandler(mockNc)
	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()
	handler.Register(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status Created, got %d", w.Code)
	}
}

func TestHandler_Logout_ValidationInvalid(t *testing.T) {
	handler := NewHandler(&common.MockNatsConn{})
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	handler.Logout(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %d", w.Code)
	}
}

func TestHandler_Logout_NATSError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return nil, nats.ErrTimeout
		},
	}
	handler := NewHandler(mockNc)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+makeTestToken("user-123", "testuser", false))
	w := httptest.NewRecorder()
	handler.Logout(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status ServiceUnavailable, got %d", w.Code)
	}
}

func TestHandler_Refresh(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			resp := map[string]interface{}{
				"response": map[string]string{"access_token": "new-token"},
			}
			respBytes, _ := json.Marshal(resp)
			return &nats.Msg{Data: respBytes}, nil
		},
	}
	handler := NewHandler(mockNc)
	bodyBytes, _ := json.Marshal(map[string]string{"refresh_token": "old-token"})
	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()
	handler.Refresh(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %d", w.Code)
	}
}

func TestValidateToken_Empty(t *testing.T) {
	res, err := ValidateToken("")
	if err != nil {
		t.Fatal(err)
	}
	if res.IsValid {
		t.Error("Expected valid=false for empty token")
	}
}

func TestValidateToken_InvalidFormat(t *testing.T) {
	res, err := ValidateToken("not-a-jwt")
	if err != nil {
		t.Fatal(err)
	}
	if res.IsValid {
		t.Error("Expected valid=false for malformed token")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	token := makeTestToken("user-123", "testuser", false)
	res, err := ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if !res.IsValid {
		t.Error("Expected valid=true for valid token")
	}
	if res.User.ID != "user-123" {
		t.Errorf("Expected user ID user-123, got %s", res.User.ID)
	}
	if res.User.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", res.User.Username)
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	token := makeTestToken("user-123", "testuser", true)
	res, err := ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if res.IsValid {
		t.Error("Expected valid=false for expired token")
	}
}

func TestProxyRequest_Fallback(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("direct response")}, nil
		},
	}
	handler := NewHandler(mockNc)
	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()
	handler.Login(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError for unparseable response, got %d", w.Code)
	}
}

func TestProxyRequest_UnmarshalError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return &nats.Msg{Data: []byte("invalid json")}, nil
		},
	}
	handler := NewHandler(mockNc)
	bodyBytes, _ := json.Marshal(map[string]string{"foo": "bar"})
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()
	handler.Login(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status InternalServerError for invalid json, got %d", w.Code)
	}
}
