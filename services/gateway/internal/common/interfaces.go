package common

import (
	"time"

	"github.com/nats-io/nats.go"
)

// NatsConn defines the subset of nats.Conn methods used across the gateway service.
type NatsConn interface {
	Publish(subject string, data []byte) error
	Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error)
	Request(subject string, data []byte, timeout time.Duration) (*nats.Msg, error)
}
