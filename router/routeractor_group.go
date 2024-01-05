package router

import (
	"sync"

	"github.com/colin1989/battery/actor"
)

type groupRouterActor struct {
	props  *actor.Props
	config RouterConfig
	state  State
	wg     *sync.WaitGroup
}

func (a *groupRouterActor) Receive(context actor.Context) {
	envelope := context.Envelope()
	message := envelope.Message
	switch m := message.(type) {
	case *actor.Started:
		a.config.OnStarted(context, a.props, a.state)
		a.wg.Done()

	case *AddRoutee:
		r := a.state.GetRoutees()
		if r.Contains(m.PID) {
			return
		}
		context.Watch(m.PID)
		r.Add(m.PID)
		a.state.SetRoutees(r)

	case *RemoveRoutee:
		r := a.state.GetRoutees()
		if !r.Contains(m.PID) {
			return
		}

		context.Unwatch(m.PID)
		r.Remove(m.PID)
		a.state.SetRoutees(r)

	case *BroadcastMessage:
		msg := m.Message
		sender := context.Sender()
		a.state.GetRoutees().ForEach(func(i int, pid *actor.PID) {
			context.Send(pid, actor.WrapEnvelopWithSender(msg, sender))
		})

	case *GetRoutees:
		r := a.state.GetRoutees()
		routees := make([]*actor.PID, r.Len())
		r.ForEach(func(i int, pid *actor.PID) {
			routees[i] = pid
		})

		context.Respond(actor.WrapEnvelop(&Routees{PIDs: routees}))
	}
}
