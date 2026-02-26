package ws

import (
	"encoding/json"
	"gateway/internal/models"
	"net"
	"sync"
	"testing"
	"time"

	apiv1 "github.com/Mathis-brgs/storm-project/services/message/api/v1"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func TestHandler_OnOpen(t *testing.T) {
	hub := NewHub()
	mockNats := &MockNatsConn{}
	handler := NewHandler(hub, mockNats)
	socket := &MockSocket{addr: "127.0.0.1:1234"}
	socket.Session().Store("userId", "user1")
	socket.Session().Store("username", "testuser")

	handler.onOpen(socket)

	// Verify user joined their private room
	if _, exists := hub.Rooms["user:user1"]; !exists {
		t.Error("User should have joined their private room")
	}

	// Wait a bit to see if heartbeat starts (though we can't easily wait 30s)
	// But we can check if it at least doesn't crash.
}

func TestHandler_OnPing(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, &MockNatsConn{})
	socket := &MockSocket{}
	payload := []byte("ping")

	handler.onPing(socket, payload)
}

func TestHandler_OnClose(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, &MockNatsConn{})
	socket := &MockSocket{addr: "1"}
	room := "group:1"
	hub.Join(room, socket)
	socket.Session().Store("room", room)

	handler.onClose(socket, nil)

	if len(hub.Rooms) != 0 {
		t.Error("User should have left the room on close")
	}
}

func TestHandler_OnMessage_Join(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, &MockNatsConn{})
	socket := &MockSocket{addr: "1"}

	msg := models.InputMessage{
		Action: models.WSActionJoin,
		Room:   "group:123",
	}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)

	if _, exists := hub.Rooms["group:123"]; !exists {
		t.Error("User should have joined the room")
	}

	room, _ := socket.Session().Load("room")
	if room != "group:123" {
		t.Errorf("Expected room group:123 in session, got %v", room)
	}
}

func TestHandler_OnMessage_Message(t *testing.T) {
	hub := NewHub()
	mockNats := &MockNatsConn{
		RequestFunc: func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
			if subject == "NEW_MESSAGE" {
				// Simuler une réponse positive du message-service
				resp := &apiv1.SendMessageResponse{
					Ok: true,
					Data: &apiv1.ChatMessage{
						Id:       1,
						SenderId: "456",
						GroupId:  123,
						Content:  "hello",
					},
				}
				respBytes, _ := proto.Marshal(resp)
				return &nats.Msg{Data: respBytes}, nil
			}
			return &nats.Msg{}, nil
		},
	}
	handler := NewHandler(hub, mockNats)
	socket := &MockSocket{addr: "1"}
	socket.Session().Store("userId", "456")

	// On rejoint la room pour pouvoir capter le broadcast
	hub.Join("group:123", socket)

	msg := models.InputMessage{
		Action:  models.WSActionMessage,
		Room:    "group:123",
		Content: "hello",
	}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)

	// Vérifier que le socket a reçu le message diffusé (Echo)
	// WriteCount devrait être 1 car l'envoyeur reçoit aussi son message
	if socket.WriteCount == 0 {
		t.Error("Expected broadcast message (Echo) to be sent to the socket")
	}

	var res models.InputMessage
	if err := json.Unmarshal(socket.LastPayload, &res); err != nil {
		t.Fatalf("Failed to unmarshal broadcast payload: %v", err)
	}
	if res.Content != "hello" {
		t.Errorf("Expected broadcast content 'hello', got %s", res.Content)
	}
}

func TestHandler_OnMessage_Message_InvalidRoom(t *testing.T) {
	hub := NewHub()
	mockNats := &MockNatsConn{}
	handler := NewHandler(hub, mockNats)
	socket := &MockSocket{addr: "1"}

	msg := models.InputMessage{
		Action: models.WSActionMessage,
		Room:   "invalid_room", // missing colon or wrong prefix
	}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)

	if mockNats.LastPublishedSubject != "" {
		t.Error("Should not have published to NATS for invalid room")
	}
}

