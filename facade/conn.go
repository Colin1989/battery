package facade

import "github.com/colin1989/battery/net/packet"

type PackDecoder interface {
	Decode(data []byte) ([]*packet.Packet, error)
}

type PackEncoder interface {
	Encode(typ packet.Type, data []byte) ([]byte, error)
}
