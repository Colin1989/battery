package codec

import (
	"github.com/colin1989/battery/net/packet"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name    string
		pType   packet.Type
		size    int
		wantErr error
		extra   byte
	}{
		// TODO: Add test cases.
		{name: "Handshake", pType: packet.Handshake, size: 1, wantErr: nil},
		{name: "HandshakeAck", pType: packet.HandshakeAck, size: 2, wantErr: nil},
		{name: "Heartbeat", pType: packet.Heartbeat, size: 3, wantErr: nil},
		{name: "Data", pType: packet.Data, size: 4, wantErr: nil},
		{name: "Kick", pType: packet.Kick, size: MaxPacketSize - 1, wantErr: nil},
		{name: "ErrInvalidPomeloHeader", pType: packet.Handshake, size: 5, wantErr: packet.ErrInvalidPomeloHeader, extra: 1},
		{name: "ErrWrongPomeloPacketType", pType: 0, size: 5, wantErr: packet.ErrWrongPomeloPacketType},
		{name: "ErrWrongPomeloPacketType", pType: 6, size: 5, wantErr: packet.ErrWrongPomeloPacketType},
		// it's will be never happen
		//{name: "ErrWrongPomeloPacketType", pType: packet.Data, size: MaxPacketSize - 1, wantErr: ErrPacketSizeExceed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := append([]byte{byte(tt.pType)}, IntToBytes(tt.size)...)
			if tt.extra != 0 {
				header = append(header, tt.extra)
			}
			typ, size, err := ParseHeader(header)
			if err != tt.wantErr {
				t.Errorf("ParseHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if size != tt.size {
				t.Errorf("ParseHeader() got = %v, want %v", size, tt.size)
			}
			if typ != tt.pType {
				t.Errorf("ParseHeader() got1 = %v, want %v", typ, tt.pType)
			}
		})
	}
}

func TestBytesToIntIntToBytes(t *testing.T) {
	for i := 0; i < 100; i++ {
		n := int(rand.Int31()) >> 8
		bytes := IntToBytes(n)
		v := BytesToInt(bytes)
		assert.Equal(t, n, v)
	}
}

func TestBytesToInt(t *testing.T) {
	bytes := []byte{0x0ff, 0x0ff, 0x0ff}
	//bytes := IntToBytes(MaxPacketSize)
	v := BytesToInt(bytes)
	assert.Equal(t, MaxPacketSize-1, v)
}
