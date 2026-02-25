package ws

import (
	"net"
	"testing"

	"github.com/lxzan/gws"
	"github.com/nats-io/nats.go"
)

func TestHub(t *testing.T) {
	t.Run("NewHub", func(t *testing.T) {
		hub := NewHub()
		if hub == nil || hub.Rooms == nil {
			t.Fatal("Failed to initialize hub")
		}
	})

	t.Run("Join and Leave", func(t *testing.T) {
		hub := NewHub()
		socket := &MockSocket{addr: "127.0.0.1:1234"}
		room := "room1"

		hub.Join(room, socket)
		if len(hub.Rooms[room]) != 1 {
			t.Errorf("Expected 1 client in room, got %d", len(hub.Rooms[room]))
		}

		hub.Leave(room, socket)
		if len(hub.Rooms) != 0 {
			t.Errorf("Expected room to be deleted after last client left, but it exists with %d clients", len(hub.Rooms[room]))
		}
	})

	t.Run("BroadcastToRoom", func(t *testing.T) {
		hub := NewHub()
		socket1 := &MockSocket{addr: "1"}
		socket2 := &MockSocket{addr: "2"}
		room := "room1"

		hub.Join(room, socket1)
		hub.Join(room, socket2)

		payload := []byte("hello")
		hub.BroadcastToRoom(room, payload)

		if socket1.WriteCount != 1 || socket2.WriteCount != 1 {
			t.Errorf("Expected 1 write per socket, got %d and %d", socket1.WriteCount, socket2.WriteCount)
		}
	})

	t.Run("StartNatsSubscription", func(t *testing.T) {
		hub := NewHub()
		var handler nats.MsgHandler
		mockNats := &MockNatsConn{
			SubscribeFunc: func(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
				handler = cb
				return &nats.Subscription{}, nil
			},
		}

		err := hub.StartNatsSubscription(mockNats)
		if err != nil {
			t.Fatalf("Failed to start NATS subscription: %v", err)
		}
		socket := &MockSocket{addr: "1"}
		hub.Join("group:123", socket)

		msg := &nats.Msg{
			Subject: "message.broadcast.group:123",
			Data:    []byte("broadcast test"),
		}
		handler(msg)

		if socket.WriteCount != 1 {
			t.Errorf("Expected broadcast to be redistribted to WS, got %d writes", socket.WriteCount)
		}
	})

	t.Run("BroadcastToNonExistentRoom", func(t *testing.T) {
		hub := NewHub()
		hub.BroadcastToRoom("ghost", []byte("ignore me"))
		// Should not panic, just return
	})

	t.Run("StartNatsSubscription Invalid Subject", func(t *testing.T) {
		hub := NewHub()
		var handler nats.MsgHandler
		mockNats := &MockNatsConn{
			SubscribeFunc: func(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
				handler = cb
				return &nats.Subscription{}, nil
			},
		}
		err := hub.StartNatsSubscription(mockNats)
		if err != nil {
			t.Fatalf("Failed to start NATS subscription: %v", err)
		}

		handler(&nats.Msg{Subject: "too.short"}) // Should handle < 3 parts
	})

	t.Run("BroadcastToRoom Write Error", func(t *testing.T) {
		hub := NewHub()
		socket := &MockSocket{
			addr: "1",
			writeFunc: func(opcode gws.Opcode, payload []byte) error {
				return net.ErrClosed
			},
		}
		hub.Join("room1", socket)
		hub.BroadcastToRoom("room1", []byte("fail"))
	})
}
