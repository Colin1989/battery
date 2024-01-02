package facade

import (
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
	"net"
)

// MessageEncoder interface
type MessageEncoder interface {
	Encode(message *message.Message) ([]byte, error)
}

type PacketDecoder interface {
	Decode(data []byte) ([]*packet.Packet, error)
}

type PacketEncoder interface {
	Encode(typ packet.Type, data []byte) ([]byte, error)
}

type PacketProcessor interface {
	ProcessPacket(IAgent, *packet.Packet) error
}

type Connector interface {
	GetNextMessage() (b []byte, err error)
	RemoteAddr() net.Addr
	net.Conn
}
