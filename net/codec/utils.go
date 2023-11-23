package codec

import "github.com/colin1989/battery/net/packet"

func ParseHeader(header []byte) (packet.Type, int, error) {
	if len(header) != 4 {
		return 0, 0, packet.ErrInvalidPomeloHeader
	}

	typ := header[0]
	if typ < packet.Handshake || packet.Kick < typ {
		return 0, 0, packet.ErrWrongPomeloPacketType
	}

	size := BytesToInt(header[1:])
	if size > MaxPacketSize {
		return 0, 0, packet.ErrPacketSizeExceed
	}

	return packet.Type(typ), size, nil
}

// BytesToInt decode packet data length byte to int(Big end)
func BytesToInt(b []byte) int {
	result := 0
	for _, v := range b {
		result = result<<8 + int(v)
	}
	return result
}

// IntToBytes encode packet data length to bytes(Big end)
func IntToBytes(v int) []byte {
	buf := make([]byte, 3)
	buf[0] = byte((v >> 16) & 0xFF)
	buf[1] = byte((v >> 8) & 0xFF)
	buf[2] = byte(v & 0xFF)
	return buf
}
