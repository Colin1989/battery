package facade

import (
	"github.com/colin1989/battery/net/packet"
	"net"
)

type PackDecoder interface {
	Decode(data []byte) ([]*packet.Packet, error)
}

type PackEncoder interface {
	Encode(typ packet.Type, data []byte) ([]byte, error)
}

type Connector interface {
	GetNextMessage() (b []byte, err error)
	RemoteAddr() net.Addr
	net.Conn
}