func TestHandler_OnMessage_Message_InvalidIDs(t *testing.T) {
	hub := NewHub()
	mockNats := &MockNatsConn{}
	handler := NewHandler(hub, mockNats)
	socket := &MockSocket{addr: "1"}
	socket.Session().Store("userId", "not-a-number")

	msg := models.InputMessage{
		Action:  models.WSActionMessage,
		Room:    "group:abc", // not-a-number group id
		Content: "hello",
	}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)

	if mockNats.LastPublishedSubject != "" {
		t.Error("Should not have published to NATS for invalid group ID")
	}
}

func TestHandler_OnMessage_NatsPublishError(t *testing.T) {
	hub := NewHub()
	mockNats := &MockNatsConn{
		PublishFunc: func(subject string, data []byte) error {
			return nats.ErrNoResponders
		},
	}
	handler := NewHandler(hub, mockNats)
	socket := &MockSocket{addr: "1"}
	socket.Session().Store("userId", "123")

	msg := models.InputMessage{
		Action:  models.WSActionMessage,
		Room:    "group:123",
		Content: "hello",
	}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)
	// Should not panic
}

func TestHandler_OnMessage_Typing(t *testing.T) {
	hub := NewHub()
	mockNats := &MockNatsConn{}
	handler := NewHandler(hub, mockNats)
	socket := &MockSocket{addr: "1"}
	socket.Session().Store("userId", "789")

	msg := models.InputMessage{
		Action: models.WSActionTyping,
		Room:   "group:123",
	}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)

	if mockNats.LastPublishedSubject != "message.broadcast.group:123" {
		t.Errorf("Expected NATS publish to message.broadcast.group:123, got %s", mockNats.LastPublishedSubject)
	}
}

func TestHandler_OnMessage_InvalidJSON(t *testing.T) {
	handler := NewHandler(NewHub(), &MockNatsConn{})
	socket := &MockSocket{}
	message := &MockMessage{payload: []byte("invalid json")}

	handler.onMessage(socket, message)
	// Should not panic, just return
}

func TestHandler_OnMessage_UnknownAction(t *testing.T) {
	handler := NewHandler(NewHub(), &MockNatsConn{})
	socket := &MockSocket{}
	msg := models.InputMessage{Action: "unknown"}
	payload, _ := json.Marshal(msg)
	message := &MockMessage{payload: payload}

	handler.onMessage(socket, message)
	// Should not panic
}

func TestHandler_OnOpen_Heartbeat(t *testing.T) {
	hub := NewHub()
	handler := NewHandler(hub, &MockNatsConn{})
	handler.HeartbeatInterval = 10 * time.Millisecond

	var pingCalled int
	var mu sync.Mutex
	socket := &MockSocket{
		addr: "1",
		pingFunc: func(payload []byte) error {
			mu.Lock()
			pingCalled++
			mu.Unlock()
			return nil
		},
	}

	handler.onOpen(socket)

	// Wait for at least one ping
	time.Sleep(25 * time.Millisecond)

	mu.Lock()
	if pingCalled == 0 {
		t.Error("Heartbeat ping was not called")
	}
	mu.Unlock()

	// Test write ping error to cover goroutine exit
	socket.pingFunc = func(payload []byte) error {
		return net.ErrClosed
	}
	time.Sleep(20 * time.Millisecond)
}

func TestHandler_OnMessage_CloseError(t *testing.T) {
	handler := NewHandler(NewHub(), &MockNatsConn{})
	socket := &MockSocket{}
	msg := models.InputMessage{Action: models.WSActionJoin, Room: "r"}
	payload, _ := json.Marshal(msg)

	message := &MockMessage{
		payload: payload,
		closeFunc: func() error {
			return net.ErrClosed
		},
	}

	handler.onMessage(socket, message)
	// Should log error and continue
}
