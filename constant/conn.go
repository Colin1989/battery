package constant

import "errors"

type AcceptorType int32

const (
	AcceptorTypeTCP AcceptorType = 1
	AcceptorTypeWS  AcceptorType = 2
)

var (
	// ErrWrongPomeloPacketType represents a wrong packet type.
	ErrWrongPomeloPacketType = errors.New("wrong packet type")

	// ErrInvalidPomeloHeader represents an invalid header
	ErrInvalidPomeloHeader = errors.New("invalid header")
)
