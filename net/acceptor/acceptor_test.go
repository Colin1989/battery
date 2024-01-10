package acceptor

import (
	"fmt"

	"github.com/colin1989/battery/net/codec"
	"github.com/colin1989/battery/net/message"
	"github.com/colin1989/battery/serializer/json"

	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
)

var (
	tapp = &testApp{
		messageEncoder: message.NewMessagesEncoder(true),
		decoder:        codec.NewPomeloPacketDecoder(),
		encoder:        codec.NewPomeloPacketEncoder(),
		serializer:     json.NewSerializer(),
	}
	system = actor.NewActorSystem()
	props  = actor.PropsFromProducer(func() actor.Actor {
		return &testGate{
			app:    tapp,
			agents: actor.PIDSet{},
		}
	})
	gate, _  = system.Root.SpawnNamed(props, constant.Gate)
	gateCtx  actor.Context
	agentPid *actor.PID
)

type tAgent struct {
	conn    facade.Connector
	message []interface{}
}

type requestMsg struct{}
type localAddr struct{}
type remoteAddr struct{}

func (ta *tAgent) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		agentPid = ctx.Self()
		fmt.Printf("tAgent started \n")
		go ta.read(ctx)
	case *actor.Stopped:
		err := ta.conn.Close()
		fmt.Printf("tAgent stopped err [%v]  \n", err)
	case *requestMsg:
		var resp interface{}
		if len(ta.message) > 0 {
			resp = ta.message[0]
			ta.message = ta.message[1:]
		}
		ctx.Respond(actor.WrapEnvelope(resp))
	case *localAddr:
		ctx.Respond(actor.WrapEnvelope(ta.conn.LocalAddr()))
	case *remoteAddr:
		ctx.Respond(actor.WrapEnvelope(ta.conn.RemoteAddr()))
	default:
		fmt.Printf("tAgent unsupported type %T msg : %+v \n", msg, msg)
		ta.message = append(ta.message, msg)
	}
}

func (ta *tAgent) read(ctx actor.Context) {
	for {
		msg, err := ta.conn.GetNextMessage()
		if err != nil {
			fmt.Printf("pid[%v] conn receive err[%v]  \n", ctx.Self().String(), err)
			ctx.Send(ctx.Self(), actor.WrapEnvelope(err))
			//ctx.Poison(ctx.Self())
			//return
			continue
		}
		ctx.Send(ctx.Self(), actor.WrapEnvelope(msg))
	}
}

type testApp struct {
	messageEncoder facade.MessageEncoder
	decoder        facade.PacketDecoder
	encoder        facade.PacketEncoder
	serializer     facade.Serializer
}

func (t *testApp) MessageEncoder() facade.MessageEncoder {
	return t.messageEncoder
}

func (t *testApp) Decoder() facade.PacketDecoder {
	return t.decoder
}

func (t *testApp) Encoder() facade.PacketEncoder {
	return t.encoder
}

func (t *testApp) Serializer() facade.Serializer {
	return t.serializer
}

type testGate struct {
	agents actor.PIDSet
	app    facade.App
}

type ReqChildCount struct {
}

func (tg *testGate) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started testGate")
		gateCtx = ctx
	case *actor.Stopped:
		fmt.Println("actor stopped testGate")
	case facade.Connector:
		conn := msg
		props := actor.PropsFromProducer(func() actor.Actor {
			return &tAgent{conn: conn}
		})
		ctx.SpawnPrefix(props, constant.AgentPrefix)
	case *ReqChildCount:
		ctx.Respond(actor.WrapEnvelope(len(ctx.Children())))
	default:
		fmt.Printf("testGate unsupported type %T msg : %+v \n", msg, msg)
	}
}
