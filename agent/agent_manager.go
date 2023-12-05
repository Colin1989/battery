package agent

import (
	"fmt"
	"github.com/colin1989/battery/actor"
	"github.com/colin1989/battery/constant"
	"github.com/colin1989/battery/message"
)

//goland:noinspection GoNameStartsWithPackageName
type AgentManager struct {
}

func NewAgentManager() actor.Actor {
	return &AgentManager{}
}

func (am *AgentManager) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	switch msg := envelope.Message.(type) {
	case *actor.Started:
		fmt.Println("actor started")
	case *actor.Stopped:
		fmt.Println("actor stopped")
	case *actor.Stopping:
		fmt.Println("actor stopping")
	case *message.NewAgent:
		props := actor.PropsFromProducer(func() actor.Actor {
			return NewAgent(msg.Conn)
		})
		pid := ctx.SpawnPrefix(props, constant.AgentPrefix)
		_ = pid
	default:
		fmt.Printf("unsupported type %T msg : %+v \n", msg, msg)
	}
}
