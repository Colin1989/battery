package codec

import (
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/packet"
)

var _ facade.PacketEncoder = (*PomeloPacketEncoder)(nil)

type PomeloPacketEncoder struct{}

func NewPomeloPacketEncoder() *PomeloPacketEncoder {
	return &PomeloPacketEncoder{}
}

// Encode create a packet.Packet from  the raw bytes slice and then encode to network bytes slice
// Protocol refs: https://github.com/NetEase/pomelo/wiki/Communication-Protocol
//
// -<type>-|--------<length>--------|-<data>-
// --------|------------------------|--------
// 1 byte packet type, 3 bytes packet data length(big end), and data segment
func (e *PomeloPacketEncoder) Encode(typ packet.Type, data []byte) ([]byte, error) {
	if err := packet.IsPacketType(typ); err != nil {
		return nil, err
	}
	if len(data) > MaxPacketSize {
		return nil, packet.ErrPacketSizeExceed
	}
	p := &packet.Packet{
		Type: typ,
		Data: data,
	}

	buf := make([]byte, HeadLength+p.Length())
	buf[0] = byte(p.Type)
	copy(buf[1:], IntToBytes(p.Length()))
	copy(buf[HeadLength:], p.Data)

	return buf, nil
}
