package actor

// A Process is an interface that defines the base contract for interaction of actors
type Process interface {
	SendUserMessage(pid *PID, envelope *MessageEnvelope)
	SendSystemMessage(pid *PID, message SystemMessage)
	Stop(pid *PID)
}

type ProcessActor interface {
	Dead()
}

// The Producer type is a function that creates a new actor
type Producer func() Actor

// Actor is the interface that defines the Receive method.
//
// Receive is sent messages to be processed from the mailbox associated with the instance of the actor
type Actor interface {
	Receive(c Context)
}

// The ReceiveFunc type is an adapter to allow the use of ordinary functions as actors to process messages
type ReceiveFunc func(c Context)

// Receive calls f(c)
func (f ReceiveFunc) Receive(c Context) {
	f(c)
}

type queue[T any] interface {
	Push(envelope T)
	Pop() (T, bool)
}
