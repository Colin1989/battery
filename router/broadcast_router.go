package router

import (
	"github.com/colin1989/battery/actor"
)

type broadcastGroupRouter struct {
	GroupRouter
}

type broadcastPoolRouter struct {
	PoolRouter
}

type broadcastRouterState struct {
	routees *actor.PIDSet
	sender  actor.SenderContext
}

func (state *broadcastRouterState) SetSender(sender actor.SenderContext) {
	state.sender = sender
}

func (state *broadcastRouterState) SetRoutees(routees *actor.PIDSet) {
	state.routees = routees
}

func (state *broadcastRouterState) GetRoutees() *actor.PIDSet {
	return state.routees
}

func (state *broadcastRouterState) RouteMessage(message *actor.MessageEnvelope) {
	state.routees.ForEach(func(i int, pid *actor.PID) {
		state.sender.Send(pid, message)
	})
}

func NewBroadcastPool(size int, opts ...actor.PropsOption) *actor.Props {
	return (&actor.Props{}).
		Configure(actor.WithSpawnFunc(spawner(&broadcastPoolRouter{PoolRouter{PoolSize: size}}))).
		Configure(opts...)
}

func NewBroadcastGroup(routees ...*actor.PID) *actor.Props {
	return (&actor.Props{}).
		Configure(
			actor.WithSpawnFunc(spawner(&broadcastGroupRouter{GroupRouter{Routees: actor.NewPIDSet(routees...)}})),
			actor.WithMailbox(actor.UnboundedLockfree()),
		)
}

func (config *broadcastPoolRouter) CreateRouterState() State {
	return &broadcastRouterState{}
}

func (config *broadcastGroupRouter) CreateRouterState() State {
	return &broadcastRouterState{}
}
