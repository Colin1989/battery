package agent

import (
	"github.com/colin1989/battery/actor"
	"net"
)

type Agent struct {
	pid  *actor.PID
	conn net.Conn
	//packetDecoder codec.PacketDecoder
	//packetEncoder codec.PacketEncoder
	//serializer    serialize.Serializer
}

func (a *Agent) Receive(ctx actor.Context) {
	envelop := ctx.Envelope()
	switch _ := envelop.Message.(type) {
	case *actor.Started:
	case *actor.Stopped:
	default:

	}
}

func NewAgent(pid *actor.PID, conn net.Conn) actor.Actor {
	return &Agent{
		pid:  pid,
		conn: conn,
		//packetDecoder: nil,
		//packetEncoder: nil,
		//serializer:    nil,
	}
}
