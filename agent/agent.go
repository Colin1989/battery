package agent

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/facade"
)

type Agent struct {
	ctx  actor.Context
	conn facade.Connector
	//packetDecoder codec.PacketDecoder
	//packetEncoder codec.PacketEncoder
	//serializer    serialize.Serializer
}

func NewAgent(conn facade.Connector) actor.Actor {
	return &Agent{
		conn: conn,
		//packetDecoder: nil,
		//packetEncoder: nil,
		//serializer:    nil,
	}
}

func (a *Agent) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started")
		go a.read(ctx)
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *actor.Stopping:
		fmt.Println("actor stopping")
	default:
		fmt.Printf("unsupported type %T msg : %+v \n", msg, msg)
	}
}

func (a *Agent) read(ctx actor.Context) {
	for {
		msg, err := a.conn.GetNextMessage()
		if err != nil {
			fmt.Printf("pid[%v] conn receive err[%v]  \n", ctx.Self().String(), err)
			ctx.Send(ctx.Self(), actor.WrapEnvelop(err))
			//ctx.Poison(ctx.Self())
			return
		}
		ctx.Send(ctx.Self(), actor.WrapEnvelop(msg))
	}
}
