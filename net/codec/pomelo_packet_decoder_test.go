package codec

import (
	"github.com/colin1989/battery/net/packet"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getMaxData() []byte {
	maxData := make([]byte, MaxPacketSize)
	for i := 0; i < MaxPacketSize; i++ {
		maxData[i] = byte(i % 0xff)
	}
	return maxData
}

func TestPomeloPacketDecoder_Decode(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []*packet.Packet
		wantErr error
	}{
		// TODO: Add test cases.
		{
			name: "Handshake",
			args: args{
				data: []byte{packet.Handshake, 0x00, 0x00, 0x00},
			},
			want: []*packet.Packet{
				{Type: packet.Handshake, Data: []byte{}},
			},
			wantErr: nil,
		},
		{
			name: "Data",
			args: args{
				data: []byte{packet.Data, 0x00, 0x00, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05},
			},
			want: []*packet.Packet{
				{Type: packet.Data, Data: []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
			},
			wantErr: nil,
		},
		{
			name: "invalid",
			args: args{
				data: []byte{0xff, 0x00, 0x00, 0x05, 0x01, 0x02, 0x03, 0x04, 0x05},
			},
			//want:    []*packet.Packet{},
			wantErr: packet.ErrWrongPomeloPacketType,
		},
		{
			name: "MaxSize",
			args: args{
				data: []byte{packet.Data, 0xFF, 0xFF, 0xFF},
			},
			want: []*packet.Packet{
				{Type: packet.Data, Data: getMaxData()},
			},
			wantErr: nil,
		},
		//{
		//	name: "MaxSize",
		//	args: args{
		//		data: []byte{packet.Data, 0xFF, 0xFF, 0xFF, 0x01, 0x02, 0x03, 0x04, 0x05},
		//	},
		//	want: []*packet.Packet{
		//		{Type: packet.Data, Data: []byte{0x01, 0x02, 0x03, 0x04, 0x05}},
		//	},
		//	wantErr: nil,
		//},
	}

	tests[len(tests)-1].args.data = append(tests[len(tests)-1].args.data, getMaxData()...)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PomeloPacketDecoder{}
			got, err := p.Decode(tt.args.data)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.Equalf(t, tt.want, got, "Decode(%v)", tt.args.data)
			}
		})
	}
}
