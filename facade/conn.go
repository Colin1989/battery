package facade

import (
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
	"net"
)

// Encoder interface
type Encoder interface {
	IsCompressionEnabled() bool
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

type Codec interface {
	PacketDecoder
	PacketEncoder
	PacketProcessor
}

type Connector interface {
	GetNextMessage() (b []byte, err error)
	RemoteAddr() net.Addr
	net.Conn
}
