package agent

import (
	"encoding/json"
	"github.com/colin1989/battery/facade"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/net/packet"
	"github.com/colin1989/battery/util/compression"
	"sync"
	"time"
)

var (
	// hbd contains the heartbeat packet data
	hbd []byte
	// hrd contains the handshake response data
	hrd []byte
	// herd contains the handshake error response data
	herd []byte
	once sync.Once
)

func hbdEncode(heartbeatTimeout time.Duration, packetEncoder facade.PacketEncoder, dataCompression bool, serializerName string) {
	hData := map[string]interface{}{
		"code": 200,
		"sys": map[string]interface{}{
			"heartbeat":  heartbeatTimeout.Seconds(),
			"dict":       message.GetDictionary(),
			"serializer": serializerName,
		},
	}

	data, err := encodeAndCompress(hData, dataCompression)
	if err != nil {
		panic(err)
	}

	hrd, err = packetEncoder.Encode(packet.Handshake, data)
	if err != nil {
		panic(err)
	}

	hbd, err = packetEncoder.Encode(packet.Heartbeat, nil)
	if err != nil {
		panic(err)
	}
}

func herdEncode(heartbeatTimeout time.Duration, packetEncoder facade.PacketEncoder, dataCompression bool, serializerName string) {
	hErrData := map[string]interface{}{
		"code": 400,
		"sys": map[string]interface{}{
			"heartbeat":  heartbeatTimeout.Seconds(),
			"dict":       message.GetDictionary(),
			"serializer": serializerName,
		},
	}

	errData, err := encodeAndCompress(hErrData, dataCompression)
	if err != nil {
		panic(err)
	}

	herd, err = packetEncoder.Encode(packet.Handshake, errData)
	if err != nil {
		panic(err)
	}
}

func encodeAndCompress(data interface{}, dataCompression bool) ([]byte, error) {
	encData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	if dataCompression {
		compressedData, err := compression.DeflateData(encData)
		if err != nil {
			return nil, err
		}

		if len(compressedData) < len(encData) {
			encData = compressedData
		}
	}
	return encData, nil
}
