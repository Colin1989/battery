package codec

import (
	"bytes"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/packet"
)

var _ facade.PacketDecoder = (*PomeloPacketDecoder)(nil)

// PomeloPacketDecoder reads and decodes network data slice following pomelo's protocol
type PomeloPacketDecoder struct{}

func NewPomeloPacketDecoder() *PomeloPacketDecoder {
	return &PomeloPacketDecoder{}
}

func (p *PomeloPacketDecoder) forward(buf *bytes.Buffer) (packet.Type, int, error) {
	header := buf.Next(HeadLength)
	return ParseHeader(header)
}

func (p *PomeloPacketDecoder) Decode(data []byte) ([]*packet.Packet, error) {
	var (
		buf     = bytes.NewBuffer(nil)
		packets []*packet.Packet
		err     error
	)

	buf.Write(data)
	if buf.Len() < HeadLength {
		return nil, constant.ErrInvalidPomeloHeader
	}

	typ, size, err := p.forward(buf)
	if err != nil {
		return nil, err
	}

	for size <= buf.Len() {
		pkt := &packet.Packet{
			Type: typ,
			Data: buf.Next(size),
		}

		packets = append(packets, pkt)

		if buf.Len() < HeadLength {
			break
		}

		typ, size, err = p.forward(buf)
		if err != nil {
			return nil, err
		}
	}

	return packets, nil
}
