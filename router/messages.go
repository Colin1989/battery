package router

import (
	"github.com/colin1989/battery/actor"
)

type ManagementMessage interface {
	ManagementMessage()
}

type BroadcastMessage struct {
	Message *actor.MessageEnvelope
}

func (*AddRoutee) ManagementMessage()        {}
func (*RemoveRoutee) ManagementMessage()     {}
func (*GetRoutees) ManagementMessage()       {}
func (*AdjustPoolSize) ManagementMessage()   {}
func (*BroadcastMessage) ManagementMessage() {}

func AddRouteeEnvelope(pid *actor.PID) *actor.MessageEnvelope {
	return actor.WrapEnvelope(&AddRoutee{
		PID: pid,
	})
}

func RemoveRouteeEnvelope(pid *actor.PID) *actor.MessageEnvelope {
	return actor.WrapEnvelope(&RemoveRoutee{
		PID: pid,
	})
}

func GetRouteesEnvelope() *actor.MessageEnvelope {
	return actor.WrapEnvelope(&GetRoutees{})
}

func BroadcastMessageEnvelope(envelope *actor.MessageEnvelope) *actor.MessageEnvelope {
	return actor.WrapEnvelope(&BroadcastMessage{Message: envelope})
}
