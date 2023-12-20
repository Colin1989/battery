package acceptor

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/facade"
)

var (
	system = actor.NewActorSystem()
	props  = actor.PropsFromProducer(func() actor.Actor {
		return &TAgentManager{agents: actor.PIDSet{}}
	})
	agentManagerPID, _ = system.Root.SpawnNamed(props, constant.Gate)
	agentPid           *actor.PID
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
		ctx.Respond(actor.WrapEnvelop(resp))
	case *localAddr:
		ctx.Respond(actor.WrapEnvelop(ta.conn.LocalAddr()))
	case *remoteAddr:
		ctx.Respond(actor.WrapEnvelop(ta.conn.RemoteAddr()))
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
			ctx.Send(ctx.Self(), actor.WrapEnvelop(err))
			//ctx.Poison(ctx.Self())
			return
			//continue
		}
		ctx.Send(ctx.Self(), actor.WrapEnvelop(msg))
	}
}

type TAgentManager struct {
	agents actor.PIDSet
}

type ReqChildCount struct {
}

func (am *TAgentManager) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started TAgentManager")
	case *actor.Stopped:
		fmt.Println("actor stopped TAgentManager")
	case *messages.NewAgent:
		props := actor.PropsFromProducer(func() actor.Actor {
			return &tAgent{
				conn:    msg.Conn,
				message: nil,
			}
		})
		ctx.SpawnPrefix(props, constant.AgentPrefix)
	case *ReqChildCount:
		ctx.Respond(actor.WrapEnvelop(len(ctx.Children())))
	default:
		fmt.Printf("TAgentManager unsupported type %T msg : %+v \n", msg, msg)
	}
}
