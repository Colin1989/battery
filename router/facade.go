package router

import "github.com/colin1989/battery/actor"

// State A type that satisfies router.Interface can be used as a router
type State interface {
	RouteMessage(message *actor.MessageEnvelope)
	SetRoutees(routees *actor.PIDSet)
	GetRoutees() *actor.PIDSet
	SetSender(sender actor.SenderContext)
}
