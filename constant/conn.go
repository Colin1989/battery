package constant

import "errors"

var (
	// ErrWrongPomeloPacketType represents a wrong packet type.
	ErrWrongPomeloPacketType = errors.New("wrong packet type")

	// ErrInvalidPomeloHeader represents an invalid header
	ErrInvalidPomeloHeader = errors.New("invalid header")
)
