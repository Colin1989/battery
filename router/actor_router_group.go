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

func (a *groupRouterActor) Receive(ctx actor.Context) {
	envelope := ctx.Envelope()
	message := envelope.Message
	switch m := message.(type) {
	case *actor.Started:
		a.config.OnStarted(ctx, a.props, a.state)
		a.wg.Done()

	case *AddRoutee:
		r := a.state.GetRoutees()
		if r.Contains(m.PID) {
			return
		}
		ctx.Watch(m.PID)
		r.Add(m.PID)
		a.state.SetRoutees(r)

	case *RemoveRoutee:
		r := a.state.GetRoutees()
		if !r.Contains(m.PID) {
			return
		}

		ctx.Unwatch(m.PID)
		r.Remove(m.PID)
		a.state.SetRoutees(r)

	case *BroadcastMessage:
		r := a.state.GetRoutees()
		msg := m.Message
		r.ForEach(func(i int, pid *actor.PID) {
			ctx.Send(pid, msg)
		})

	case *GetRoutees:
		r := a.state.GetRoutees()
		routees := make([]*actor.PID, r.Len())
		r.ForEach(func(i int, pid *actor.PID) {
			routees[i] = pid
		})

		ctx.Respond(actor.WrapEnvelope(&Routees{PIDs: routees}))
	case *actor.Terminated:
		r := a.state.GetRoutees()
		if r.Remove(m.Who) {
			a.state.SetRoutees(r)
		}
	case *actor.DeadLetterResponse:
		r := a.state.GetRoutees()
		if r.Remove(m.Target) {
			a.state.SetRoutees(r)
		}
	}
}
