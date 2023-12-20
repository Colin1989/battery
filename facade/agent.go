package facade

import "net"

type IAgent interface {
	PID() string
	RemoteAddr() net.Addr
	Close()
	send([]byte) error
	SetStatus(int32)
	CheckStatus(int32) bool
}
