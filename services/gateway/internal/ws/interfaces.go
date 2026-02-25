package ws

import (
	"net"

	"github.com/lxzan/gws"
	"github.com/nats-io/nats.go"
)

// NatsConn defines the subset of nats.Conn methods used by the WS package.
type NatsConn interface {
	Publish(subject string, data []byte) error
	Subscribe(subject string, cb nats.MsgHandler) (*nats.Subscription, error)
}

// Socket defines the subset of gws.Conn methods used by the WS package.
type Socket interface {
	RemoteAddr() net.Addr
	WriteMessage(opcode gws.Opcode, payload []byte) error
	WritePing(payload []byte) error
	WritePong(payload []byte) error
	Session() gws.SessionStorage
}

// WSMessage defines the subset of gws.Message methods used.
type WSMessage interface {
	Bytes() []byte
	Close() error
}
