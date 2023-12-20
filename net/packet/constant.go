package packet

import "errors"

// Type represents the network packet's type such as: handshake and so on.
type Type byte

const (
	_ Type = iota
	// Handshake represents a handshake: request(client) <====> handshake response(server)
	Handshake = 0x01
	// HandshakeAck represents a handshake ack from client to server
	HandshakeAck = 0x02
	// Heartbeat represents a heartbeat
	Heartbeat = 0x03
	// Data represents a common data packet
	Data = 0x04
	// Kick represents a kick off packet
	Kick = 0x05 // disconnect message from server
)

func IsPacketType(typ Type) error {
	if typ >= Handshake && typ <= Kick {
		return nil
	}
	return ErrWrongPomeloPacketType
}

type (

	// HandshakeClientData represents information about the client sent on the handshake.
	HandshakeClientData struct {
		Platform    string `json:"platform"`
		LibVersion  string `json:"libVersion"`
		BuildNumber string `json:"clientBuildNumber"`
		Version     string `json:"clientVersion"`
	}

	// HandshakeData represents information about the handshake sent by the client.
	// `sys` corresponds to information independent from the app and `user` information
	// that depends on the app and is customized by the user.
	HandshakeData struct {
		Sys  HandshakeClientData    `json:"sys"`
		User map[string]interface{} `json:"user,omitempty"`
	}
)

var (
	// ErrWrongPomeloPacketType represents a wrong packet type.
	ErrWrongPomeloPacketType = errors.New("wrong packet type")

	// ErrInvalidPomeloHeader represents an invalid header
	ErrInvalidPomeloHeader = errors.New("invalid header")

	// ErrPacketSizeExceed is the error used for encode/decode.
	ErrPacketSizeExceed = errors.New("codec: packet size exceed")
)
