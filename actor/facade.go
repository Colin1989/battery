package actor

// A Process is an interface that defines the base contract for interaction of actors
type Process interface {
	Send(pid *PID, message interface{})
	Stop(pid *PID)
}

type ProcessActor interface {
	Dead()
}

// Actor is the interface that defines the Receive method.
//
// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
type Actor interface {
	Receive(c Context)
}

type queue interface {
	Push(interface{})
	Pop() interface{}
}
