package main

import (
	"encoding/json"
	"gateway/internal/common"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestSetupServer(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				type respWrapper struct {
					Response struct {
						IsValid bool `json:"valid"`
						User    struct {
							ID       string `json:"id"`
							Username string `json:"username"`
						} `json:"user"`
					} `json:"response"`
				}
				var resp respWrapper
				resp.Response.IsValid = true
				resp.Response.User.ID = "user-123"
				resp.Response.User.Username = "testuser"
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{}, nil
		},
	}
	r := SetupServer(mockNc)

	ts := httptest.NewServer(r)
	defer ts.Close()

	// 1. Health check
	res, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %d", res.StatusCode)
	}

	// 2. Check route registration for various modules
	routes := []struct {
		method string
		path   string
	}{
		{"POST", "/auth/register"},
		{"POST", "/auth/login"},
		{"GET", "/users/123"},
		{"POST", "/api/messages"},
	}

	for _, rt := range routes {
		req, _ := http.NewRequest(rt.method, ts.URL+rt.path, nil)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Errorf("Error requesting %s %s: %v", rt.method, rt.path, err)
			continue
		}
		if res.StatusCode == http.StatusNotFound {
			t.Errorf("Route %s %s not found", rt.method, rt.path)
		}
	}

	// 3. Test /ws auth logic (failure cases)
	// No token
	res, _ = http.Get(ts.URL + "/ws")
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for /ws without token, got %d", res.StatusCode)
	}

	// Invalid token (we'll change the mock behavior for a second server or just use a different path if we had one, but let's just test positive case here since we can't easily change the mock in this test structure without a specialized mock)
}

func TestSetupServer_WS_AuthFail(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				resp := map[string]interface{}{
					"response": map[string]interface{}{
						"valid": false,
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{}, nil
		},
	}
	r := SetupServer(mockNc)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/ws?token=invalid", nil)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for /ws with invalid token, got %d", res.StatusCode)
	}
}

func TestSetupServer_WS_NATSError(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			return nil, nats.ErrTimeout
		},
	}
	r := SetupServer(mockNc)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/ws?token=valid", nil)
	res, _ := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected 503 for /ws on NATS error, got %d", res.StatusCode)
	}
}

func TestSetupServer_WS_Bearer(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				resp := map[string]interface{}{
					"response": map[string]interface{}{
						"valid": true,
						"user":  map[string]string{"id": "user-123"},
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{}, nil
		},
	}
	r := SetupServer(mockNc)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/ws", nil)
	req.Header.Set("Authorization", "Bearer token123")
	res, _ := http.DefaultClient.Do(req)
	// it will fail to upgrade but should pass auth
	if res.StatusCode == http.StatusUnauthorized {
		t.Error("Bearer token not recognized in /ws")
	}
}

func TestSetupServer_WS_Valid(t *testing.T) {
	mockNc := &common.MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "auth.validate" {
				resp := map[string]interface{}{
					"response": map[string]interface{}{
						"valid": true,
						"user":  map[string]string{"id": "user-123", "username": "testuser"},
					},
				}
				respBytes, _ := json.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{}, nil
		},
	}
	r := SetupServer(mockNc)
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL+"/ws?token=valid", nil)
	res, _ := http.DefaultClient.Do(req)
	// It will return whatever the upgrader writes, but it won't be 401.
	if res.StatusCode == http.StatusUnauthorized {
		t.Error("Valid token should not be unauthorized")
	}
}
