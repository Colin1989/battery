package codec

import (
	"github.com/colin1989/battery/net/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BenchmarkPomeloPacketEncoder_Encode(b *testing.B) {
	encoder := NewPomeloPacketEncoder()
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	for i := 0; i < b.N; i++ {
		b.Run("", func(b2 *testing.B) {
			_, err := encoder.Encode(packet.Data, data)
			if err != nil {
				b.Error(err)
			}
		})
	}
}

func helperConcatBytes(packetType packet.Type, length, data []byte) []byte {
	if data == nil {
		return nil
	}

	bytes := []byte{}
	bytes = append(bytes, byte(packetType))
	bytes = append(bytes, length...)
	bytes = append(bytes, data...)
	return bytes
}

var tooBigData = make([]byte, 1<<25)

var encodeTables = map[string]struct {
	packetType packet.Type
	length     []byte
	data       []byte
	err        error
}{
	"test_encode_handshake":    {packet.Handshake, []byte{0x00, 0x00, 0x02}, []byte{0x01, 0x00}, nil},
	"test_invalid_packet_type": {0xff, nil, nil, packet.ErrWrongPomeloPacketType},
	"test_too_big_packet":      {packet.Data, nil, tooBigData, packet.ErrPacketSizeExceed},
}

func TestEncode(t *testing.T) {
	t.Parallel()

	for name, table := range encodeTables {
		t.Run(name, func(t *testing.T) {
			ppe := NewPomeloPacketEncoder()

			encoded, err := ppe.Encode(table.packetType, table.data)
			if table.err != nil {
				assert.Equal(t, table.err, err)
			} else {
				expectedEncoded := helperConcatBytes(table.packetType, table.length, table.data)
				assert.Equal(t, expectedEncoded, encoded)
			}
		})
	}
}
