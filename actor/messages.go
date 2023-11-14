package actor

// An AutoReceiveMessage is a special kind of user message that will be handled in some way automatically by the actor
type AutoReceiveMessage interface {
	AutoReceiveMessage()
}

// A SystemMessage message is reserved for specific lifecycle messages used by the actor system
type SystemMessage interface {
	SystemMessage()
}

// NotInfluenceReceiveTimeout messages will not reset the ReceiveTimeout timer of an actor that receives the message
type NotInfluenceReceiveTimeout interface {
	NotInfluenceReceiveTimeout()
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

// Failure message is sent to an actor parent when an exception is thrown by one of its methods
type Failure struct {
	Who    *PID
	Reason interface{}
	//RestartStats *RestartStatistics
	Message interface{}
}

// ResumeMailbox is message sent by the actor system to resume mailbox processing.
//
// This will not be forwarded to the Receive method
type ResumeMailbox struct{}

// SuspendMailbox is message sent by the actor system to suspend mailbox processing.
//
// This will not be forwarded to the Receive method
type SuspendMailbox struct{}

func (*Restarting) AutoReceiveMessage() {}
func (*Stopping) AutoReceiveMessage()   {}
func (*Stopped) AutoReceiveMessage()    {}
func (*PoisonPill) AutoReceiveMessage() {}

func (*Started) SystemMessage()        {}
func (*Stop) SystemMessage()           {}
func (*Terminated) SystemMessage()     {}
func (*Failure) SystemMessage()        {}
func (*Restart) SystemMessage()        {}
func (*ReceiveTimeout) SystemMessage() {}
func (*SuspendMailbox) SystemMessage() {}
func (*ResumeMailbox) SystemMessage()  {}

var (
	//restartingMessage AutoReceiveMessage = &Restarting{}
	//stoppingMessage   AutoReceiveMessage = &Stopping{}
	//stoppedMessage    AutoReceiveMessage = &Stopped{}
	//poisonPillMessage AutoReceiveMessage = &PoisonPill{}

	restartMessage        SystemMessage = &Restart{}
	startedMessage        SystemMessage = &Started{}
	stopMessage           SystemMessage = &Stop{}
	receiveTimeoutMessage SystemMessage = &ReceiveTimeout{}
	resumeMailboxMessage  SystemMessage = &ResumeMailbox{}
	suspendMailboxMessage SystemMessage = &SuspendMailbox{}
)

func startedMessageEnvelope() *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: &Started{},
		Sender:  nil,
	}
}

func restartingMessage() *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: &Restarting{},
		Sender:  nil,
	}
}

func stoppingMessage() *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: &Stopping{},
		Sender:  nil,
	}
}

func stoppedMessage() *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: &Stopped{},
		Sender:  nil,
	}
}

func poisonPillMessage() *MessageEnvelope {
	return &MessageEnvelope{
		Header:  nil,
		Message: &PoisonPill{},
		Sender:  nil,
	}
}
