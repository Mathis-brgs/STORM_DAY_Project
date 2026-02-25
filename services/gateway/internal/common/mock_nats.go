package common

import (
	"time"

	"github.com/nats-io/nats.go"
)

// MockNatsConn is a mock implementation of the NatsConn interface for testing.
type MockNatsConn struct {
	PublishFunc   func(subject string, data []byte) error
	SubscribeFunc func(subject string, cb nats.MsgHandler) (*nats.Subscription, error)
	RequestFunc   func(subject string, data []byte, timeout time.Duration) (*nats.Msg, error)
}

func (m *MockNatsConn) Publish(subject string, data []byte) error {
	if m.PublishFunc != nil {
		return m.PublishFunc(subject, data)
	}
	return nil
}

func (m *MockNatsConn) Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error) {
	if m.SubscribeFunc != nil {
		return m.SubscribeFunc(subject, cb)
	}
	return nil, nil
}

func (m *MockNatsConn) Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	if m.RequestFunc != nil {
		return m.RequestFunc(subject, data, timeout)
	}
	return &nats.Msg{}, nil
}
