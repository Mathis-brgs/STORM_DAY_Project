package ws

import (
	"gateway/internal/common"
	"net"

	"github.com/lxzan/gws"
)

// NatsConn is an alias for common.NatsConn for backwards compatibility within the ws package.
type NatsConn = common.NatsConn

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
