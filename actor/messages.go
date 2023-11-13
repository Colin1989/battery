package actor

// A SystemMessage message is reserved for specific lifecycle messages used by the actor system
type SystemMessage interface {
	SystemMessage()
}

// A Stopping message is sent to an actor prior to the actor being stopped
type Stopping struct{}

// A Stopped message is sent to the actor once it has been stopped. A stopped actor will receive no further messages
type Stopped struct{}

// A Started message is sent to an actor once it has been started and ready to begin receiving messages.
type Started struct{}

// Restart is message sent by the actor system to control the lifecycle of an actor
type Restart struct{}

// A ReceiveTimeout message is sent to an actor after the Context.ReceiveTimeout duration has expired
type ReceiveTimeout struct{}

// A Restarting message is sent to an actor when the actor is being restarted by the system due to a failure
type Restarting struct{}

// ResumeMailbox is message sent by the actor system to resume mailbox processing.
//
// This will not be forwarded to the Receive method
type ResumeMailbox struct{}

func (*Stopped) SystemMessage() {}

func makeMessage[T any]() *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: new(T),
		Sender:  nil,
	}
}
