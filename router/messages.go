package router

import (
	"github.com/colin1989/battery/actor"
)

type ManagementMessage interface {
	ManagementMessage()
}

type BroadcastMessage struct {
	Message interface{}
}

func (*AddRoutee) ManagementMessage()        {}
func (*RemoveRoutee) ManagementMessage()     {}
func (*GetRoutees) ManagementMessage()       {}
func (*AdjustPoolSize) ManagementMessage()   {}
func (*BroadcastMessage) ManagementMessage() {}

func AddRouteeEnvelope(pid *actor.PID) *actor.MessageEnvelope {
	return actor.WrapEnvelop(&AddRoutee{
		PID: pid,
	})
}

func RemoveRouteeEnvelope(pid *actor.PID) *actor.MessageEnvelope {
	return actor.WrapEnvelop(&RemoveRoutee{
		PID: pid,
	})
}
