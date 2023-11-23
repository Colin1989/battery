package packet

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPacket(t *testing.T) {
	p := New()
	assert.NotNil(t, p)
}

func TestPacket_String(t *testing.T) {
	type fields struct {
		Type Type
		Data []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
		{
			name:   "Handshake",
			fields: fields{Type: Handshake, Data: []byte{0x01}},
			want:   fmt.Sprintf("Type: %d, Length: %d, Data: %s", Handshake, 1, string([]byte{0x01})),
		},
		{
			name:   "Data",
			fields: fields{Type: Data, Data: []byte{0x01, 0x02, 0x03}},
			want:   fmt.Sprintf("Type: %d, Length: %d, Data: %s", Data, 3, string([]byte{0x01, 0x02, 0x03})),
		},
		{
			name:   "Kick",
			fields: fields{Type: Kick, Data: []byte{0x05, 0x02, 0x03, 0x04}},
			want:   fmt.Sprintf("Type: %d, Length: %d, Data: %s", Kick, 4, string([]byte{0x05, 0x02, 0x03, 0x04})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Packet{
				Type: tt.fields.Type,
				Data: tt.fields.Data,
			}
			assert.Equalf(t, tt.want, p.String(), "String()")
		})
	}
}
