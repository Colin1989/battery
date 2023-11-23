package acceptor

import (
	"github.com/colin1989/battery/actor"
	"net"
)

type Connector interface {
	GetNextMessage() (b []byte, err error)
	RemoteAddr() net.Addr
	net.Conn
}

type connProducer func(conn Connector) actor.Producer
