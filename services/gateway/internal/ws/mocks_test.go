package ws

import (
	"net"
	"sync"
	"time"

	"github.com/lxzan/gws"
	"github.com/nats-io/nats.go"
)

type MockNatsConn struct {
	PublishFunc          func(subject string, data []byte) error
	SubscribeFunc        func(subject string, cb nats.MsgHandler) (*nats.Subscription, error)
	RequestFunc          func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error)
	LastPublishedSubject string
	LastPublishedData    []byte
}

func (m *MockNatsConn) Publish(subject string, data []byte) error {
	m.LastPublishedSubject = subject
	m.LastPublishedData = data
	if m.PublishFunc != nil {
		return m.PublishFunc(subject, data)
	}
	return nil
}

func (m *MockNatsConn) Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(subject, cb)
	}
	return &nats.Subscription{}, nil
}

func (m *MockNatsConn) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	if m.RequestFunc != nil {
		return m.RequestFunc(subject, data, timeout)
	}
	return &nats.Msg{}, nil
}

type MockAddr struct {
	addr string
}

func (m MockAddr) Network() string { return "tcp" }
func (m MockAddr) String() string  { return m.addr }

type MockSocket struct {
	addr               string
	writeFunc          func(opcode gws.Opcode, payload []byte) error
	pingFunc           func(payload []byte) error
	pongFunc           func(payload []byte) error
	session            *MockSession
	WriteCount         int
	LastPayload        []byte
	LastOpcode         gws.Opcode
	RemoteAddrOverride string
}

func (m *MockSocket) RemoteAddr() net.Addr {
	if m.RemoteAddrOverride != "" {
		return MockAddr{addr: m.RemoteAddrOverride}
	}
	return MockAddr{addr: m.addr}
}
func (m *MockSocket) WriteMessage(opcode gws.Opcode, payload []byte) error {
	m.WriteCount++
	m.LastPayload = payload
	m.LastOpcode = opcode
	if m.writeFunc != nil {
		return m.writeFunc(opcode, payload)
	}
	return nil
}
func (m *MockSocket) WritePing(payload []byte) error {
	if m.pingFunc != nil {
		return m.pingFunc(payload)
	}
	return nil
}
func (m *MockSocket) WritePong(payload []byte) error {
	if m.pongFunc != nil {
		return m.pongFunc(payload)
	}
	return nil
}
func (m *MockSocket) Session() gws.SessionStorage {
	if m.session == nil {
		m.session = &MockSession{data: make(map[string]any)}
	}
	return m.session
}

type MockSession struct {
	mu   sync.RWMutex
	data map[string]any
}

func (s *MockSession) Load(key string) (value any, exist bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exist = s.data[key]
	return
}

func (s *MockSession) Store(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *MockSession) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *MockSession) Range(f func(key string, value any) bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.data {
		if !f(k, v) {
			break
		}
	}
}

func (s *MockSession) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

type MockMessage struct {
	payload   []byte
	closed    bool
	closeFunc func() error
}

func (m *MockMessage) Bytes() []byte { return m.payload }
func (m *MockMessage) Close() error {
	m.closed = true
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}
